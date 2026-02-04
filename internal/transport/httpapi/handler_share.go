package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"tucha/internal/application/service"
	"tucha/internal/domain/vo"
)

// ShareHandler handles folder sharing operations: share, unshare, info, incoming,
// mount, unmount, and reject.
type ShareHandler struct {
	auth      *service.AuthService
	shares    *service.ShareService
	presenter *Presenter
}

// NewShareHandler creates a new ShareHandler.
func NewShareHandler(
	auth *service.AuthService,
	shares *service.ShareService,
	presenter *Presenter,
) *ShareHandler {
	return &ShareHandler{
		auth:      auth,
		shares:    shares,
		presenter: presenter,
	}
}

// HandleShare handles POST /api/v2/folder/share - create a share invitation.
// The invite is sent as a JSON string in the "invite" form field.
func (h *ShareHandler) HandleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	homePath := r.FormValue("home")
	inviteJSON := r.FormValue("invite")
	if homePath == "" || inviteJSON == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	var invite struct {
		Email  string `json:"email"`
		Access string `json:"access"`
	}
	if err := json.Unmarshal([]byte(inviteJSON), &invite); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	if invite.Email == "" || invite.Access == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	access, err := vo.ParseAccessLevel(invite.Access)
	if err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	path := vo.NewCloudPath(homePath)
	share, err := h.shares.Share(authed.UserID, path, invite.Email, access)
	if err != nil {
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

	writeSuccess(w, authed.Email, share.InviteToken)
}

// HandleUnshare handles POST /api/v2/folder/unshare - remove a share invitation.
func (h *ShareHandler) HandleUnshare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	homePath := r.FormValue("home")
	inviteJSON := r.FormValue("invite")
	if homePath == "" || inviteJSON == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	var invite struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal([]byte(inviteJSON), &invite); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	if invite.Email == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	if err := h.shares.Unshare(authed.UserID, path, invite.Email); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeHomeError(w, authed.Email, 404, "not_exists")
			return
		}
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, path.String())
}

// HandleSharedInfo handles GET /api/v2/folder/shared/info - list share members for a folder.
func (h *ShareHandler) HandleSharedInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	homePath := r.URL.Query().Get("home")
	if homePath == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	path := vo.NewCloudPath(homePath)
	shares, err := h.shares.GetShareInfo(authed.UserID, path)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	var members []ShareMember
	for i := range shares {
		members = append(members, h.presenter.ShareToMember(&shares[i]))
	}

	writeSuccess(w, authed.Email, members)
}

// HandleIncoming handles GET /api/v2/folder/shared/incoming - list pending incoming invites.
func (h *ShareHandler) HandleIncoming(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	shares, err := h.shares.ListIncoming(authed.Email)
	if err != nil {
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	var invites []IncomingInvite
	for i := range shares {
		// Resolve owner email from owner ID.
		ownerEmail := ""
		owner, _ := h.auth.ResolveUser(shares[i].OwnerID)
		if owner != nil {
			ownerEmail = owner.Email
		}
		invites = append(invites, h.presenter.ShareToIncomingInvite(&shares[i], ownerEmail))
	}

	writeSuccess(w, authed.Email, invites)
}

// HandleMount handles POST /api/v2/folder/mount - accept and mount a shared folder.
func (h *ShareHandler) HandleMount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	name := r.FormValue("name")
	inviteToken := r.FormValue("invite_token")
	if name == "" || inviteToken == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	if err := h.shares.Mount(authed.UserID, name, inviteToken); err != nil {
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

// HandleUnmount handles POST /api/v2/folder/unmount - unmount a shared folder.
func (h *ShareHandler) HandleUnmount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
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

	cloneCopy := r.FormValue("clone_copy") == "true"

	if err := h.shares.Unmount(authed.UserID, homePath, cloneCopy); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeHomeError(w, authed.Email, 404, "not_exists")
			return
		}
		writeHomeError(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, "ok")
}

// HandleReject handles POST /api/v2/folder/invites/reject - reject a share invitation.
func (h *ShareHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeHomeError(w, authed.Email, 400, "invalid")
		return
	}

	inviteToken := r.FormValue("invite_token")
	if inviteToken == "" {
		writeHomeError(w, authed.Email, 400, "required")
		return
	}

	if err := h.shares.Reject(authed.UserID, inviteToken); err != nil {
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
