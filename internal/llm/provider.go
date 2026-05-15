package llm

import (
	"context"
	"encoding/json"
)

// Role represents the role of a message in the conversation.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// ContentBlockType identifies the kind of content block.
type ContentBlockType string

const (
	ContentBlockTypeText       ContentBlockType = "text"
	ContentBlockTypeToolUse    ContentBlockType = "tool_use"
	ContentBlockTypeToolResult ContentBlockType = "tool_result"
)

// ContentBlock is a single structured block within a message.
// Only the fields relevant to the block type are populated.
type ContentBlock struct {
	// Type is always set.
	Type ContentBlockType `json:"type"`

	// Text is set for type "text".
	Text string `json:"text,omitempty"`

	// ToolUseID, ToolName, and ToolInput are set for type "tool_use".
	ToolUseID string          `json:"id,omitempty"`
	ToolName  string          `json:"name,omitempty"`
	ToolInput json.RawMessage `json:"input,omitempty"`

	// ToolResultID and ToolResultContent are set for type "tool_result".
	ToolResultID      string `json:"tool_use_id,omitempty"`
	ToolResultContent string `json:"content,omitempty"`
	IsError           bool   `json:"is_error,omitempty"`
}

// Message is a single entry in the conversation history.
// Content holds structured blocks; for simple text messages a single
// ContentBlockTypeText block is used.
type Message struct {
	Role    Role           `json:"role"`
	Content []ContentBlock `json:"content"`
}

// NewTextMessage creates a Message with a single text block.
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role:    role,
		Content: []ContentBlock{{Type: ContentBlockTypeText, Text: text}},
	}
}

// NewToolUseMessage creates an assistant Message containing tool_use blocks
// (and optionally a preceding text block).
func NewToolUseMessage(text string, calls []ToolCall) Message {
	var blocks []ContentBlock
	if text != "" {
		blocks = append(blocks, ContentBlock{Type: ContentBlockTypeText, Text: text})
	}
	for _, c := range calls {
		blocks = append(blocks, ContentBlock{
			Type:      ContentBlockTypeToolUse,
			ToolUseID: c.ID,
			ToolName:  c.Name,
			ToolInput: c.Input,
		})
	}
	return Message{Role: RoleAssistant, Content: blocks}
}

// NewToolResultMessage creates a user Message containing tool_result blocks.
func NewToolResultMessage(results []ToolResult) Message {
	blocks := make([]ContentBlock, 0, len(results))
	for _, r := range results {
		blocks = append(blocks, ContentBlock{
			Type:              ContentBlockTypeToolResult,
			ToolResultID:      r.ToolUseID,
			ToolResultContent: r.Content,
			IsError:           r.IsError,
		})
	}
	return Message{Role: RoleUser, Content: blocks}
}

// ToolResult holds the output of a single tool execution.
type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// ToolDefinition describes a tool available to the LLM.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// InputSchema is a JSON Schema object describing the tool's input parameters.
	InputSchema interface{} `json:"input_schema"`
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input []byte `json:"input"` // raw JSON input
}

// Request is the payload sent to an LLM provider.
type Request struct {
	Model        string
	SystemPrompt string
	Messages     []Message
	Tools        []ToolDefinition
}

// Response is the payload returned by an LLM provider.
type Response struct {
	// Text is the assistant's text reply, if any.
	Text string
	// ToolCalls contains any tool invocations requested by the model.
	ToolCalls []ToolCall
	// StopReason indicates why the model stopped generating.
	StopReason string
}

// LLMProvider is the abstraction over any LLM backend.
type LLMProvider interface {
	Send(ctx context.Context, req Request) (Response, error)
}
