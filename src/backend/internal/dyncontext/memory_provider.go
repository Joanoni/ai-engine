package dyncontext

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// MemoryProvider renders all .md files from .ai-engine/memory/ as a Markdown
// block injected into every agent's system prompt (Layer 4).
// Files are sorted by filename and read on every LLM turn (live recomputation).
type MemoryProvider struct{}

// NewMemoryProvider creates a new MemoryProvider.
func NewMemoryProvider() MemoryProvider {
	return MemoryProvider{}
}

func (MemoryProvider) Name() string { return "memory" }

func (MemoryProvider) Render(_ context.Context, sb *sandbox.Sandbox) (string, error) {
	memDir := filepath.Join(sb.WorkspacePath(), ".ai-engine", "memory")

	entries, err := os.ReadDir(memDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		log.Printf("dyncontext/memory: failed to read memory dir: %v", err)
		return "", nil
	}

	// Collect .md files sorted by filename.
	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdFiles = append(mdFiles, e.Name())
		}
	}
	if len(mdFiles) == 0 {
		return "", nil
	}
	sort.Strings(mdFiles)

	var sb2 strings.Builder
	sb2.WriteString("# Agent Memory\n\n")
	sb2.WriteString("> The files below are the **current, live content** of the persistent memory store, recomputed before every LLM turn.\n")
	sb2.WriteString("> Other agents may have modified these files since you last wrote to them — the content shown here is always the authoritative, up-to-date state.\n")
	sb2.WriteString("> **When constructing an `update_memory` diff, always copy the `search` string verbatim from the file content shown below.** Never rely on what you previously wrote; the file may have changed.\n")
	sb2.WriteString("> If the exact string you want to replace is not present in the content below, use `write_memory` to overwrite the file entirely instead of using `update_memory`.\n\n")
	for _, name := range mdFiles {
		data, err := os.ReadFile(filepath.Join(memDir, name))
		if err != nil {
			log.Printf("dyncontext/memory: failed to read %q: %v", name, err)
			continue
		}
		fmt.Fprintf(&sb2, "\n## %s\n%s\n", name, string(data))
	}

	return strings.TrimRight(sb2.String(), "\n"), nil
}
