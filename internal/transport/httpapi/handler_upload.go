package httpapi

import (
	"fmt"
	"io"
	"net/http"

	"tucha/internal/application/service"
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

	token, err := h.auth.Validate(r.URL.Query().Get("token"))
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hash, err := h.uploads.Upload(data)
	if err != nil {
		http.Error(w, "Failed to store content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, hash.String())
}
