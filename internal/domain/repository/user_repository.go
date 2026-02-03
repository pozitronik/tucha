package repository

import (
	"tucha/internal/domain/entity"
)

// UserRepository persists and retrieves user accounts.
type UserRepository interface {
	// Upsert creates or updates a user by email. Returns the user ID.
	Upsert(email, password string, isAdmin bool, quotaBytes int64) (int64, error)

	// Create inserts a new user. Returns the user ID.
	Create(user *entity.User) (int64, error)

	// GetByID retrieves a user by ID. Returns nil, nil if not found.
	GetByID(id int64) (*entity.User, error)

	// GetByEmail retrieves a user by email. Returns nil, nil if not found.
	GetByEmail(email string) (*entity.User, error)

	// List returns all users.
	List() ([]entity.User, error)

	// Update modifies an existing user's fields.
	Update(user *entity.User) error

	// Delete removes a user by ID.
	Delete(id int64) error
}
