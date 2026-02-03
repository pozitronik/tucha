package vo

import (
	"fmt"
	"regexp"
	"strings"
)

// hashPattern matches a 40-character uppercase hex string.
var hashPattern = regexp.MustCompile(`^[0-9A-F]{40}$`)

// ContentHash represents a validated 40-character uppercase hex hash.
type ContentHash struct {
	value string
}

// NewContentHash creates a ContentHash from a raw string.
// Returns an error if the string is not a valid 40-char hex hash.
func NewContentHash(raw string) (ContentHash, error) {
	upper := strings.ToUpper(raw)
	if !hashPattern.MatchString(upper) {
		return ContentHash{}, fmt.Errorf("invalid content hash: must be 40 uppercase hex characters, got %q", raw)
	}
	return ContentHash{value: upper}, nil
}

// MustContentHash creates a ContentHash, panicking on invalid input.
// Use only when the hash is known to be valid (e.g., from the database).
func MustContentHash(raw string) ContentHash {
	h, err := NewContentHash(raw)
	if err != nil {
		panic(err)
	}
	return h
}

// String returns the 40-character uppercase hex string.
func (h ContentHash) String() string {
	return h.value
}

// ShardPrefix returns the two-level directory prefix for content-addressable storage.
// Example: "C172..." -> ("C1", "72")
func (h ContentHash) ShardPrefix() (string, string) {
	return h.value[0:2], h.value[2:4]
}

// IsZero returns true if the hash is unset.
func (h ContentHash) IsZero() bool {
	return h.value == ""
}
