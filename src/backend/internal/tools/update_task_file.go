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

// UpdateTaskFile is a leader tool that overwrites the Markdown checklist task
// file for a sub-agent at .ai-engine/chats/{session_id}/{agent_name}/tasks.md.
type UpdateTaskFile struct {
	sb        *sandbox.Sandbox
	bus       *events.Bus
	sessionID string
}

func NewUpdateTaskFile(sb *sandbox.Sandbox, bus *events.Bus, sessionID string) *UpdateTaskFile {
	return &UpdateTaskFile{sb: sb, bus: bus, sessionID: sessionID}
}

func (t *UpdateTaskFile) Name() string { return "update_task_file" }

func (t *UpdateTaskFile) Description() string {
	return "Updates (overwrites) the task file for a sub-agent in the current session."
}

func (t *UpdateTaskFile) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "Name of the agent whose task file to update."
			},
			"tasks": {
				"type": "array",
				"description": "Updated list of tasks for the checklist.",
				"items": {
					"type": "object",
					"properties": {
						"description": { "type": "string" },
						"status": { "type": "string", "enum": ["pending", "in_progress", "done"] }
					},
					"required": ["description", "status"]
				}
			}
		},
		"required": ["agent", "tasks"]
	}`)
}

type updateTaskFileInput struct {
	Agent string     `json:"agent"`
	Tasks []taskItem `json:"tasks"`
}

func (t *UpdateTaskFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in updateTaskFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("update_task_file: invalid input: %w", err)
	}

	content := buildTaskMarkdown(in.Tasks)

	relPath := filepath.Join(".ai-engine", "chats", t.sessionID, in.Agent, "tasks.md")
	absPath, err := t.sb.ResolvePath(relPath)
	if err != nil {
		return "", fmt.Errorf("update_task_file: path resolution failed: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("update_task_file: failed to create directories: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("update_task_file: task file for agent %q does not exist — call create_task_file first", in.Agent)
	}

	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("update_task_file: failed to write file: %w", err)
	}

	t.bus.Publish(events.Event{
		Type:      events.EventTypeTasksUpdated,
		SessionID: t.sessionID,
		AgentName: in.Agent,
		Payload:   map[string]string{"path": relPath},
	})

	return fmt.Sprintf("Task file updated for agent %q at %s", in.Agent, relPath), nil
}
