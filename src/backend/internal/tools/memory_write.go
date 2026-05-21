package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// WriteMemory writes a Markdown file to .ai-engine/memory/.
type WriteMemory struct {
	sb *sandbox.Sandbox
}

// NewWriteMemory creates a new WriteMemory tool.
func NewWriteMemory(sb *sandbox.Sandbox) *WriteMemory {
	return &WriteMemory{sb: sb}
}

func (t *WriteMemory) Name() string { return "write_memory" }

func (t *WriteMemory) Description() string {
	return "Writes (or overwrites) a Markdown file in the persistent agent memory store (.ai-engine/memory/). The file is injected into every agent's system prompt on the next LLM turn. filename must end with .md (appended automatically if missing)."
}

func (t *WriteMemory) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"filename": {
				"type": "string",
				"description": "Name of the memory file (e.g. \"decisions.md\"). Must not contain path separators."
			},
			"content": {
				"type": "string",
				"description": "Markdown content to write to the file."
			}
		},
		"required": ["filename", "content"]
	}`)
}

type writeMemoryInput struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func (t *WriteMemory) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in writeMemoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("write_memory: invalid input: %w", err)
	}

	filename, err := sanitizeMemoryFilename(in.Filename)
	if err != nil {
		return "", fmt.Errorf("write_memory: %w", err)
	}

	memDir := filepath.Join(t.sb.WorkspacePath(), ".ai-engine", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return "", fmt.Errorf("write_memory: failed to create memory directory: %w", err)
	}

	dest := filepath.Join(memDir, filename)
	if err := os.WriteFile(dest, []byte(in.Content), 0o644); err != nil {
		return "", fmt.Errorf("write_memory: failed to write file: %w", err)
	}

	return fmt.Sprintf("Memory written: %s", filename), nil
}

// sanitizeMemoryFilename validates and normalises a memory filename.
// It rejects path traversal attempts and ensures the .md extension.
func sanitizeMemoryFilename(name string) (string, error) {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return "", fmt.Errorf("filename must not contain path separators or '..'")
	}
	if name == "" {
		return "", fmt.Errorf("filename must not be empty")
	}
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	return name, nil
}
