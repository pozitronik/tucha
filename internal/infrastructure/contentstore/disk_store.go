// Package contentstore implements content-addressable file storage on disk.
package contentstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"tucha/internal/domain/vo"
)

// DiskStore manages content-addressable file storage using two-level directory sharding.
// Storage structure: <base>/C1/72/C172C6E2FF47284FF33F348FEA7EECE532F6C051
type DiskStore struct {
	baseDir string
}

// NewDiskStore creates a new content store at the given base directory.
// The directory is created if it does not exist.
func NewDiskStore(baseDir string) (*DiskStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating content directory: %w", err)
	}
	return &DiskStore{baseDir: baseDir}, nil
}

// Write stores data from the reader under the given hash.
// Returns the number of bytes written.
func (s *DiskStore) Write(hash vo.ContentHash, r io.Reader) (int64, error) {
	p := s.path(hash)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("creating shard directory: %w", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return 0, fmt.Errorf("creating content file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		os.Remove(p)
		return 0, fmt.Errorf("writing content: %w", err)
	}

	return n, nil
}

// Open returns a file handle for the content identified by hash.
// Returns os.ErrNotExist if the content does not exist.
func (s *DiskStore) Open(hash vo.ContentHash) (*os.File, error) {
	return os.Open(s.path(hash))
}

// Delete removes the content file for the given hash.
// No error is returned if the file does not exist.
func (s *DiskStore) Delete(hash vo.ContentHash) error {
	err := os.Remove(s.path(hash))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Exists checks whether content with the given hash exists on disk.
func (s *DiskStore) Exists(hash vo.ContentHash) bool {
	_, err := os.Stat(s.path(hash))
	return err == nil
}

// path returns the filesystem path for the given hash using two-level sharding.
func (s *DiskStore) path(hash vo.ContentHash) string {
	h := hash.String()
	if len(h) < 4 {
		return filepath.Join(s.baseDir, h)
	}
	l1, l2 := hash.ShardPrefix()
	return filepath.Join(s.baseDir, l1, l2, h)
}
