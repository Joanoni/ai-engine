package dyncontext

import (
	"context"
	"log"
	"strings"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// DynamicContextProvider computes a context string to be injected into the
// agent system prompt at runtime, just before each LLM call.
type DynamicContextProvider interface {
	Name() string
	Render(ctx context.Context, sb *sandbox.Sandbox) (string, error)
}

// Registry holds a list of enabled DynamicContextProviders and renders them
// all into a single string to be injected as Layer 4 of the system prompt.
type Registry struct {
	providers []DynamicContextProvider
}

// NewRegistry creates a Registry containing only the providers whose names
// appear in enabledNames. If enabledNames is empty, all registered providers
// are included (opt-out model).
func NewRegistry(all []DynamicContextProvider, enabledNames []string) *Registry {
	if len(enabledNames) == 0 {
		return &Registry{providers: all}
	}
	enabled := make(map[string]bool, len(enabledNames))
	for _, n := range enabledNames {
		enabled[n] = true
	}
	var filtered []DynamicContextProvider
	for _, p := range all {
		if enabled[p.Name()] {
			filtered = append(filtered, p)
		}
	}
	return &Registry{providers: filtered}
}

// RenderAll calls every provider in order, concatenates their output separated
// by "---", and returns the combined string. Provider errors are logged and
// skipped — they are non-fatal.
func (r *Registry) RenderAll(ctx context.Context, sb *sandbox.Sandbox) string {
	if len(r.providers) == 0 {
		return ""
	}
	var parts []string
	for _, p := range r.providers {
		out, err := p.Render(ctx, sb)
		if err != nil {
			log.Printf("dyncontext: provider %q error: %v", p.Name(), err)
			continue
		}
		if out != "" {
			parts = append(parts, out)
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}
