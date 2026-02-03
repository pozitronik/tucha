package service

import (
	"tucha/internal/domain/repository"
)

// QuotaService checks storage quota usage on a per-user basis.
type QuotaService struct {
	nodes repository.NodeRepository
	users repository.UserRepository
}

// NewQuotaService creates a new QuotaService.
func NewQuotaService(nodes repository.NodeRepository, users repository.UserRepository) *QuotaService {
	return &QuotaService{nodes: nodes, users: users}
}

// SpaceUsage holds the result of a quota check.
type SpaceUsage struct {
	Overquota  bool
	BytesTotal int64
	BytesUsed  int64
}

// GetUsage returns the current storage usage for the given user.
func (s *QuotaService) GetUsage(userID int64) (*SpaceUsage, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	used, err := s.nodes.TotalSize(userID)
	if err != nil {
		return nil, err
	}
	return &SpaceUsage{
		Overquota:  used > user.QuotaBytes,
		BytesTotal: user.QuotaBytes,
		BytesUsed:  used,
	}, nil
}

// CheckQuota returns true if adding additionalBytes would exceed the user's quota.
func (s *QuotaService) CheckQuota(userID int64, additionalBytes int64) (bool, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, ErrNotFound
	}

	used, err := s.nodes.TotalSize(userID)
	if err != nil {
		return false, err
	}
	return used+additionalBytes > user.QuotaBytes, nil
}
