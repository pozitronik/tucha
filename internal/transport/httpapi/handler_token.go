package httpapi

import (
	"net/http"

	"tucha/internal/application/service"
)

const tokenTTLSeconds = 86400

// TokenHandler handles OAuth2 password grant authentication.
type TokenHandler struct {
	tokens   *service.TokenService
	email    string
	password string
	userID   int64
}

// NewTokenHandler creates a new TokenHandler.
func NewTokenHandler(tokens *service.TokenService, email, password string, userID int64) *TokenHandler {
	return &TokenHandler{
		tokens:   tokens,
		email:    email,
		password: password,
		userID:   userID,
	}
}

// HandleToken handles POST /token.
func (h *TokenHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, OAuthToken{
			Error:            "invalid_request",
			ErrorCode:        1,
			ErrorDescription: "Failed to parse form data",
		})
		return
	}

	clientID := r.FormValue("client_id")
	grantType := r.FormValue("grant_type")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if clientID != "cloud-win" {
		writeJSON(w, http.StatusOK, OAuthToken{
			Error:            "invalid_client",
			ErrorCode:        2,
			ErrorDescription: "Unknown client_id",
		})
		return
	}

	if grantType != "password" {
		writeJSON(w, http.StatusOK, OAuthToken{
			Error:            "unsupported_grant_type",
			ErrorCode:        3,
			ErrorDescription: "Only password grant is supported",
		})
		return
	}

	if username != h.email || password != h.password {
		writeJSON(w, http.StatusOK, OAuthToken{
			Error:            "invalid_grant",
			ErrorCode:        4,
			ErrorDescription: "Invalid credentials",
		})
		return
	}

	token, err := h.tokens.Create(h.userID, tokenTTLSeconds)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, OAuthToken{
			Error:            "server_error",
			ErrorCode:        5,
			ErrorDescription: "Failed to create token",
		})
		return
	}

	writeJSON(w, http.StatusOK, OAuthToken{
		ExpiresIn:        tokenTTLSeconds,
		RefreshToken:     token.RefreshToken,
		AccessToken:      token.AccessToken,
		Error:            "",
		ErrorCode:        0,
		ErrorDescription: "",
	})
}
