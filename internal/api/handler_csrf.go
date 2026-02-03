package api

import (
	"net/http"
)

// handleCSRF handles GET /api/v2/tokens/csrf.
// Returns the CSRF token associated with the caller's access token.
func (h *Handlers) handleCSRF(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateAPI(r)
	if err != nil {
		writeEnvelope(w, "", 403, "user")
		return
	}
	if token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	email := h.cfg.User.Email
	writeSuccess(w, email, map[string]string{
		"token": token.CSRFToken,
	})
}
