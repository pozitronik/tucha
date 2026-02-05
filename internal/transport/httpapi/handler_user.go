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
	adminAuth *service.AdminAuthService
	users     *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(adminAuth *service.AdminAuthService, users *service.UserService) *UserHandler {
	return &UserHandler{adminAuth: adminAuth, users: users}
}

// HandleUserAdd handles POST /admin/user/add - create a new user.
func (h *UserHandler) HandleUserAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.adminAuth.Validate(extractAdminToken(r)) {
		writeEnvelope(w, "", 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, "", 400, "invalid")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		writeEnvelope(w, "", 400, "required")
		return
	}

	var quotaBytes int64
	var err error
	if qStr := r.FormValue("quota_bytes"); qStr != "" {
		quotaBytes, err = strconv.ParseInt(qStr, 10, 64)
		if err != nil {
			writeEnvelope(w, "", 400, "invalid")
			return
		}
	}

	user, err := h.users.Create(email, password, false, quotaBytes)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			writeEnvelope(w, "", 400, "exists")
			return
		}
		writeEnvelope(w, "", 500, "unknown")
		return
	}

	writeSuccess(w, "", userToInfo(user))
}

// HandleUserList handles GET /admin/user/list - list all users.
func (h *UserHandler) HandleUserList(w http.ResponseWriter, r *http.Request) {
	if !h.adminAuth.Validate(extractAdminToken(r)) {
		writeEnvelope(w, "", 403, "forbidden")
		return
	}

	users, err := h.users.ListWithUsage()
	if err != nil {
		writeEnvelope(w, "", 500, "unknown")
		return
	}

	infos := make([]UserInfo, 0, len(users))
	for _, u := range users {
		infos = append(infos, UserInfo{
			ID:         u.ID,
			Email:      u.Email,
			Password:   u.Password,
			QuotaBytes: u.QuotaBytes,
			BytesUsed:  u.BytesUsed,
			Created:    u.Created,
		})
	}

	writeSuccess(w, "", infos)
}

// HandleUserEdit handles POST /admin/user/edit - update user fields.
func (h *UserHandler) HandleUserEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.adminAuth.Validate(extractAdminToken(r)) {
		writeEnvelope(w, "", 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, "", 400, "invalid")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeEnvelope(w, "", 400, "required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeEnvelope(w, "", 400, "invalid")
		return
	}

	user := &entity.User{ID: id}

	if v := r.FormValue("email"); v != "" {
		user.Email = v
	}
	if v := r.FormValue("password"); v != "" {
		user.Password = v
	}
	if qStr := r.FormValue("quota_bytes"); qStr != "" {
		user.QuotaBytes, err = strconv.ParseInt(qStr, 10, 64)
		if err != nil {
			writeEnvelope(w, "", 400, "invalid")
			return
		}
	}

	if err := h.users.Update(user); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeEnvelope(w, "", 404, "not_found")
			return
		}
		writeEnvelope(w, "", 500, "unknown")
		return
	}

	writeSuccess(w, "", "ok")
}

// HandleUserRemove handles POST /admin/user/remove - delete a user.
func (h *UserHandler) HandleUserRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.adminAuth.Validate(extractAdminToken(r)) {
		writeEnvelope(w, "", 403, "forbidden")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeEnvelope(w, "", 400, "invalid")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeEnvelope(w, "", 400, "required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeEnvelope(w, "", 400, "invalid")
		return
	}

	if err := h.users.Delete(id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeEnvelope(w, "", 404, "not_found")
			return
		}
		writeEnvelope(w, "", 500, "unknown")
		return
	}

	writeSuccess(w, "", "ok")
}

// userToInfo converts a User pointer to a UserInfo DTO.
func userToInfo(u *entity.User) UserInfo {
	return UserInfo{
		ID:         u.ID,
		Email:      u.Email,
		Password:   u.Password,
		QuotaBytes: u.QuotaBytes,
		Created:    u.Created,
	}
}
