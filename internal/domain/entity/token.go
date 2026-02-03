package entity

import "time"

// Token represents an authentication token stored in the database.
type Token struct {
	ID           int64
	UserID       int64
	AccessToken  string
	RefreshToken string
	CSRFToken    string
	ExpiresAt    int64
	Created      int64
}

// IsExpired returns true if the token has passed its expiration time.
func (t *Token) IsExpired() bool {
	return time.Now().Unix() > t.ExpiresAt
}
