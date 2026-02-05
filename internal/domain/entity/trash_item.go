package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// TrashItem represents a node that has been soft-deleted into the trashbin.
// It preserves all original node fields plus deletion metadata.
type TrashItem struct {
	ID          int64
	UserID      int64
	Name        string
	Home        vo.CloudPath
	Type        vo.NodeType
	Size        int64
	Hash        vo.ContentHash
	MTime       int64
	Rev         int64
	GRev        int64
	Tree        string
	DeletedAt   int64
	DeletedFrom string
	DeletedBy   int64
	Created     int64
}

// IsFile returns true if this trash item was a file.
func (t *TrashItem) IsFile() bool {
	return t.Type.IsFile()
}

// IsFolder returns true if this trash item was a folder.
func (t *TrashItem) IsFolder() bool {
	return t.Type.IsFolder()
}

// HasContent returns true if this trash item has an associated content hash.
func (t *TrashItem) HasContent() bool {
	return t.IsFile() && !t.Hash.IsZero()
}
