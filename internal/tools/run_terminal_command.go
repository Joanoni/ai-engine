package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/swarmit/ai-engine/internal/sandbox"
)

// defaultCommandTimeout is applied when the agent does not specify timeout_seconds.
const defaultCommandTimeout = 30 * time.Second

// RunTerminalCommand executes a shell command inside the workspace sandbox.
type RunTerminalCommand struct {
	sb *sandbox.Sandbox
}

// NewRunTerminalCommand creates a new RunTerminalCommand tool.
func NewRunTerminalCommand(sb *sandbox.Sandbox) *RunTerminalCommand {
	return &RunTerminalCommand{sb: sb}
}

func (t *RunTerminalCommand) Name() string { return "run_terminal_command" }

func (t *RunTerminalCommand) Description() string {
	return "Executes a shell command inside the workspace. The working directory defaults to the workspace root but can be overridden with the optional workdir parameter. Returns stdout and stderr combined. Commands that exceed timeout_seconds are killed and their partial output is returned with a [TIMEOUT] prefix — this is not an error."
}

func (t *RunTerminalCommand) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The shell command to execute."
			},
			"workdir": {
				"type": "string",
				"description": "Optional subdirectory within the workspace to use as working directory. Defaults to workspace root."
			},
			"timeout_seconds": {
				"type": "integer",
				"description": "Maximum seconds to wait for the command to complete. Defaults to 30. Commands that exceed this limit are killed and partial output is returned with a [TIMEOUT] prefix."
			}
		},
		"required": ["command"]
	}`)
}

type runTerminalCommandInput struct {
	Command        string `json:"command"`
	Workdir        string `json:"workdir"`
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

	var workdir string
	if in.Workdir != "" {
		resolved, err := t.sb.ResolvePath(in.Workdir)
		if err != nil {
			return "", fmt.Errorf("run_terminal_command: invalid workdir: %w", err)
		}
		workdir = resolved
	} else {
		workdir = t.sb.WorkspacePath()
	}

	// Determine effective timeout.
	timeout := defaultCommandTimeout
	if in.TimeoutSeconds > 0 {
		timeout = time.Duration(in.TimeoutSeconds) * time.Second
	}

	// Create a child context with the command timeout.
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd.exe", "/C", in.Command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", in.Command)
	}
	cmd.Dir = workdir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		// Timeout — process was killed by the context deadline.
		if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
			return fmt.Sprintf("[TIMEOUT after %s]\n%s", timeout, output), nil
		}

		// Non-zero exit code — return output without error so the agent can decide.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Sprintf("Exit code %d:\n%s", exitErr.ExitCode(), output), nil
		}

		// Command could not be started at all (e.g., executable not found).
		// This is a real engine error — fail fast.
		return output, fmt.Errorf("run_terminal_command: failed to start command: %w", err)
	}
	return output, nil
}
