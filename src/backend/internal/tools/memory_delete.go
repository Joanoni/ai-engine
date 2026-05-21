package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// DeleteMemory removes a file from .ai-engine/memory/.
type DeleteMemory struct {
	sb *sandbox.Sandbox
}

// NewDeleteMemory creates a new DeleteMemory tool.
func NewDeleteMemory(sb *sandbox.Sandbox) *DeleteMemory {
	return &DeleteMemory{sb: sb}
}

func (t *DeleteMemory) Name() string { return "delete_memory" }

func (t *DeleteMemory) Description() string {
	return "Deletes a Markdown file from the persistent agent memory store (.ai-engine/memory/). Non-fatal if the file does not exist."
}

func (t *DeleteMemory) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"filename": {
				"type": "string",
				"description": "Name of the memory file to delete (e.g. \"decisions.md\"). Must not contain path separators."
			}
		},
		"required": ["filename"]
	}`)
}

type deleteMemoryInput struct {
	Filename string `json:"filename"`
}

func (t *DeleteMemory) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in deleteMemoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("delete_memory: invalid input: %w", err)
	}

	filename, err := sanitizeMemoryFilename(in.Filename)
	if err != nil {
		return "", fmt.Errorf("delete_memory: %w", err)
	}

	dest := filepath.Join(t.sb.WorkspacePath(), ".ai-engine", "memory", filename)

	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("delete_memory: failed to delete file: %w", err)
	}

	return fmt.Sprintf("Memory deleted: %s", filename), nil
}
