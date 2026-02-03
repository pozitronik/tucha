package vo

import "fmt"

// NodeType represents the type of a filesystem node.
type NodeType string

const (
	// NodeTypeFile represents a file node.
	NodeTypeFile NodeType = "file"
	// NodeTypeFolder represents a folder node.
	NodeTypeFolder NodeType = "folder"
)

// ParseNodeType converts a raw string to a NodeType.
// Returns an error for unknown values.
func ParseNodeType(raw string) (NodeType, error) {
	switch raw {
	case "file":
		return NodeTypeFile, nil
	case "folder":
		return NodeTypeFolder, nil
	default:
		return "", fmt.Errorf("unknown node type: %q", raw)
	}
}

// String returns the string representation of the node type.
func (t NodeType) String() string {
	return string(t)
}

// IsFile returns true if the node type is "file".
func (t NodeType) IsFile() bool {
	return t == NodeTypeFile
}

// IsFolder returns true if the node type is "folder".
func (t NodeType) IsFolder() bool {
	return t == NodeTypeFolder
}
