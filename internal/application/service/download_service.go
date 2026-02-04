package service

import (
	"os"

	"tucha/internal/application/port"
	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
)

// DownloadService resolves a cloud path to a content file on disk.
type DownloadService struct {
	nodes   repository.NodeRepository
	storage port.ContentStorage
}

// NewDownloadService creates a new DownloadService.
func NewDownloadService(
	nodes repository.NodeRepository,
	storage port.ContentStorage,
) *DownloadService {
	return &DownloadService{
		nodes:   nodes,
		storage: storage,
	}
}

// DownloadResult holds the file handle and metadata needed to serve a download.
type DownloadResult struct {
	File *os.File
	Node *entity.Node
}

// ResolveByNode opens content for an already-resolved node.
// Used by public weblink downloads where the node is looked up separately.
// Returns ErrNotFound if the node is not a file.
func (s *DownloadService) ResolveByNode(node *entity.Node) (*DownloadResult, error) {
	if node == nil || !node.IsFile() {
		return nil, ErrNotFound
	}

	f, err := s.storage.Open(node.Hash)
	if err != nil {
		return nil, err
	}

	return &DownloadResult{
		File: f,
		Node: node,
	}, nil
}

// Resolve looks up a file node by path and opens the associated content.
// Returns ErrNotFound if the path does not exist or is not a file.
func (s *DownloadService) Resolve(userID int64, path vo.CloudPath) (*DownloadResult, error) {
	node, err := s.nodes.Get(userID, path)
	if err != nil {
		return nil, err
	}
	if node == nil || !node.IsFile() {
		return nil, ErrNotFound
	}

	f, err := s.storage.Open(node.Hash)
	if err != nil {
		return nil, err
	}

	return &DownloadResult{
		File: f,
		Node: node,
	}, nil
}
