package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// WriteFile creates or fully overwrites a file within the workspace.
type WriteFile struct {
	sb *sandbox.Sandbox
}

// NewWriteFile creates a new WriteFile tool.
func NewWriteFile(sb *sandbox.Sandbox) *WriteFile {
	return &WriteFile{sb: sb}
}

func (t *WriteFile) Name() string { return "write_file" }

func (t *WriteFile) Description() string {
	return "Creates or fully overwrites a file within the workspace with the given content. Creates parent directories as needed."
}

func (t *WriteFile) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path to the file within the workspace."
			},
			"content": {
				"type": "string",
				"description": "The complete content to write to the file."
			}
		},
		"required": ["path", "content"]
	}`)
}

type writeFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (t *WriteFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in writeFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("write_file: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}

	// Ensure parent directories exist.
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("write_file: failed to create directories: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(in.Content), 0644); err != nil {
		return "", fmt.Errorf("write_file: failed to write file: %w", err)
	}

	return fmt.Sprintf("File written successfully: %s", in.Path), nil
}
