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

// CreateTaskFile is a leader tool that creates a Markdown checklist task file
// for a sub-agent at .ai-engine/chats/{session_id}/{agent_name}/tasks.md.
type CreateTaskFile struct {
	sb        *sandbox.Sandbox
	bus       *events.Bus
	sessionID string
}

func NewCreateTaskFile(sb *sandbox.Sandbox, bus *events.Bus, sessionID string) *CreateTaskFile {
	return &CreateTaskFile{sb: sb, bus: bus, sessionID: sessionID}
}

func (t *CreateTaskFile) Name() string { return "create_task_file" }

func (t *CreateTaskFile) Description() string {
	return "Creates a Markdown checklist task file for a sub-agent in the current session."
}

func (t *CreateTaskFile) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "Name of the agent to create the task file for."
			},
			"tasks": {
				"type": "array",
				"description": "List of tasks to include in the checklist.",
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

type taskItem struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

type createTaskFileInput struct {
	Agent string     `json:"agent"`
	Tasks []taskItem `json:"tasks"`
}

func (t *CreateTaskFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in createTaskFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("create_task_file: invalid input: %w", err)
	}

	content := buildTaskMarkdown(in.Tasks)

	relPath := filepath.Join(".ai-engine", "chats", t.sessionID, in.Agent, "tasks.md")
	absPath, err := t.sb.ResolvePath(relPath)
	if err != nil {
		return "", fmt.Errorf("create_task_file: path resolution failed: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("create_task_file: failed to create directories: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("create_task_file: failed to write file: %w", err)
	}

	t.bus.Publish(events.Event{
		Type:      events.EventTypeTasksUpdated,
		SessionID: t.sessionID,
		AgentName: in.Agent,
		Payload:   map[string]string{"path": relPath},
	})

	return fmt.Sprintf("Task file created for agent %q at %s", in.Agent, relPath), nil
}

// buildTaskMarkdown converts a list of task items into a Markdown checklist.
func buildTaskMarkdown(tasks []taskItem) string {
	out := "# Tasks\n\n"
	for _, task := range tasks {
		var marker string
		switch task.Status {
		case "done":
			marker = "[x]"
		case "in_progress":
			marker = "[-]"
		default:
			marker = "[ ]"
		}
		out += fmt.Sprintf("- %s %s\n", marker, task.Description)
	}
	return out
}
