package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// DefaultPIDFile returns the default PID file path relative to the config file.
func DefaultPIDFile(configPath string) string {
	dir := filepath.Dir(configPath)
	return filepath.Join(dir, "tucha.pid")
}

// CheckNotRunning verifies that no server is currently running.
// Returns an error if a server is already running.
func CheckNotRunning(pidFile string) error {
	pid, err := ReadPID(pidFile)
	if err != nil {
		// PID file doesn't exist or is unreadable - that's fine
		return nil
	}

	if IsProcessRunning(pid) {
		return fmt.Errorf("server already running (PID: %d)", pid)
	}

	// Stale PID file, remove it
	_ = RemovePID(pidFile)
	return nil
}

// ServerStatus checks if the server is running and returns status information.
func ServerStatus(pidFile string) (running bool, pid int, err error) {
	pid, err = ReadPID(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	if IsProcessRunning(pid) {
		return true, pid, nil
	}

	// Stale PID file
	_ = RemovePID(pidFile)
	return false, 0, nil
}

// StopServer stops the running server.
func StopServer(pidFile string) error {
	pid, err := ReadPID(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("server is not running (no PID file)")
		}
		return fmt.Errorf("reading PID file: %w", err)
	}

	if !IsProcessRunning(pid) {
		_ = RemovePID(pidFile)
		return fmt.Errorf("server is not running (stale PID file removed)")
	}

	if err := StopProcess(pid); err != nil {
		return fmt.Errorf("stopping process: %w", err)
	}

	// Wait for process to stop (up to 10 seconds)
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		if !IsProcessRunning(pid) {
			_ = RemovePID(pidFile)
			return nil
		}
	}

	return fmt.Errorf("process did not stop within timeout")
}

// RunServerWithGracefulShutdown starts the HTTP server with graceful shutdown support.
// It writes a PID file and removes it on shutdown.
func RunServerWithGracefulShutdown(addr string, handler http.Handler, pidFile string, shutdownTimeout time.Duration) error {
	// Write PID file
	if err := WritePID(pidFile); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}
	defer RemovePID(pidFile)

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Channel for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Channel for server errors
	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		_ = sig // Signal received, proceed with graceful shutdown
	case err := <-errChan:
		return err
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return srv.Shutdown(ctx)
}
