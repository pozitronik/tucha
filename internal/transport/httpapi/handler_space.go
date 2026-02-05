package httpapi

import (
	"net/http"

	"github.com/pozitronik/tucha/internal/application/service"
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
	authed := authenticate(w, r, h.auth)
	if authed == nil {
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
