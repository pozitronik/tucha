package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tucha/internal/application/service"
	"tucha/internal/domain/vo"
)

// FileHandler handles file metadata, CRUD operations, and file registration by hash.
type FileHandler struct {
	auth      *service.AuthService
	files     *service.FileService
	presenter *Presenter
	email     string
	userID    int64
}

// NewFileHandler creates a new FileHandler.
func NewFileHandler(
	auth *service.AuthService,
	files *service.FileService,
	presenter *Presenter,
	email string,
	userID int64,
) *FileHandler {
	return &FileHandler{
		auth:      auth,
		files:     files,
		presenter: presenter,
		email:     email,
		userID:    userID,
	}
}

// HandleFile handles GET /api/v2/file - file/folder metadata.
func (h *FileHandler) HandleFile(w http.ResponseWriter, r *http.Request) {
	token, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || token == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.files.Get(h.userID, path)
	if err != nil {
		writeHomeError(w, h.email, 500, "unknown")
		return
	}
	if node == nil {
		writeHomeError(w, h.email, 404, "not_exists")
		return
	}

	var subCount *FolderCount
	if node.IsFolder() {
		sf, sfi, _ := h.files.CountChildren(h.userID, node.Home)
		subCount = &FolderCount{Folders: sf, Files: sfi}
	}

	item := h.presenter.NodeToFolderItem(node, subCount)
	writeSuccess(w, h.email, item)
}

// HandleFileAdd handles POST /api/v2/file/add - register file by hash (dedup).
func (h *FileHandler) HandleFileAdd(w http.ResponseWriter, r *http.Request) {
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
	hashStr := r.FormValue("hash")
	sizeStr := r.FormValue("size")

	if homePath == "" || hashStr == "" || sizeStr == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		writeHomeError(w, h.email, 400, "invalid")
		return
	}

	hash, err := vo.NewContentHash(hashStr)
	if err != nil {
		writeHomeError(w, h.email, 400, "invalid")
		return
	}

	conflict, err := vo.ParseConflictMode(r.FormValue("conflict"))
	if err != nil {
		writeHomeError(w, h.email, 400, "invalid")
		return
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.files.AddByHash(h.userID, path, hash, size, conflict)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContentNotFound):
			writeEnvelope(w, h.email, 400, map[string]interface{}{
				"home": map[string]string{"error": "not_exists"},
			})
		case errors.Is(err, service.ErrOverQuota):
			writeEnvelope(w, h.email, 507, map[string]interface{}{
				"home": map[string]string{"error": "overquota"},
			})
		case errors.Is(err, service.ErrAlreadyExists):
			writeHomeError(w, h.email, 400, "exists")
		default:
			writeHomeError(w, h.email, 500, "unknown")
		}
		return
	}

	writeSuccess(w, h.email, node.Home.String())
}

// HandleFileRemove handles POST /api/v2/file/remove - delete file/folder.
func (h *FileHandler) HandleFileRemove(w http.ResponseWriter, r *http.Request) {
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
	h.files.Remove(h.userID, path)

	writeSuccess(w, h.email, path.String())
}

// HandleFileRename handles POST /api/v2/file/rename.
// Note: client sends home without leading "/" for rename; CloudPath normalizes.
func (h *FileHandler) HandleFileRename(w http.ResponseWriter, r *http.Request) {
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
	newName := r.FormValue("name")
	if homePath == "" || newName == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	// Protocol inconsistency: client sends home without leading "/" for rename.
	if !strings.HasPrefix(homePath, "/") {
		homePath = "/" + homePath
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.files.Rename(h.userID, path, newName)
	if err != nil {
		writeHomeError(w, h.email, 400, "not_exists")
		return
	}

	writeSuccess(w, h.email, node.Home.String())
}

// HandleFileMove handles POST /api/v2/file/move.
// Note: client sends both without leading "/" for move.
func (h *FileHandler) HandleFileMove(w http.ResponseWriter, r *http.Request) {
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
	folder := r.FormValue("folder")
	if homePath == "" || folder == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	// Protocol inconsistency: client sends both without leading "/" for move.
	if !strings.HasPrefix(homePath, "/") {
		homePath = "/" + homePath
	}
	if !strings.HasPrefix(folder, "/") {
		folder = "/" + folder
	}

	srcPath := vo.NewCloudPath(homePath)
	targetFolder := vo.NewCloudPath(folder)
	node, err := h.files.Move(h.userID, srcPath, targetFolder)
	if err != nil {
		writeHomeError(w, h.email, 400, "not_exists")
		return
	}

	writeSuccess(w, h.email, node.Home.String())
}

// HandleFileCopy handles POST /api/v2/file/copy.
// Note: both home and folder have explicit leading "/" from client.
func (h *FileHandler) HandleFileCopy(w http.ResponseWriter, r *http.Request) {
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
	folder := r.FormValue("folder")
	if homePath == "" || folder == "" {
		writeHomeError(w, h.email, 400, "required")
		return
	}

	srcPath := vo.NewCloudPath(homePath)
	targetFolder := vo.NewCloudPath(folder)
	node, err := h.files.Copy(h.userID, srcPath, targetFolder)
	if err != nil {
		writeHomeError(w, h.email, 400, "not_exists")
		return
	}

	writeSuccess(w, h.email, node.Home.String())
}
