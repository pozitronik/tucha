//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// daemonize starts the server as a background daemon process.
// On Unix, this forks and detaches from the terminal.
func daemonize(configPath string) error {
	// Re-execute ourselves without --background
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "-config", configPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session, detach from terminal
	}

	// Detach from parent's stdin/stdout/stderr
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}
