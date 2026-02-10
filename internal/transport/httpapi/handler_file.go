package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// mountKindShared is the kind value used for mount-root folders.
const mountKindShared = "shared"

// FileHandler handles file metadata, CRUD operations, and file registration by hash.
type FileHandler struct {
	auth      *service.AuthService
	files     *service.FileService
	trash     *service.TrashService
	shares    *service.ShareService
	presenter *Presenter
}

// NewFileHandler creates a new FileHandler.
func NewFileHandler(
	auth *service.AuthService,
	files *service.FileService,
	trash *service.TrashService,
	shares *service.ShareService,
	presenter *Presenter,
) *FileHandler {
	return &FileHandler{
		auth:      auth,
		files:     files,
		trash:     trash,
		shares:    shares,
		presenter: presenter,
	}
}

// HandleFile handles GET /api/v2/file - file/folder metadata.
func (h *FileHandler) HandleFile(w http.ResponseWriter, r *http.Request) {
	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.files.Get(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	if node == nil {
		// Path not found in user's own tree -- try mounted shares.
		h.handleMountedFile(w, authed, path)
		return
	}

	var subCount *FolderCount
	if node.IsFolder() {
		sf, sfi, _ := h.files.CountChildren(authed.UserID, node.Home)
		subCount = &FolderCount{Folders: sf, Files: sfi}
	}

	item := h.presenter.NodeToFolderItem(node, subCount)
	writeSuccess(w, authed.Email, item)
}

// handleMountedFile resolves the path through mounted shares and returns file/folder metadata.
// Returns 404 if the path does not match any mount.
func (h *FileHandler) handleMountedFile(w http.ResponseWriter, authed *service.AuthenticatedUser, path vo.CloudPath) {
	resolution, err := h.shares.ResolveMount(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}
	if resolution == nil {
		writeHomeError(w, authed.Email, 404, "not_exists")
		return
	}

	node, err := h.files.Get(resolution.Share.OwnerID, resolution.OwnerPath)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}
	if node == nil {
		writeHomeError(w, authed.Email, 404, "not_exists")
		return
	}

	var subCount *FolderCount
	if node.IsFolder() {
		sf, sfi, _ := h.files.CountChildren(resolution.Share.OwnerID, node.Home)
		subCount = &FolderCount{Folders: sf, Files: sfi}
	}

	item := h.presenter.NodeToFolderItem(node, subCount)

	// Remap home path from owner's namespace to mount namespace.
	ownerPrefix := resolution.OwnerPath.String()
	mountPrefix := path.String()
	item.Home = strings.Replace(item.Home, ownerPrefix, mountPrefix, 1)

	// At the mount root, report kind as "shared".
	if resolution.OwnerPath.String() == resolution.Share.Home.String() && node.IsFolder() {
		item.Kind = mountKindShared
	}

	writeSuccess(w, authed.Email, item)
}

// HandleFileAdd handles POST /api/v2/file/add - register file by hash (dedup).
func (h *FileHandler) HandleFileAdd(w http.ResponseWriter, r *http.Request) {
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
	hashStr := r.FormValue("hash")
	sizeStr := r.FormValue("size")

	if homePath == "" || hashStr == "" || sizeStr == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	hash, err := vo.NewContentHash(hashStr)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	conflict, err := vo.ParseConflictMode(r.FormValue("conflict"))
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	path := vo.NewCloudPath(homePath)
	targetUserID := authed.UserID
	targetPath := path

	// Check if the path falls under a mounted share.
	resolution, mErr := h.shares.ResolveMount(authed.UserID, path)
	if mErr != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}
	if resolution != nil {
		if resolution.Share.Access == vo.AccessReadOnly {
			writeHomeError(w, authed.Email, 403, "readonly")
			return
		}
		targetUserID = resolution.Share.OwnerID
		targetPath = resolution.OwnerPath
	}

	node, err := h.files.AddByHash(targetUserID, targetPath, hash, size, conflict)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrContentNotFound):
			writeEnvelope(w, authed.Email, 400, map[string]interface{}{
				"home": map[string]string{"error": "not_exists"},
			})
		case errors.Is(err, service.ErrOverQuota):
			writeEnvelope(w, authed.Email, 507, map[string]interface{}{
				"home": map[string]string{"error": "overquota"},
			})
		case errors.Is(err, service.ErrAlreadyExists):
			writeHomeError(w, authed.Email, 400, "exists")
		default:
			writeHomeError(w, authed.Email, 500, "unknown")
		}
		return
	}

	// Remap the response path to the mount namespace if needed.
	responsePath := node.Home.String()
	if resolution != nil {
		ownerPrefix := resolution.OwnerPath.Parent().String()
		mountPrefix := path.Parent().String()
		responsePath = strings.Replace(responsePath, ownerPrefix, mountPrefix, 1)
	}

	writeSuccess(w, authed.Email, responsePath)
}

