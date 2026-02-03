package api

import (
	"net/http"

	"tucha/internal/auth"
	"tucha/internal/config"
	"tucha/internal/content"
	"tucha/internal/storage"
)

// Handlers holds all dependencies needed by HTTP handlers.
type Handlers struct {
	cfg      *config.Config
	auth     *auth.Authenticator
	tokens   *storage.TokenStore
	nodes    *storage.NodeStore
	contents *storage.ContentStore
	store    *content.Store
	userID   int64
}

// NewHandlers creates a new Handlers instance with all required dependencies.
func NewHandlers(
	cfg *config.Config,
	authenticator *auth.Authenticator,
	tokens *storage.TokenStore,
	nodes *storage.NodeStore,
	contents *storage.ContentStore,
	store *content.Store,
	userID int64,
) *Handlers {
	return &Handlers{
		cfg:      cfg,
		auth:     authenticator,
		tokens:   tokens,
		nodes:    nodes,
		contents: contents,
		store:    store,
		userID:   userID,
	}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	// OAuth token endpoint.
	mux.HandleFunc("/token", h.handleToken)

	// CSRF token.
	mux.HandleFunc("/api/v2/tokens/csrf", h.handleCSRF)

	// Dispatcher.
	mux.HandleFunc("/api/v2/dispatcher/", h.handleDispatcher)
	mux.HandleFunc("/d", h.handleOAuthDispatcher)
	mux.HandleFunc("/u", h.handleOAuthDispatcher)

	// Folder listing and file metadata.
	mux.HandleFunc("/api/v2/folder", h.handleFolder)
	mux.HandleFunc("/api/v2/file", h.handleFile)

	// Folder operations.
	mux.HandleFunc("/api/v2/folder/add", h.handleFolderAdd)

	// File operations.
	mux.HandleFunc("/api/v2/file/remove", h.handleFileRemove)
	mux.HandleFunc("/api/v2/file/rename", h.handleFileRename)
	mux.HandleFunc("/api/v2/file/move", h.handleFileMove)
	mux.HandleFunc("/api/v2/file/copy", h.handleFileCopy)
	mux.HandleFunc("/api/v2/file/add", h.handleFileAdd)

	// Upload and download.
	mux.HandleFunc("/upload/", h.handleUpload)
	mux.HandleFunc("/get/", h.handleDownload)

	// User space/quota.
	mux.HandleFunc("/api/v2/user/space", h.handleSpace)
}
