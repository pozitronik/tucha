package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"tucha/internal/model"
)

// TokenStore handles token CRUD operations.
type TokenStore struct {
	db *sql.DB
}

// NewTokenStore creates a new TokenStore using the given database connection.
func NewTokenStore(db *DB) *TokenStore {
	return &TokenStore{db: db.Conn()}
}

// Create generates a new token set for the given user and stores it.
// Returns the created Token.
func (s *TokenStore) Create(userID int64, ttlSeconds int) (*model.Token, error) {
	accessToken, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	csrfToken, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("generating CSRF token: %w", err)
	}

	now := time.Now().Unix()
	expiresAt := now + int64(ttlSeconds)

	res, err := s.db.Exec(
		`INSERT INTO tokens (user_id, access_token, refresh_token, csrf_token, expires_at, created)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, accessToken, refreshToken, csrfToken, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting token: %w", err)
	}

	id, _ := res.LastInsertId()

	return &model.Token{
		ID:           id,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CSRFToken:    csrfToken,
		ExpiresAt:    expiresAt,
		Created:      now,
	}, nil
}

// LookupAccess finds a token by its access_token value.
// Returns nil if not found or expired.
func (s *TokenStore) LookupAccess(accessToken string) (*model.Token, error) {
	t := &model.Token{}
	err := s.db.QueryRow(
		`SELECT id, user_id, access_token, refresh_token, csrf_token, expires_at, created
		 FROM tokens WHERE access_token = ?`,
		accessToken,
	).Scan(&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.CSRFToken, &t.ExpiresAt, &t.Created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up token: %w", err)
	}

	if time.Now().Unix() > t.ExpiresAt {
		return nil, nil
	}

	return t, nil
}

// LookupCSRF returns the CSRF token associated with the given access token.
// Returns empty string if not found.
func (s *TokenStore) LookupCSRF(accessToken string) (string, error) {
	t, err := s.LookupAccess(accessToken)
	if err != nil {
		return "", err
	}
	if t == nil {
		return "", nil
	}
	return t.CSRFToken, nil
}

// randomHex generates n random bytes and returns them as a hex string.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
