package entity

import (
	"testing"
	"tucha/internal/domain/vo"
)

func TestNode_IsFile(t *testing.T) {
	file := &Node{Type: vo.NodeTypeFile}
	if !file.IsFile() {
		t.Error("file node: IsFile() = false, want true")
	}
	folder := &Node{Type: vo.NodeTypeFolder}
	if folder.IsFile() {
		t.Error("folder node: IsFile() = true, want false")
	}
}

func TestNode_IsFolder(t *testing.T) {
	folder := &Node{Type: vo.NodeTypeFolder}
	if !folder.IsFolder() {
		t.Error("folder node: IsFolder() = false, want true")
	}
	file := &Node{Type: vo.NodeTypeFile}
	if file.IsFolder() {
		t.Error("file node: IsFolder() = true, want false")
	}
}

func TestNode_HasContent(t *testing.T) {
	hash := vo.MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")

	tests := []struct {
		name string
		node *Node
		want bool
	}{
		{"file with hash", &Node{Type: vo.NodeTypeFile, Hash: hash}, true},
		{"file without hash", &Node{Type: vo.NodeTypeFile}, false},
		{"folder with hash", &Node{Type: vo.NodeTypeFolder, Hash: hash}, false},
		{"folder without hash", &Node{Type: vo.NodeTypeFolder}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.HasContent(); got != tt.want {
				t.Errorf("HasContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_IsRoot(t *testing.T) {
	parentID := int64(1)
	tests := []struct {
		name     string
		parentID *int64
		want     bool
	}{
		{"nil parent", nil, true},
		{"non-nil parent", &parentID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{ParentID: tt.parentID}
			if got := n.IsRoot(); got != tt.want {
				t.Errorf("IsRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
