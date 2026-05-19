package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrFinishWork is a sentinel error returned by FinishWork.Execute to signal
// the runner loop to stop cleanly. It is not a real error.
var ErrFinishWork = errors.New("finish_work")

// FinishWork signals that the agent has completed its assigned work.
type FinishWork struct{}

func NewFinishWork() *FinishWork { return &FinishWork{} }

func (t *FinishWork) Name() string { return "finish_work" }

func (t *FinishWork) Description() string {
	return "Signals that the agent has completed all assigned tasks. Call this when the work is done."
}

func (t *FinishWork) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"result": {
				"type": "string",
				"description": "The final response or result message to return to the caller."
			}
		},
		"required": ["result"]
	}`)
}

type finishWorkInput struct {
	Result string `json:"result"`
}

// Execute returns the result string wrapped in ErrFinishWork so the runner
// can detect a clean stop and extract the result message.
func (t *FinishWork) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in finishWorkInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("finish_work: invalid input: %w", err)
	}
	// Return the result as the string and ErrFinishWork as the sentinel.
	return in.Result, ErrFinishWork
}
