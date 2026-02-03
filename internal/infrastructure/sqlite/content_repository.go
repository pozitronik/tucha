package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"tucha/internal/domain/vo"
)

// ContentRepository implements repository.ContentRepository using SQLite.
type ContentRepository struct {
	db *sql.DB
}

// NewContentRepository creates a ContentRepository from the given database connection.
func NewContentRepository(db *DB) *ContentRepository {
	return &ContentRepository{db: db.Conn()}
}

// Insert registers a new content entry or increments the reference count
// if the hash already exists. Returns true if newly inserted.
func (r *ContentRepository) Insert(hash vo.ContentHash, size int64) (bool, error) {
	now := time.Now().Unix()
	res, err := r.db.Exec(
		`INSERT INTO contents (hash, size, ref_count, created) VALUES (?, ?, 1, ?)
		 ON CONFLICT(hash) DO UPDATE SET ref_count = ref_count + 1`,
		hash.String(), size, now,
	)
	if err != nil {
		return false, fmt.Errorf("inserting content: %w", err)
	}
	affected, _ := res.RowsAffected()
	return affected == 1, nil
}

// Decrement decreases the reference count for the given hash.
// Returns true if the reference count reached zero (content can be deleted from disk).
func (r *ContentRepository) Decrement(hash vo.ContentHash) (bool, error) {
	_, err := r.db.Exec(
		"UPDATE contents SET ref_count = ref_count - 1 WHERE hash = ?",
		hash.String(),
	)
	if err != nil {
		return false, fmt.Errorf("decrementing content ref: %w", err)
	}

	var refCount int64
	err = r.db.QueryRow(
		"SELECT ref_count FROM contents WHERE hash = ?",
		hash.String(),
	).Scan(&refCount)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking ref count: %w", err)
	}

	if refCount <= 0 {
		_, err = r.db.Exec("DELETE FROM contents WHERE hash = ?", hash.String())
		if err != nil {
			return false, fmt.Errorf("deleting content record: %w", err)
		}
		return true, nil
	}
	return false, nil
}

// Exists checks whether content with the given hash is registered.
func (r *ContentRepository) Exists(hash vo.ContentHash) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM contents WHERE hash = ?)",
		hash.String(),
	).Scan(&exists)
	return exists, err
}
