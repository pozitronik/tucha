package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"  debug  ", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo}, // default
		{"", LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "INFO"}, // unknown defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
			}
		})
	}
}

func TestNew_Stdout(t *testing.T) {
	logger, err := New("info", "stdout", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	if logger.level != LevelInfo {
		t.Errorf("level = %v, want %v", logger.level, LevelInfo)
	}
}

func TestNew_DefaultOutput(t *testing.T) {
	logger, err := New("warn", "", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	if logger.level != LevelWarn {
		t.Errorf("level = %v, want %v", logger.level, LevelWarn)
	}
}

func TestNew_File(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New("debug", "file", logPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("test message")
	logger.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	if !strings.Contains(string(content), "[INFO] test message") {
		t.Errorf("log file content = %q, want to contain [INFO] test message", content)
	}
}

func TestNew_Both(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New("info", "both", logPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("dual output test")
	logger.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	if !strings.Contains(string(content), "[INFO] dual output test") {
		t.Errorf("log file content = %q, want to contain message", content)
	}
}

func TestNew_FileRequired(t *testing.T) {
	_, err := New("info", "file", "")
	if err == nil {
		t.Error("expected error when file path is empty for 'file' output")
	}

	_, err = New("info", "both", "")
	if err == nil {
		t.Error("expected error when file path is empty for 'both' output")
	}
}

func TestNew_UnknownOutput(t *testing.T) {
	_, err := New("info", "invalid", "")
	if err == nil {
		t.Error("expected error for unknown output type")
	}
}

func TestLevelFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New("warn", "file", logPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")
	logger.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	if strings.Contains(string(content), "debug msg") {
		t.Error("debug message should be filtered at warn level")
	}
	if strings.Contains(string(content), "info msg") {
		t.Error("info message should be filtered at warn level")
	}
	if !strings.Contains(string(content), "warn msg") {
		t.Error("warn message should be logged at warn level")
	}
	if !strings.Contains(string(content), "error msg") {
		t.Error("error message should be logged at warn level")
	}
}

func TestLogFormat(t *testing.T) {
	// Override timeNow for deterministic output
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	originalTimeNow := timeNow
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = originalTimeNow }()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New("debug", "file", logPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("hello %s", "world")
	logger.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	expected := "2024-01-15 10:30:45 [INFO] hello world\n"
	if string(content) != expected {
		t.Errorf("log content = %q, want %q", content, expected)
	}
}

func TestClose_NoFile(t *testing.T) {
	logger, err := New("info", "stdout", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should not error when no file is open
	if err := logger.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestLoggerInterface(t *testing.T) {
	// Verify StandardLogger satisfies port.Logger interface
	logger := &StandardLogger{
		level:  LevelDebug,
		logger: nil,
	}

	// Just verify the methods exist with correct signatures
	_ = logger.Debug
	_ = logger.Info
	_ = logger.Warn
	_ = logger.Error
}
