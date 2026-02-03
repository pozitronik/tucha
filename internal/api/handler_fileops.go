package api

import (
	"net/http"
	"strings"
)

// handleFolderAdd handles POST /api/v2/folder/add - create folder.
func (h *Handlers) handleFolderAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	if homePath == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	// Check if already exists.
	exists, err := h.nodes.Exists(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}
	if exists {
		writeHomeError(w, email, 400, "exists")
		return
	}

	node, err := h.nodes.CreateFolder(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 400, "invalid")
		return
	}

	writeSuccess(w, email, node.Home)
}

// handleFileRemove handles POST /api/v2/file/remove - delete file/folder.
// Per protocol, this always returns success even if the path does not exist.
func (h *Handlers) handleFileRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	if homePath == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	// Strip trailing slash for folder paths (normalize).
	cleanPath := strings.TrimRight(homePath, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}

	// Look up the node to handle content cleanup for files.
	node, _ := h.nodes.Get(h.userID, cleanPath)
	if node != nil && node.Type == "file" && node.Hash != "" {
		// Decrement content reference; delete content file if unreferenced.
		deleted, _ := h.contents.Decrement(node.Hash)
		if deleted {
			h.store.Delete(node.Hash)
		}
	}

	// Delete the node (always returns success per protocol).
	h.nodes.Delete(h.userID, cleanPath)

	writeSuccess(w, email, cleanPath)
}

// handleFileRename handles POST /api/v2/file/rename.
// Note: home does NOT have a leading "/" prepended (unlike create/delete).
func (h *Handlers) handleFileRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	newName := r.FormValue("name")

	if homePath == "" || newName == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	// The client sends home without leading "/" for rename; normalize.
	if !strings.HasPrefix(homePath, "/") {
		homePath = "/" + homePath
	}

	node, err := h.nodes.Rename(h.userID, homePath, newName)
	if err != nil {
		writeHomeError(w, email, 400, "not_exists")
		return
	}

	writeSuccess(w, email, node.Home)
}

// handleFileMove handles POST /api/v2/file/move.
// Note: No leading "/" prefix on either parameter.
func (h *Handlers) handleFileMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	folder := r.FormValue("folder")

	if homePath == "" || folder == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	// The client sends both without leading "/" for move; normalize.
	if !strings.HasPrefix(homePath, "/") {
		homePath = "/" + homePath
	}
	if !strings.HasPrefix(folder, "/") {
		folder = "/" + folder
	}

	node, err := h.nodes.Move(h.userID, homePath, folder)
	if err != nil {
		writeHomeError(w, email, 400, "not_exists")
		return
	}

	writeSuccess(w, email, node.Home)
}

// handleFileCopy handles POST /api/v2/file/copy.
// Note: Both home and folder have explicit leading "/" prefix.
func (h *Handlers) handleFileCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.cfg.User.Email, 400, "invalid")
		return
	}

	email := h.cfg.User.Email
	homePath := r.FormValue("home")
	folder := r.FormValue("folder")

	if homePath == "" || folder == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	node, err := h.nodes.Copy(h.userID, homePath, folder)
	if err != nil {
		writeHomeError(w, email, 400, "not_exists")
		return
	}

	writeSuccess(w, email, node.Home)
}
