package httpapi

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/pozitronik/tucha/internal/application/service"
)

//go:embed admin.html
var adminHTML []byte

// AdminHandler serves the admin panel and handles admin login/logout.
type AdminHandler struct {
	adminAuth *service.AdminAuthService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(adminAuth *service.AdminAuthService) *AdminHandler {
	return &AdminHandler{adminAuth: adminAuth}
}

// HandleAdmin serves the admin panel SPA.
func (h *AdminHandler) HandleAdmin(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(adminHTML)
}

// HandleLogin handles POST /admin/login - authenticate admin and issue bearer token.
func (h *AdminHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": 400,
			"body":   "invalid",
		})
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	token, err := h.adminAuth.Login(login, password)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"status": 403,
			"body":   "forbidden",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": 200,
		"body": map[string]string{
			"token": token,
		},
	})
}

// HandleLogout handles POST /admin/logout - invalidate token and clear cookie.
func (h *AdminHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := extractAdminToken(r)
	if token != "" {
		h.adminAuth.Logout(token)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": 200,
		"body":   "ok",
	})
}

// extractAdminToken reads the admin bearer token from Authorization header or cookie.
func extractAdminToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := r.Cookie("admin_token"); err == nil {
		return cookie.Value
	}
	return ""
}
