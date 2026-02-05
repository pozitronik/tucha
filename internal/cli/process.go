package cli

import (
	"os"
)

// findProcess wraps os.FindProcess for testability and consistent behavior.
func findProcess(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}
