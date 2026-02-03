package service

import "errors"

// Sentinel errors used across application services.
var (
	// ErrAlreadyExists indicates the target path is already occupied.
	ErrAlreadyExists = errors.New("already exists")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrOverQuota indicates the operation would exceed the storage quota.
	ErrOverQuota = errors.New("over quota")

	// ErrContentNotFound indicates the content hash is not available in storage.
	ErrContentNotFound = errors.New("content not found")
)
