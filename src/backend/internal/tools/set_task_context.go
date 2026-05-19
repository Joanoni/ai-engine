package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swarmit/ai-engine/internal/events"
	"github.com/swarmit/ai-engine/internal/sandbox"
)

// SetTaskContext is a leader tool that writes a task context file for a
// sub-agent at .ai-engine/chats/{session_id}/{agent}/task_context.md.
// The runner will inject this content as Layer 3 of the agent's system prompt.
type SetTaskContext struct {
	sb        *sandbox.Sandbox
	bus       *events.Bus
	sessionID string
}

func NewSetTaskContext(sb *sandbox.Sandbox, bus *events.Bus, sessionID string) *SetTaskContext {
	return &SetTaskContext{sb: sb, bus: bus, sessionID: sessionID}
}

func (t *SetTaskContext) Name() string { return "set_task_context" }

func (t *SetTaskContext) Description() string {
	return "Sets the task context for a sub-agent in the current session. The content is injected as the 'Current Task' section of the agent's system prompt when create_chat is called."
}

func (t *SetTaskContext) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "Name of the agent to set the task context for."
			},
			"content": {
				"type": "string",
				"description": "Markdown content describing the current task for the agent."
			}
		},
		"required": ["agent", "content"]
	}`)
}

type setTaskContextInput struct {
	Agent   string `json:"agent"`
	Content string `json:"content"`
}

func (t *SetTaskContext) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in setTaskContextInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("set_task_context: invalid input: %w", err)
	}

	relPath := filepath.Join(".ai-engine", "chats", t.sessionID, in.Agent, "task_context.md")
	absPath, err := t.sb.ResolvePath(relPath)
	if err != nil {
		return "", fmt.Errorf("set_task_context: path resolution failed: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("set_task_context: failed to create directories: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(in.Content), 0o644); err != nil {
		return "", fmt.Errorf("set_task_context: failed to write file: %w", err)
	}

	t.bus.Publish(events.Event{
		Type:      events.EventTypeTasksUpdated,
		SessionID: t.sessionID,
		AgentName: in.Agent,
		Payload:   map[string]string{"path": relPath},
	})

	return fmt.Sprintf("Task context set for agent %q at %s", in.Agent, relPath), nil
}
