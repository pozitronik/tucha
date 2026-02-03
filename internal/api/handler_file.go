package api

import (
	"net/http"

	"tucha/internal/model"
)

// handleFile handles GET /api/v2/file - file/folder metadata.
func (h *Handlers) handleFile(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	email := h.cfg.User.Email
	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		writeHomeError(w, email, 400, "required")
		return
	}

	node, err := h.nodes.Get(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}
	if node == nil {
		writeHomeError(w, email, 404, "not_exists")
		return
	}

	item := model.FolderItem{
		Name: node.Name,
		Home: node.Home,
		Type: node.Type,
		Kind: node.Type,
		Size: node.Size,
		Rev:  node.Rev,
		GRev: node.GRev,
	}

	if node.Type == "file" {
		item.Hash = node.Hash
		item.MTime = node.MTime
		item.VirusScan = "pass"
	} else {
		item.Tree = node.Tree
		folderCount, fileCount, _ := h.nodes.CountChildren(h.userID, node.Home)
		item.Count = &model.FolderCount{
			Folders: folderCount,
			Files:   fileCount,
		}
	}

	writeSuccess(w, email, item)
}
