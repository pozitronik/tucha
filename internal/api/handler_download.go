package api

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// handleDownload handles GET /get/{path...} - download binary.
//
// The download endpoint blocks browser-like User-Agents (e.g., Mozilla/*).
// The client sets UA to "cloud-win".
// Serves via http.ServeContent for Range support.
func (h *Handlers) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Block browser-like User-Agents.
	ua := r.Header.Get("User-Agent")
	if strings.Contains(ua, "Mozilla") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	token, err := h.auth.ValidateShard(r)
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract the cloud path from the URL.
	// URL format: /get/<url_encoded_path>
	rawPath := strings.TrimPrefix(r.URL.Path, "/get")
	if rawPath == "" || rawPath == "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// URL-decode the path.
	cloudPath, err := url.PathUnescape(rawPath)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Look up the file node.
	node, err := h.nodes.Get(h.userID, cloudPath)
	if err != nil || node == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if node.Type != "file" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Open the content file.
	f, err := h.store.Open(node.Hash)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Serve with Range support.
	http.ServeContent(w, r, node.Name, time.Unix(node.MTime, 0), f)
}
