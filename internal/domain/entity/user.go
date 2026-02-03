// Package entity defines domain entities with behavior.
package entity

// User represents a registered user account.
type User struct {
	ID         int64
	Email      string
	Password   string
	IsAdmin    bool
	QuotaBytes int64
	Created    int64
}
