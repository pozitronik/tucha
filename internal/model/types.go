// Package model defines data structures used across the application.
package model

// User represents a registered user account.
type User struct {
	ID       int64  `json:"-"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Created  int64  `json:"-"`
}

// Node represents a file or folder in the virtual filesystem.
type Node struct {
	ID       int64  `json:"-"`
	UserID   int64  `json:"-"`
	ParentID *int64 `json:"-"`
	Name     string `json:"name"`
	Home     string `json:"home"`
	Type     string `json:"type"` // "file" or "folder"
	Size     int64  `json:"size"`
	Hash     string `json:"hash,omitempty"`
	MTime    int64  `json:"mtime,omitempty"`
	Rev      int64  `json:"rev"`
	GRev     int64  `json:"grev"`
	Tree     string `json:"tree,omitempty"`
	Created  int64  `json:"-"`
}

// FolderCount holds the count of files and folders within a directory.
type FolderCount struct {
	Folders int `json:"folders"`
	Files   int `json:"files"`
}

// FolderItem represents a single item in a folder listing response.
// It merges file and folder fields; omitted fields are zero-valued.
type FolderItem struct {
	Name      string       `json:"name"`
	Home      string       `json:"home"`
	Type      string       `json:"type"`
	Kind      string       `json:"kind"`
	Size      int64        `json:"size"`
	Hash      string       `json:"hash,omitempty"`
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

// Token represents an authentication token stored in the database.
type Token struct {
	ID           int64
	UserID       int64
	AccessToken  string
	RefreshToken string
	CSRFToken    string
	ExpiresAt    int64
	Created      int64
}

// Content represents a content-addressable storage entry.
type Content struct {
	Hash     string
	Size     int64
	RefCount int64
	Created  int64
}

// SpaceInfo represents the user's storage quota information.
type SpaceInfo struct {
	Overquota  bool  `json:"overquota"`
	BytesTotal int64 `json:"bytes_total"`
	BytesUsed  int64 `json:"bytes_used"`
}
