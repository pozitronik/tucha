package entity

import (
	"testing"
	"tucha/internal/domain/vo"
)

func TestTrashItem_IsFile(t *testing.T) {
	file := &TrashItem{Type: vo.NodeTypeFile}
	if !file.IsFile() {
		t.Error("file trash item: IsFile() = false, want true")
	}
	folder := &TrashItem{Type: vo.NodeTypeFolder}
	if folder.IsFile() {
		t.Error("folder trash item: IsFile() = true, want false")
	}
}

func TestTrashItem_IsFolder(t *testing.T) {
	folder := &TrashItem{Type: vo.NodeTypeFolder}
	if !folder.IsFolder() {
		t.Error("folder trash item: IsFolder() = false, want true")
	}
	file := &TrashItem{Type: vo.NodeTypeFile}
	if file.IsFolder() {
		t.Error("file trash item: IsFolder() = true, want false")
	}
}

func TestTrashItem_HasContent(t *testing.T) {
	hash := vo.MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")

	tests := []struct {
		name string
		item *TrashItem
		want bool
	}{
		{"file with hash", &TrashItem{Type: vo.NodeTypeFile, Hash: hash}, true},
		{"file without hash", &TrashItem{Type: vo.NodeTypeFile}, false},
		{"folder", &TrashItem{Type: vo.NodeTypeFolder, Hash: hash}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.HasContent(); got != tt.want {
				t.Errorf("HasContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
