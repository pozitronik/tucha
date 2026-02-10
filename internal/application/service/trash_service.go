package service

import (
	"github.com/pozitronik/tucha/internal/application/port"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// TrashService handles soft-deletion (trash), listing, restoring, and emptying the trashbin.
type TrashService struct {
	nodes    repository.NodeRepository
	trash    repository.TrashRepository
	contents repository.ContentRepository
	storage  port.ContentStorage
	shares   repository.ShareRepository
}

// NewTrashService creates a new TrashService.
func NewTrashService(
	nodes repository.NodeRepository,
	trash repository.TrashRepository,
	contents repository.ContentRepository,
	storage port.ContentStorage,
	shares repository.ShareRepository,
) *TrashService {
	return &TrashService{
		nodes:    nodes,
		trash:    trash,
		contents: contents,
		storage:  storage,
		shares:   shares,
	}
}

// Trash soft-deletes a node by moving it (and its descendants) to the trash table,
// then hard-deleting from nodes. Content ref counts are NOT decremented -- content
// remains available while in trash.
// When a shared folder is trashed, mounted RW shares are cloned into the mount
// user's tree (so they keep the data), then all share records are removed.
// No-op if the path does not exist (per protocol).
func (s *TrashService) Trash(userID int64, path vo.CloudPath, deletedBy int64) error {
	node, descendants, err := s.nodes.GetWithDescendants(userID, path)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}

	// Clone mounted RW shares before the source tree is deleted.
	affectedShares := s.cloneMountedRWShares(userID, path)

	if err := s.trash.Insert(userID, node, descendants, deletedBy); err != nil {
		return err
	}

	if err := s.nodes.Delete(userID, path); err != nil {
		return err
	}

	// Remove all share records pointing at the trashed subtree.
	s.deleteShares(affectedShares)

	return nil
}

// cloneMountedRWShares finds shares affected by trashing the given path.
// For each mounted RW share, it clones the shared content into the mount user's
// tree so the data persists after the source is deleted.
// Returns the list of affected shares for later cleanup.
func (s *TrashService) cloneMountedRWShares(ownerID int64, path vo.CloudPath) []entity.Share {
	shares, err := s.shares.ListByOwnerPathPrefix(ownerID, path)
	if err != nil {
		return nil
	}

	for i := range shares {
		share := &shares[i]
		if !share.IsAccepted() || share.Access != vo.AccessReadWrite || share.MountUserID == nil {
			continue
		}
		mountPath := vo.NewCloudPath(share.MountHome)
		_ = s.nodes.EnsurePath(*share.MountUserID, mountPath)
		_ = cloneTree(s.nodes, s.contents, ownerID, share.Home, *share.MountUserID, mountPath)
	}

	return shares
}

// deleteShares removes the given share records from the repository.
func (s *TrashService) deleteShares(shares []entity.Share) {
	for i := range shares {
		_ = s.shares.Delete(shares[i].ID)
	}
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
