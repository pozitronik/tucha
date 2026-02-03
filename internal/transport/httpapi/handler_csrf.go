package httpapi

import (
	"net/http"

	"tucha/internal/application/service"
)

// CSRFHandler handles CSRF token retrieval.
type CSRFHandler struct {
	auth  *service.AuthService
	email string
}

// NewCSRFHandler creates a new CSRFHandler.
func NewCSRFHandler(auth *service.AuthService, email string) *CSRFHandler {
	return &CSRFHandler{auth: auth, email: email}
}

// HandleCSRF handles GET /api/v2/tokens/csrf.
func (h *CSRFHandler) HandleCSRF(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	writeSuccess(w, h.email, map[string]string{
		"token": token.CSRFToken,
	})
}
