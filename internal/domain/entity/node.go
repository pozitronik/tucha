package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// Node represents a file or folder in the virtual filesystem.
type Node struct {
	ID       int64
	UserID   int64
	ParentID *int64
	Name     string
	Home     vo.CloudPath
	Type     vo.NodeType
	Size     int64
	Hash     vo.ContentHash
	MTime    int64
	Rev      int64
	GRev     int64
	Tree     string
	Weblink  string
	Created  int64
}

// IsFile returns true if this node is a file.
func (n *Node) IsFile() bool {
	return n.Type.IsFile()
}

// IsFolder returns true if this node is a folder.
func (n *Node) IsFolder() bool {
	return n.Type.IsFolder()
}

// HasContent returns true if this file node has an associated content hash.
func (n *Node) HasContent() bool {
	return n.IsFile() && !n.Hash.IsZero()
}

// IsRoot returns true if this node is the root folder (no parent).
func (n *Node) IsRoot() bool {
	return n.ParentID == nil
}
