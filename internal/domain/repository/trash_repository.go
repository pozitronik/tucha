package repository

import (
	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// TrashRepository persists and retrieves soft-deleted nodes.
type TrashRepository interface {
	// Insert copies a node and its descendants into the trash table.
	// deletedFrom records the original parent path, deletedBy records who performed the deletion.
	Insert(userID int64, node *entity.Node, descendants []entity.Node, deletedBy int64) error

	// List returns all trash items for a given user.
	List(userID int64) ([]entity.TrashItem, error)

	// GetByPathAndRev finds a specific trash item by its original path and revision.
	// Returns nil, nil if not found.
	GetByPathAndRev(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error)

	// Delete removes a single trash item by ID.
	Delete(id int64) error

	// DeleteAll removes all trash items for a user and returns the deleted items
	// so callers can clean up associated content.
	DeleteAll(userID int64) ([]entity.TrashItem, error)
}
