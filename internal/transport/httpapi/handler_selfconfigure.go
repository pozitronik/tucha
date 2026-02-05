package httpapi

import (
	"net/http"

	"github.com/pozitronik/tucha/internal/config"
)

// SelfConfigureHandler returns the configured endpoint URLs for service discovery.
// This endpoint is unauthenticated because clients need it before they can authenticate.
type SelfConfigureHandler struct {
	endpoints config.EndpointsConfig
}

// NewSelfConfigureHandler creates a new SelfConfigureHandler.
func NewSelfConfigureHandler(endpoints config.EndpointsConfig) *SelfConfigureHandler {
	return &SelfConfigureHandler{endpoints: endpoints}
}

// selfConfigureResponse is the JSON structure returned by /self-configure.
type selfConfigureResponse struct {
	API        string `json:"api"`
	OAuth      string `json:"oauth"`
	Dispatcher string `json:"dispatcher"`
	Upload     string `json:"upload"`
	Download   string `json:"download"`
}

// HandleSelfConfigure handles GET /.
// The "/" pattern in http.ServeMux is a catch-all, so we reject non-root paths explicitly.
func (h *SelfConfigureHandler) HandleSelfConfigure(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := selfConfigureResponse{
		API:        h.endpoints.API,
		OAuth:      h.endpoints.OAuth,
		Dispatcher: h.endpoints.Dispatcher,
		Upload:     h.endpoints.Upload,
		Download:   h.endpoints.Download,
	}

	writeJSON(w, http.StatusOK, resp)
}
