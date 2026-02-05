package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// Content represents a content-addressable storage entry.
type Content struct {
	Hash     vo.ContentHash
	Size     int64
	RefCount int64
	Created  int64
}

// IsUnreferenced returns true if no file nodes reference this content.
func (c *Content) IsUnreferenced() bool {
	return c.RefCount <= 0
}
