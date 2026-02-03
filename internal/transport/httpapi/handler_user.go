package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"tucha/internal/application/service"
	"tucha/internal/domain/entity"
)

// UserHandler handles admin user management CRUD operations.
type UserHandler struct {
	auth  *service.AuthService
	users *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(auth *service.AuthService, users *service.UserService) *UserHandler {
	return &UserHandler{auth: auth, users: users}
}

// HandleUserAdd handles POST /api/v2/admin/user/add - create a new user.
func (h *UserHandler) HandleUserAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}
	if !authed.IsAdmin {
		writeEnvelope(w, authed.Email, 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, authed.Email, 400, "invalid")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		writeEnvelope(w, authed.Email, 400, "required")
		return
	}

	isAdmin := r.FormValue("is_admin") == "1"

	var quotaBytes int64
	if qStr := r.FormValue("quota_bytes"); qStr != "" {
		quotaBytes, err = strconv.ParseInt(qStr, 10, 64)
		if err != nil {
			writeEnvelope(w, authed.Email, 400, "invalid")
			return
		}
	}

	user, err := h.users.Create(email, password, isAdmin, quotaBytes)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			writeEnvelope(w, authed.Email, 400, "exists")
			return
		}
		writeEnvelope(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, userToInfo(user))
}

// HandleUserList handles GET /api/v2/admin/user/list - list all users.
func (h *UserHandler) HandleUserList(w http.ResponseWriter, r *http.Request) {
	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}
	if !authed.IsAdmin {
		writeEnvelope(w, authed.Email, 403, "forbidden")
		return
	}

	users, err := h.users.List()
	if err != nil {
		writeEnvelope(w, authed.Email, 500, "unknown")
		return
	}

	infos := make([]UserInfo, 0, len(users))
	for i := range users {
		infos = append(infos, userEntityToInfo(&users[i]))
	}

	writeSuccess(w, authed.Email, infos)
}

// HandleUserEdit handles POST /api/v2/admin/user/edit - update user fields.
func (h *UserHandler) HandleUserEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}
	if !authed.IsAdmin {
		writeEnvelope(w, authed.Email, 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, authed.Email, 400, "invalid")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeEnvelope(w, authed.Email, 400, "required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeEnvelope(w, authed.Email, 400, "invalid")
		return
	}

	user := &entity.User{ID: id}

	if v := r.FormValue("email"); v != "" {
		user.Email = v
	}
	if v := r.FormValue("password"); v != "" {
		user.Password = v
	}
	user.IsAdmin = r.FormValue("is_admin") == "1"
	if qStr := r.FormValue("quota_bytes"); qStr != "" {
		user.QuotaBytes, err = strconv.ParseInt(qStr, 10, 64)
		if err != nil {
			writeEnvelope(w, authed.Email, 400, "invalid")
			return
		}
	}

	if err := h.users.Update(user); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeEnvelope(w, authed.Email, 404, "not_found")
			return
		}
		writeEnvelope(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, "ok")
}

// HandleUserRemove handles POST /api/v2/admin/user/remove - delete a user.
func (h *UserHandler) HandleUserRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authed, err := h.auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeEnvelope(w, "", 403, "user")
		return
	}
	if !authed.IsAdmin {
		writeEnvelope(w, authed.Email, 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, authed.Email, 400, "invalid")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeEnvelope(w, authed.Email, 400, "required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeEnvelope(w, authed.Email, 400, "invalid")
		return
	}

	if err := h.users.Delete(authed.UserID, id); err != nil {
		if errors.Is(err, service.ErrSelfDelete) {
			writeEnvelope(w, authed.Email, 400, "self_delete")
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			writeEnvelope(w, authed.Email, 404, "not_found")
			return
		}
		writeEnvelope(w, authed.Email, 500, "unknown")
		return
	}

	writeSuccess(w, authed.Email, "ok")
}

// userToInfo converts a User pointer to a UserInfo DTO.
func userToInfo(u *entity.User) UserInfo {
	return UserInfo{
		ID:         u.ID,
		Email:      u.Email,
		IsAdmin:    u.IsAdmin,
		QuotaBytes: u.QuotaBytes,
		Created:    u.Created,
	}
}

// userEntityToInfo converts a User pointer to a UserInfo DTO (same as userToInfo, for slice iteration).
func userEntityToInfo(u *entity.User) UserInfo {
	return userToInfo(u)
}
