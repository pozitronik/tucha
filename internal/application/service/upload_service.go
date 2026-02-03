package service

import (
	"bytes"

	"tucha/internal/application/port"
	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
)

// UploadService handles binary uploads: hash computation, disk storage, and DB registration.
type UploadService struct {
	hasher   port.Hasher
	storage  port.ContentStorage
	contents repository.ContentRepository
}

// NewUploadService creates a new UploadService.
func NewUploadService(
	hasher port.Hasher,
	storage port.ContentStorage,
	contents repository.ContentRepository,
) *UploadService {
	return &UploadService{
		hasher:   hasher,
		storage:  storage,
		contents: contents,
	}
}

// Upload processes a binary upload: computes the hash, stores the content, and registers it.
// Returns the content hash.
func (s *UploadService) Upload(data []byte) (vo.ContentHash, error) {
	hash := s.hasher.Compute(data)
	size := int64(len(data))

	if _, err := s.storage.Write(hash, bytes.NewReader(data)); err != nil {
		return vo.ContentHash{}, err
	}

	if _, err := s.contents.Insert(hash, size); err != nil {
		return vo.ContentHash{}, err
	}

	return hash, nil
}
