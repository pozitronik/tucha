//go:build windows

package cli

import (
	"os/exec"
	"strconv"
	"strings"
)

// IsProcessRunning checks if a process with the given PID is running.
// On Windows, this uses tasklist to check process existence.
func IsProcessRunning(pid int) bool {
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// tasklist output contains the PID if the process exists
	return strings.Contains(string(output), strconv.Itoa(pid))
}

// StopProcess terminates the process with the given PID.
// On Windows, this uses taskkill.
func StopProcess(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	return cmd.Run()
}
