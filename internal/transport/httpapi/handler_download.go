package httpapi

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"tucha/internal/application/service"
	"tucha/internal/domain/vo"
)

// DownloadHandler handles binary downloads.
type DownloadHandler struct {
	auth      *service.AuthService
	downloads *service.DownloadService
}

// NewDownloadHandler creates a new DownloadHandler.
func NewDownloadHandler(auth *service.AuthService, downloads *service.DownloadService) *DownloadHandler {
	return &DownloadHandler{
		auth:      auth,
		downloads: downloads,
	}
}

// HandleDownload handles GET /get/{path...} - download binary.
// Blocks browser-like User-Agents and serves with Range support.
func (h *DownloadHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ua := r.Header.Get("User-Agent")
	if strings.Contains(ua, "Mozilla") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("token"))
	if err != nil || authed == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawPath := strings.TrimPrefix(r.URL.Path, "/get")
	if rawPath == "" || rawPath == "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	cloudPath, err := url.PathUnescape(rawPath)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	path := vo.NewCloudPath(cloudPath)
	result, err := h.downloads.Resolve(authed.UserID, path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer result.File.Close()

	http.ServeContent(w, r, result.Node.Name, time.Unix(result.Node.MTime, 0), result.File)
}
