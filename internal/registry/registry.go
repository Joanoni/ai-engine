package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentType represents the role of an agent in the tree.
type AgentType string

const (
	AgentTypeLeader   AgentType = "leader"
	AgentTypeExecutor AgentType = "executor"
)

// AgentDefinition holds the parsed metadata and system prompt for an agent.
type AgentDefinition struct {
	Name         string    `json:"name"`
	Type         AgentType `json:"type"`
	Team         []string  `json:"team"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"-"`
	TaskContext  string    `json:"-"` // runtime-only: per-session task context
}

// Registry loads agent definitions from the workspace. It is not cached —
// every call reads from disk to support hot-reload.
type Registry struct {
	workspacePath string
}

// New creates a new Registry for the given workspace path.
func New(workspacePath string) *Registry {
	return &Registry{workspacePath: workspacePath}
}

// LoadAgent loads the agent definition for the given agent name.
// It reads agent.json and system_prompt.md from .ai-engine/agents/{name}/.
func (r *Registry) LoadAgent(name string) (*AgentDefinition, error) {
	agentDir := filepath.Join(r.workspacePath, ".ai-engine", "agents", name)

	// Parse agent.json.
	agentJSONPath := filepath.Join(agentDir, "agent.json")
	data, err := os.ReadFile(agentJSONPath)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to read agent.json for %q: %w", name, err)
	}

	var def AgentDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("registry: failed to parse agent.json for %q: %w", name, err)
	}

	// Read system_prompt.md.
	promptPath := filepath.Join(agentDir, "system_prompt.md")
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to read system_prompt.md for %q: %w", name, err)
	}
	def.SystemPrompt = string(promptData)

	return &def, nil
}

// LoadEngineContext reads the optional .ai-engine/engine_context.md file from
// the workspace. Returns ("", nil) if the file does not exist.
func (r *Registry) LoadEngineContext() (string, error) {
	path := filepath.Join(r.workspacePath, ".ai-engine", "engine_context.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("registry: failed to read engine_context.md: %w", err)
	}
	return string(data), nil
}

// ListAgents returns the names of all agents defined in the workspace.
func (r *Registry) ListAgents() ([]string, error) {
	agentsDir := filepath.Join(r.workspacePath, ".ai-engine", "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to list agents directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// AgentTreeNode is a lightweight agent descriptor for the /agents endpoint.
type AgentTreeNode struct {
	Name string    `json:"name"`
	Type AgentType `json:"type"`
	Team []string  `json:"team"`
}

// LoadAgentTree traverses the agent tree starting from rootName and returns
// all reachable agents as a flat list. Uses BFS to avoid cycles.
func (r *Registry) LoadAgentTree(rootName string) ([]AgentTreeNode, error) {
	visited := make(map[string]bool)
	queue := []string{rootName}
	var result []AgentTreeNode

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		if visited[name] {
			continue
		}
		visited[name] = true

		def, err := r.LoadAgent(name)
		if err != nil {
			return nil, fmt.Errorf("registry: LoadAgentTree failed for %q: %w", name, err)
		}

		result = append(result, AgentTreeNode{
			Name: def.Name,
			Type: def.Type,
			Team: def.Team,
		})

		for _, member := range def.Team {
			if !visited[member] {
				queue = append(queue, member)
			}
		}
	}

	return result, nil
}
