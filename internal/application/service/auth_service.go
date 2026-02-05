// Package service provides application-level business logic.
// Services orchestrate domain entities, repositories, and ports.
package service

import (
	"github.com/pozitronik/tucha/internal/domain/repository"
)

// AuthenticatedUser holds the resolved user context from a validated token.
type AuthenticatedUser struct {
	UserID    int64
	Email     string
	IsAdmin   bool
	CSRFToken string
}

// AuthService validates access tokens and resolves user context.
type AuthService struct {
	tokens repository.TokenRepository
	users  repository.UserRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(tokens repository.TokenRepository, users repository.UserRepository) *AuthService {
	return &AuthService{tokens: tokens, users: users}
}

// ResolveUser looks up a user by ID.
// Returns nil, nil if the user does not exist.
func (s *AuthService) ResolveUser(userID int64) (*AuthenticatedUser, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return &AuthenticatedUser{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}, nil
}

// Validate checks an access token string and returns the authenticated user context.
// Returns nil, nil if the token is not found, expired, or the user no longer exists.
func (s *AuthService) Validate(accessToken string) (*AuthenticatedUser, error) {
	if accessToken == "" {
		return nil, nil
	}

	token, err := s.tokens.LookupAccess(accessToken)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, nil
	}

	if token.IsExpired() {
		// Clean up expired token from database
		_ = s.tokens.Delete(token.ID)
		return nil, nil
	}

	user, err := s.users.GetByID(token.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &AuthenticatedUser{
		UserID:    user.ID,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CSRFToken: token.CSRFToken,
	}, nil
}
