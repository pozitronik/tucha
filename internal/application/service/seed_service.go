package service

import (
	"fmt"
	"log"

	"tucha/internal/domain/repository"
	"tucha/internal/domain/vo"
)

// SeedService initializes the configured user and root node.
type SeedService struct {
	users             repository.UserRepository
	nodes             repository.NodeRepository
	defaultQuotaBytes int64
}

// NewSeedService creates a new SeedService.
func NewSeedService(users repository.UserRepository, nodes repository.NodeRepository, defaultQuotaBytes int64) *SeedService {
	return &SeedService{users: users, nodes: nodes, defaultQuotaBytes: defaultQuotaBytes}
}

// Seed ensures the configured user exists and has a root node.
// Returns the user ID.
func (s *SeedService) Seed(email, password string, isAdmin bool) (int64, error) {
	userID, err := s.users.Upsert(email, password, isAdmin, s.defaultQuotaBytes)
	if err != nil {
		return 0, fmt.Errorf("seeding user: %w", err)
	}

	root := vo.NewCloudPath("/")
	exists, err := s.nodes.Exists(userID, root)
	if err != nil {
		return 0, fmt.Errorf("checking root node: %w", err)
	}

	if !exists {
		if _, err := s.nodes.CreateRootNode(userID); err != nil {
			return 0, fmt.Errorf("creating root node: %w", err)
		}
		log.Printf("Created root node for user %s", email)
	}

	return userID, nil
}
