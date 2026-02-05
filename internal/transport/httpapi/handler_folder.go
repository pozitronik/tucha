package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// FolderHandler handles folder listing and creation.
type FolderHandler struct {
	auth      *service.AuthService
	folders   *service.FolderService
	publish   *service.PublishService
	presenter *Presenter
}

// NewFolderHandler creates a new FolderHandler.
func NewFolderHandler(
	auth *service.AuthService,
	folders *service.FolderService,
	publish *service.PublishService,
	presenter *Presenter,
) *FolderHandler {
	return &FolderHandler{
		auth:      auth,
		folders:   folders,
		publish:   publish,
		presenter: presenter,
	}
}

// HandleFolder handles GET /api/v2/folder - directory listing with pagination.
// Supports two modes:
//   - Authenticated: home= parameter with access_token
//   - Public: weblink= parameter, no auth required
func (h *FolderHandler) HandleFolder(w http.ResponseWriter, r *http.Request) {
	weblinkParam := r.URL.Query().Get("weblink")
	if weblinkParam != "" {
		h.handlePublicFolder(w, r, weblinkParam)
		return
	}

	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		homePath = "/"
	}
	path := vo.NewCloudPath(homePath)

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 65535
	}

	folder, err := h.folders.Get(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}
	if folder == nil {
		writeHomeError(w, authed.Email, 404, "not_exists")
		return
	}

	children, err := h.folders.ListChildren(authed.UserID, path, offset, limit)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	folderCount, fileCount, err := h.folders.CountChildren(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	items := make([]FolderItem, 0, len(children))
	for i := range children {
		child := &children[i]
		var subCount *FolderCount
		if child.IsFolder() {
			sf, sfi, _ := h.folders.CountChildren(authed.UserID, child.Home)
			subCount = &FolderCount{Folders: sf, Files: sfi}
		}
		items = append(items, h.presenter.NodeToFolderItem(child, subCount))
	}

	count := FolderCount{Folders: folderCount, Files: fileCount}
	listing := h.presenter.BuildFolderListing(folder, count, items)

	writeSuccess(w, authed.Email, listing)
}

// handlePublicFolder serves a folder listing resolved via weblink (no auth).
// Paths in the response are relative to the published folder root.
func (h *FolderHandler) handlePublicFolder(w http.ResponseWriter, r *http.Request, weblinkParam string) {
	// Clean the weblink value: strip leading/trailing slashes.
	weblinkID := strings.Trim(weblinkParam, "/")
	if weblinkID == "" {
		writeHomeError(w, "", 404, "not_exists")
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 65535
	}

	folder, err := h.publish.ResolveWeblink(weblinkID, "")
	if err != nil {
		writeHomeError(w, "", 404, "not_exists")
		return
	}

	children, err := h.folders.ListChildren(folder.UserID, folder.Home, offset, limit)
	if err != nil {
		writeHomeError(w, "", 500, "unknown")
		return
	}

	folderCount, fileCount, err := h.folders.CountChildren(folder.UserID, folder.Home)
	if err != nil {
		writeHomeError(w, "", 500, "unknown")
		return
	}

	// Build items with paths relative to the published folder root.
	prefix := folder.Home.String()
	items := make([]FolderItem, 0, len(children))
	for i := range children {
		child := &children[i]
		var subCount *FolderCount
		if child.IsFolder() {
			sf, sfi, _ := h.folders.CountChildren(folder.UserID, child.Home)
			subCount = &FolderCount{Folders: sf, Files: sfi}
		}
		item := h.presenter.NodeToFolderItem(child, subCount)
		item.Home = strings.TrimPrefix(child.Home.String(), prefix)
		if item.Home == "" {
			item.Home = "/"
		}
		items = append(items, item)
	}

	count := FolderCount{Folders: folderCount, Files: fileCount}
	listing := h.presenter.BuildFolderListing(folder, count, items)
	listing.Home = "/"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": 200,
		"body":   listing,
	})
}

// HandleFolderAdd handles POST /api/v2/folder/add - create folder.
func (h *FolderHandler) HandleFolderAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	homePath := r.FormValue("home")
	if homePath == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.folders.CreateFolder(authed.UserID, path)
	if err != nil {
		if err == service.ErrAlreadyExists {
			writeHomeError(w, authed.Email, 400, "exists")
			return
		}
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	writeSuccess(w, authed.Email, node.Home.String())
}
