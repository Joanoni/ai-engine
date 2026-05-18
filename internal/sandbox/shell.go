package sandbox

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Shell is a persistent shell process (cmd.exe on Windows, sh on Unix) that
// stays alive for the duration of a single agent execution. All run_terminal_command
// calls share the same process, so state (working directory, env vars) is preserved
// between calls.
//
// On Windows, the shell process is associated with a Windows Job Object so that
// all child processes are automatically killed when the shell is closed — even
// processes that were started in the background with &.
type Shell struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	pw     *os.File // write end of the combined stdout/stderr pipe
	mu     sync.Mutex
	closed bool
	job    uintptr // Windows Job Object handle; 0 on non-Windows
}

// NewShell starts a new persistent shell with workdir as the initial working directory.
func NewShell(workdir string) (*Shell, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		cmd = exec.Command("sh")
	}
	cmd.Dir = workdir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("shell: stdin pipe: %w", err)
	}

	// Use a single os.Pipe so both stdout and stderr are merged into one reader.
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("shell: os.Pipe: %w", err)
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pr.Close()
		pw.Close()
		return nil, fmt.Errorf("shell: start: %w", err)
	}

	// Close the write end in the parent process — the child owns it now.
	// If we keep pw open in the parent, reads from pr will never get EOF.
	pw.Close()

	s := &Shell{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(pr),
		pw:     pw,
	}

	// On Windows, create a Job Object and assign the shell process to it.
	// When the Job Object handle is closed (in Shell.Close), the OS kills
	// the entire process tree automatically.
	if runtime.GOOS == "windows" {
		if err := s.attachJobObject(); err != nil {
			// Non-fatal: continue without Job Object protection.
			_ = err
		}
	}

	return s, nil
}

// Exec runs a command in the persistent shell and returns the combined output.
// If the command does not complete within timeout, the shell is killed and
// "[TIMEOUT after Xs]" is prepended to any partial output collected.
func (s *Shell) Exec(command string, timeout time.Duration) (output string, timedOut bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return "shell: process has been closed", false
	}

	// Use a unique sentinel so we know when the command output ends.
	const sentinel = "__AI_ENGINE_CMD_DONE_7f3a9b2c__"

	// Write the command followed by an unconditional echo of the sentinel.
	var line string
	if runtime.GOOS == "windows" {
		line = fmt.Sprintf("%s\r\necho %s\r\n", command, sentinel)
	} else {
		line = fmt.Sprintf("%s\necho %s\n", command, sentinel)
	}

	if _, err := io.WriteString(s.stdin, line); err != nil {
		return fmt.Sprintf("shell: write error: %v", err), false
	}

	// Read output lines until we see the sentinel or timeout.
	var buf strings.Builder
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			l, err := s.stdout.ReadString('\n')
			if strings.TrimSpace(l) == sentinel {
				return
			}
			if l != "" {
				buf.WriteString(l)
			}
			if err != nil {
				return
			}
		}
	}()

	select {
	case <-done:
		return buf.String(), false
	case <-time.After(timeout):
		// Kill the shell (and via Job Object, all its children).
		s.closed = true
		_ = s.cmd.Process.Kill()
		closeJobObject(s.job)
		return fmt.Sprintf("[TIMEOUT after %s]\n%s", timeout, buf.String()), true
	}
}

// Close terminates the shell process and the associated Job Object.
// Safe to call multiple times.
func (s *Shell) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	_ = s.stdin.Close()
	_ = s.cmd.Process.Kill()
	_ = s.cmd.Wait()
	closeJobObject(s.job)
}
