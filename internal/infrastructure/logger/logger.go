// Package logger provides a leveled logging implementation.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents a logging severity level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// ParseLevel converts a string to a Level.
// Returns LevelInfo for unrecognized strings.
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// String returns the level name.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// StandardLogger is a leveled logger backed by stdlib log.Logger.
type StandardLogger struct {
	level  Level
	logger *log.Logger
	file   *os.File // kept for closing, nil if not using file output
	mu     sync.Mutex
}

// New creates a new StandardLogger.
// level: minimum level to log (debug, info, warn, error)
// output: where to write logs (stdout, file, both)
// filePath: path to log file (required if output is "file" or "both")
func New(level, output, filePath string) (*StandardLogger, error) {
	lvl := ParseLevel(level)
	out := strings.ToLower(strings.TrimSpace(output))
	if out == "" {
		out = "stdout"
	}

	var w io.Writer
	var file *os.File

	switch out {
	case "stdout":
		w = os.Stdout
	case "file":
		if filePath == "" {
			return nil, fmt.Errorf("file path required when output is 'file'")
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
		w = f
		file = f
	case "both":
		if filePath == "" {
			return nil, fmt.Errorf("file path required when output is 'both'")
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
		w = io.MultiWriter(os.Stdout, f)
		file = f
	default:
		return nil, fmt.Errorf("unknown output type: %q (expected stdout, file, or both)", out)
	}

	return &StandardLogger{
		level:  lvl,
		logger: log.New(w, "", 0), // no prefix, custom format
		file:   file,
	}, nil
}

// Close releases resources (closes log file if open).
func (l *StandardLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Debug logs at DEBUG level.
func (l *StandardLogger) Debug(msg string, args ...any) {
	l.log(LevelDebug, msg, args...)
}

// Info logs at INFO level.
func (l *StandardLogger) Info(msg string, args ...any) {
	l.log(LevelInfo, msg, args...)
}

// Warn logs at WARN level.
func (l *StandardLogger) Warn(msg string, args ...any) {
	l.log(LevelWarn, msg, args...)
}

// Error logs at ERROR level.
func (l *StandardLogger) Error(msg string, args ...any) {
	l.log(LevelError, msg, args...)
}

// log writes a formatted message if the level is enabled.
func (l *StandardLogger) log(level Level, msg string, args ...any) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format: 2006-01-02 15:04:05 [LEVEL] message
	formatted := fmt.Sprintf(msg, args...)
	l.logger.Printf("%s [%s] %s", timeNow().Format("2006-01-02 15:04:05"), level, formatted)
}

// timeNow is a variable for testing purposes.
var timeNow = time.Now
