package tools

import (
	"bufio"
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

	f, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("read_file: failed to open file: %w", err)
	}
	defer f.Close()

	offset := in.Offset
	if offset <= 0 {
		offset = 1
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	// Increase buffer for long lines.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineNum := 0
	collected := 0
	for scanner.Scan() {
		lineNum++
		if lineNum < offset {
			continue
		}
		if in.Limit > 0 && collected >= in.Limit {
			break
		}
		fmt.Fprintf(&sb, "%4d | %s\n", lineNum, scanner.Text())
		collected++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read_file: scan error: %w", err)
	}

	if collected == 0 && lineNum < offset {
		return "", fmt.Errorf("read_file: offset %d exceeds file length %d", offset, lineNum)
	}

	return sb.String(), nil
}
