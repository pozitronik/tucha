package httpapi

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// ThumbnailHandler handles thumbnail requests.
type ThumbnailHandler struct {
	auth       *service.AuthService
	thumbnails *service.ThumbnailService
}

// NewThumbnailHandler creates a new ThumbnailHandler.
func NewThumbnailHandler(auth *service.AuthService, thumbnails *service.ThumbnailService) *ThumbnailHandler {
	return &ThumbnailHandler{
		auth:       auth,
		thumbnails: thumbnails,
	}
}

// HandleThumbnail handles GET /thumb/{preset}/{path...}.
// URL format: /thumb/<preset>/<cloud_path>?client_id=cloud-win&token=<access_token>
func (h *ThumbnailHandler) HandleThumbnail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate
	authed, err := h.auth.Validate(r.URL.Query().Get("token"))
	if err != nil || authed == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse path: /thumb/<preset>/<cloud_path>
	rawPath := strings.TrimPrefix(r.URL.Path, "/thumb/")
	if rawPath == "" || rawPath == "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Split into preset and path
	parts := strings.SplitN(rawPath, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	preset := parts[0]
	encodedPath := "/" + parts[1]

	cloudPath, err := url.PathUnescape(encodedPath)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	path := vo.NewCloudPath(cloudPath)

	result, err := h.thumbnails.Generate(authed.UserID, path, preset)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Set caching headers (thumbnails can be cached for a while)
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	_, _ = w.Write(result.Data)
}
