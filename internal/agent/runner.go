package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/swarmit/ai-engine/internal/chatlog"
	"github.com/swarmit/ai-engine/internal/events"
	"github.com/swarmit/ai-engine/internal/llm"
	"github.com/swarmit/ai-engine/internal/sandbox"
	"github.com/swarmit/ai-engine/internal/tools"
)

// Runner executes the agent loop: send message → process tool calls → repeat
// until finish_work is called.
type Runner struct {
	agent          *Agent
	provider       llm.LLMProvider
	tools          *tools.Registry
	chat           *Chat
	bus            *events.Bus
	engineContext  string
	sb             *sandbox.Sandbox
	logger         *chatlog.Logger
	maxToolRetries int
	maxToolCalls   int
}

// NewRunner creates a new Runner for the given agent with all required dependencies.
func NewRunner(a *Agent, provider llm.LLMProvider, toolRegistry *tools.Registry, bus *events.Bus, engineContext string, sb *sandbox.Sandbox, maxToolRetries int, maxToolCalls int) *Runner {
	logger := chatlog.NewLogger(sb.WorkspacePath(), a.SessionID, a.Definition.Name)
	return &Runner{
		agent:          a,
		provider:       provider,
		tools:          toolRegistry,
		chat:           NewChat(),
		bus:            bus,
		engineContext:  engineContext,
		sb:             sb,
		logger:         logger,
		maxToolRetries: maxToolRetries,
		maxToolCalls:   maxToolCalls,
	}
}

// composeSystemPrompt builds the final 3-layer system prompt:
//
//	[1] ENGINE CONTEXT  (optional prefix)
//	[2] AGENT ROLE      (always present)
//	[3] CURRENT TASK    (optional suffix)
func composeSystemPrompt(engineCtx, agentRole, taskCtx string) string {
	var b strings.Builder

	if engineCtx != "" {
		b.WriteString(engineCtx)
		b.WriteString("\n\n---\n\n")
	}

	b.WriteString(agentRole)

	if taskCtx != "" {
		b.WriteString("\n\n---\n\n# Current Task\n\n")
		b.WriteString(taskCtx)
	}

	return b.String()
}

