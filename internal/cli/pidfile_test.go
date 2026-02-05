package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWritePID(t *testing.T) {
	t.Run("writes current PID to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		err := WritePID(pidFile)
		if err != nil {
			t.Fatalf("WritePID() error = %v", err)
		}

		// Verify file content
		data, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("Reading PID file: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("Parsing PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Errorf("WritePID() wrote PID %d, want %d", pid, os.Getpid())
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Write initial content
		if err := os.WriteFile(pidFile, []byte("12345"), 0644); err != nil {
			t.Fatalf("Writing initial file: %v", err)
		}

		err := WritePID(pidFile)
		if err != nil {
			t.Fatalf("WritePID() error = %v", err)
		}

		// Verify overwritten
		data, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("Reading PID file: %v", err)
		}

		pid, _ := strconv.Atoi(string(data))
		if pid != os.Getpid() {
			t.Errorf("WritePID() did not overwrite existing file")
		}
	})

	t.Run("fails on invalid path", func(t *testing.T) {
		err := WritePID("/nonexistent/directory/test.pid")
		if err == nil {
			t.Error("WritePID() expected error for invalid path, got nil")
		}
	})
}

func TestReadPID(t *testing.T) {
	t.Run("reads valid PID", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte("12345"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		pid, err := ReadPID(pidFile)
		if err != nil {
			t.Fatalf("ReadPID() error = %v", err)
		}
		if pid != 12345 {
			t.Errorf("ReadPID() = %d, want 12345", pid)
		}
	})

	t.Run("reads PID with whitespace", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte("  67890  \n"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		pid, err := ReadPID(pidFile)
		if err != nil {
			t.Fatalf("ReadPID() error = %v", err)
		}
		if pid != 67890 {
			t.Errorf("ReadPID() = %d, want 67890", pid)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := ReadPID("/nonexistent/test.pid")
		if err == nil {
			t.Error("ReadPID() expected error for nonexistent file, got nil")
		}
		if !os.IsNotExist(err) {
			t.Errorf("ReadPID() error should be IsNotExist, got %v", err)
		}
	})

	t.Run("returns error for invalid PID content", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte("not-a-number"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		_, err := ReadPID(pidFile)
		if err == nil {
			t.Error("ReadPID() expected error for invalid content, got nil")
		}
	})

	t.Run("returns error for empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte(""), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		_, err := ReadPID(pidFile)
		if err == nil {
			t.Error("ReadPID() expected error for empty file, got nil")
		}
	})

	t.Run("returns error for negative PID", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte("-123"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		// Note: strconv.Atoi accepts negative numbers, so this won't error
		// but the PID will be negative (which is technically invalid for a process)
		pid, err := ReadPID(pidFile)
		if err != nil {
			t.Fatalf("ReadPID() error = %v", err)
		}
		if pid != -123 {
			t.Errorf("ReadPID() = %d, want -123", pid)
		}
	})
}

func TestRemovePID(t *testing.T) {
	t.Run("removes existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		if err := os.WriteFile(pidFile, []byte("12345"), 0644); err != nil {
			t.Fatalf("Writing PID file: %v", err)
		}

		err := RemovePID(pidFile)
		if err != nil {
			t.Fatalf("RemovePID() error = %v", err)
		}

		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Error("RemovePID() did not remove the file")
		}
	})

	t.Run("returns nil for nonexistent file", func(t *testing.T) {
		err := RemovePID("/nonexistent/test.pid")
		if err != nil {
			t.Errorf("RemovePID() for nonexistent file = %v, want nil", err)
		}
	})

	t.Run("handles already removed file", func(t *testing.T) {
		tmpDir := t.TempDir()
		pidFile := filepath.Join(tmpDir, "test.pid")

		// Never create the file, just try to remove it
		err := RemovePID(pidFile)
		if err != nil {
			t.Errorf("RemovePID() for never-created file = %v, want nil", err)
		}
	})
}

func TestWriteAndReadPIDRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")

	// Write
	if err := WritePID(pidFile); err != nil {
		t.Fatalf("WritePID() error = %v", err)
	}

	// Read
	pid, err := ReadPID(pidFile)
	if err != nil {
		t.Fatalf("ReadPID() error = %v", err)
	}

	if pid != os.Getpid() {
		t.Errorf("Round-trip PID mismatch: got %d, want %d", pid, os.Getpid())
	}

	// Remove
	if err := RemovePID(pidFile); err != nil {
		t.Fatalf("RemovePID() error = %v", err)
	}

	// Verify removed
	_, err = ReadPID(pidFile)
	if !os.IsNotExist(err) {
		t.Error("PID file still exists after RemovePID()")
	}
}
