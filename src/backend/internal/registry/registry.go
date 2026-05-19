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

// agentJSON is the optional on-disk format for agent.json.
// Only description and model are used; name/type/team are ignored for backward compat.
type agentJSON struct {
	Description string `json:"description"`
	Model       string `json:"model"`
}

// AgentNode is the in-memory representation of a single agent in the tree.
type AgentNode struct {
	Name         string       // directory name
	Type         AgentType    // "leader" if len(Children) > 0, "executor" otherwise
	Description  string       // from agent.json (optional)
	Model        string       // from agent.json (optional)
	SystemPrompt string       // from system_prompt.md
	DirPath      string       // absolute path on disk (internal use only)
	Children     []*AgentNode // direct sub-agents (subdirs with system_prompt.md)
	Parent       *AgentNode   // reference to parent node
	TaskContext  string       // runtime-only: per-session task context (injected by create_chat)
}

// Registry loads agent definitions from the workspace.
type Registry struct {
	workspacePath string
}

// New creates a new Registry for the given workspace path.
func New(workspacePath string) *Registry {
	return &Registry{workspacePath: workspacePath}
}

// LoadTree reads the agent tree recursively from .ai-engine/agents/{rootName}/.
// Returns the root AgentNode with the full tree populated.
func (r *Registry) LoadTree(rootName string) (*AgentNode, error) {
	rootDir := filepath.Join(r.workspacePath, ".ai-engine", "agents", rootName)
	return loadNode(rootDir, rootName, nil)
}

// loadNode recursively loads an AgentNode from the given directory.
func loadNode(dir, name string, parent *AgentNode) (*AgentNode, error) {
	node := &AgentNode{
		Name:    name,
		DirPath: dir,
		Parent:  parent,
	}

	// Read system_prompt.md (required).
	promptPath := filepath.Join(dir, "system_prompt.md")
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to read system_prompt.md for %q: %w", name, err)
	}
	node.SystemPrompt = string(promptData)

	// Read agent.json (optional — only description and model).
	agentJSONPath := filepath.Join(dir, "agent.json")
	if data, err := os.ReadFile(agentJSONPath); err == nil {
		var aj agentJSON
		if jsonErr := json.Unmarshal(data, &aj); jsonErr != nil {
			return nil, fmt.Errorf("registry: failed to parse agent.json for %q: %w", name, jsonErr)
		}
		node.Description = aj.Description
		node.Model = aj.Model
	}
	// If agent.json does not exist, leave Description and Model empty — no error.

	// Discover child directories that contain system_prompt.md.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("registry: failed to read directory %q: %w", dir, err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		childDir := filepath.Join(dir, e.Name())
		childPrompt := filepath.Join(childDir, "system_prompt.md")
		if _, err := os.Stat(childPrompt); os.IsNotExist(err) {
			// Subdirectory without system_prompt.md — not an agent, skip.
			continue
		}
		child, err := loadNode(childDir, e.Name(), node)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	// Derive type from children.
	if len(node.Children) > 0 {
		node.Type = AgentTypeLeader
	} else {
		node.Type = AgentTypeExecutor
	}

	return node, nil
}

// FindNode performs a BFS search in the in-memory tree by agent name.
// Returns nil if not found.
func FindNode(root *AgentNode, name string) *AgentNode {
	if root == nil {
		return nil
	}
	queue := []*AgentNode{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.Name == name {
			return node
		}
		queue = append(queue, node.Children...)
	}
	return nil
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

// AgentTreeNode is a lightweight agent descriptor for the /agents endpoint.
// Kept for backward compatibility with the frontend.
type AgentTreeNode struct {
	Name string    `json:"name"`
	Type AgentType `json:"type"`
	Team []string  `json:"team"`
}

// LoadAgentTree builds the in-memory tree via LoadTree and returns a flat list
// of AgentTreeNode for the /agents HTTP endpoint. Frontend depends on this format.
func (r *Registry) LoadAgentTree(rootName string) ([]AgentTreeNode, error) {
	root, err := r.LoadTree(rootName)
	if err != nil {
		return nil, fmt.Errorf("registry: LoadAgentTree failed: %w", err)
	}

	var result []AgentTreeNode
	var walk func(node *AgentNode)
	walk = func(node *AgentNode) {
		team := make([]string, 0, len(node.Children))
		for _, c := range node.Children {
			team = append(team, c.Name)
		}
		result = append(result, AgentTreeNode{
			Name: node.Name,
			Type: node.Type,
			Team: team,
		})
		for _, c := range node.Children {
			walk(c)
		}
	}
	walk(root)

	return result, nil
}
