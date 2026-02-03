package api

import (
	"net/http"

	"tucha/internal/model"
)

const tokenTTLSeconds = 86400

// handleToken handles POST /token (OAuth2 password grant).
func (h *Handlers) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, model.OAuthToken{
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
		writeJSON(w, http.StatusOK, model.OAuthToken{
			Error:            "invalid_client",
			ErrorCode:        2,
			ErrorDescription: "Unknown client_id",
		})
		return
	}

	if grantType != "password" {
		writeJSON(w, http.StatusOK, model.OAuthToken{
			Error:            "unsupported_grant_type",
			ErrorCode:        3,
			ErrorDescription: "Only password grant is supported",
		})
		return
	}

	if username != h.cfg.User.Email || password != h.cfg.User.Password {
		writeJSON(w, http.StatusOK, model.OAuthToken{
			Error:            "invalid_grant",
			ErrorCode:        4,
			ErrorDescription: "Invalid credentials",
		})
		return
	}

	token, err := h.tokens.Create(h.userID, tokenTTLSeconds)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.OAuthToken{
			Error:            "server_error",
			ErrorCode:        5,
			ErrorDescription: "Failed to create token",
		})
		return
	}

	writeJSON(w, http.StatusOK, model.OAuthToken{
		ExpiresIn:        tokenTTLSeconds,
		RefreshToken:     token.RefreshToken,
		AccessToken:      token.AccessToken,
		Error:            "",
		ErrorCode:        0,
		ErrorDescription: "",
	})
}
