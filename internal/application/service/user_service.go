package service

import (
	"fmt"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
)

// UserService handles user CRUD operations.
type UserService struct {
	users             repository.UserRepository
	nodes             repository.NodeRepository
	defaultQuotaBytes int64
}

// NewUserService creates a new UserService.
func NewUserService(users repository.UserRepository, nodes repository.NodeRepository, defaultQuotaBytes int64) *UserService {
	return &UserService{users: users, nodes: nodes, defaultQuotaBytes: defaultQuotaBytes}
}

// Create adds a new user and creates their root node.
func (s *UserService) Create(email, password string, isAdmin bool, quotaBytes int64) (*entity.User, error) {
	if quotaBytes <= 0 {
		quotaBytes = s.defaultQuotaBytes
	}

	existing, err := s.users.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	user := &entity.User{
		Email:      email,
		Password:   password,
		IsAdmin:    isAdmin,
		QuotaBytes: quotaBytes,
	}

	id, err := s.users.Create(user)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	user.ID = id

	if _, err := s.nodes.CreateRootNode(id); err != nil {
		return nil, fmt.Errorf("creating root node: %w", err)
	}

	return user, nil
}

// UserWithUsage pairs a user with their current disk usage.
type UserWithUsage struct {
	entity.User
	BytesUsed int64
}

// List returns all users.
func (s *UserService) List() ([]entity.User, error) {
	return s.users.List()
}

// ListWithUsage returns all users with their current disk usage.
func (s *UserService) ListWithUsage() ([]UserWithUsage, error) {
	users, err := s.users.List()
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	result := make([]UserWithUsage, 0, len(users))
	for _, u := range users {
		used, err := s.nodes.TotalSize(u.ID)
		if err != nil {
			return nil, fmt.Errorf("getting usage for user %d: %w", u.ID, err)
		}
		result = append(result, UserWithUsage{User: u, BytesUsed: used})
	}

	return result, nil
}

// Update modifies an existing user's fields.
// Zero-value fields are treated as "no change": empty Email and Password
// preserve the existing values, and QuotaBytes <= 0 keeps the current quota.
// IsAdmin is always applied because the caller must set it explicitly.
func (s *UserService) Update(user *entity.User) error {
	existing, err := s.users.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	if user.Email != "" {
		existing.Email = user.Email
	}
	if user.Password != "" {
		existing.Password = user.Password
	}
	existing.IsAdmin = user.IsAdmin
	if user.QuotaBytes > 0 {
		existing.QuotaBytes = user.QuotaBytes
	}

	return s.users.Update(existing)
}

// Delete removes a user by ID.
func (s *UserService) Delete(targetID int64) error {
	existing, err := s.users.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	return s.users.Delete(targetID)
}
