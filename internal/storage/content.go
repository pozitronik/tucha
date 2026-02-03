package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// ContentStore handles content reference counting in the database.
type ContentStore struct {
	db *sql.DB
}

// NewContentStore creates a new ContentStore using the given database connection.
func NewContentStore(db *DB) *ContentStore {
	return &ContentStore{db: db.Conn()}
}

// Insert registers a new content entry or increments the reference count
// if the hash already exists. Returns true if the content was newly inserted.
func (s *ContentStore) Insert(hash string, size int64) (bool, error) {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO contents (hash, size, ref_count, created) VALUES (?, ?, 1, ?)
		 ON CONFLICT(hash) DO UPDATE SET ref_count = ref_count + 1`,
		hash, size, now,
	)
	if err != nil {
		return false, fmt.Errorf("inserting content: %w", err)
	}
	affected, _ := res.RowsAffected()
	return affected == 1, nil
}

// Decrement decreases the reference count for the given hash.
// Returns true if the reference count reached zero (content can be deleted).
func (s *ContentStore) Decrement(hash string) (bool, error) {
	_, err := s.db.Exec(
		"UPDATE contents SET ref_count = ref_count - 1 WHERE hash = ?",
		hash,
	)
	if err != nil {
		return false, fmt.Errorf("decrementing content ref: %w", err)
	}

	var refCount int64
	err = s.db.QueryRow(
		"SELECT ref_count FROM contents WHERE hash = ?",
		hash,
	).Scan(&refCount)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking ref count: %w", err)
	}

	if refCount <= 0 {
		_, err = s.db.Exec("DELETE FROM contents WHERE hash = ?", hash)
		if err != nil {
			return false, fmt.Errorf("deleting content record: %w", err)
		}
		return true, nil
	}
	return false, nil
}

// Exists checks whether content with the given hash exists in the store.
func (s *ContentStore) Exists(hash string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM contents WHERE hash = ?)",
		hash,
	).Scan(&exists)
	return exists, err
}
