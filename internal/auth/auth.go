// Package auth provides token validation helpers for API and shard authentication.
package auth

import (
	"net/http"

	"tucha/internal/model"
	"tucha/internal/storage"
)

// Authenticator validates access tokens from incoming requests.
type Authenticator struct {
	tokens *storage.TokenStore
}

// New creates a new Authenticator with the given token store.
func New(tokens *storage.TokenStore) *Authenticator {
	return &Authenticator{tokens: tokens}
}

// ValidateAPI extracts and validates the access_token query parameter
// used by API v2 endpoints. Returns the token or nil if invalid.
func (a *Authenticator) ValidateAPI(r *http.Request) (*model.Token, error) {
	accessToken := r.URL.Query().Get("access_token")
	if accessToken == "" {
		return nil, nil
	}
	return a.tokens.LookupAccess(accessToken)
}

// ValidateShard extracts and validates the token query parameter
// used by upload/download shard endpoints. Returns the token or nil if invalid.
func (a *Authenticator) ValidateShard(r *http.Request) (*model.Token, error) {
	accessToken := r.URL.Query().Get("token")
	if accessToken == "" {
		return nil, nil
	}
	return a.tokens.LookupAccess(accessToken)
}
