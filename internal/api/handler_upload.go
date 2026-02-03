package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"tucha/internal/hash"
)

// handleUpload handles PUT /upload/ - upload binary, return hash.
//
// The client sends the raw file content as the PUT body.
// The response is the 40-character uppercase hex hash as plain text.
func (h *Handlers) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateShard(r)
	if err != nil || token == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read the entire body to compute hash.
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Compute mrCloud hash.
	fileHash := hash.Compute(data)

	// Store the content on disk.
	size := int64(len(data))
	if _, err := h.store.Write(fileHash, bytes.NewReader(data)); err != nil {
		http.Error(w, "Failed to store content", http.StatusInternalServerError)
		return
	}

	// Register in content database.
	if _, err := h.contents.Insert(fileHash, size); err != nil {
		http.Error(w, "Failed to register content", http.StatusInternalServerError)
		return
	}

	// Return the hash as plain text.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, fileHash)
}
