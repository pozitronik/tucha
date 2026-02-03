package service

import (
	"fmt"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
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

// List returns all users.
func (s *UserService) List() ([]entity.User, error) {
	return s.users.List()
}

// Update modifies an existing user's fields.
func (s *UserService) Update(user *entity.User) error {
	existing, err := s.users.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	return s.users.Update(user)
}

// Delete removes a user by ID. Self-deletion is blocked.
func (s *UserService) Delete(callerID, targetID int64) error {
	if callerID == targetID {
		return ErrSelfDelete
	}

	existing, err := s.users.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	return s.users.Delete(targetID)
}
