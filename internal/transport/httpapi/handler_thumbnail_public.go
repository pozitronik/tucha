package httpapi

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/infrastructure/thumbnail"
)

// PublicThumbnailHandler handles public (unauthenticated) thumbnail requests for weblinks.
type PublicThumbnailHandler struct {
	publish    *service.PublishService
	thumbnails *service.ThumbnailService
}

// NewPublicThumbnailHandler creates a new PublicThumbnailHandler.
func NewPublicThumbnailHandler(
	publish *service.PublishService,
	thumbnails *service.ThumbnailService,
) *PublicThumbnailHandler {
	return &PublicThumbnailHandler{
		publish:    publish,
		thumbnails: thumbnails,
	}
}

// HandlePublicThumbnail handles GET /public/thumb/{preset}/{weblink_seg1}/{weblink_seg2}[/{subpath...}].
// Serves thumbnails for published files without authentication.
func (h *PublicThumbnailHandler) HandlePublicThumbnail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse: /public/thumb/{preset}/{seg1}/{seg2}[/{subpath...}]
	raw := strings.TrimPrefix(r.URL.Path, "/public/thumb/")
	if raw == "" || raw == "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Split into preset, weblink parts, and optional subpath
	parts := strings.SplitN(raw, "/", 4)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	preset := parts[0]
	weblinkID := parts[1] + "/" + parts[2]
	subPath := ""
	if len(parts) == 4 {
		var err error
		subPath, err = url.PathUnescape(parts[3])
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	// Resolve the weblink to a node
	node, err := h.publish.ResolveWeblink(weblinkID, subPath)
	if err != nil || node == nil || !node.IsFile() {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Check if file is a supported image format
	if !thumbnail.IsSupportedFormat(node.Name) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Generate thumbnail
	result, err := h.thumbnails.GenerateByHash(node.Hash, node.Name, preset)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Set caching headers
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	_, _ = w.Write(result.Data)
}
