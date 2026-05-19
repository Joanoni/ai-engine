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

const aiEngineDir = ".ai-engine"

// ListFiles lists files and directories at a given path within the workspace.
type ListFiles struct {
	sb *sandbox.Sandbox
}

// NewListFiles creates a new ListFiles tool.
func NewListFiles(sb *sandbox.Sandbox) *ListFiles {
	return &ListFiles{sb: sb}
}

func (t *ListFiles) Name() string { return "list_files" }

func (t *ListFiles) Description() string {
	return "Lists files and directories at a given path within the workspace. Supports recursive listing."
}

func (t *ListFiles) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path within the workspace to list. Use '.' for the workspace root."
			},
			"recursive": {
				"type": "boolean",
				"description": "If true, lists files recursively. Default: false."
			}
		},
		"required": ["path"]
	}`)
}

type listFilesInput struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

func (t *ListFiles) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in listFilesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("list_files: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("list_files: %w", err)
	}

	var entries []string

	if in.Recursive {
		err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(absPath, path)
			if rel == "." {
				return nil
			}
			parts := strings.Split(rel, string(filepath.Separator))
			for _, part := range parts {
				if part == aiEngineDir {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
			if info.IsDir() {
				entries = append(entries, rel+"/")
			} else {
				entries = append(entries, rel)
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("list_files: walk error: %w", err)
		}
	} else {
		dirEntries, err := os.ReadDir(absPath)
		if err != nil {
			return "", fmt.Errorf("list_files: failed to read directory: %w", err)
		}
		for _, e := range dirEntries {
			if e.Name() == aiEngineDir {
				continue
			}
			if e.IsDir() {
				entries = append(entries, e.Name()+"/")
			} else {
				entries = append(entries, e.Name())
			}
		}
	}

	return strings.Join(entries, "\n"), nil
}
