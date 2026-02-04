package httpapi

import (
	"log"
	"net/http"

	"tucha/internal/application/service"
)

const tokenTTLSeconds = 86400

// TokenHandler handles OAuth2 password grant authentication.
type TokenHandler struct {
	tokens *service.TokenService
}

// NewTokenHandler creates a new TokenHandler.
func NewTokenHandler(tokens *service.TokenService) *TokenHandler {
	return &TokenHandler{tokens: tokens}
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

	log.Printf("Auth attempt: email=%q password_len=%d", username, len(password))

	token, err := h.tokens.Authenticate(username, password, tokenTTLSeconds)
	if err != nil {
		log.Printf("Auth failed: email=%q err=%v", username, err)
		writeJSON(w, http.StatusOK, OAuthToken{
			Error:            "invalid_grant",
			ErrorCode:        4,
			ErrorDescription: "Invalid credentials",
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
