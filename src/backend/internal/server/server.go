package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/swarmit/ai-engine/internal/agent"
	"github.com/swarmit/ai-engine/internal/config"
	"github.com/swarmit/ai-engine/internal/dyncontext"
	"github.com/swarmit/ai-engine/internal/events"
	"github.com/swarmit/ai-engine/internal/llm"
	"github.com/swarmit/ai-engine/internal/registry"
	"github.com/swarmit/ai-engine/internal/sandbox"
	"github.com/swarmit/ai-engine/internal/session"
	"github.com/swarmit/ai-engine/internal/sessionstore"
	"github.com/swarmit/ai-engine/internal/tokenstore"
	"github.com/swarmit/ai-engine/internal/tools"
)

// upgrader accepts connections from any origin. This is intentional for V1
// (local use only). For network-exposed deployments, restrict CheckOrigin
// to allowed origins via config.
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
	port       int
	mux        *http.ServeMux
	sessions   *session.Manager
	bus        *events.Bus
	reg        *registry.Registry
	sb         *sandbox.Sandbox
	provider   llm.LLMProvider
	staticFS   fs.FS // optional — nil means no static serving
	store      *sessionstore.Store
	tokenStore *tokenstore.Store
	version    string
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
	store *sessionstore.Store,
	tokenStore *tokenstore.Store,
	version string,
) *Server {
	s := &Server{
		port:       port,
		mux:        http.NewServeMux(),
		sessions:   sessions,
		bus:        bus,
		reg:        reg,
		sb:         sb,
		provider:   provider,
		staticFS:   staticFS,
		store:      store,
		tokenStore: tokenStore,
		version:    version,
	}
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/version", s.handleVersion)
	s.mux.HandleFunc("/agents", s.handleAgents)
	s.mux.HandleFunc("/sessions", s.handleSessions)
	s.mux.HandleFunc("/sessions/", s.handleSessionEvents)
	s.mux.HandleFunc("/tokens", s.handleTokens)
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

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"version": s.version})
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

	var (
		sessionMu     sync.Mutex
		sessionActive bool
	)

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

			sessionMu.Lock()
			if sessionActive {
				sessionMu.Unlock()
				s.sendError(writeCh, "", "a session is already running; please wait for it to finish")
				continue
			}
			sessionActive = true
			sessionMu.Unlock()

			// Hot-reload config on every session start.
			cfg, err := config.Load(s.sb.WorkspacePath())
			if err != nil {
				s.sendError(writeCh, "", fmt.Sprintf("failed to load config: %v", err))
				continue
			}

			// Load optional engine context (Layer 1 of the 4-layer system prompt).
			engineContext, err := s.reg.LoadEngineContext()
			if err != nil {
				s.sendError(writeCh, "", fmt.Sprintf("failed to load engine context: %v", err))
				continue
			}

			// Build the dynamic context registry (Layer 4 of the system prompt).
			allProviders := []dyncontext.DynamicContextProvider{
				dyncontext.WorkspaceTreeProvider{},
			}
			dynCtxReg := dyncontext.NewRegistry(allProviders, cfg.DynamicContext.Providers)

			// Create a new session.
			sess := s.sessions.Create()

			// Persist session start
			prompt := payload.Text
			if err := s.store.StartSession(sess.ID, prompt); err != nil {
				log.Printf("server: [session=%s] failed to start session store: %v", sess.ID, err)
			}
			if err := s.tokenStore.StartSession(sess.ID); err != nil {
				log.Printf("server: [session=%s] failed to start token store session: %v", sess.ID, err)
			}

			// Subscribe to bus to persist events
			var persistID events.SubscriptionID
			persistID = s.bus.Subscribe(func(ev events.Event) {
				if ev.SessionID != sess.ID {
					return
				}
				if appendErr := s.store.AppendEvent(sess.ID, ev); appendErr != nil {
						log.Printf("server: [session=%s] failed to append event: %v", sess.ID, appendErr)
					}
					// Finish session on terminal events and self-remove
					if ev.Type == events.EventTypeSessionFinished {
						if finishErr := s.store.FinishSession(sess.ID, "done"); finishErr != nil {
							log.Printf("server: [session=%s] failed to finish session: %v", sess.ID, finishErr)
						}
						s.bus.Unsubscribe(persistID)
					} else if ev.Type == events.EventTypeError {
						if finishErr := s.store.FinishSession(sess.ID, "error"); finishErr != nil {
							log.Printf("server: [session=%s] failed to finish session: %v", sess.ID, finishErr)
						}
					s.bus.Unsubscribe(persistID)
				}
			})

			s.bus.Publish(events.Event{
				Type:      events.EventTypeSessionStarted,
				SessionID: sess.ID,
			})

			// Subscribe to bus events and forward them to this connection.
			// We use a per-session subscription so we only forward events for
			// this session (or all events in V1 since there's one session per conn).
			forwardID := s.bus.Subscribe(func(ev events.Event) {
				select {
				case writeCh <- ev:
				default:
				}
			})
			defer s.bus.Unsubscribe(forwardID)

			// Load the agent tree (hot-reload per session).
			rootNode, err := s.reg.LoadTree(cfg.RootAgent)
			if err != nil {
				s.sendError(writeCh, sess.ID, fmt.Sprintf("failed to load agent tree: %v", err))
				continue
			}
			if rootNode.Model == "" {
				rootNode.Model = cfg.DefaultModel
			}

			// Build the SubAgentRunner factory (closes over server deps).
			defaultModel := cfg.DefaultModel
			maxRetries := cfg.MaxToolRetries
			maxToolCalls := cfg.MaxToolCalls
			var runAgent tools.SubAgentRunner
			runAgent = func(ctx context.Context, node *registry.AgentNode, sessionID string, message string) (string, error) {
				if node.Model == "" {
					node.Model = defaultModel
				}
				// Map agent node type to the appropriate tool set.
				subToolSet := tools.ToolSetExecutor
				if node.Type == registry.AgentTypeLeader {
					subToolSet = tools.ToolSetLeader
				}
				subToolReg := tools.NewRegistry(s.sb, nil, s.bus, node, runAgent, sessionID, subToolSet, node.Name)
				subAgentInst := agent.New(node, sessionID)
				subRunner := agent.NewRunner(subAgentInst, s.provider, subToolReg, s.bus, engineContext, s.sb, maxRetries, maxToolCalls, s.tokenStore, dynCtxReg)
				return subRunner.Run(ctx, message)
			}

			// Build the tool registry for the root agent.
			toolReg := tools.NewRegistry(s.sb, nil, s.bus, rootNode, runAgent, sess.ID, tools.ToolSetLeader, rootNode.Name)

			// Create the root agent instance and runner.
			rootAgent := agent.New(rootNode, sess.ID)
			runner := agent.NewRunner(rootAgent, s.provider, toolReg, s.bus, engineContext, s.sb, maxRetries, maxToolCalls, s.tokenStore, dynCtxReg)

			// Run the agent in a goroutine so we can keep reading from the socket.
			go func(sessionID string) {
				defer func() {
					sessionMu.Lock()
					sessionActive = false
					sessionMu.Unlock()
				}()

				result, err := runner.Run(r.Context(), payload.Text)
				if err != nil {
					s.bus.Publish(events.Event{
						Type:      events.EventTypeError,
						SessionID: sessionID,
						Payload:   map[string]string{"error": err.Error()},
					})
					return
				}

				s.bus.Publish(events.Event{
					Type:      events.EventTypeSessionFinished,
					SessionID: sessionID,
					Payload:   map[string]string{"result": result},
				})

				s.sessions.Delete(sessionID)
			}(sess.ID)

		case "ping":
			// heartbeat from frontend — no-op

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

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sessions, err := s.store.ListSessions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleSessionEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Path: /sessions/{id}/events or /sessions/{id}/tokens
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id := parts[0]
	if _, err := uuid.Parse(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid session ID"})
		return
	}
	if parts[1] == "tokens" {
		json.NewEncoder(w).Encode(s.tokenStore.ReadSession(id))
		return
	}
	if parts[1] == "logs" {
		s.handleSessionLogs(w, r, id)
		return
	}
	if parts[1] != "events" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	evts, err := s.store.ReadEvents(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Write([]byte("["))
	for i, ev := range evts {
		if i > 0 {
			w.Write([]byte(","))
		}
		w.Write(ev)
	}
	w.Write([]byte("]"))
}

func (s *Server) handleSessionLogs(w http.ResponseWriter, _ *http.Request, sessionID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	logsDir := filepath.Join(s.sb.WorkspacePath(), ".ai-engine", "logs", sessionID)
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			json.NewEncoder(w).Encode(map[string]interface{}{"agents": map[string]interface{}{}})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	result := make(map[string][]json.RawMessage)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentName := entry.Name()
		logPath := filepath.Join(logsDir, agentName, "chat.jsonl")
		data, err := os.ReadFile(logPath)
		if err != nil {
			continue
		}
		var lines []json.RawMessage
		for _, line := range bytes.Split(data, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			lines = append(lines, json.RawMessage(line))
		}
		result[agentName] = lines
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"agents": result})
}

func (s *Server) handleTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(s.tokenStore.ReadProject())
}

// sendError publishes an error event to the write channel.
func (s *Server) sendError(writeCh chan<- events.Event, sessionID string, msg string) {
	writeCh <- events.Event{
		Type:      events.EventTypeError,
		SessionID: sessionID,
		Payload:   map[string]string{"error": msg},
	}
}
