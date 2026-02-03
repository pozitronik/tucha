// Package content implements a content-addressable filesystem store.
// Files are stored by their hash in a two-level sharded directory structure:
//
//	<base>/C1/72/C172C6E2FF47284FF33F348FEA7EECE532F6C051
//
// First 2 chars -> second 2 chars -> full 40-char hash filename.
package content

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Store manages content-addressable file storage on disk.
type Store struct {
	baseDir string
}

// NewStore creates a new content store at the given base directory.
// The directory is created if it does not exist.
func NewStore(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating content directory: %w", err)
	}
	return &Store{baseDir: baseDir}, nil
}

// Write stores data from the reader under the given hash.
// The hash must be a 40-character hex string.
// Returns the number of bytes written.
func (s *Store) Write(hash string, r io.Reader) (int64, error) {
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

// Open returns a ReadSeekCloser for the content identified by hash.
// Returns os.ErrNotExist if the content does not exist.
func (s *Store) Open(hash string) (*os.File, error) {
	return os.Open(s.path(hash))
}

// Delete removes the content file for the given hash.
// No error is returned if the file does not exist.
func (s *Store) Delete(hash string) error {
	err := os.Remove(s.path(hash))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Exists checks whether content with the given hash exists on disk.
func (s *Store) Exists(hash string) bool {
	_, err := os.Stat(s.path(hash))
	return err == nil
}

// path returns the filesystem path for the given hash using two-level sharding.
func (s *Store) path(hash string) string {
	if len(hash) < 4 {
		return filepath.Join(s.baseDir, hash)
	}
	return filepath.Join(s.baseDir, hash[0:2], hash[2:4], hash)
}
