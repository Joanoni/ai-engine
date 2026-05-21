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

// UpdateMemory applies search/replace diff blocks to a memory file.
type UpdateMemory struct {
	sb *sandbox.Sandbox
}

// NewUpdateMemory creates a new UpdateMemory tool.
func NewUpdateMemory(sb *sandbox.Sandbox) *UpdateMemory {
	return &UpdateMemory{sb: sb}
}

func (t *UpdateMemory) Name() string { return "update_memory" }

func (t *UpdateMemory) Description() string {
	return "Applies one or more search/replace operations to an existing memory file in .ai-engine/memory/. Each block finds the FIRST occurrence of the search string and replaces it. Returns an error if a search string is not found."
}

func (t *UpdateMemory) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"filename": {
				"type": "string",
				"description": "Name of the memory file to patch (e.g. \"decisions.md\"). Must not contain path separators."
			},
			"diff": {
				"type": "array",
				"description": "List of search/replace blocks to apply in order.",
				"items": {
					"type": "object",
					"properties": {
						"search": {
							"type": "string",
							"description": "Exact string to search for. Only the FIRST occurrence is replaced."
						},
						"replace": {
							"type": "string",
							"description": "String to replace the matched content with."
						}
					},
					"required": ["search", "replace"]
				}
			}
		},
		"required": ["filename", "diff"]
	}`)
}

type updateMemoryInput struct {
	Filename string      `json:"filename"`
	Diff     []diffBlock `json:"diff"`
}

func (t *UpdateMemory) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in updateMemoryInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("update_memory: invalid input: %w", err)
	}

	filename, err := sanitizeMemoryFilename(in.Filename)
	if err != nil {
		return "", fmt.Errorf("update_memory: %w", err)
	}

	dest := filepath.Join(t.sb.WorkspacePath(), ".ai-engine", "memory", filename)

	data, err := os.ReadFile(dest)
	if err != nil {
		return "", fmt.Errorf("update_memory: failed to read file: %w", err)
	}

	content := string(data)
	for i, block := range in.Diff {
		if !strings.Contains(content, block.Search) {
			return "", fmt.Errorf("update_memory: block %d: search string not found in file", i+1)
		}
		content = strings.Replace(content, block.Search, block.Replace, 1)
	}

	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("update_memory: failed to write file: %w", err)
	}

	return fmt.Sprintf("Applied %d diff block(s) to memory: %s", len(in.Diff), filename), nil
}
