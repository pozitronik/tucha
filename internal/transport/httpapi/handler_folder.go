package httpapi

import (
	"net/http"
	"strconv"

	"tucha/internal/application/service"
	"tucha/internal/domain/vo"
)

// FolderHandler handles folder listing and creation.
type FolderHandler struct {
	auth      *service.AuthService
	folders   *service.FolderService
	presenter *Presenter
	email     string
	userID    int64
}

// NewFolderHandler creates a new FolderHandler.
func NewFolderHandler(
	auth *service.AuthService,
	folders *service.FolderService,
	presenter *Presenter,
	email string,
	userID int64,
) *FolderHandler {
	return &FolderHandler{
		auth:      auth,
		folders:   folders,
		presenter: presenter,
		email:     email,
		userID:    userID,
	}
}

// HandleFolder handles GET /api/v2/folder - directory listing with pagination.
func (h *FolderHandler) HandleFolder(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
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

	folder, err := h.folders.Get(h.userID, path)
	if err != nil {
		writeHomeError(w, h.email, 500, "unknown")
		return
	}
	if folder == nil {
		writeHomeError(w, h.email, 404, "not_exists")
		return
	}

	children, err := h.folders.ListChildren(h.userID, path, offset, limit)
	if err != nil {
		writeHomeError(w, h.email, 500, "unknown")
		return
	}

	folderCount, fileCount, err := h.folders.CountChildren(h.userID, path)
	if err != nil {
		writeHomeError(w, h.email, 500, "unknown")
		return
	}

	items := make([]FolderItem, 0, len(children))
	for i := range children {
		child := &children[i]
		var subCount *FolderCount
		if child.IsFolder() {
			sf, sfi, _ := h.folders.CountChildren(h.userID, child.Home)
			subCount = &FolderCount{Folders: sf, Files: sfi}
		}
		items = append(items, h.presenter.NodeToFolderItem(child, subCount))
	}

	count := FolderCount{Folders: folderCount, Files: fileCount}
	listing := h.presenter.BuildFolderListing(folder, count, items)

	writeSuccess(w, h.email, listing)
}

// HandleFolderAdd handles POST /api/v2/folder/add - create folder.
func (h *FolderHandler) HandleFolderAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, h.email, 400, "invalid")
		return
	}

	homePath := r.FormValue("home")
	if homePath == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.folders.CreateFolder(h.userID, path)
	if err != nil {
		if err == service.ErrAlreadyExists {
			writeHomeError(w, h.email, 400, "exists")
			return
		}
		writeHomeError(w, h.email, 400, "invalid")
		return
	}

	writeSuccess(w, h.email, node.Home.String())
}
