package tools

import (
	"context"
	"encoding/json"
)

// Tool is the interface that every tool must implement.
type Tool interface {
	Name() string
	Description() string
	// InputSchema returns a JSON Schema object describing the tool's input.
	InputSchema() json.RawMessage
	// Execute runs the tool with the given JSON-encoded input.
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}
