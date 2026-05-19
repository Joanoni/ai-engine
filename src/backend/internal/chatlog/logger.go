package chatlog

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// LogEntry represents a single line in chat.jsonl.
// The Role field determines which fields are populated.
type LogEntry struct {
	Timestamp string `json:"ts"`   // ISO 8601 UTC
	Turn      int    `json:"turn"` // increments per LLM round-trip (0 = init)
	Role      string `json:"role"` // see role constants below

	// role=agent_init — written once at agent start (turn=0)
	AgentName string `json:"agent_name,omitempty"` // e.g. "backend-executor"
	AgentType string `json:"agent_type,omitempty"` // "leader" | "executor"
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"` // also present in llm_request

	// role=user — initial user message (turn=0) and nudge messages
	Content string `json:"content,omitempty"`

	// role=llm_request — written before every provider.Send() call
	// Model is also set here (same field as agent_init)
	SystemPrompt        string        `json:"system_prompt,omitempty"`         // full composed prompt (all layers)
	SystemLayers        *SystemLayers `json:"system_layers,omitempty"`         // each layer individually
	Messages            []MessageLog  `json:"messages,omitempty"`              // full message history sent to LLM
	Tools               []ToolLog     `json:"tools,omitempty"`                 // tool definitions sent to LLM
	MessageCount        int           `json:"message_count,omitempty"`         // len(messages)
	TotalToolCallsSoFar int           `json:"total_tool_calls_so_far,omitempty"`
	ConsecutiveErrors   int           `json:"consecutive_errors,omitempty"`

	// role=llm_response — written after every provider.Send() call
	Text         string          `json:"text,omitempty"`
	ToolCalls    []ToolCallEntry `json:"tool_calls,omitempty"`
	StopReason   string          `json:"stop_reason,omitempty"`
	InputTokens  int             `json:"input_tokens,omitempty"`
	OutputTokens int             `json:"output_tokens,omitempty"`

	// role=tool_result
	ToolUseID  string `json:"tool_use_id,omitempty"`
	Tool       string `json:"tool,omitempty"`
	Success    *bool  `json:"success,omitempty"` // pointer so false is serialized
	Output     string `json:"output,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"` // tool execution time

	// role=error
	Message string `json:"message,omitempty"`

	// role=finish
	Result string `json:"result,omitempty"`
}

// SystemLayers holds each layer of the composed system prompt individually.
type SystemLayers struct {
	EngineContext  string `json:"engine_context,omitempty"`  // Layer 1
	AgentRole      string `json:"agent_role,omitempty"`      // Layer 2
	TeamContext    string `json:"team_context,omitempty"`    // Layer 3
	DynamicContext string `json:"dynamic_context,omitempty"` // Layer 4
	TaskContext    string `json:"task_context,omitempty"`    // Layer 5
}

// MessageLog is a serializable snapshot of a single message in the conversation history.
type MessageLog struct {
	Role    string       `json:"role"`
	Content []ContentLog `json:"content"`
}

// ContentLog is a serializable snapshot of a single content block.
type ContentLog struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// tool_use fields
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
	// tool_result fields
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ToolLog is a serializable snapshot of a tool definition.
type ToolLog struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToolCallEntry represents a single tool call within an assistant message.
type ToolCallEntry struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// Logger writes LogEntry lines to a chat.jsonl file.
type Logger struct {
	workspacePath string
	sessionID     string
	agentName     string
	file          *os.File
	mu            sync.Mutex
}

// NewLogger creates a new Logger. It only stores the parameters; it does NOT open the file.
func NewLogger(workspacePath, sessionID, agentName string) *Logger {
	return &Logger{
		workspacePath: workspacePath,
		sessionID:     sessionID,
		agentName:     agentName,
	}
}

// Open creates the log file (and parent dirs). Must be called before WriteEntry.
func (l *Logger) Open() error {
	dir := filepath.Join(l.workspacePath, ".ai-engine", "logs", l.sessionID, l.agentName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("chatlog: failed to create log directory %q: %v", dir, err)
		return err
	}

	logPath := filepath.Join(dir, "chat.jsonl")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("chatlog: failed to open log file %q: %v", logPath, err)
		return err
	}

	l.file = f
	return nil
}

// WriteEntry serializes entry to JSON and appends it with a newline.
func (l *Logger) WriteEntry(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("chatlog: failed to marshal log entry: %v", err)
		return err
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		log.Printf("chatlog: failed to write log entry: %v", err)
		return err
	}

	return nil
}

// Close closes the underlying file handle.
func (l *Logger) Close() {
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}
}
