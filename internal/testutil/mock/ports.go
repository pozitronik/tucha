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
