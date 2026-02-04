// Package httpapi implements HTTP handlers for the Tucha cloud API v2 protocol.
package httpapi

// FolderCount holds the count of files and folders within a directory.
type FolderCount struct {
	Folders int `json:"folders"`
	Files   int `json:"files"`
}

// FolderItem represents a single item in a folder listing response.
type FolderItem struct {
	Name      string       `json:"name"`
	Home      string       `json:"home"`
	Type      string       `json:"type"`
	Kind      string       `json:"kind"`
	Size      int64        `json:"size"`
	Hash      string       `json:"hash,omitempty"`
	Weblink   string       `json:"weblink,omitempty"`
	MTime     int64        `json:"mtime,omitempty"`
	Rev       int64        `json:"rev,omitempty"`
	GRev      int64        `json:"grev,omitempty"`
	Tree      string       `json:"tree,omitempty"`
	Count     *FolderCount `json:"count,omitempty"`
	VirusScan string       `json:"virus_scan,omitempty"`
}

// FolderListing represents the response body for a folder listing.
type FolderListing struct {
	Count FolderCount  `json:"count"`
	Tree  string       `json:"tree"`
	Name  string       `json:"name"`
	GRev  int64        `json:"grev"`
	Size  int64        `json:"size"`
	Sort  SortInfo     `json:"sort"`
	Kind  string       `json:"kind"`
	Rev   int64        `json:"rev"`
	Type  string       `json:"type"`
	Home  string       `json:"home"`
	List  []FolderItem `json:"list"`
}

// SortInfo describes the sort parameters for a folder listing.
type SortInfo struct {
	Order string `json:"order"`
	Type  string `json:"type"`
}

// OAuthToken represents an OAuth2 token response.
type OAuthToken struct {
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	AccessToken      string `json:"access_token"`
	Error            string `json:"error"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

// SpaceInfo represents the user's storage quota information.
type SpaceInfo struct {
	Overquota  bool  `json:"overquota"`
	BytesTotal int64 `json:"bytes_total"`
	BytesUsed  int64 `json:"bytes_used"`
}

// TrashFolderItem represents a trashed item in the trashbin listing response.
type TrashFolderItem struct {
	FolderItem
	DeletedAt   int64  `json:"deleted_at"`
	DeletedFrom string `json:"deleted_from"`
	DeletedBy   int64  `json:"deleted_by"`
}

// ShareMember represents a member in a folder share info response.
type ShareMember struct {
	Email  string `json:"email"`
	Status string `json:"status"`
	Access string `json:"access"`
	Name   string `json:"name"`
}

// IncomingInvite represents a pending incoming share invitation.
type IncomingInvite struct {
	Owner       InviteOwner `json:"owner"`
	Tree        string      `json:"tree,omitempty"`
	Access      string      `json:"access"`
	Name        string      `json:"name"`
	Home        string      `json:"home,omitempty"`
	Size        int64       `json:"size"`
	InviteToken string      `json:"invite_token"`
}

// InviteOwner represents the owner info within an incoming invite.
type InviteOwner struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserInfo represents a user in admin API responses (password omitted).
type UserInfo struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	IsAdmin    bool   `json:"is_admin"`
	QuotaBytes int64  `json:"quota_bytes"`
	BytesUsed  int64  `json:"bytes_used"`
	Created    int64  `json:"created"`
}
