package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// TrashHandler handles trashbin operations: listing, restoring, and emptying.
type TrashHandler struct {
	auth      *service.AuthService
	trash     *service.TrashService
	presenter *Presenter
}

// NewTrashHandler creates a new TrashHandler.
func NewTrashHandler(
	auth *service.AuthService,
	trash *service.TrashService,
	presenter *Presenter,
) *TrashHandler {
	return &TrashHandler{
		auth:      auth,
		trash:     trash,
		presenter: presenter,
	}
}

// HandleTrashList handles GET /api/v2/trashbin - list trashed items.
func (h *TrashHandler) HandleTrashList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	items, err := h.trash.List(authed.UserID)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	dtoItems := make([]TrashFolderItem, 0, len(items))
	for i := range items {
		dtoItems = append(dtoItems, h.presenter.TrashItemToDTO(&items[i]))
	}

	writeSuccess(w, authed.Email, map[string]interface{}{"list": dtoItems})
}

// HandleTrashRestore handles POST /api/v2/trashbin/restore - restore a trashed item.
// Body: path=<url_encoded_original_path>&restore_revision=<rev>&conflict=<mode>
func (h *TrashHandler) HandleTrashRestore(w http.ResponseWriter, r *http.Request) {
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

	pathStr := r.FormValue("path")
	revStr := r.FormValue("restore_revision")
	if pathStr == "" || revStr == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	rev, err := strconv.ParseInt(revStr, 10, 64)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	// Parse conflict mode, default to "rename" per API spec
	conflict, err := vo.ParseConflictMode(r.FormValue("conflict"))
	if err != nil {
		conflict = vo.ConflictRename
	}

	path := vo.NewCloudPath(pathStr)
	if err := h.trash.Restore(authed.UserID, path, rev, conflict); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeHomeError(w, authed.Email, 404, "not_exists")
		case errors.Is(err, service.ErrAlreadyExists):
			writeHomeError(w, authed.Email, 400, "exists")
		default:
			writeHomeError(w, authed.Email, 500, "unknown")
		}
		return
	}

	writeSuccess(w, authed.Email, path.String())
}

// HandleTrashEmpty handles POST /api/v2/trashbin/empty - permanently delete all trashed items.
func (h *TrashHandler) HandleTrashEmpty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed := authenticate(w, r, h.auth)
	if authed == nil {
		return
	}

	if err := h.trash.Empty(authed.UserID); err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, "ok")
}
