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

// contextKey is an unexported type for context keys in this package.
type contextKey string

// agentStackKey is the context key used to propagate the active agent call stack.
const agentStackKey contextKey = "agent_stack"

// agentStackFromContext returns the current agent call stack stored in ctx,
// or nil if none is present.
func agentStackFromContext(ctx context.Context) []string {
	if v := ctx.Value(agentStackKey); v != nil {
		if stack, ok := v.([]string); ok {
			return stack
		}
	}
	return nil
}

// contextWithAgentStack returns a new context with the given agent stack stored.
func contextWithAgentStack(ctx context.Context, stack []string) context.Context {
	return context.WithValue(ctx, agentStackKey, stack)
}

// SubAgentRunner is a function that runs a sub-agent to completion and returns
// its result. It is injected into CreateChat to avoid a circular import between
// the tools and agent packages.
type SubAgentRunner func(ctx context.Context, def *registry.AgentNode, sessionID string, message string) (string, error)

// CreateChat is a leader tool that delegates a task to a sub-agent by creating
// a new Runner for it, running it to completion, and returning the result.
type CreateChat struct {
	sb          *sandbox.Sandbox
	currentNode *registry.AgentNode // the agent that owns this tool instance
	runAgent    SubAgentRunner
	sessionID   string
	agentName   string // name of the agent that owns this tool instance
}

func NewCreateChat(
	sb *sandbox.Sandbox,
	currentNode *registry.AgentNode,
	runAgent SubAgentRunner,
	sessionID string,
) *CreateChat {
	return &CreateChat{
		sb:          sb,
		currentNode: currentNode,
		runAgent:    runAgent,
		sessionID:   sessionID,
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

	// Check for recursive call — if the target agent is already in the call stack,
	// return an error to prevent infinite recursion.
	stack := agentStackFromContext(ctx)
	for _, name := range stack {
		if name == in.Agent {
			return "", fmt.Errorf("create_chat: recursive call detected — agent %q is already in the call stack %v", in.Agent, stack)
		}
	}

	// Resolve the target agent from direct children only (in-memory, no disk I/O).
	var targetNode *registry.AgentNode
	for _, child := range t.currentNode.Children {
		if child.Name == in.Agent {
			targetNode = child
			break
		}
	}
	if targetNode == nil {
		return "", fmt.Errorf("create_chat: agent %q is not a direct team member of %q", in.Agent, t.currentNode.Name)
	}

	// Attempt to load the task context for this agent/session (optional).
	relPath := filepath.Join(".ai-engine", "chats", t.sessionID, in.Agent, "task_context.md")
	absPath, err := t.sb.ResolvePath(relPath)
	if err == nil {
		data, readErr := os.ReadFile(absPath)
		if readErr == nil {
			targetNode.TaskContext = string(data)
		}
		// If the file does not exist, leave TaskContext empty — no error.
	}

	// Push the current agent onto the stack before running the sub-agent.
	newStack := append(stack, t.agentName)
	ctx = contextWithAgentStack(ctx, newStack)

	// Delegate execution to the injected runner factory.
	result, err := t.runAgent(ctx, targetNode, t.sessionID, in.Message)
	if err != nil {
		return "", fmt.Errorf("create_chat: sub-agent %q failed: %w", in.Agent, err)
	}

	return result, nil
}
