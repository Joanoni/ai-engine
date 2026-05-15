package sandbox

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Sandbox wraps all filesystem and terminal operations, ensuring that all
// paths are resolved relative to the workspace and that path traversal
// attacks are prevented.
type Sandbox struct {
	workspacePath string
}

// New creates a new Sandbox. workspacePath must be a non-empty path to the
// workspace root directory.
func New(workspacePath string) (*Sandbox, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("sandbox: workspace path is required")
	}

	abs, err := filepath.Abs(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("sandbox: failed to resolve workspace path: %w", err)
	}

	return &Sandbox{workspacePath: abs}, nil
}

// WorkspacePath returns the absolute workspace path.
func (s *Sandbox) WorkspacePath() string {
	return s.workspacePath
}

// ResolvePath takes a relative path provided by an agent and returns the
// absolute path within the workspace. It returns an error if the resolved
// path would escape the workspace (path traversal attack).
func (s *Sandbox) ResolvePath(relPath string) (string, error) {
	// Clean the path to remove any ".." components before joining.
	cleaned := filepath.Clean(relPath)

	// Join with workspace root.
	abs := filepath.Join(s.workspacePath, cleaned)

	// Ensure the resolved path is still inside the workspace.
	if !strings.HasPrefix(abs, s.workspacePath+string(filepath.Separator)) && abs != s.workspacePath {
		return "", fmt.Errorf("sandbox: path %q escapes workspace boundary", relPath)
	}

	return abs, nil
}
