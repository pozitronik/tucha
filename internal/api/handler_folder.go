package api

import (
	"net/http"
	"strconv"

	"tucha/internal/model"
)

// handleFolder handles GET /api/v2/folder - directory listing with pagination.
func (h *Handlers) handleFolder(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.ValidateAPI(r)
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	email := h.cfg.User.Email
	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		homePath = "/"
	}

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 65535
	}

	// Get the folder node itself.
	folder, err := h.nodes.Get(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}
	if folder == nil {
		writeHomeError(w, email, 404, "not_exists")
		return
	}

	// Get children.
	children, err := h.nodes.ListChildren(h.userID, homePath, offset, limit)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	// Count children.
	folderCount, fileCount, err := h.nodes.CountChildren(h.userID, homePath)
	if err != nil {
		writeHomeError(w, email, 500, "unknown")
		return
	}

	// Build list items.
	list := make([]model.FolderItem, 0, len(children))
	for _, child := range children {
		item := model.FolderItem{
			Name: child.Name,
			Home: child.Home,
			Type: child.Type,
			Kind: child.Type,
			Size: child.Size,
			Rev:  child.Rev,
			GRev: child.GRev,
		}
		if child.Type == "file" {
			item.Hash = child.Hash
			item.MTime = child.MTime
			item.VirusScan = "pass"
		} else {
			item.Tree = child.Tree
			// Count sub-children for each subfolder.
			subFolders, subFiles, _ := h.nodes.CountChildren(h.userID, child.Home)
			item.Count = &model.FolderCount{
				Folders: subFolders,
				Files:   subFiles,
			}
		}
		list = append(list, item)
	}

	listing := model.FolderListing{
		Count: model.FolderCount{
			Folders: folderCount,
			Files:   fileCount,
		},
		Tree: folder.Tree,
		Name: folder.Name,
		GRev: folder.GRev,
		Size: folder.Size,
		Sort: model.SortInfo{Order: "asc", Type: "name"},
		Kind: "folder",
		Rev:  folder.Rev,
		Type: "folder",
		Home: folder.Home,
		List: list,
	}

	writeSuccess(w, email, listing)
}
