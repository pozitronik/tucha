// Package vo provides value objects for the domain layer.
package vo

import (
	"path"
	"strings"
)

// CloudPath represents a normalized cloud filesystem path.
// It is immutable and always starts with "/".
type CloudPath struct {
	value string
}

// NewCloudPath creates a CloudPath from a raw string, normalizing it.
// Empty or whitespace-only strings resolve to "/".
func NewCloudPath(raw string) CloudPath {
	return CloudPath{value: normalize(raw)}
}

// String returns the normalized path string.
func (p CloudPath) String() string {
	return p.value
}

// Name returns the last element of the path (the file or folder name).
func (p CloudPath) Name() string {
	return path.Base(p.value)
}

// Parent returns the parent directory path.
// The parent of "/" is "/".
func (p CloudPath) Parent() CloudPath {
	return CloudPath{value: path.Dir(p.value)}
}

// IsRoot returns true if this is the root path "/".
func (p CloudPath) IsRoot() bool {
	return p.value == "/"
}

// Join appends a child name to this path and returns a new CloudPath.
func (p CloudPath) Join(name string) CloudPath {
	return CloudPath{value: normalize(path.Join(p.value, name))}
}

// HasPrefix returns true if this path starts with prefix followed by "/".
// Used for identifying descendants during rename/move.
func (p CloudPath) HasPrefix(prefix CloudPath) bool {
	return strings.HasPrefix(p.value, prefix.value+"/")
}

// normalize cleans the path and ensures it starts with "/".
func normalize(raw string) string {
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return "/"
	}
	raw = path.Clean(raw)
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return raw
}
