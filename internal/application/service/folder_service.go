package service

import (
	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
)

// FolderService handles folder listing and creation.
type FolderService struct {
	nodes repository.NodeRepository
}

// NewFolderService creates a new FolderService.
func NewFolderService(nodes repository.NodeRepository) *FolderService {
	return &FolderService{nodes: nodes}
}

// Get retrieves a single node by path.
// Returns nil, nil if not found.
func (s *FolderService) Get(userID int64, path vo.CloudPath) (*entity.Node, error) {
	return s.nodes.Get(userID, path)
}

// ListChildren returns the children of the given folder, with offset/limit pagination.
func (s *FolderService) ListChildren(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
	return s.nodes.ListChildren(userID, path, offset, limit)
}

// CountChildren returns the count of folders and files that are direct children.
func (s *FolderService) CountChildren(userID int64, path vo.CloudPath) (folders, files int, err error) {
	return s.nodes.CountChildren(userID, path)
}

// CreateFolder creates a new folder at the given path.
// Returns an error if the path already exists.
func (s *FolderService) CreateFolder(userID int64, path vo.CloudPath) (*entity.Node, error) {
	exists, err := s.nodes.Exists(userID, path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}
	return s.nodes.CreateFolder(userID, path)
}
