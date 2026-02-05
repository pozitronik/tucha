package repository

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// ContentRepository manages content reference counting in the database.
type ContentRepository interface {
	// Insert registers a new content entry or increments the reference count
	// if the hash already exists. Returns true if newly inserted.
	Insert(hash vo.ContentHash, size int64) (bool, error)

	// Decrement decreases the reference count for the given hash.
	// Returns true if the reference count reached zero (content can be deleted from disk).
	Decrement(hash vo.ContentHash) (bool, error)

	// Exists checks whether content with the given hash is registered.
	Exists(hash vo.ContentHash) (bool, error)
}
