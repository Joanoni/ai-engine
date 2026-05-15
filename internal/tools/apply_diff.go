package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// ApplyDiff applies search/replace blocks to an existing file.
type ApplyDiff struct {
	sb *sandbox.Sandbox
}

// NewApplyDiff creates a new ApplyDiff tool.
func NewApplyDiff(sb *sandbox.Sandbox) *ApplyDiff {
	return &ApplyDiff{sb: sb}
}

func (t *ApplyDiff) Name() string { return "apply_diff" }

func (t *ApplyDiff) Description() string {
	return "Applies one or more search/replace operations to an existing file. Each block specifies an exact string to find and the replacement string."
}

func (t *ApplyDiff) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path to the file within the workspace."
			},
			"diff": {
				"type": "array",
				"description": "List of search/replace blocks to apply in order.",
				"items": {
					"type": "object",
					"properties": {
						"search": {
							"type": "string",
							"description": "Exact string to search for in the file."
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
		"required": ["path", "diff"]
	}`)
}

type diffBlock struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
}

type applyDiffInput struct {
	Path string      `json:"path"`
	Diff []diffBlock `json:"diff"`
}

func (t *ApplyDiff) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in applyDiffInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("apply_diff: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("apply_diff: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("apply_diff: failed to read file: %w", err)
	}

	content := string(data)

	for i, block := range in.Diff {
		if !strings.Contains(content, block.Search) {
			return "", fmt.Errorf("apply_diff: block %d: search string not found in file", i+1)
		}
		content = strings.Replace(content, block.Search, block.Replace, 1)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("apply_diff: failed to write file: %w", err)
	}

	return fmt.Sprintf("Applied %d diff block(s) to %s", len(in.Diff), in.Path), nil
}
