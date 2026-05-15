package agent

import "github.com/swarmit/ai-engine/internal/registry"

// AgentType mirrors registry.AgentType for use within the agent package.
type AgentType = registry.AgentType

const (
	TypeLeader   = registry.AgentTypeLeader
	TypeExecutor = registry.AgentTypeExecutor
)

// Agent represents a running agent instance bound to a session.
type Agent struct {
	Definition *registry.AgentDefinition
	SessionID  string
}

// New creates a new Agent from a definition and session ID.
func New(def *registry.AgentDefinition, sessionID string) *Agent {
	return &Agent{
		Definition: def,
		SessionID:  sessionID,
	}
}
