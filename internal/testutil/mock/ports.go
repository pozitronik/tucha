package mock

import (
	"io"
	"os"

	"tucha/internal/domain/vo"
)

// ContentStorageMock is a test double for port.ContentStorage.
type ContentStorageMock struct {
	WriteFunc  func(hash vo.ContentHash, r io.Reader) (int64, error)
	OpenFunc   func(hash vo.ContentHash) (*os.File, error)
	DeleteFunc func(hash vo.ContentHash) error
	ExistsFunc func(hash vo.ContentHash) bool
}

func (m *ContentStorageMock) Write(hash vo.ContentHash, r io.Reader) (int64, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(hash, r)
	}
	return 0, nil
}

func (m *ContentStorageMock) Open(hash vo.ContentHash) (*os.File, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(hash)
	}
	return nil, os.ErrNotExist
}

func (m *ContentStorageMock) Delete(hash vo.ContentHash) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(hash)
	}
	return nil
}

func (m *ContentStorageMock) Exists(hash vo.ContentHash) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(hash)
	}
	return false
}

// HasherMock is a test double for port.Hasher.
type HasherMock struct {
	ComputeFunc       func(data []byte) vo.ContentHash
	ComputeReaderFunc func(r io.Reader, size int64) (vo.ContentHash, error)
	FixedHash         vo.ContentHash
}

func (m *HasherMock) Compute(data []byte) vo.ContentHash {
	if m.ComputeFunc != nil {
		return m.ComputeFunc(data)
	}
	return m.FixedHash
}

func (m *HasherMock) ComputeReader(r io.Reader, size int64) (vo.ContentHash, error) {
	if m.ComputeReaderFunc != nil {
		return m.ComputeReaderFunc(r, size)
	}
	return m.FixedHash, nil
}

// LogEntry represents a captured log message for testing.
type LogEntry struct {
	Level string
	Msg   string
	Args  []any
}

// LoggerMock is a test double for port.Logger.
type LoggerMock struct {
	DebugFunc func(msg string, args ...any)
	InfoFunc  func(msg string, args ...any)
	WarnFunc  func(msg string, args ...any)
	ErrorFunc func(msg string, args ...any)
	Captured  []LogEntry
}

func (m *LoggerMock) Debug(msg string, args ...any) {
	m.Captured = append(m.Captured, LogEntry{Level: "DEBUG", Msg: msg, Args: args})
	if m.DebugFunc != nil {
		m.DebugFunc(msg, args...)
	}
}

func (m *LoggerMock) Info(msg string, args ...any) {
	m.Captured = append(m.Captured, LogEntry{Level: "INFO", Msg: msg, Args: args})
	if m.InfoFunc != nil {
		m.InfoFunc(msg, args...)
	}
}

func (m *LoggerMock) Warn(msg string, args ...any) {
	m.Captured = append(m.Captured, LogEntry{Level: "WARN", Msg: msg, Args: args})
	if m.WarnFunc != nil {
		m.WarnFunc(msg, args...)
	}
}

func (m *LoggerMock) Error(msg string, args ...any) {
	m.Captured = append(m.Captured, LogEntry{Level: "ERROR", Msg: msg, Args: args})
	if m.ErrorFunc != nil {
		m.ErrorFunc(msg, args...)
	}
}
