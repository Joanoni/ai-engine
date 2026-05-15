package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/swarmit/ai-engine/internal/agent"
	"github.com/swarmit/ai-engine/internal/config"
	"github.com/swarmit/ai-engine/internal/events"
	"github.com/swarmit/ai-engine/internal/llm"
	"github.com/swarmit/ai-engine/internal/registry"
	"github.com/swarmit/ai-engine/internal/sandbox"
	"github.com/swarmit/ai-engine/internal/session"
	"github.com/swarmit/ai-engine/internal/tools"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// incomingMessage is the envelope for messages received over WebSocket.
type incomingMessage struct {
	Event     string          `json:"event"`
	SessionID string          `json:"session_id"`
	Payload   json.RawMessage `json:"payload"`
}

// userMessagePayload is the payload for "user.message" events.
type userMessagePayload struct {
	Text string `json:"text"`
}

// Server is an HTTP/WebSocket server that drives the agent runtime.
type Server struct {
	port     int
	mux      *http.ServeMux
	sessions *session.Manager
	bus      *events.Bus
	reg      *registry.Registry
	sb       *sandbox.Sandbox
	provider llm.LLMProvider
	staticFS fs.FS // optional — nil means no static serving
}

// New creates a new Server with all required dependencies.
func New(
	port int,
	sessions *session.Manager,
	bus *events.Bus,
	reg *registry.Registry,
	sb *sandbox.Sandbox,
	provider llm.LLMProvider,
	staticFS fs.FS,
) *Server {
	s := &Server{
		port:     port,
		mux:      http.NewServeMux(),
		sessions: sessions,
		bus:      bus,
		reg:      reg,
		sb:       sb,
		provider: provider,
		staticFS: staticFS,
	}
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/agents", s.handleAgents)
	if staticFS != nil {
		s.mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}
	return s
}

// Start begins listening for connections. It blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server: listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("server: websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("server: new WebSocket connection from %s", r.RemoteAddr)

	// writeCh serialises all writes to the connection from multiple goroutines.
	writeCh := make(chan events.Event, 64)
	defer close(writeCh)

	go func() {
		for ev := range writeCh {
			data, err := json.Marshal(ev)
			if err != nil {
				log.Printf("server: failed to marshal event: %v", err)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("server: websocket write error: %v", err)
				return
			}
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg incomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("server: invalid message: %v", err)
			continue
		}

		switch msg.Event {
		case "user.message":
			var payload userMessagePayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("server: invalid user.message payload: %v", err)
				continue
			}

			// Hot-reload config on every session start.
			cfg, err := config.Load(s.sb.WorkspacePath())
			if err != nil {
				s.sendError(writeCh, "", fmt.Sprintf("failed to load config: %v", err))
				continue
			}
	
			// Load optional engine context (Layer 1 of the 3-layer system prompt).
			engineContext, err := s.reg.LoadEngineContext()
			if err != nil {
				s.sendError(writeCh, "", fmt.Sprintf("failed to load engine context: %v", err))
				continue
			}

			// Create a new session.
			sess := s.sessions.Create()

			s.bus.Publish(events.Event{
				Type:      events.EventTypeSessionStarted,
				SessionID: sess.ID,
			})

			// Subscribe to bus events and forward them to this connection.
			// We use a per-session subscription so we only forward events for
			// this session (or all events in V1 since there's one session per conn).
			s.bus.Subscribe(func(ev events.Event) {
				writeCh <- ev
			})

			// Load root agent definition.
			rootDef, err := s.reg.LoadAgent(cfg.RootAgent)
			if err != nil {
				s.sendError(writeCh, sess.ID, fmt.Sprintf("failed to load root agent: %v", err))
				continue
			}
			if rootDef.Model == "" {
				rootDef.Model = cfg.DefaultModel
			}

			// Build the SubAgentRunner factory (closes over server deps).
				defaultModel := cfg.DefaultModel
				maxRetries := cfg.MaxToolRetries
				maxToolCalls := cfg.MaxToolCalls
				var runAgent tools.SubAgentRunner
				runAgent = func(ctx context.Context, def *registry.AgentDefinition, sessionID string, message string) (string, error) {
					if def.Model == "" {
						def.Model = defaultModel
					}
					subToolReg := tools.NewRegistry(s.sb, s.bus, s.reg, runAgent, sessionID)
					subAgentInst := agent.New(def, sessionID)
					subRunner := agent.NewRunner(subAgentInst, s.provider, subToolReg, s.bus, engineContext, s.sb, maxRetries, maxToolCalls)
					return subRunner.Run(ctx, message)
				}
		
				// Build the tool registry for the root agent.
				toolReg := tools.NewRegistry(s.sb, s.bus, s.reg, runAgent, sess.ID)
		
				// Create the root agent instance and runner.
				rootAgent := agent.New(rootDef, sess.ID)
				runner := agent.NewRunner(rootAgent, s.provider, toolReg, s.bus, engineContext, s.sb, maxRetries, maxToolCalls)

			// Run the agent in a goroutine so we can keep reading from the socket.
			go func(sessionID string) {
				result, err := runner.Run(r.Context(), payload.Text)
				if err != nil {
					s.sendError(writeCh, sessionID, err.Error())
					return
				}

				writeCh <- events.Event{
					Type:      events.EventTypeSessionFinished,
					SessionID: sessionID,
					Payload:   map[string]string{"result": result},
				}

				s.sessions.Delete(sessionID)
			}(sess.ID)

		default:
			log.Printf("server: unknown event type %q", msg.Event)
		}
	}
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	cfg, err := config.Load(s.sb.WorkspacePath())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load config: %v", err)})
		return
	}

	agents, err := s.reg.LoadAgentTree(cfg.RootAgent)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load agent tree: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
}

// sendError publishes an error event to the write channel.
func (s *Server) sendError(writeCh chan<- events.Event, sessionID string, msg string) {
	writeCh <- events.Event{
		Type:      events.EventTypeError,
		SessionID: sessionID,
		Payload:   map[string]string{"error": msg},
	}
}
