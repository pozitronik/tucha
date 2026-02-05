package service

import (
	"tucha/internal/application/port"
	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
)

// TrashService handles soft-deletion (trash), listing, restoring, and emptying the trashbin.
type TrashService struct {
	nodes    repository.NodeRepository
	trash    repository.TrashRepository
	contents repository.ContentRepository
	storage  port.ContentStorage
}

// NewTrashService creates a new TrashService.
func NewTrashService(
	nodes repository.NodeRepository,
	trash repository.TrashRepository,
	contents repository.ContentRepository,
	storage port.ContentStorage,
) *TrashService {
	return &TrashService{
		nodes:    nodes,
		trash:    trash,
		contents: contents,
		storage:  storage,
	}
}

// Trash soft-deletes a node by moving it (and its descendants) to the trash table,
// then hard-deleting from nodes. Content ref counts are NOT decremented -- content
// remains available while in trash.
// No-op if the path does not exist (per protocol).
func (s *TrashService) Trash(userID int64, path vo.CloudPath, deletedBy int64) error {
	node, descendants, err := s.nodes.GetWithDescendants(userID, path)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}

	if err := s.trash.Insert(userID, node, descendants, deletedBy); err != nil {
		return err
	}

	return s.nodes.Delete(userID, path)
}

// List returns all items in the user's trashbin.
func (s *TrashService) List(userID int64) ([]entity.TrashItem, error) {
	return s.trash.List(userID)
}

// Restore moves a trash item back into the active filesystem.
// The item is identified by its original path and revision.
// conflict determines how to handle an existing node at the target path.
func (s *TrashService) Restore(userID int64, path vo.CloudPath, rev int64, conflict vo.ConflictMode) error {
	item, err := s.trash.GetByPathAndRev(userID, path, rev)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrNotFound
	}

	// Ensure parent directory exists.
	parentPath := path.Parent()
	if err := s.nodes.EnsurePath(userID, parentPath); err != nil {
		return err
	}

	// Handle conflict at target path.
	exists, err := s.nodes.Exists(userID, path)
	if err != nil {
		return err
	}
	if exists {
		if conflict == vo.ConflictStrict {
			return ErrAlreadyExists
		}
		if err := s.nodes.Delete(userID, path); err != nil {
			return err
		}
	}

	// Recreate the node in the active filesystem.
	if item.IsFile() {
		_, err = s.nodes.CreateFile(userID, path, item.Hash, item.Size)
	} else {
		_, err = s.nodes.CreateFolder(userID, path)
	}
	if err != nil {
		return err
	}

	return s.trash.Delete(item.ID)
}

// Empty permanently deletes all items in the user's trashbin and cleans up
// unreferenced content from disk.
func (s *TrashService) Empty(userID int64) error {
	items, err := s.trash.DeleteAll(userID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.HasContent() {
			deleted, err := s.contents.Decrement(item.Hash)
			if err != nil {
				continue
			}
			if deleted {
				_ = s.storage.Delete(item.Hash)
			}
		}
	}

	return nil
}
