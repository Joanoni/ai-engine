package dyncontext

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// WorkspaceTreeProvider renders the current file/directory tree of the
// workspace (excluding .ai-engine/) as a Markdown code block.
type WorkspaceTreeProvider struct{}

func (WorkspaceTreeProvider) Name() string { return "workspace_tree" }

func (WorkspaceTreeProvider) Render(_ context.Context, sb *sandbox.Sandbox) (string, error) {
	root := sb.WorkspacePath()
	var lines []string
	err := walkTree(root, root, &lines)
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("## Workspace Tree\n\n```\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	b.WriteString("```")
	return b.String(), nil
}

func walkTree(root, dir string, lines *[]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		// Skip .ai-engine/ at the workspace root level.
		if dir == root && e.Name() == ".ai-engine" {
			continue
		}
		rel, err := filepath.Rel(root, filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		depth := strings.Count(rel, string(filepath.Separator))
		indent := strings.Repeat("  ", depth)
		if e.IsDir() {
			*lines = append(*lines, indent+e.Name()+"/")
			if err := walkTree(root, filepath.Join(dir, e.Name()), lines); err != nil {
				return err
			}
		} else {
			*lines = append(*lines, indent+e.Name())
		}
	}
	return nil
}
