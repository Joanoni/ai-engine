package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// DeleteFile deletes a file within the workspace.
type DeleteFile struct {
	sb *sandbox.Sandbox
}

// NewDeleteFile creates a new DeleteFile tool.
func NewDeleteFile(sb *sandbox.Sandbox) *DeleteFile {
	return &DeleteFile{sb: sb}
}

func (t *DeleteFile) Name() string { return "delete_file" }

func (t *DeleteFile) Description() string {
	return "Deletes a file within the workspace."
}

func (t *DeleteFile) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path to the file within the workspace."
			}
		},
		"required": ["path"]
	}`)
}

type deleteFileInput struct {
	Path string `json:"path"`
}

func (t *DeleteFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in deleteFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("delete_file: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("delete_file: %w", err)
	}

	if err := os.Remove(absPath); err != nil {
		return "", fmt.Errorf("delete_file: failed to delete file: %w", err)
	}

	return fmt.Sprintf("File deleted: %s", in.Path), nil
}
