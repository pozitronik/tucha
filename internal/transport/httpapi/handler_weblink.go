package httpapi

import (
	"net/http"
	"strings"
	"time"

	"tucha/internal/application/service"
)

// WeblinkDownloadHandler serves public (unauthenticated) downloads for published files.
type WeblinkDownloadHandler struct {
	publish   *service.PublishService
	downloads *service.DownloadService
}

// NewWeblinkDownloadHandler creates a new WeblinkDownloadHandler.
func NewWeblinkDownloadHandler(
	publish *service.PublishService,
	downloads *service.DownloadService,
) *WeblinkDownloadHandler {
	return &WeblinkDownloadHandler{
		publish:   publish,
		downloads: downloads,
	}
}

// HandleWeblinkDownload handles GET /weblink/{seg1}/{seg2}[/{subpath...}] - public download.
// The weblink ID is composed of two path segments: "{seg1}/{seg2}".
// An optional subpath after the weblink ID resolves files within a published folder.
// No authentication is required.
func (h *WeblinkDownloadHandler) HandleWeblinkDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse: /weblink/{seg1}/{seg2}[/{subpath...}]
	raw := strings.TrimPrefix(r.URL.Path, "/weblink/")
	parts := strings.SplitN(raw, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	weblinkID := parts[0] + "/" + parts[1]
	subPath := ""
	if len(parts) == 3 {
		subPath = parts[2]
	}

	node, err := h.publish.ResolveWeblink(weblinkID, subPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	result, err := h.downloads.ResolveByNode(node)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer result.File.Close()

	http.ServeContent(w, r, result.Node.Name, time.Unix(result.Node.MTime, 0), result.File)
}
