package dyncontext

import (
	"context"
	"os"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// ignoredDirs is the set of directory names that are never included in the workspace tree.
var ignoredDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	"vendor":      true,
	"dist":        true,
	"build":       true,
	"__pycache__": true,
	".venv":       true,
	"venv":        true,
	".ai-engine":  true,
}

const maxTreeDepth = 6

// WorkspaceTreeProvider renders the current file/directory tree of the
// workspace (excluding .ai-engine/) as a Markdown code block.
type WorkspaceTreeProvider struct{}

func (WorkspaceTreeProvider) Name() string { return "workspace_tree" }

func (WorkspaceTreeProvider) Render(_ context.Context, sb *sandbox.Sandbox) (string, error) {
	root := sb.WorkspacePath()
	var lines []string
	err := walkTree(root, root, &lines, 0)
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("## Workspace File Tree\n\n" +
		"This is the complete file and directory structure of your workspace, " +
		"updated before every LLM turn. Use it to understand the project layout " +
		"before reading or writing files.\n\n" +
		"**Do NOT call `list_files` to explore the root or any directory already " +
		"visible here — the tree below is always up to date.** Only use `list_files` " +
		"if you need to inspect a specific subdirectory that is truncated or not " +
		"shown (depth limit: 6 levels).\n\n```\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	b.WriteString("```")
	return b.String(), nil
}

func walkTree(root, dir string, lines *[]string, depth int) error {
	if depth > maxTreeDepth {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		// Skip ignored directories.
		if e.IsDir() && ignoredDirs[name] {
			continue
		}
		indent := strings.Repeat("  ", depth)
		if e.IsDir() {
			*lines = append(*lines, indent+name+"/")
			if err := walkTree(root, dir+string(os.PathSeparator)+name, lines, depth+1); err != nil {
				return err
			}
		} else {
			*lines = append(*lines, indent+name)
		}
	}
	return nil
}
