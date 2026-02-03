package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"tucha/internal/application/service"
)

// DispatchHandler handles shard endpoint discovery.
type DispatchHandler struct {
	auth        *service.AuthService
	externalURL string
}

// NewDispatchHandler creates a new DispatchHandler.
func NewDispatchHandler(auth *service.AuthService, externalURL string) *DispatchHandler {
	return &DispatchHandler{
		auth:        auth,
		externalURL: strings.TrimRight(externalURL, "/"),
	}
}

// HandleDispatcher handles POST /api/v2/dispatcher/.
func (h *DispatchHandler) HandleDispatcher(w http.ResponseWriter, r *http.Request) {
	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	shards := map[string]interface{}{
		"get":                []map[string]string{{"url": h.externalURL + "/get/"}},
		"upload":             []map[string]string{{"url": h.externalURL + "/upload/"}},
		"thumbnails":         []map[string]string{{"url": h.externalURL + "/thumb/"}},
		"weblink_get":        []map[string]string{{"url": h.externalURL + "/weblink/"}},
		"weblink_view":       []map[string]string{{"url": h.externalURL + "/weblink/"}},
		"weblink_video":      []map[string]string{{"url": h.externalURL + "/weblink/"}},
		"weblink_thumbnails": []map[string]string{{"url": h.externalURL + "/thumb/"}},
		"video":              []map[string]string{{"url": h.externalURL + "/video/"}},
		"view_direct":        []map[string]string{{"url": h.externalURL + "/view/"}},
		"stock":              []map[string]string{{"url": h.externalURL + "/stock/"}},
		"public_upload":      []map[string]string{{"url": h.externalURL + "/upload/"}},
		"auth":               []map[string]string{{"url": h.externalURL + "/"}},
		"web":                []map[string]string{{"url": h.externalURL + "/"}},
	}

	writeSuccess(w, authed.Email, shards)
}

// HandleOAuthDispatcher handles GET /d and GET /u (OAuth dispatcher).
func (h *DispatchHandler) HandleOAuthDispatcher(w http.ResponseWriter, r *http.Request) {
	authed, err := h.auth.Validate(r.URL.Query().Get("token"))
	if err != nil || authed == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path

	var shardURL string
	switch {
	case strings.HasSuffix(path, "/d"):
		shardURL = h.externalURL + "/get/"
	case strings.HasSuffix(path, "/u"):
		shardURL = h.externalURL + "/upload/"
	default:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s 127.0.0.1 1", shardURL)
}
