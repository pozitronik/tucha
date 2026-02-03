package api

import (
	"net/http"
	"strconv"
)

// handleFileAdd handles POST /api/v2/file/add - register file by hash (dedup).
//
// Protocol:
//   - status 200: hash exists in storage, file created (deduplication success)
//   - status 400: hash not found in storage (upload required)
func (h *Handlers) handleFileAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	hash := r.FormValue("hash")
	sizeStr := r.FormValue("size")

	if homePath == "" || hash == "" || sizeStr == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		writeHomeError(w, email, 400, "invalid")
		return
	}

	// Check if content exists in storage (deduplication check).
	contentExists, err := h.contents.Exists(hash)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	if !contentExists {
		// Also check the on-disk content store.
		if !h.store.Exists(hash) {
			// Hash not found - client needs to upload first.
			writeEnvelope(w, email, 400, map[string]interface{}{
				"home": map[string]string{
					"error": "not_exists",
				},
			})
			return
		}
	}

	// Content exists; create the file node.
	// Increment content reference count.
	if _, err := h.contents.Insert(hash, size); err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	// Check quota.
	used, err := h.nodes.TotalSize(h.userID)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}
	if used+size > h.cfg.Storage.QuotaBytes {
		writeEnvelope(w, email, 507, map[string]interface{}{
			"home": map[string]string{
				"error": "overquota",
			},
		})
		return
	}

	// Handle conflict mode.
	conflict := r.FormValue("conflict")
	exists, err := h.nodes.Exists(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}
	if exists {
		if conflict == "strict" {
			writeHomeError(w, email, 400, "exists")
			return
		}
		// For "rename" or default: delete old node first, then create new.
		h.nodes.Delete(h.userID, homePath)
	}

	node, err := h.nodes.CreateFile(h.userID, homePath, hash, size)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	writeSuccess(w, email, node.Home)
}
