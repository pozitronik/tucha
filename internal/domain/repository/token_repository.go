package repository

import (
	"tucha/internal/domain/entity"
)

// TokenRepository persists and retrieves authentication tokens.
type TokenRepository interface {
	// Create generates a new token set for the given user and stores it.
	Create(userID int64, ttlSeconds int) (*entity.Token, error)

	// LookupAccess finds a token by its access_token value.
	// Returns nil, nil if not found. Does NOT check expiration -- that is the caller's responsibility.
	LookupAccess(accessToken string) (*entity.Token, error)
}
