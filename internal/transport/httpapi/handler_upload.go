package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// UploadHandler handles binary uploads.
type UploadHandler struct {
	auth    *service.AuthService
	uploads *service.UploadService
	files   *service.FileService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(auth *service.AuthService, uploads *service.UploadService, files *service.FileService) *UploadHandler {
	return &UploadHandler{auth: auth, uploads: uploads, files: files}
}

// HandleUpload handles PUT /upload/ - upload binary, return hash.
// The real API encodes the target path in the URL: PUT /upload/home=/path/to/file.txt?token=...
// When a home parameter is present, the handler also registers the file node and records
// a version history entry (matching real API behavior where upload = store + register).
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

	// If the URL path contains a home parameter, also register the file node.
	// This matches the real API where PUT /upload/home=/path stores AND registers.
	if homePath := parseUploadHome(r.URL.Path); homePath != "" {
		path := vo.NewCloudPath(homePath)
		// Use ConflictReplace so re-uploads overwrite the existing node and append a version entry.
		_, _ = h.files.AddByHash(authed.UserID, path, hash, int64(len(data)), vo.ConflictReplace)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, hash.String())
}

// parseUploadHome extracts the home path from an upload URL path.
// The real API uses: /upload/home=/path/to/file.txt
// Returns empty string if no home parameter is found.
func parseUploadHome(urlPath string) string {
	// Strip the /upload/ prefix.
	rest := strings.TrimPrefix(urlPath, "/upload/")
	if rest == "" || rest == urlPath {
		return ""
	}

	// The remainder may contain key=value pairs separated by &
	// (e.g., "home=/file.txt&cloud_domain=2"). Parse the home value.
	for _, part := range strings.Split(rest, "&") {
		if strings.HasPrefix(part, "home=") {
			value := strings.TrimPrefix(part, "home=")
			decoded, err := url.PathUnescape(value)
			if err != nil {
				return value
			}
			return decoded
		}
	}

	return ""
}
