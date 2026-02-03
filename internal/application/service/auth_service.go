// Package service provides application-level business logic.
// Services orchestrate domain entities, repositories, and ports.
package service

import (
	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
)

// AuthService validates access tokens without depending on HTTP.
type AuthService struct {
	tokens repository.TokenRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(tokens repository.TokenRepository) *AuthService {
	return &AuthService{tokens: tokens}
}

// Validate checks an access token string and returns the associated token entity.
// Returns nil, nil if the token is not found or expired.
func (s *AuthService) Validate(accessToken string) (*entity.Token, error) {
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
		return nil, nil
	}

	return token, nil
}
