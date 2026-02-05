package service

import (
	"tucha/internal/application/port"
	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
	"tucha/internal/infrastructure/thumbnail"
)

// ThumbnailService handles thumbnail generation for images.
type ThumbnailService struct {
	nodes     repository.NodeRepository
	storage   port.ContentStorage
	generator *thumbnail.Generator
}

// NewThumbnailService creates a new ThumbnailService.
func NewThumbnailService(
	nodes repository.NodeRepository,
	storage port.ContentStorage,
	generator *thumbnail.Generator,
) *ThumbnailService {
	return &ThumbnailService{
		nodes:     nodes,
		storage:   storage,
		generator: generator,
	}
}

// ThumbnailResult holds the thumbnail data and metadata.
type ThumbnailResult struct {
	Data        []byte
	ContentType string
}

// Generate creates a thumbnail for the file at the given path.
// Returns ErrNotFound if the file doesn't exist or is not an image.
func (s *ThumbnailService) Generate(userID int64, path vo.CloudPath, presetName string) (*ThumbnailResult, error) {
	node, err := s.nodes.Get(userID, path)
	if err != nil {
		return nil, err
	}
	if node == nil || !node.IsFile() {
		return nil, ErrNotFound
	}

	// Check if file is a supported image format
	if !thumbnail.IsSupportedFormat(node.Name) {
		return nil, ErrNotFound
	}

	// Get preset
	preset := thumbnail.GetPreset(presetName)
	if preset == nil {
		// Default to a common preset if not found
		preset = thumbnail.GetPreset("xw14")
	}

	// Open the source file
	file, err := s.storage.Open(node.Hash)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Generate thumbnail
	result, err := s.generator.Generate(file, node.Hash.String(), *preset)
	if err != nil {
		return nil, err
	}

	return &ThumbnailResult{
		Data:        result.Data,
		ContentType: result.ContentType,
	}, nil
}

// GenerateByHash creates a thumbnail for content identified by hash.
// Used for weblink thumbnails where the node is resolved separately.
func (s *ThumbnailService) GenerateByHash(hash vo.ContentHash, filename string, presetName string) (*ThumbnailResult, error) {
	// Check if file is a supported image format
	if !thumbnail.IsSupportedFormat(filename) {
		return nil, ErrNotFound
	}

	// Get preset
	preset := thumbnail.GetPreset(presetName)
	if preset == nil {
		preset = thumbnail.GetPreset("xw14")
	}

	// Open the source file
	file, err := s.storage.Open(hash)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Generate thumbnail
	result, err := s.generator.Generate(file, hash.String(), *preset)
	if err != nil {
		return nil, err
	}

	return &ThumbnailResult{
		Data:        result.Data,
		ContentType: result.ContentType,
	}, nil
}
