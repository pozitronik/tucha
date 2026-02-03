package api

import (
	"fmt"
	"net/http"
	"strings"
)

// handleDispatcher handles POST /api/v2/dispatcher/.
// Returns shard URLs; in our single-server setup, all shards point to self.
func (h *Handlers) handleDispatcher(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	baseURL := strings.TrimRight(h.cfg.Server.ExternalURL, "/")
	email := h.cfg.User.Email

	// Build shard map. All shards point to ourselves.
	shards := map[string]interface{}{
		"get":                []map[string]string{{"url": baseURL + "/get/"}},
		"upload":             []map[string]string{{"url": baseURL + "/upload/"}},
		"thumbnails":         []map[string]string{{"url": baseURL + "/thumb/"}},
		"weblink_get":        []map[string]string{{"url": baseURL + "/weblink/"}},
		"weblink_view":       []map[string]string{{"url": baseURL + "/weblink/"}},
		"weblink_video":      []map[string]string{{"url": baseURL + "/weblink/"}},
		"weblink_thumbnails": []map[string]string{{"url": baseURL + "/thumb/"}},
		"video":              []map[string]string{{"url": baseURL + "/video/"}},
		"view_direct":        []map[string]string{{"url": baseURL + "/view/"}},
		"stock":              []map[string]string{{"url": baseURL + "/stock/"}},
		"public_upload":      []map[string]string{{"url": baseURL + "/upload/"}},
		"auth":               []map[string]string{{"url": baseURL + "/"}},
		"web":                []map[string]string{{"url": baseURL + "/"}},
	}

	writeSuccess(w, email, shards)
}

// handleOAuthDispatcher handles GET /d and GET /u (OAuth dispatcher).
// Returns plain text: "URL IP COUNT" (space-separated).
func (h *Handlers) handleOAuthDispatcher(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateShard(r)
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	baseURL := strings.TrimRight(h.cfg.Server.ExternalURL, "/")
	path := r.URL.Path

	var shardURL string
	switch {
	case strings.HasSuffix(path, "/d"):
		shardURL = baseURL + "/get/"
	case strings.HasSuffix(path, "/u"):
		shardURL = baseURL + "/upload/"
	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s 127.0.0.1 1", shardURL)
}
