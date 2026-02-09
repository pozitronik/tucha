package httpapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pozitronik/tucha/internal/application/service"
)

// UploadHandler handles binary uploads.
type UploadHandler struct {
	auth    *service.AuthService
	uploads *service.UploadService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(auth *service.AuthService, uploads *service.UploadService) *UploadHandler {
	return &UploadHandler{auth: auth, uploads: uploads}
}

// HandleUpload handles PUT /upload/ - upload binary, return hash.
func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("token"))
	if err != nil || authed == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fast reject via Content-Length header if file size limit is set.
	if authed.FileSizeLimit > 0 && r.ContentLength > 0 && r.ContentLength > authed.FileSizeLimit {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Check actual body size against limit.
	if authed.FileSizeLimit > 0 && int64(len(data)) > authed.FileSizeLimit {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	hash, err := h.uploads.Upload(data)
	if err != nil {
		http.Error(w, "Failed to store content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, hash.String())
}
