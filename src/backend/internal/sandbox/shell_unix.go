//go:build !windows

package sandbox

// attachJobObject is a no-op on non-Windows platforms.
// On Unix, exec.CommandContext with SIGKILL already kills the process group.
func (s *Shell) attachJobObject() error {
	return nil
}

// closeJobObject is a no-op on non-Windows platforms.
func closeJobObject(handle uintptr) {}
