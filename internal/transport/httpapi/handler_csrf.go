package httpapi

import (
	"net/http"

	"tucha/internal/application/service"
)

// CSRFHandler handles CSRF token retrieval.
type CSRFHandler struct {
	auth *service.AuthService
}

// NewCSRFHandler creates a new CSRFHandler.
func NewCSRFHandler(auth *service.AuthService) *CSRFHandler {
	return &CSRFHandler{auth: auth}
}

// HandleCSRF handles GET /api/v2/tokens/csrf.
func (h *CSRFHandler) HandleCSRF(w http.ResponseWriter, r *http.Request) {
	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	writeSuccess(w, authed.Email, map[string]string{
		"token": authed.CSRFToken,
	})
}
