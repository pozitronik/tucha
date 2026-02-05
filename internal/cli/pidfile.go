package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WritePID writes the current process ID to the specified file.
func WritePID(path string) error {
	pid := os.Getpid()
	content := strconv.Itoa(pid)
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadPID reads the process ID from the specified file.
// Returns 0 and an error if the file doesn't exist or contains invalid data.
func ReadPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	return pid, nil
}

// RemovePID removes the PID file at the specified path.
// Returns nil if the file doesn't exist.
func RemovePID(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
