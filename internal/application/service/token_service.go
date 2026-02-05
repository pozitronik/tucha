package service

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
)

// TokenService handles token creation and credential-based authentication.
type TokenService struct {
	tokens repository.TokenRepository
	users  repository.UserRepository
}

// NewTokenService creates a new TokenService.
func NewTokenService(tokens repository.TokenRepository, users repository.UserRepository) *TokenService {
	return &TokenService{tokens: tokens, users: users}
}

// Create generates a new token set for the given user.
func (s *TokenService) Create(userID int64, ttlSeconds int) (*entity.Token, error) {
	return s.tokens.Create(userID, ttlSeconds)
}

// Authenticate validates credentials against the user repository and creates a token.
// Returns ErrNotFound if the email does not exist, or credentials do not match.
func (s *TokenService) Authenticate(email, password string, ttlSeconds int) (*entity.Token, error) {
	user, err := s.users.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Password != password {
		return nil, ErrNotFound
	}

	return s.tokens.Create(user.ID, ttlSeconds)
}
