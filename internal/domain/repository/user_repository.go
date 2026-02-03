package repository

import (
	"tucha/internal/domain/entity"
)

// UserRepository persists and retrieves user accounts.
type UserRepository interface {
	// Upsert creates or updates a user by email. Returns the user ID.
	Upsert(email, password string) (int64, error)

	// GetByEmail retrieves a user by email. Returns nil, nil if not found.
	GetByEmail(email string) (*entity.User, error)
}
