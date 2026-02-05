package service

import (
	"github.com/pozitronik/tucha/internal/application/port"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// FileService handles file CRUD, deduplication, quota checks, conflict resolution,
// and content lifecycle management.
type FileService struct {
	nodes    repository.NodeRepository
	contents repository.ContentRepository
	storage  port.ContentStorage
	quota    *QuotaService
}

// NewFileService creates a new FileService.
func NewFileService(
	nodes repository.NodeRepository,
	contents repository.ContentRepository,
	storage port.ContentStorage,
	quota *QuotaService,
) *FileService {
	return &FileService{
		nodes:    nodes,
		contents: contents,
		storage:  storage,
		quota:    quota,
	}
}

// Get retrieves a single node by path.
func (s *FileService) Get(userID int64, path vo.CloudPath) (*entity.Node, error) {
	return s.nodes.Get(userID, path)
}

// CountChildren returns the count of folders and files that are direct children.
func (s *FileService) CountChildren(userID int64, path vo.CloudPath) (folders, files int, err error) {
	return s.nodes.CountChildren(userID, path)
}

// AddByHash registers a file by its content hash (deduplication endpoint).
// Returns the created node or an error.
func (s *FileService) AddByHash(userID int64, path vo.CloudPath, hash vo.ContentHash, size int64, conflict vo.ConflictMode) (*entity.Node, error) {
	// Check if content exists in DB or on disk.
	dbExists, err := s.contents.Exists(hash)
	if err != nil {
		return nil, err
	}
	if !dbExists && !s.storage.Exists(hash) {
		return nil, ErrContentNotFound
	}

	// Increment content reference count.
	if _, err := s.contents.Insert(hash, size); err != nil {
		return nil, err
	}

	// Check quota.
	overQuota, err := s.quota.CheckQuota(userID, size)
	if err != nil {
		return nil, err
	}
	if overQuota {
		return nil, ErrOverQuota
	}

	// Handle conflict.
	exists, err := s.nodes.Exists(userID, path)
	if err != nil {
		return nil, err
	}
	if exists {
		if conflict == vo.ConflictStrict {
			return nil, ErrAlreadyExists
		}
		// Delete existing node before creating the replacement.
		if err := s.nodes.Delete(userID, path); err != nil {
			return nil, err
		}
	}

	if err := s.nodes.EnsurePath(userID, path.Parent()); err != nil {
		return nil, err
	}
	return s.nodes.CreateFile(userID, path, hash, size)
}

// Remove deletes a file or folder, handling content reference counting and disk cleanup.
// Always succeeds per protocol (no error if path does not exist).
func (s *FileService) Remove(userID int64, path vo.CloudPath) error {
	node, _ := s.nodes.Get(userID, path)
	if node != nil && node.HasContent() {
		deleted, _ := s.contents.Decrement(node.Hash)
		if deleted {
			_ = s.storage.Delete(node.Hash)
		}
	}

	return s.nodes.Delete(userID, path)
}

// Rename changes the name of a file or folder.
func (s *FileService) Rename(userID int64, path vo.CloudPath, newName string) (*entity.Node, error) {
	return s.nodes.Rename(userID, path, newName)
}

// Move moves a file or folder to a target directory.
func (s *FileService) Move(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	if err := s.nodes.EnsurePath(userID, targetFolder); err != nil {
		return nil, err
	}
	return s.nodes.Move(userID, srcPath, targetFolder)
}

// Copy duplicates a file or folder into a target directory.
func (s *FileService) Copy(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	if err := s.nodes.EnsurePath(userID, targetFolder); err != nil {
		return nil, err
	}
	return s.nodes.Copy(userID, srcPath, targetFolder)
}
