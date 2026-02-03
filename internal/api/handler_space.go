package api

import (
	"net/http"

	"tucha/internal/model"
)

// handleSpace handles GET /api/v2/user/space - storage quota information.
func (h *Handlers) handleSpace(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	email := h.cfg.User.Email
	bytesUsed, err := h.nodes.TotalSize(h.userID)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	space := model.SpaceInfo{
		Overquota:  bytesUsed > h.cfg.Storage.QuotaBytes,
		BytesTotal: h.cfg.Storage.QuotaBytes,
		BytesUsed:  bytesUsed,
	}

	writeSuccess(w, email, space)
}
