package httpapi

import (
	"net/http"

	"tucha/internal/application/service"
)

// SpaceHandler handles storage quota information.
type SpaceHandler struct {
	auth   *service.AuthService
	quota  *service.QuotaService
	email  string
	userID int64
}

// NewSpaceHandler creates a new SpaceHandler.
func NewSpaceHandler(auth *service.AuthService, quota *service.QuotaService, email string, userID int64) *SpaceHandler {
	return &SpaceHandler{
		auth:   auth,
		quota:  quota,
		email:  email,
		userID: userID,
	}
}

// HandleSpace handles GET /api/v2/user/space - storage quota information.
func (h *SpaceHandler) HandleSpace(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	usage, err := h.quota.GetUsage(h.userID)
	if err != nil {
		writeHomeError(w, h.email, 500, "unknown")
		return
	}

	space := SpaceInfo{
		Overquota:  usage.Overquota,
		BytesTotal: usage.BytesTotal,
		BytesUsed:  usage.BytesUsed,
	}

	writeSuccess(w, h.email, space)
}
