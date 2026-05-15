package tools

import (
	"fmt"

	"github.com/swarmit/ai-engine/internal/events"
	"github.com/swarmit/ai-engine/internal/registry"
	"github.com/swarmit/ai-engine/internal/sandbox"
)

// AgentToolSet defines which set of tools an agent type receives.
type AgentToolSet string

const (
	ToolSetLeader   AgentToolSet = "leader"
	ToolSetExecutor AgentToolSet = "executor"
)

// Registry maps tool names to Tool implementations and provides tool lists
// for each agent type.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a Registry pre-populated with all tools.
//
// Parameters:
//   - sb        : sandbox for filesystem operations (executor tools)
//   - bus       : event bus for publishing events (leader tools)
//   - reg       : agent registry for loading sub-agent definitions (create_chat)
//   - runAgent  : factory function that runs a sub-agent (injected to avoid circular import)
//   - sessionID : current session identifier (leader tools)
func NewRegistry(
	sb *sandbox.Sandbox,
	bus *events.Bus,
	reg *registry.Registry,
	runAgent SubAgentRunner,
	sessionID string,
) *Registry {
	r := &Registry{tools: make(map[string]Tool)}

	// Register executor tools.
	r.register(NewRunTerminalCommand(sb))
	r.register(NewListFiles(sb))
	r.register(NewReadFile(sb))
	r.register(NewWriteFile(sb))
	r.register(NewApplyDiff(sb))
	r.register(NewSearchFiles(sb))
	r.register(NewDeleteFile(sb))

	// Register shared tools.
	r.register(NewFinishWork())

	// Register leader tools.
	r.register(NewCreateChat(sb, reg, runAgent, sessionID))
	r.register(NewCreateTaskFile(sb, bus, sessionID))
	r.register(NewUpdateTaskFile(sb, bus, sessionID))
	r.register(NewSetTaskContext(sb, bus, sessionID))

	return r
}

func (r *Registry) register(t Tool) {
	r.tools[t.Name()] = t
}

// Get returns the Tool with the given name, or an error if not found.
func (r *Registry) Get(name string) (Tool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tools: unknown tool %q", name)
	}
	return t, nil
}

// ToolsForLeader returns the tool list available to leader agents.
func (r *Registry) ToolsForLeader() []Tool {
	names := []string{"create_chat", "create_task_file", "update_task_file", "set_task_context", "list_files", "read_file", "finish_work"}
	return r.byNames(names)
}

// ToolsForExecutor returns the tool list available to executor agents.
func (r *Registry) ToolsForExecutor() []Tool {
	names := []string{
		"run_terminal_command",
		"list_files",
		"read_file",
		"write_file",
		"apply_diff",
		"search_files",
		"delete_file",
		"finish_work",
	}
	return r.byNames(names)
}

func (r *Registry) byNames(names []string) []Tool {
	out := make([]Tool, 0, len(names))
	for _, name := range names {
		if t, ok := r.tools[name]; ok {
			out = append(out, t)
		}
	}
	return out
}
