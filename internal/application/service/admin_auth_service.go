package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// AdminAuthService handles admin panel authentication using config-based credentials.
// Tokens are stored in memory; a server restart invalidates all admin sessions.
type AdminAuthService struct {
	login    string
	password string
	mu       sync.RWMutex
	tokens   map[string]bool
}

// NewAdminAuthService creates a new AdminAuthService with the given config credentials.
func NewAdminAuthService(login, password string) *AdminAuthService {
	return &AdminAuthService{
		login:    login,
		password: password,
		tokens:   make(map[string]bool),
	}
}

// Login validates credentials against the config and returns a bearer token.
func (s *AdminAuthService) Login(login, password string) (string, error) {
	if login != s.login || password != s.password {
		return "", ErrForbidden
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	s.mu.Lock()
	s.tokens[token] = true
	s.mu.Unlock()

	return token, nil
}

// Validate checks whether the given bearer token is active.
func (s *AdminAuthService) Validate(token string) bool {
	if token == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tokens[token]
}

// Logout invalidates the given bearer token.
func (s *AdminAuthService) Logout(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}
