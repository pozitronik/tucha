package service

import (
	"tucha/internal/domain/repository"
)

// QuotaService checks storage quota usage.
type QuotaService struct {
	nodes      repository.NodeRepository
	quotaBytes int64
}

// NewQuotaService creates a new QuotaService.
func NewQuotaService(nodes repository.NodeRepository, quotaBytes int64) *QuotaService {
	return &QuotaService{nodes: nodes, quotaBytes: quotaBytes}
}

// SpaceUsage holds the result of a quota check.
type SpaceUsage struct {
	Overquota  bool
	BytesTotal int64
	BytesUsed  int64
}

// GetUsage returns the current storage usage for the given user.
func (s *QuotaService) GetUsage(userID int64) (*SpaceUsage, error) {
	used, err := s.nodes.TotalSize(userID)
	if err != nil {
		return nil, err
	}
	return &SpaceUsage{
		Overquota:  used > s.quotaBytes,
		BytesTotal: s.quotaBytes,
		BytesUsed:  used,
	}, nil
}

// CheckQuota returns true if adding additionalBytes would exceed the quota.
func (s *QuotaService) CheckQuota(userID int64, additionalBytes int64) (bool, error) {
	used, err := s.nodes.TotalSize(userID)
	if err != nil {
		return false, err
	}
	return used+additionalBytes > s.quotaBytes, nil
}
