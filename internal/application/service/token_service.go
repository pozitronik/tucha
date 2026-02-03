package service

import (
	"tucha/internal/domain/entity"
	"tucha/internal/domain/repository"
)

// TokenService handles token creation.
type TokenService struct {
	tokens repository.TokenRepository
}

// NewTokenService creates a new TokenService.
func NewTokenService(tokens repository.TokenRepository) *TokenService {
	return &TokenService{tokens: tokens}
}

// Create generates a new token set for the given user.
func (s *TokenService) Create(userID int64, ttlSeconds int) (*entity.Token, error) {
	return s.tokens.Create(userID, ttlSeconds)
}
