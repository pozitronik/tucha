package sqlite

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pozitronik/tucha/internal/domain/entity"
)

// TokenRepository implements repository.TokenRepository using SQLite.
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a TokenRepository from the given database connection.
func NewTokenRepository(db *DB) *TokenRepository {
	return &TokenRepository{db: db.Conn()}
}

// Create generates a new token set for the given user and stores it.
func (r *TokenRepository) Create(userID int64, ttlSeconds int) (*entity.Token, error) {
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

	res, err := r.db.Exec(
		`INSERT INTO tokens (user_id, access_token, refresh_token, csrf_token, expires_at, created)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, accessToken, refreshToken, csrfToken, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting token: %w", err)
	}

	id, _ := res.LastInsertId()

	return &entity.Token{
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
// Returns nil, nil if not found. Does NOT check expiration.
func (r *TokenRepository) LookupAccess(accessToken string) (*entity.Token, error) {
	t := &entity.Token{}
	err := r.db.QueryRow(
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
	return t, nil
}

// Delete removes a token by its ID.
func (r *TokenRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting token: %w", err)
	}
	return nil
}

// randomHex generates n random bytes and returns them as a hex string.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
