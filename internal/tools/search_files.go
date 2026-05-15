package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// SearchFiles performs regex search across files in a directory.
type SearchFiles struct {
	sb *sandbox.Sandbox
}

// NewSearchFiles creates a new SearchFiles tool.
func NewSearchFiles(sb *sandbox.Sandbox) *SearchFiles {
	return &SearchFiles{sb: sb}
}

func (t *SearchFiles) Name() string { return "search_files" }

func (t *SearchFiles) Description() string {
	return "Performs a regex search across files in a directory within the workspace. Returns matching lines with file path and line number context."
}

func (t *SearchFiles) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Relative path to the directory to search in."
			},
			"regex": {
				"type": "string",
				"description": "Regular expression pattern to search for."
			},
			"file_pattern": {
				"type": "string",
				"description": "Optional glob pattern to filter files (e.g., '*.go'). If omitted, all files are searched."
			}
		},
		"required": ["path", "regex"]
	}`)
}

type searchFilesInput struct {
	Path        string `json:"path"`
	Regex       string `json:"regex"`
	FilePattern string `json:"file_pattern"`
}

func (t *SearchFiles) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in searchFilesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("search_files: invalid input: %w", err)
	}

	absPath, err := t.sb.ResolvePath(in.Path)
	if err != nil {
		return "", fmt.Errorf("search_files: %w", err)
	}

	re, err := regexp.Compile(in.Regex)
	if err != nil {
		return "", fmt.Errorf("search_files: invalid regex: %w", err)
	}

	var results strings.Builder

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Apply file pattern filter.
		if in.FilePattern != "" {
			matched, err := filepath.Match(in.FilePattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		lines := strings.Split(string(data), "\n")
		rel, _ := filepath.Rel(absPath, path)

		for i, line := range lines {
			if re.MatchString(line) {
				fmt.Fprintf(&results, "%s:%d: %s\n", rel, i+1, line)
			}
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search_files: walk error: %w", err)
	}

	if results.Len() == 0 {
		return "No matches found.", nil
	}

	return results.String(), nil
}
