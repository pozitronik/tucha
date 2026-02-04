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
	userH *UserHandler,
	adminH *AdminHandler,
	trashH *TrashHandler,
	publishH *PublishHandler,
	weblinkH *WeblinkDownloadHandler,
	shareH *ShareHandler,
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
	mux.HandleFunc("/upload", uploadH.HandleUpload)
	mux.HandleFunc("/get/", downloadH.HandleDownload)

	// User space/quota.
	mux.HandleFunc("/api/v2/user/space", spaceH.HandleSpace)

	// Admin panel and authentication.
	mux.HandleFunc("/admin", adminH.HandleAdmin)
	mux.HandleFunc("/admin/login", adminH.HandleLogin)
	mux.HandleFunc("/admin/logout", adminH.HandleLogout)

	// Admin user management.
	mux.HandleFunc("/admin/user/add", userH.HandleUserAdd)
	mux.HandleFunc("/admin/user/list", userH.HandleUserList)
	mux.HandleFunc("/admin/user/edit", userH.HandleUserEdit)
	mux.HandleFunc("/admin/user/remove", userH.HandleUserRemove)

	// Trashbin.
	mux.HandleFunc("/api/v2/trashbin", trashH.HandleTrashList)
	mux.HandleFunc("/api/v2/trashbin/restore", trashH.HandleTrashRestore)
	mux.HandleFunc("/api/v2/trashbin/empty", trashH.HandleTrashEmpty)

	// Publishing / weblinks.
	mux.HandleFunc("/api/v2/file/publish", publishH.HandlePublish)
	mux.HandleFunc("/api/v2/file/unpublish", publishH.HandleUnpublish)
	mux.HandleFunc("/api/v2/folder/shared/links", publishH.HandleSharedLinks)
	mux.HandleFunc("/api/v2/clone", publishH.HandleClone)
	mux.HandleFunc("/public/", weblinkH.HandleWeblinkDownload)

	// Folder sharing / invites.
	mux.HandleFunc("/api/v2/folder/share", shareH.HandleShare)
	mux.HandleFunc("/api/v2/folder/unshare", shareH.HandleUnshare)
	mux.HandleFunc("/api/v2/folder/shared/info", shareH.HandleSharedInfo)
	mux.HandleFunc("/api/v2/folder/shared/incoming", shareH.HandleIncoming)
	mux.HandleFunc("/api/v2/folder/mount", shareH.HandleMount)
	mux.HandleFunc("/api/v2/folder/unmount", shareH.HandleUnmount)
	mux.HandleFunc("/api/v2/folder/invites/reject", shareH.HandleReject)
}
