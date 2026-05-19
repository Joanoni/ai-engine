package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// defaultCommandTimeout is applied when the agent does not specify timeout_seconds.
const defaultCommandTimeout = 30 * time.Second

// RunTerminalCommand executes a shell command using the agent's persistent Shell.
type RunTerminalCommand struct {
	shell *sandbox.Shell
}

// NewRunTerminalCommand creates a new RunTerminalCommand tool backed by the given Shell.
func NewRunTerminalCommand(shell *sandbox.Shell) *RunTerminalCommand {
	return &RunTerminalCommand{shell: shell}
}

func (t *RunTerminalCommand) Name() string { return "run_terminal_command" }

func (t *RunTerminalCommand) Description() string {
	return "Executes a shell command in the agent's persistent shell session. Working directory and environment variables are preserved between calls. Commands that exceed timeout_seconds are killed (along with all child processes) and partial output is returned with a [TIMEOUT] prefix — this is not an error."
}

func (t *RunTerminalCommand) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The shell command to execute."
			},
			"timeout_seconds": {
				"type": "integer",
				"description": "Maximum seconds to wait for the command to complete. Defaults to 30. Commands that exceed this limit are killed (including all child processes) and partial output is returned with a [TIMEOUT] prefix."
			}
		},
		"required": ["command"]
	}`)
}

type runTerminalCommandInput struct {
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

func (t *RunTerminalCommand) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in runTerminalCommandInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("run_terminal_command: invalid input: %w", err)
	}
	if in.Command == "" {
		return "", fmt.Errorf("run_terminal_command: command is required")
	}

	if t.shell == nil {
		return "", fmt.Errorf("run_terminal_command: shell not initialised")
	}

	timeout := defaultCommandTimeout
	if in.TimeoutSeconds > 0 {
		timeout = time.Duration(in.TimeoutSeconds) * time.Second
	}

	output, timedOut := t.shell.Exec(in.Command, timeout)
	_ = timedOut // timedOut is already reflected in the [TIMEOUT] prefix
	return output, nil
}
