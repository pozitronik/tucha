package httpapi

import (
	"net/http"

	"github.com/pozitronik/tucha/internal/application/service"
)

// authenticate validates the access_token query parameter and returns the authenticated user.
// If authentication fails, it writes a 403 error response and returns nil.
// Callers should return immediately when nil is returned.
func authenticate(w http.ResponseWriter, r *http.Request, auth *service.AuthService) *service.AuthenticatedUser {
	authed, err := auth.Validate(r.URL.Query().Get("access_token"))
	if err != nil || authed == nil {
		writeAuthError(w)
		return nil
	}
	return authed
}
