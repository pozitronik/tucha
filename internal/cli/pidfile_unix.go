//go:build !windows

package cli

import (
	"syscall"
)

// IsProcessRunning checks if a process with the given PID is running.
// On Unix systems, this uses signal 0 which checks process existence without sending a signal.
func IsProcessRunning(pid int) bool {
	process, err := findProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, sending signal 0 checks if the process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// StopProcess sends a termination signal to the process with the given PID.
// On Unix systems, this sends SIGTERM.
func StopProcess(pid int) error {
	process, err := findProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(syscall.SIGTERM)
}
