package chatlog

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// LogEntry represents a single line in chat.jsonl.
type LogEntry struct {
	Timestamp string `json:"ts"`              // ISO 8601 UTC
	Turn      int    `json:"turn"`            // increments per LLM round-trip
	Role      string `json:"role"`            // "user" | "assistant" | "tool_result" | "error" | "finish"

	// role=user
	Content string `json:"content,omitempty"`

	// role=assistant
	Text      string          `json:"text,omitempty"`
	ToolCalls []ToolCallEntry `json:"tool_calls,omitempty"`

	// role=tool_result
	ToolUseID string `json:"tool_use_id,omitempty"`
	Tool      string `json:"tool,omitempty"`
	Success   *bool  `json:"success,omitempty"` // pointer so false is serialized
	Output    string `json:"output,omitempty"`

	// role=error
	Message string `json:"message,omitempty"`

	// role=finish
	Result string `json:"result,omitempty"`
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
