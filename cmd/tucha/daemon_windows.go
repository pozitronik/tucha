//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// daemonize starts the server as a background process.
// On Windows, this starts the process with hidden window.
func daemonize(configPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "-config", configPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}

	// Detach from parent's stdin/stdout/stderr
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}
