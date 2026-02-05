package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPIDFile(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		want       string
	}{
		{
			name:       "simple config path",
			configPath: "/etc/tucha/config.yaml",
			want:       "/etc/tucha/tucha.pid",
		},
		{
			name:       "current directory config",
			configPath: "config.yaml",
			want:       "tucha.pid",
		},
		{
			name:       "nested path",
			configPath: "/home/user/projects/tucha/configs/server.yaml",
			want:       "/home/user/projects/tucha/configs/tucha.pid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultPIDFile(tt.configPath)
			// Normalize paths for comparison (handles OS differences)
			if filepath.Clean(got) != filepath.Clean(tt.want) {
				t.Errorf("DefaultPIDFile(%q) = %q, want %q", tt.configPath, got, tt.want)
			}
		})
	}
}

func TestCheckNotRunning(t *testing.T) {
	t.Run("returns nil when PID file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "nonexistent.pid")

		err := CheckNotRunning(pidFile)
		if err != nil {
			t.Errorf("CheckNotRunning() error = %v, want nil", err)
		}
	})

	t.Run("returns nil for stale PID file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write a PID that definitely doesn't exist (very high number)
		if err := os.WriteFile(pidFile, []byte("9999999"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		err := CheckNotRunning(pidFile)
		if err != nil {
			t.Errorf("CheckNotRunning() error = %v, want nil for stale PID", err)
		}

		// Verify stale PID file was removed
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Error("CheckNotRunning() should have removed stale PID file")
		}
	})

	t.Run("returns error when server is running", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write current process PID (which is definitely running)
		if err := WritePID(pidFile); err != nil {
			t.Fatalf("WritePID() error = %v", err)
		}

		err := CheckNotRunning(pidFile)
		if err == nil {
			t.Error("CheckNotRunning() expected error for running process, got nil")
		}
		if !strings.Contains(err.Error(), "already running") {
			t.Errorf("CheckNotRunning() error = %q, want to contain 'already running'", err.Error())
		}
	})

	t.Run("handles invalid PID content gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write invalid content
		if err := os.WriteFile(pidFile, []byte("not-a-pid"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		// Should not error - treats unreadable PID as "not running"
		err := CheckNotRunning(pidFile)
		if err != nil {
			t.Errorf("CheckNotRunning() error = %v, want nil for invalid PID", err)
		}
	})
}

func TestServerStatus(t *testing.T) {
	t.Run("returns not running when PID file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "nonexistent.pid")

		running, pid, err := ServerStatus(pidFile)
		if err != nil {
			t.Fatalf("ServerStatus() error = %v", err)
		}
		if running {
			t.Error("ServerStatus() running = true, want false")
		}
		if pid != 0 {
			t.Errorf("ServerStatus() pid = %d, want 0", pid)
		}
	})

	t.Run("returns not running for stale PID file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write a PID that definitely doesn't exist
		if err := os.WriteFile(pidFile, []byte("9999999"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		running, pid, err := ServerStatus(pidFile)
		if err != nil {
			t.Fatalf("ServerStatus() error = %v", err)
		}
		if running {
			t.Error("ServerStatus() running = true for stale PID, want false")
		}
		if pid != 0 {
			t.Errorf("ServerStatus() pid = %d for stale PID, want 0", pid)
		}

		// Verify stale PID file was removed
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Error("ServerStatus() should have removed stale PID file")
		}
	})

	t.Run("returns running for active process", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write current process PID
		if err := WritePID(pidFile); err != nil {
			t.Fatalf("WritePID() error = %v", err)
		}

		running, pid, err := ServerStatus(pidFile)
		if err != nil {
			t.Fatalf("ServerStatus() error = %v", err)
		}
		if !running {
			t.Error("ServerStatus() running = false, want true")
		}
		if pid != os.Getpid() {
			t.Errorf("ServerStatus() pid = %d, want %d", pid, os.Getpid())
		}
	})

	t.Run("returns error for unreadable PID file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write invalid content
		if err := os.WriteFile(pidFile, []byte("invalid"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		_, _, err := ServerStatus(pidFile)
		if err == nil {
			t.Error("ServerStatus() expected error for invalid PID, got nil")
		}
	})
}

func TestStopServer(t *testing.T) {
	t.Run("returns error when PID file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "nonexistent.pid")

		err := StopServer(pidFile)
		if err == nil {
			t.Error("StopServer() expected error for nonexistent PID file, got nil")
		}
		if !strings.Contains(err.Error(), "not running") {
			t.Errorf("StopServer() error = %q, want to contain 'not running'", err.Error())
		}
	})

	t.Run("returns error for stale PID file and removes it", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write a PID that definitely doesn't exist
		if err := os.WriteFile(pidFile, []byte("9999999"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		err := StopServer(pidFile)
		if err == nil {
			t.Error("StopServer() expected error for stale PID, got nil")
		}
		if !strings.Contains(err.Error(), "stale PID file") {
			t.Errorf("StopServer() error = %q, want to contain 'stale PID file'", err.Error())
		}

		// Verify stale PID file was removed
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Error("StopServer() should have removed stale PID file")
		}
	})

	t.Run("returns error for unreadable PID file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write invalid content
		if err := os.WriteFile(pidFile, []byte("not-a-pid"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		err := StopServer(pidFile)
		if err == nil {
			t.Error("StopServer() expected error for invalid PID, got nil")
		}
		if !strings.Contains(err.Error(), "reading PID file") {
			t.Errorf("StopServer() error = %q, want to contain 'reading PID file'", err.Error())
		}
	})
}
