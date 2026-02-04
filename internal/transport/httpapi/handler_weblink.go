package httpapi

import (
	"net/http"
	"strings"
	"time"

	"tucha/internal/application/service"
	"tucha/internal/domain/entity"
)

// WeblinkDownloadHandler serves public (unauthenticated) access to published nodes.
// Files are served as binary downloads; folders return a JSON listing.
type WeblinkDownloadHandler struct {
	publish   *service.PublishService
	downloads *service.DownloadService
	folders   *service.FolderService
	presenter *Presenter
}

// NewWeblinkDownloadHandler creates a new WeblinkDownloadHandler.
func NewWeblinkDownloadHandler(
	publish *service.PublishService,
	downloads *service.DownloadService,
	folders *service.FolderService,
	presenter *Presenter,
) *WeblinkDownloadHandler {
	return &WeblinkDownloadHandler{
		publish:   publish,
		downloads: downloads,
		folders:   folders,
		presenter: presenter,
	}
}

// HandleWeblinkDownload handles GET /public/{seg1}/{seg2}[/{subpath...}] - public access.
// The weblink ID is composed of two path segments: "{seg1}/{seg2}".
// An optional subpath after the weblink ID resolves items within a published folder.
// Files are served as binary; folders return a JSON listing with relative paths.
// No authentication is required.
func (h *WeblinkDownloadHandler) HandleWeblinkDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse: /public/{seg1}/{seg2}[/{subpath...}]
	raw := strings.TrimPrefix(r.URL.Path, "/public/")
	parts := strings.SplitN(raw, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	weblinkID := parts[0] + "/" + parts[1]
	subPath := ""
	if len(parts) == 3 {
		subPath = parts[2]
	}

	node, err := h.publish.ResolveWeblink(weblinkID, subPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if node.IsFile() {
		h.serveFile(w, r, node)
		return
	}

	h.serveFolderListing(w, node)
}

// serveFile streams the file content as a binary download.
func (h *WeblinkDownloadHandler) serveFile(w http.ResponseWriter, r *http.Request, node *entity.Node) {
	result, err := h.downloads.ResolveByNode(node)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer result.File.Close()

	http.ServeContent(w, r, result.Node.Name, time.Unix(result.Node.MTime, 0), result.File)
}

// serveFolderListing returns a JSON listing of the folder's children with relative paths.
func (h *WeblinkDownloadHandler) serveFolderListing(w http.ResponseWriter, node *entity.Node) {
	children, err := h.folders.ListChildren(node.UserID, node.Home, 0, 65535)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	folderCount, fileCount, err := h.folders.CountChildren(node.UserID, node.Home)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Build items with paths relative to the published folder root.
	prefix := node.Home.String()
	items := make([]FolderItem, 0, len(children))
	for i := range children {
		child := &children[i]
		var subCount *FolderCount
		if child.IsFolder() {
			sf, sfi, _ := h.folders.CountChildren(node.UserID, child.Home)
			subCount = &FolderCount{Folders: sf, Files: sfi}
		}
		item := h.presenter.NodeToFolderItem(child, subCount)
		// Strip owner's path prefix to produce relative home.
		item.Home = strings.TrimPrefix(child.Home.String(), prefix)
		if item.Home == "" {
			item.Home = "/"
		}
		items = append(items, item)
	}

	count := FolderCount{Folders: folderCount, Files: fileCount}
	listing := h.presenter.BuildFolderListing(node, count, items)
	// The listing's own home should also be relative (root of published tree).
	listing.Home = "/"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": 200,
		"body":   listing,
	})
}
