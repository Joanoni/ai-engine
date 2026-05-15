package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swarmit/ai-engine/internal/registry"
	"github.com/swarmit/ai-engine/internal/sandbox"
)

// SubAgentRunner is a function that runs a sub-agent to completion and returns
// its result. It is injected into CreateChat to avoid a circular import between
// the tools and agent packages.
type SubAgentRunner func(ctx context.Context, def *registry.AgentDefinition, sessionID string, message string) (string, error)

// CreateChat is a leader tool that delegates a task to a sub-agent by creating
// a new Runner for it, running it to completion, and returning the result.
type CreateChat struct {
	sb        *sandbox.Sandbox
	reg       *registry.Registry
	runAgent  SubAgentRunner
	sessionID string
}

func NewCreateChat(
	sb *sandbox.Sandbox,
	reg *registry.Registry,
	runAgent SubAgentRunner,
	sessionID string,
) *CreateChat {
	return &CreateChat{
		sb:        sb,
		reg:       reg,
		runAgent:  runAgent,
		sessionID: sessionID,
	}
}

func (t *CreateChat) Name() string { return "create_chat" }

func (t *CreateChat) Description() string {
	return "Creates a new chat session with a sub-agent to delegate a task. Returns the agent's result."
}

func (t *CreateChat) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent": {
				"type": "string",
				"description": "Name of the agent to create a chat with."
			},
			"message": {
				"type": "string",
				"description": "Initial message to send to the agent."
			}
		},
		"required": ["agent", "message"]
	}`)
}

type createChatInput struct {
	Agent   string `json:"agent"`
	Message string `json:"message"`
}

func (t *CreateChat) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in createChatInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("create_chat: invalid input: %w", err)
	}

	// Load the target agent definition from the registry.
	def, err := t.reg.LoadAgent(in.Agent)
	if err != nil {
		return "", fmt.Errorf("create_chat: failed to load agent %q: %w", in.Agent, err)
	}

	// Attempt to load the task context for this agent/session (optional).
	relPath := filepath.Join(".ai-engine", "chats", t.sessionID, in.Agent, "task_context.md")
	absPath, err := t.sb.ResolvePath(relPath)
	if err == nil {
		data, readErr := os.ReadFile(absPath)
		if readErr == nil {
			def.TaskContext = string(data)
		}
		// If the file does not exist, leave def.TaskContext empty — no error.
	}

	// Delegate execution to the injected runner factory.
	result, err := t.runAgent(ctx, def, t.sessionID, in.Message)
	if err != nil {
		return "", fmt.Errorf("create_chat: sub-agent %q failed: %w", in.Agent, err)
	}

	return result, nil
}