// Run starts the agent execution loop with the given initial user message.
// It returns the result string from finish_work, or an error on failure.
func (r *Runner) Run(ctx context.Context, userMessage string) (string, error) {
	agentName := r.agent.Definition.Name
	sessionID := r.agent.SessionID

	// Open the chat log; errors are non-fatal (already logged internally).
	_ = r.logger.Open()
	defer r.logger.Close()

	r.bus.Publish(events.Event{
		Type:      events.EventTypeAgentStarted,
		SessionID: sessionID,
		AgentName: agentName,
		Payload:   map[string]string{"message": userMessage},
	})

	// Log the initial user message.
	r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Turn:      0,
		Role:      "user",
		Content:   userMessage,
	})

	// Seed the conversation with the initial user message.
	r.chat.AddText(llm.RoleUser, userMessage)

	// Build the tool definitions for this agent type.
	var agentTools []tools.Tool
	switch r.agent.Definition.Type {
	case TypeLeader:
		agentTools = r.tools.ToolsForLeader()
	default:
		agentTools = r.tools.ToolsForExecutor()
	}

	toolDefs := make([]llm.ToolDefinition, 0, len(agentTools))
	for _, t := range agentTools {
		var schema interface{}
		_ = schema // schema is passed as raw JSON
		toolDefs = append(toolDefs, llm.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}

	turn := 0
	consecutiveErrors := 0
	totalToolCalls := 0

	for {
		// Increment turn counter at the start of each LLM call iteration.
		turn++

		// Compose the 3-layer system prompt on every LLM call.
		systemPrompt := composeSystemPrompt(r.engineContext, r.agent.Definition.SystemPrompt, r.agent.Definition.TaskContext)

		req := llm.Request{
			Model:        r.agent.Definition.Model,
			SystemPrompt: systemPrompt,
			Messages:     r.chat.Messages(),
			Tools:        toolDefs,
		}

		resp, err := r.provider.Send(ctx, req)
		if err != nil {
			r.bus.Publish(events.Event{
				Type:      events.EventTypeError,
				SessionID: sessionID,
				AgentName: agentName,
				Payload:   map[string]string{"error": err.Error()},
			})
			r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Turn:      turn,
				Role:      "error",
				Message:   err.Error(),
			})
			return "", fmt.Errorf("runner: LLM call failed: %w", err)
		}

		// Build ToolCalls slice for logging.
		var toolCallEntries []chatlog.ToolCallEntry
		for _, tc := range resp.ToolCalls {
			toolCallEntries = append(toolCallEntries, chatlog.ToolCallEntry{
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Input,
			})
		}
		r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Turn:      turn,
			Role:      "assistant",
			Text:      resp.Text,
			ToolCalls: toolCallEntries,
		})

		// No tool calls — the model returned a plain text response.
		// Append it as an assistant message, then inject a user nudge so the
		// next LLM call does not start with an assistant-last history (which
		// the Anthropic API rejects with 400 "conversation must end with a
		// user message").
		if len(resp.ToolCalls) == 0 {
			if resp.Text != "" {
				r.chat.AddText(llm.RoleAssistant, resp.Text)
			}
			r.chat.AddText(llm.RoleUser, "Continue. Use a tool to proceed or call finish_work if you are done.")
			continue
		}

		// Append the assistant message with tool_use blocks to history.
		r.chat.AddMessage(llm.NewToolUseMessage(resp.Text, resp.ToolCalls))

		// Execute each tool call sequentially.
		var toolResults []llm.ToolResult
		var finishResult string
		var finished bool

		for _, call := range resp.ToolCalls {
			// Enforce the per-agent tool call limit.
			totalToolCalls++
			if r.maxToolCalls > 0 && totalToolCalls > r.maxToolCalls {
				return "", fmt.Errorf("runner: agent %q exceeded the maximum tool call limit (%d); session terminated to prevent runaway loops", agentName, r.maxToolCalls)
			}

			r.bus.Publish(events.Event{
				Type:      events.EventTypeToolCalled,
				SessionID: sessionID,
				AgentName: agentName,
				Payload:   map[string]string{"tool": call.Name, "id": call.ID},
			})

			tool, err := r.tools.Get(call.Name)
			if err != nil {
				r.bus.Publish(events.Event{
					Type:      events.EventTypeError,
					SessionID: sessionID,
					AgentName: agentName,
					Payload:   map[string]string{"error": err.Error(), "tool": call.Name},
				})
				return "", fmt.Errorf("runner: unknown tool %q: %w", call.Name, err)
			}

			result, execErr := tool.Execute(ctx, call.Input)

			if errors.Is(execErr, tools.ErrFinishWork) {
				// Clean stop — record the result and break out of tool loop.
				finishResult = result
				finished = true

				success := true
				r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Turn:      turn,
					Role:      "tool_result",
					ToolUseID: call.ID,
					Tool:      call.Name,
					Success:   &success,
					Output:    result,
				})

				toolResults = append(toolResults, llm.ToolResult{
					ToolUseID: call.ID,
					Content:   result,
					IsError:   false,
				})

				r.bus.Publish(events.Event{
					Type:      events.EventTypeToolResult,
					SessionID: sessionID,
					AgentName: agentName,
					Payload:   map[string]string{"tool": call.Name, "id": call.ID, "result": result},
				})
				break
			}

			if execErr != nil {
					// Publish error event for frontend observability.
					r.bus.Publish(events.Event{
						Type:      events.EventTypeError,
						SessionID: sessionID,
						AgentName: agentName,
						Payload:   map[string]string{"error": execErr.Error(), "tool": call.Name},
					})
	
					// Log the failure.
					failure := false
					r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Turn:      turn,
						Role:      "tool_result",
						ToolUseID: call.ID,
						Tool:      call.Name,
						Success:   &failure,
						Output:    execErr.Error(),
					})
					r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Turn:      turn,
						Role:      "error",
						Tool:      call.Name,
						Message:   execErr.Error(),
					})
	
					// Feed the error back to the agent as a tool result.
					toolResults = append(toolResults, llm.ToolResult{
						ToolUseID: call.ID,
						Content:   fmt.Sprintf("Tool error: %s", execErr.Error()),
						IsError:   true,
					})
	
					consecutiveErrors++
					if consecutiveErrors >= r.maxToolRetries {
						// Exceeded retry limit — terminate.
						return "", fmt.Errorf("runner: tool %q failed %d consecutive times, last error: %w", call.Name, consecutiveErrors, execErr)
					}
					// Do NOT break — continue processing remaining tool calls in this batch.
					continue
				}
	
				// Tool succeeded — reset consecutive error counter.
				consecutiveErrors = 0
	
				success := true
			r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Turn:      turn,
				Role:      "tool_result",
				ToolUseID: call.ID,
				Tool:      call.Name,
				Success:   &success,
				Output:    result,
			})

			toolResults = append(toolResults, llm.ToolResult{
				ToolUseID: call.ID,
				Content:   result,
				IsError:   false,
			})

			r.bus.Publish(events.Event{
				Type:      events.EventTypeToolResult,
				SessionID: sessionID,
				AgentName: agentName,
				Payload:   map[string]string{"tool": call.Name, "id": call.ID, "result": result},
			})
		}

		// Append tool results as a user message.
		r.chat.AddMessage(llm.NewToolResultMessage(toolResults))

		if finished {
			r.bus.Publish(events.Event{
				Type:      events.EventTypeAgentFinished,
				SessionID: sessionID,
				AgentName: agentName,
				Payload:   map[string]string{"result": finishResult},
			})

			r.logger.WriteEntry(chatlog.LogEntry{ //nolint:errcheck
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Turn:      turn,
				Role:      "finish",
				Result:    finishResult,
			})

			// Archive task_context.md if it exists.
			r.archiveTaskContext(sessionID, agentName)

			return finishResult, nil
		}
	}
}

// archiveTaskContext moves .ai-engine/chats/{sessionID}/{agentName}/task_context.md
// to .ai-engine/history/{sessionID}/{agentName}/task_context.md.
// Silently skips if the source file does not exist.
func (r *Runner) archiveTaskContext(sessionID, agentName string) {
	if r.sb == nil {
		return
	}

	srcRel := filepath.Join(".ai-engine", "chats", sessionID, agentName, "task_context.md")
	dstRel := filepath.Join(".ai-engine", "history", sessionID, agentName, "task_context.md")

	srcAbs, err := r.sb.ResolvePath(srcRel)
	if err != nil {
		return
	}
	dstAbs, err := r.sb.ResolvePath(dstRel)
	if err != nil {
		return
	}

	// Check source exists.
	if _, err := os.Stat(srcAbs); os.IsNotExist(err) {
		return
	}

	// Ensure destination directory exists.
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return
	}

	// Try atomic rename first; fall back to copy+delete for cross-device moves.
	if err := os.Rename(srcAbs, dstAbs); err != nil {
		data, readErr := os.ReadFile(srcAbs)
		if readErr != nil {
			return
		}
		if writeErr := os.WriteFile(dstAbs, data, 0o644); writeErr != nil {
			return
		}
		_ = os.Remove(srcAbs)
	}
}
