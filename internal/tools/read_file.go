package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// ReadFile reads file content with line numbers, supporting offset and limit.
type ReadFile struct {
	sb *sandbox.Sandbox
}

// NewReadFile creates a new ReadFile tool.
func NewReadFile(sb *sandbox.Sandbox) *ReadFile {
	return &ReadFile{sb: sb}
}

func (t *ReadFile) Name() string { return "read_file" }

func (t *ReadFile) Description() string {
	return "Reads the content of a file within the workspace. Returns lines with line numbers. Supports offset (1-based) and limit parameters."
}

func (t *ReadFile) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path to the file within the workspace."
			},
			"offset": {
				"type": "integer",
				"description": "1-based line number to start reading from. Default: 1."
			},
			"limit": {
				"type": "integer",
				"description": "Maximum number of lines to return. Default: all lines."
			}
		},
		"required": ["path"]
	}`)
}

type readFileInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (t *ReadFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in readFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("read_file: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("read_file: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read_file: failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Apply offset (1-based).
	offset := in.Offset
	if offset <= 0 {
		offset = 1
	}
	if offset > len(lines) {
		return "", fmt.Errorf("read_file: offset %d exceeds file length %d", offset, len(lines))
	}
	lines = lines[offset-1:]

	// Apply limit.
	if in.Limit > 0 && in.Limit < len(lines) {
		lines = lines[:in.Limit]
	}

	// Format with line numbers.
	var sb strings.Builder
	for i, line := range lines {
		lineNum := offset + i
		fmt.Fprintf(&sb, "%4d | %s\n", lineNum, line)
	}

	return sb.String(), nil
}
