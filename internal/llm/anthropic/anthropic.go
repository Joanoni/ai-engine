package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/swarmit/ai-engine/internal/llm"
)

// Provider implements llm.LLMProvider using the Anthropic Messages API.
type Provider struct {
	client *sdk.Client
}

// New creates a new Anthropic Provider. It reads the API key from the
// ANTHROPIC_API_KEY environment variable.
func New() (*Provider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic: ANTHROPIC_API_KEY environment variable is not set")
	}

	client := sdk.NewClient(option.WithAPIKey(apiKey))
	return &Provider{client: client}, nil
}

// Send sends a request to the Anthropic Messages API and returns the response.
func (p *Provider) Send(ctx context.Context, req llm.Request) (llm.Response, error) {
	// Build the messages slice.
	messages := make([]sdk.MessageParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		param, err := convertMessage(m)
		if err != nil {
			return llm.Response{}, fmt.Errorf("anthropic: failed to convert message: %w", err)
		}
		messages = append(messages, param)
	}

	// Build tool definitions.
	tools := make([]sdk.ToolParam, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, sdk.ToolParam{
			Name:        sdk.F(t.Name),
			Description: sdk.F(t.Description),
			InputSchema: sdk.F[interface{}](t.InputSchema),
		})
	}

	// Build the request params.
	params := sdk.MessageNewParams{
		Model:     sdk.F(sdk.Model(req.Model)),
		MaxTokens: sdk.F(int64(16000)),
		Messages:  sdk.F(messages),
	}

	if req.SystemPrompt != "" {
		params.System = sdk.F([]sdk.TextBlockParam{
			{Text: sdk.F(req.SystemPrompt), Type: sdk.F(sdk.TextBlockParamTypeText)},
		})
	}

	if len(tools) > 0 {
		params.Tools = sdk.F(tools)
	}

	// Call the API.
	msg, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return llm.Response{}, fmt.Errorf("anthropic: API call failed: %w", err)
	}

	// Detect truncated responses — if the model hit the token limit mid-output
	// the tool call inputs will be incomplete, which causes infinite retry loops.
	if msg.StopReason == sdk.MessageStopReasonMaxTokens {
		return llm.Response{}, fmt.Errorf("anthropic: response truncated (stop_reason=max_tokens); the requested output exceeded the token limit — break the task into smaller steps")
	}

	// Parse the response.
	var resp llm.Response
	resp.StopReason = string(msg.StopReason)

	for _, block := range msg.Content {
		switch block.Type {
		case sdk.ContentBlockTypeText:
			resp.Text += block.Text
		case sdk.ContentBlockTypeToolUse:
			inputBytes, err := json.Marshal(block.Input)
			if err != nil {
				return llm.Response{}, fmt.Errorf("anthropic: failed to marshal tool input for %q: %w", block.Name, err)
			}
			resp.ToolCalls = append(resp.ToolCalls, llm.ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: inputBytes,
			})
		}
	}

	return resp, nil
}

// convertMessage converts an llm.Message (with structured content blocks) into
// the Anthropic SDK MessageParam format.
func convertMessage(m llm.Message) (sdk.MessageParam, error) {
	switch m.Role {
	case llm.RoleUser:
		blocks := make([]sdk.MessageParamContentUnion, 0, len(m.Content))
		for _, b := range m.Content {
			switch b.Type {
			case llm.ContentBlockTypeText:
				blocks = append(blocks, sdk.NewTextBlock(b.Text))
			case llm.ContentBlockTypeToolResult:
				blocks = append(blocks, sdk.NewToolResultBlock(b.ToolResultID, b.ToolResultContent, b.IsError))
			default:
				return sdk.MessageParam{}, fmt.Errorf("anthropic: unexpected block type %q in user message", b.Type)
			}
		}
		return sdk.NewUserMessage(blocks...), nil

	case llm.RoleAssistant:
		blocks := make([]sdk.MessageParamContentUnion, 0, len(m.Content))
		for _, b := range m.Content {
			switch b.Type {
			case llm.ContentBlockTypeText:
				blocks = append(blocks, sdk.NewTextBlock(b.Text))
			case llm.ContentBlockTypeToolUse:
				var inputObj interface{}
				if len(b.ToolInput) > 0 {
					if err := json.Unmarshal(b.ToolInput, &inputObj); err != nil {
						return sdk.MessageParam{}, fmt.Errorf("anthropic: failed to unmarshal tool input for %q: %w", b.ToolName, err)
					}
				}
				blocks = append(blocks, sdk.NewToolUseBlockParam(b.ToolUseID, b.ToolName, inputObj))
			default:
				return sdk.MessageParam{}, fmt.Errorf("anthropic: unexpected block type %q in assistant message", b.Type)
			}
		}
		return sdk.NewAssistantMessage(blocks...), nil

	default:
		return sdk.MessageParam{}, fmt.Errorf("anthropic: unknown role %q", m.Role)
	}
}
