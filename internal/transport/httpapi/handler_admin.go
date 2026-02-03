package httpapi

import (
	_ "embed"
	"net/http"
)

//go:embed admin.html
var adminHTML []byte

// AdminHandler serves the embedded admin panel HTML.
type AdminHandler struct{}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// HandleAdmin serves the admin panel SPA.
func (h *AdminHandler) HandleAdmin(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(adminHTML)
}
