package httpapi

import (
	"errors"
	"net/http"

	"tucha/internal/application/service"
	"tucha/internal/domain/vo"
)

// PublishHandler handles publishing/unpublishing nodes, listing shared links, and cloning.
type PublishHandler struct {
	auth      *service.AuthService
	publish   *service.PublishService
	presenter *Presenter
}

// NewPublishHandler creates a new PublishHandler.
func NewPublishHandler(
	auth *service.AuthService,
	publish *service.PublishService,
	presenter *Presenter,
) *PublishHandler {
	return &PublishHandler{
		auth:      auth,
		publish:   publish,
		presenter: presenter,
	}
}

// HandlePublish handles POST /api/v2/file/publish - assign a public weblink.
func (h *PublishHandler) HandlePublish(w http.ResponseWriter, r *http.Request) {
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
	weblink, err := h.publish.Publish(authed.UserID, path)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeHomeError(w, authed.Email, 404, "not_exists")
			return
		}
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, weblink)
}

// HandleUnpublish handles POST /api/v2/file/unpublish - remove a public weblink.
func (h *PublishHandler) HandleUnpublish(w http.ResponseWriter, r *http.Request) {
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

	weblinkID := r.FormValue("weblink")
	if weblinkID == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	if err := h.publish.Unpublish(authed.UserID, weblinkID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeHomeError(w, authed.Email, 404, "not_exists")
		case errors.Is(err, service.ErrForbidden):
			writeHomeError(w, authed.Email, 403, "forbidden")
		default:
			writeHomeError(w, authed.Email, 500, "unknown")
		}
		return
	}

	writeSuccess(w, authed.Email, "ok")
}

// HandleSharedLinks handles GET /api/v2/folder/shared/links - list published items.
func (h *PublishHandler) HandleSharedLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	nodes, err := h.publish.ListPublished(authed.UserID)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	items := make([]FolderItem, 0, len(nodes))
	for i := range nodes {
		items = append(items, h.presenter.NodeToFolderItem(&nodes[i], nil))
	}

	writeSuccess(w, authed.Email, map[string]interface{}{"list": items})
}

// HandleClone handles POST /api/v2/clone - clone a published item into the caller's tree.
func (h *PublishHandler) HandleClone(w http.ResponseWriter, r *http.Request) {
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

	weblinkID := r.FormValue("weblink")
	folder := r.FormValue("folder")
	if weblinkID == "" || folder == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	conflict, err := vo.ParseConflictMode(r.FormValue("conflict"))
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	targetFolder := vo.NewCloudPath(folder)
	node, err := h.publish.Clone(authed.UserID, weblinkID, targetFolder, conflict)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeHomeError(w, authed.Email, 404, "not_exists")
		case errors.Is(err, service.ErrForbidden):
			writeHomeError(w, authed.Email, 403, "forbidden")
		case errors.Is(err, service.ErrAlreadyExists):
			writeHomeError(w, authed.Email, 400, "exists")
		default:
			writeHomeError(w, authed.Email, 500, "unknown")
		}
		return
	}

	writeSuccess(w, authed.Email, node.Home.String())
}
