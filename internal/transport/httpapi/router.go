package httpapi

import "net/http"

// RegisterRoutes registers all API routes on the given mux.
func RegisterRoutes(
	mux *http.ServeMux,
	tokenH *TokenHandler,
	csrfH *CSRFHandler,
	dispatchH *DispatchHandler,
	folderH *FolderHandler,
	fileH *FileHandler,
	uploadH *UploadHandler,
	downloadH *DownloadHandler,
	spaceH *SpaceHandler,
	selfConfigH *SelfConfigureHandler,
) {
	// Service discovery (unauthenticated).
	mux.HandleFunc("/", selfConfigH.HandleSelfConfigure)

	// OAuth token endpoint.
	mux.HandleFunc("/token", tokenH.HandleToken)

	// CSRF token.
	mux.HandleFunc("/api/v2/tokens/csrf", csrfH.HandleCSRF)

	// Dispatcher.
	mux.HandleFunc("/api/v2/dispatcher/", dispatchH.HandleDispatcher)
	mux.HandleFunc("/d", dispatchH.HandleOAuthDispatcher)
	mux.HandleFunc("/u", dispatchH.HandleOAuthDispatcher)

	// Folder listing and creation.
	mux.HandleFunc("/api/v2/folder", folderH.HandleFolder)
	mux.HandleFunc("/api/v2/folder/add", folderH.HandleFolderAdd)

	// File metadata and operations.
	mux.HandleFunc("/api/v2/file", fileH.HandleFile)
	mux.HandleFunc("/api/v2/file/remove", fileH.HandleFileRemove)
	mux.HandleFunc("/api/v2/file/rename", fileH.HandleFileRename)
	mux.HandleFunc("/api/v2/file/move", fileH.HandleFileMove)
	mux.HandleFunc("/api/v2/file/copy", fileH.HandleFileCopy)
	mux.HandleFunc("/api/v2/file/add", fileH.HandleFileAdd)

	// Upload and download.
	mux.HandleFunc("/upload/", uploadH.HandleUpload)
	mux.HandleFunc("/get/", downloadH.HandleDownload)

	// User space/quota.
	mux.HandleFunc("/api/v2/user/space", spaceH.HandleSpace)
}
