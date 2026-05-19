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
	tools     map[string]Tool
	agentType AgentToolSet // used to validate tool access in Get()
}

// NewRegistry creates a Registry pre-populated with all tools.
//
// Parameters:
//   - sb          : sandbox for filesystem operations (executor tools)
//   - shell       : persistent shell for run_terminal_command (may be nil; inject later via SetShell)
//   - bus         : event bus for publishing events (leader tools)
//   - currentNode : the AgentNode that owns this registry (used by create_chat)
//   - runAgent    : factory function that runs a sub-agent (injected to avoid circular import)
//   - sessionID   : current session identifier (leader tools)
//   - agentType   : tool set to enforce in Get() — ToolSetLeader or ToolSetExecutor
//   - agentName   : name of the owning agent (used by create_chat for recursion detection)
func NewRegistry(
	sb *sandbox.Sandbox,
	shell *sandbox.Shell,
	bus *events.Bus,
	currentNode *registry.AgentNode,
	runAgent SubAgentRunner,
	sessionID string,
	agentType AgentToolSet,
	agentName ...string,
) *Registry {
	r := &Registry{tools: make(map[string]Tool), agentType: agentType}

	// Register executor tools.
	r.register(NewRunTerminalCommand(shell))
	r.register(NewListFiles(sb))
	r.register(NewReadFile(sb))
	r.register(NewWriteFile(sb))
	r.register(NewApplyDiff(sb))
	r.register(NewSearchFiles(sb))
	r.register(NewDeleteFile(sb))

	// Register shared tools.
	r.register(NewFinishWork())

	// Register leader tools.
	ownerName := ""
	if len(agentName) > 0 {
		ownerName = agentName[0]
	}
	cc := NewCreateChat(sb, currentNode, runAgent, sessionID)
	cc.agentName = ownerName
	r.register(cc)
	r.register(NewCreateTaskFile(sb, bus, sessionID))
	r.register(NewUpdateTaskFile(sb, bus, sessionID))
	r.register(NewSetTaskContext(sb, bus, sessionID))

	return r
}

// SetShell injects the persistent shell into the RunTerminalCommand tool.
// Must be called before the agent starts executing tool calls.
func (r *Registry) SetShell(shell *sandbox.Shell) {
	if t, ok := r.tools["run_terminal_command"]; ok {
		if rtc, ok := t.(*RunTerminalCommand); ok {
			rtc.shell = shell
		}
	}
}

func (r *Registry) register(t Tool) {
	r.tools[t.Name()] = t
}

// Get returns the Tool with the given name, or an error if not found or not
// allowed for this registry's agent type.
func (r *Registry) Get(name string) (Tool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tools: unknown tool %q", name)
	}
	// Validate the tool is allowed for this agent type.
	if !r.isAllowed(name) {
		return nil, fmt.Errorf("tools: tool %q is not available for %s agents", name, r.agentType)
	}
	return t, nil
}

// isAllowed reports whether the named tool is permitted for this registry's agent type.
// Returns true for unknown agent types (safe default for tests).
func (r *Registry) isAllowed(name string) bool {
	var allowed []string
	switch r.agentType {
	case ToolSetLeader:
		allowed = []string{"create_chat", "create_task_file", "update_task_file", "set_task_context", "list_files", "read_file", "finish_work"}
	case ToolSetExecutor:
		allowed = []string{"run_terminal_command", "list_files", "read_file", "write_file", "apply_diff", "search_files", "delete_file", "finish_work"}
	default:
		return true // unknown type — allow all (safe default for tests)
	}
	for _, n := range allowed {
		if n == name {
			return true
		}
	}
	return false
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
