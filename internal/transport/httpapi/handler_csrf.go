package httpapi

import (
	"net/http"

	"github.com/pozitronik/tucha/internal/application/service"
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
	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	writeSuccess(w, authed.Email, map[string]string{
		"token": authed.CSRFToken,
	})
}