// HandleFileRemove handles POST /api/v2/file/remove - delete file/folder.
func (h *FileHandler) HandleFileRemove(w http.ResponseWriter, r *http.Request) {
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
	_ = h.trash.Trash(authed.UserID, path, authed.UserID)

	writeSuccess(w, authed.Email, path.String())
}

// HandleFileRename handles POST /api/v2/file/rename.
// Note: client sends home without leading "/" for rename; CloudPath normalizes.
func (h *FileHandler) HandleFileRename(w http.ResponseWriter, r *http.Request) {
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
	newName := r.FormValue("name")
	if homePath == "" || newName == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	// Protocol inconsistency: client sends home without leading "/" for rename.
	if !strings.HasPrefix(homePath, "/") {
		homePath = "/" + homePath
	}

	path := vo.NewCloudPath(homePath)
	node, err := h.files.Rename(authed.UserID, path, newName)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "not_exists")
		return
	}

	writeSuccess(w, authed.Email, node.Home.String())
}

// HandleFileMove handles POST /api/v2/file/move.
// Note: client sends both without leading "/" for move.
func (h *FileHandler) HandleFileMove(w http.ResponseWriter, r *http.Request) {
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
	folder := r.FormValue("folder")
	if homePath == "" || folder == "" {
		writeHomeError(w, authed.Email, 400, "required")
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
	node, err := h.files.Move(authed.UserID, srcPath, targetFolder)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "not_exists")
		return
	}

	writeSuccess(w, authed.Email, node.Home.String())
}

// HandleFileHistory handles GET /api/v2/file/history - file version history.
func (h *FileHandler) HandleFileHistory(w http.ResponseWriter, r *http.Request) {
	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)

	// Verify the file exists before returning history.
	node, err := h.files.Get(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}
	if node == nil {
		writeHomeError(w, authed.Email, 404, "not_exists")
		return
	}

	versions, err := h.files.History(authed.UserID, path, authed.VersionHistory)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	items := make([]FileVersionItem, 0, len(versions))
	for _, v := range versions {
		item := FileVersionItem{
			Name: v.Name,
			Home: v.Home.String(),
			Size: v.Size,
			Time: v.Time,
		}
		if authed.VersionHistory {
			item.Hash = v.Hash.String()
			item.Rev = v.Rev
		}
		items = append(items, item)
	}

	writeSuccess(w, authed.Email, items)
}

// HandleFileCopy handles POST /api/v2/file/copy.
// Note: both home and folder have explicit leading "/" from client.
func (h *FileHandler) HandleFileCopy(w http.ResponseWriter, r *http.Request) {
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
	folder := r.FormValue("folder")
	if homePath == "" || folder == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	srcPath := vo.NewCloudPath(homePath)
	targetFolder := vo.NewCloudPath(folder)
	node, err := h.files.Copy(authed.UserID, srcPath, targetFolder)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "not_exists")
		return
	}

	writeSuccess(w, authed.Email, node.Home.String())
}
