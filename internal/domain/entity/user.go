// Package entity defines domain entities with behavior.
package entity

// User represents a registered user account.
type User struct {
	ID             int64
	Email          string
	Password       string
	IsAdmin        bool
	QuotaBytes     int64
	FileSizeLimit  int64 // 0 = unlimited
	VersionHistory bool  // true = paid tier
	Created        int64
}
