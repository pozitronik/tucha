// Package entity defines domain entities with behavior.
package entity

// User represents a registered user account.
type User struct {
	ID       int64
	Email    string
	Password string
	Created  int64
}
