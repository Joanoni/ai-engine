//go:build windows

package sandbox

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// attachJobObject creates a Windows Job Object, configures it to kill all
// child processes when the last handle is closed, and assigns the shell
// process to it. The Job Object handle is stored in s.job.
func (s *Shell) attachJobObject() error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("shell: CreateJobObject: %w", err)
	}

	// Configure: kill all processes in the job when the last handle closes.
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		windows.CloseHandle(job)
		return fmt.Errorf("shell: SetInformationJobObject: %w", err)
	}

	// Open a handle to the shell process using its PID.
	pid := uint32(s.cmd.Process.Pid)
	handle, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, pid)
	if err != nil {
		windows.CloseHandle(job)
		return fmt.Errorf("shell: OpenProcess: %w", err)
	}
	defer windows.CloseHandle(handle)

	// Assign the shell process to the job.
	if err := windows.AssignProcessToJobObject(job, handle); err != nil {
		windows.CloseHandle(job)
		return fmt.Errorf("shell: AssignProcessToJobObject: %w", err)
	}

	s.job = uintptr(job)
	return nil
}

// closeJobObject closes the Windows Job Object handle, triggering OS-level
// kill of all processes in the job tree.
func closeJobObject(handle uintptr) {
	if handle != 0 {
		windows.CloseHandle(windows.Handle(handle))
	}
}
