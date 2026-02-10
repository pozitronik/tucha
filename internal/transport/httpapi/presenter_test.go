package httpapi

import (
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

func validHash() vo.ContentHash {
	return vo.MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")
}

func TestPresenter_NodeToFolderItem_file(t *testing.T) {
	p := NewPresenter()
	node := &entity.Node{
		Name: "photo.jpg",
		Home: vo.NewCloudPath("/photos/photo.jpg"),
		Type: vo.NodeTypeFile,
		Size: 12345,
		Hash: validHash(),
		Rev:  3,
		GRev: 5,
	}

	item := p.NodeToFolderItem(node, nil)

	if item.Name != "photo.jpg" {
		t.Errorf("Name = %q, want %q", item.Name, "photo.jpg")
	}
	if item.Type != "file" {
		t.Errorf("Type = %q, want %q", item.Type, "file")
	}
	if item.Hash != validHash().String() {
		t.Errorf("Hash = %q, want %q", item.Hash, validHash().String())
	}
	if item.VirusScan != "pass" {
		t.Errorf("VirusScan = %q, want %q", item.VirusScan, "pass")
	}
	if item.Count != nil {
		t.Errorf("Count should be nil for file, got %v", item.Count)
	}
	if item.Tree != "" {
		t.Errorf("Tree should be empty for file, got %q", item.Tree)
	}
}

func TestPresenter_NodeToFolderItem_folder(t *testing.T) {
	p := NewPresenter()
	node := &entity.Node{
		Name: "documents",
		Home: vo.NewCloudPath("/documents"),
		Type: vo.NodeTypeFolder,
		Tree: "abc123",
		Rev:  2,
		GRev: 4,
	}
	count := &FolderCount{Folders: 3, Files: 10}

	item := p.NodeToFolderItem(node, count)

	if item.Type != "folder" {
		t.Errorf("Type = %q, want %q", item.Type, "folder")
	}
	if item.Tree != "abc123" {
		t.Errorf("Tree = %q, want %q", item.Tree, "abc123")
	}
	if item.Count == nil {
		t.Fatal("Count should not be nil for folder")
	}
	if item.Count.Folders != 3 || item.Count.Files != 10 {
		t.Errorf("Count = %+v, want Folders=3, Files=10", item.Count)
	}
	if item.Hash != "" {
		t.Errorf("Hash should be empty for folder, got %q", item.Hash)
	}
	if item.VirusScan != "" {
		t.Errorf("VirusScan should be empty for folder, got %q", item.VirusScan)
	}
}

func TestPresenter_NodeToFolderItem_folder_nilCount(t *testing.T) {
	p := NewPresenter()
	node := &entity.Node{
		Name: "empty",
		Home: vo.NewCloudPath("/empty"),
		Type: vo.NodeTypeFolder,
	}

	item := p.NodeToFolderItem(node, nil)
	if item.Count != nil {
		t.Errorf("Count should be nil when subCount is nil, got %v", item.Count)
	}
}

func TestPresenter_TrashItemToDTO(t *testing.T) {
	p := NewPresenter()
	item := &entity.TrashItem{
		Name:        "deleted.txt",
		Home:        vo.NewCloudPath("/deleted.txt"),
		Type:        vo.NodeTypeFile,
		Size:        999,
		Hash:        validHash(),
		Rev:         7,
		DeletedAt:   1700000000,
		DeletedFrom: "/documents",
		DeletedBy:   42,
	}

	dto := p.TrashItemToDTO(item)
	if dto.Name != "deleted.txt" {
		t.Errorf("Name = %q, want %q", dto.Name, "deleted.txt")
	}
	if dto.DeletedAt != 1700000000 {
		t.Errorf("DeletedAt = %d, want 1700000000", dto.DeletedAt)
	}
	if dto.DeletedFrom != "/documents" {
		t.Errorf("DeletedFrom = %q, want %q", dto.DeletedFrom, "/documents")
	}
	if dto.DeletedBy != 42 {
		t.Errorf("DeletedBy = %d, want 42", dto.DeletedBy)
	}
	if dto.VirusScan != "pass" {
		t.Errorf("VirusScan = %q, want %q", dto.VirusScan, "pass")
	}
}

func TestPresenter_ShareToMember(t *testing.T) {
	p := NewPresenter()

	tests := []struct {
		access     vo.AccessLevel
		wantAccess string
	}{
		{vo.AccessReadOnly, "r"},
		{vo.AccessReadWrite, "rw"},
	}
	for _, tt := range tests {
		t.Run(tt.wantAccess, func(t *testing.T) {
			share := &entity.Share{
				InvitedEmail: "user@example.com",
				Access:       tt.access,
				Status:       vo.SharePending,
			}
			m := p.ShareToMember(share)
			if m.Email != "user@example.com" {
				t.Errorf("Email = %q", m.Email)
			}
			if m.Access != tt.wantAccess {
				t.Errorf("Access = %q, want %q", m.Access, tt.wantAccess)
			}
			if m.Status != "pending" {
				t.Errorf("Status = %q, want %q", m.Status, "pending")
			}
			if m.Name != "user@example.com" {
				t.Errorf("Name = %q, want email as name", m.Name)
			}
		})
	}
}

func TestPresenter_ShareToIncomingInvite_pending(t *testing.T) {
	p := NewPresenter()
	share := &entity.Share{
		Home:        vo.NewCloudPath("/shared/folder"),
		Status:      vo.SharePending,
		Access:      vo.AccessReadWrite,
		InviteToken: "invite-abc",
	}

	inv := p.ShareToIncomingInvite(share, "owner@example.com")
	if inv.Owner.Email != "owner@example.com" {
		t.Errorf("Owner.Email = %q", inv.Owner.Email)
	}
	if inv.Access != "rw" {
		t.Errorf("Access = %q, want %q", inv.Access, "rw")
	}
	if inv.Name != "folder" {
		t.Errorf("Name = %q, want %q", inv.Name, "folder")
	}
	if inv.Home != "" {
		t.Errorf("Home = %q, want empty for pending invite", inv.Home)
	}
	if inv.IsMounted {
		t.Error("IsMounted = true, want false for pending invite")
	}
	if inv.InviteToken != "invite-abc" {
		t.Errorf("InviteToken = %q", inv.InviteToken)
	}
}

func TestPresenter_ShareToIncomingInvite_accepted(t *testing.T) {
	p := NewPresenter()
	mountUserID := int64(42)
	share := &entity.Share{
		Home:        vo.NewCloudPath("/shared/folder"),
		Status:      vo.ShareAccepted,
		Access:      vo.AccessReadOnly,
		InviteToken: "invite-xyz",
		MountHome:   "/MountedFolder",
		MountUserID: &mountUserID,
	}

	inv := p.ShareToIncomingInvite(share, "owner@example.com")
	if inv.Home != "/MountedFolder" {
		t.Errorf("Home = %q, want %q", inv.Home, "/MountedFolder")
	}
	if !inv.IsMounted {
		t.Error("IsMounted = false, want true for accepted invite")
	}
	if inv.Name != "folder" {
		t.Errorf("Name = %q, want %q", inv.Name, "folder")
	}
}

func TestPresenter_MountedShareToFolderItem(t *testing.T) {
	p := NewPresenter()
	share := &entity.Share{
		ID:      1,
		OwnerID: 10,
		Home:    vo.NewCloudPath("/Documents"),
	}
	ownerFolder := &entity.Node{
		Name: "Documents",
		Home: vo.NewCloudPath("/Documents"),
		Type: vo.NodeTypeFolder,
		Size: 5000,
		Tree: "tree-abc",
		Rev:  3,
		GRev: 7,
	}
	subCount := &FolderCount{Folders: 2, Files: 10}

	item := p.MountedShareToFolderItem(share, ownerFolder, subCount, "/SharedDocs")

	if item.Name != "SharedDocs" {
		t.Errorf("Name = %q, want %q", item.Name, "SharedDocs")
	}
	if item.Home != "/SharedDocs" {
		t.Errorf("Home = %q, want %q", item.Home, "/SharedDocs")
	}
	if item.Type != "folder" {
		t.Errorf("Type = %q, want %q", item.Type, "folder")
	}
	if item.Kind != "shared" {
		t.Errorf("Kind = %q, want %q", item.Kind, "shared")
	}
	if item.Size != 5000 {
		t.Errorf("Size = %d, want 5000", item.Size)
	}
	if item.Tree != "tree-abc" {
		t.Errorf("Tree = %q, want %q", item.Tree, "tree-abc")
	}
	if item.Count == nil || item.Count.Folders != 2 || item.Count.Files != 10 {
		t.Errorf("Count = %+v, want Folders=2 Files=10", item.Count)
	}
}

func TestPresenter_MountedShareToFolderItem_nilOwner(t *testing.T) {
	p := NewPresenter()
	share := &entity.Share{
		ID:      1,
		OwnerID: 10,
		Home:    vo.NewCloudPath("/Deleted"),
	}

	item := p.MountedShareToFolderItem(share, nil, nil, "/MountedDeleted")

	if item.Name != "MountedDeleted" {
		t.Errorf("Name = %q, want %q", item.Name, "MountedDeleted")
	}
	if item.Home != "/MountedDeleted" {
		t.Errorf("Home = %q, want %q", item.Home, "/MountedDeleted")
	}
	if item.Kind != "shared" {
		t.Errorf("Kind = %q, want %q", item.Kind, "shared")
	}
	if item.Size != 0 {
		t.Errorf("Size = %d, want 0", item.Size)
	}
	if item.Tree != "" {
		t.Errorf("Tree = %q, want empty", item.Tree)
	}
}

func TestPresenter_BuildFolderListing(t *testing.T) {
	p := NewPresenter()
	folder := &entity.Node{
		Name: "docs",
		Home: vo.NewCloudPath("/docs"),
		Type: vo.NodeTypeFolder,
		Tree: "tree-hash",
		Rev:  1,
		GRev: 2,
		Size: 0,
	}
	count := FolderCount{Folders: 2, Files: 5}
	items := []FolderItem{
		{Name: "a.txt", Type: "file"},
		{Name: "sub", Type: "folder"},
	}

	listing := p.BuildFolderListing(folder, count, items)

	if listing.Sort.Order != "asc" {
		t.Errorf("Sort.Order = %q, want %q", listing.Sort.Order, "asc")
	}
	if listing.Sort.Type != "name" {
		t.Errorf("Sort.Type = %q, want %q", listing.Sort.Type, "name")
	}
	if listing.Kind != "folder" {
		t.Errorf("Kind = %q, want %q", listing.Kind, "folder")
	}
	if listing.Type != "folder" {
		t.Errorf("Type = %q, want %q", listing.Type, "folder")
	}
	if listing.Home != "/docs" {
		t.Errorf("Home = %q, want %q", listing.Home, "/docs")
	}
	if len(listing.List) != 2 {
		t.Errorf("List length = %d, want 2", len(listing.List))
	}
	if listing.Count.Folders != 2 || listing.Count.Files != 5 {
		t.Errorf("Count = %+v, want Folders=2 Files=5", listing.Count)
	}
}
