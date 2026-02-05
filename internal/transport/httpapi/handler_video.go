package httpapi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
)

// VideoHandler handles video streaming for published files.
type VideoHandler struct {
	publish     *service.PublishService
	downloads   *service.DownloadService
	externalURL string
}

// NewVideoHandler creates a new VideoHandler.
func NewVideoHandler(
	publish *service.PublishService,
	downloads *service.DownloadService,
	externalURL string,
) *VideoHandler {
	return &VideoHandler{
		publish:     publish,
		downloads:   downloads,
		externalURL: strings.TrimRight(externalURL, "/"),
	}
}

// HandleHLSPlaylist handles GET /video/0p/{base64_weblink}.m3u8 - HLS playlist generation.
// The weblink is Base64-encoded in the URL.
func (h *VideoHandler) HandleHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse: /video/0p/{base64_weblink}.m3u8
	raw := strings.TrimPrefix(r.URL.Path, "/video/")
	if !strings.HasPrefix(raw, "0p/") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	raw = strings.TrimPrefix(raw, "0p/")
	if !strings.HasSuffix(raw, ".m3u8") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Extract and decode the Base64 weblink
	b64Weblink := strings.TrimSuffix(raw, ".m3u8")
	weblinkBytes, err := base64.StdEncoding.DecodeString(b64Weblink)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	weblinkID := string(weblinkBytes)

	// Resolve the weblink to a node
	node, err := h.publish.ResolveWeblink(weblinkID, "")
	if err != nil || node == nil || !node.IsFile() {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Check if file is a video format
	if !isVideoFormat(node.Name) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Generate HLS playlist pointing to the video file
	// This is a simple single-segment playlist for direct streaming
	videoURL := fmt.Sprintf("%s/public/%s", h.externalURL, weblinkID)

	// Calculate approximate duration (we don't know the actual duration without parsing the video)
	// Using a large value allows the player to seek properly
	duration := 36000 // 10 hours max

	playlist := fmt.Sprintf(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:%d
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:%d,
%s
#EXT-X-ENDLIST
`, duration, duration, videoURL)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(playlist))
}

// HandleVideoStream handles GET /video/* - routes to HLS playlist or direct video access.
func (h *VideoHandler) HandleVideoStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw := strings.TrimPrefix(r.URL.Path, "/video/")

	// Route HLS playlist requests to the appropriate handler
	if strings.HasPrefix(raw, "0p/") {
		h.HandleHLSPlaylist(w, r)
		return
	}

	// Parse: /video/{seg1}/{seg2}[/{subpath...}]
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
	if err != nil || node == nil || !node.IsFile() {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Serve the video file
	result, err := h.downloads.ResolveByNode(node)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer result.File.Close()

	http.ServeContent(w, r, result.Node.Name, time.Unix(result.Node.MTime, 0), result.File)
}

// isVideoFormat checks if the filename indicates a video format.
func isVideoFormat(filename string) bool {
	lower := strings.ToLower(filename)
	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg"}
	for _, ext := range videoExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
