package httpapi

import (
	"net/http"

	"tucha/internal/application/service"
)

// SpaceHandler handles storage quota information.
type SpaceHandler struct {
	auth  *service.AuthService
	quota *service.QuotaService
}

// NewSpaceHandler creates a new SpaceHandler.
func NewSpaceHandler(auth *service.AuthService, quota *service.QuotaService) *SpaceHandler {
	return &SpaceHandler{
		auth:  auth,
		quota: quota,
	}
}

// HandleSpace handles GET /api/v2/user/space - storage quota information.
func (h *SpaceHandler) HandleSpace(w http.ResponseWriter, r *http.Request) {
	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	usage, err := h.quota.GetUsage(authed.UserID)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	space := SpaceInfo{
		Overquota:  usage.Overquota,
		BytesTotal: usage.BytesTotal,
		BytesUsed:  usage.BytesUsed,
	}

	writeSuccess(w, authed.Email, space)
}
