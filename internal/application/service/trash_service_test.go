package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestTrashService_Trash_file(t *testing.T) {
	node := mock.NewTestFileNode(1, "/file.txt", mock.ValidHash(), 100)
	var insertCalled bool

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return node, nil, nil
			},
		},
		&mock.TrashRepositoryMock{
			InsertFunc: func(userID int64, n *entity.Node, descendants []entity.Node, deletedBy int64) error {
				insertCalled = true
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Trash(1, vo.NewCloudPath("/file.txt"), 1)
	if err != nil {
		t.Fatalf("Trash: %v", err)
	}
	if !insertCalled {
		t.Error("trash Insert was not called")
	}
}

func TestTrashService_Trash_folderWithDescendants(t *testing.T) {
	folder := mock.NewTestNode(1, "/folder", vo.NodeTypeFolder)
	children := []entity.Node{
		*mock.NewTestFileNode(1, "/folder/a.txt", mock.ValidHash(), 50),
	}

	var capturedDescendants []entity.Node
	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return folder, children, nil
			},
		},
		&mock.TrashRepositoryMock{
			InsertFunc: func(userID int64, n *entity.Node, descendants []entity.Node, deletedBy int64) error {
				capturedDescendants = descendants
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Trash(1, vo.NewCloudPath("/folder"), 1)
	if err != nil {
		t.Fatalf("Trash: %v", err)
	}
	if len(capturedDescendants) != 1 {
		t.Errorf("descendants = %d, want 1", len(capturedDescendants))
	}
}

func TestTrashService_Trash_nonexistent(t *testing.T) {
	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return nil, nil, nil
			},
		},
		&mock.TrashRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Trash(1, vo.NewCloudPath("/nonexistent"), 1)
	if err != nil {
		t.Fatalf("Trash(nonexistent) should be no-op, got: %v", err)
	}
}

func TestTrashService_List(t *testing.T) {
	items := []entity.TrashItem{
		*mock.NewTestTrashItem(1, "/deleted.txt", vo.NodeTypeFile),
	}

	svc := NewTrashService(
		&mock.NodeRepositoryMock{},
		&mock.TrashRepositoryMock{
			ListFunc: func(userID int64) ([]entity.TrashItem, error) { return items, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	got, err := svc.List(1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestTrashService_Restore_success(t *testing.T) {
	item := mock.NewTestTrashItem(1, "/file.txt", vo.NodeTypeFile)
	item.Hash = mock.ValidHash()
	item.Size = 100

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) { return false, nil },
		},
		&mock.TrashRepositoryMock{
			GetByPathAndRevFunc: func(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
				return item, nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Restore(1, vo.NewCloudPath("/file.txt"), 0, vo.ConflictRename)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
}

func TestTrashService_Restore_notFound(t *testing.T) {
	svc := NewTrashService(
		&mock.NodeRepositoryMock{},
		&mock.TrashRepositoryMock{
			GetByPathAndRevFunc: func(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Restore(1, vo.NewCloudPath("/missing"), 0, vo.ConflictRename)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Restore(not found) error = %v, want ErrNotFound", err)
	}
}

func TestTrashService_Restore_conflictStrict(t *testing.T) {
	item := mock.NewTestTrashItem(1, "/file.txt", vo.NodeTypeFile)

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) { return true, nil },
		},
		&mock.TrashRepositoryMock{
			GetByPathAndRevFunc: func(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
				return item, nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Restore(1, vo.NewCloudPath("/file.txt"), 0, vo.ConflictStrict)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("Restore(strict conflict) error = %v, want ErrAlreadyExists", err)
	}
}

func TestTrashService_Empty_contentCleanup(t *testing.T) {
	hash := mock.ValidHash()
	items := []entity.TrashItem{
		{ID: 1, Type: vo.NodeTypeFile, Hash: hash, Size: 100},
		{ID: 2, Type: vo.NodeTypeFolder},
	}

	var decremented, diskDeleted bool
	svc := NewTrashService(
		&mock.NodeRepositoryMock{},
		&mock.TrashRepositoryMock{
			DeleteAllFunc: func(userID int64) ([]entity.TrashItem, error) { return items, nil },
		},
		&mock.ContentRepositoryMock{
			DecrementFunc: func(h vo.ContentHash) (bool, error) {
				decremented = true
				return true, nil
			},
		},
		&mock.ContentStorageMock{
			DeleteFunc: func(h vo.ContentHash) error {
				diskDeleted = true
				return nil
			},
		},
		&mock.ShareRepositoryMock{},
	)

	err := svc.Empty(1)
	if err != nil {
		t.Fatalf("Empty: %v", err)
	}
	if !decremented {
		t.Error("content ref count was not decremented")
	}
	if !diskDeleted {
		t.Error("content was not deleted from disk")
	}
}

func TestTrashService_Trash_deletesShareRecords(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	pendingShare := entity.Share{ID: 10, OwnerID: 1, Home: vo.NewCloudPath("/shared"), Access: vo.AccessReadOnly, Status: vo.SharePending}
	mountUserID := int64(2)
	mountedRWShare := entity.Share{ID: 20, OwnerID: 1, Home: vo.NewCloudPath("/shared"), Access: vo.AccessReadWrite, Status: vo.ShareAccepted, MountHome: "/Mounted", MountUserID: &mountUserID}

	var deletedIDs []int64
	var ensurePathCalled bool

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return folder, nil, nil
			},
			EnsurePathFunc: func(userID int64, path vo.CloudPath) error {
				ensurePathCalled = true
				return nil
			},
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return nil, nil
			},
		},
		&mock.TrashRepositoryMock{
			InsertFunc: func(userID int64, n *entity.Node, descendants []entity.Node, deletedBy int64) error {
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{
			ListByOwnerPathPrefixFunc: func(ownerID int64, path vo.CloudPath) ([]entity.Share, error) {
				return []entity.Share{pendingShare, mountedRWShare}, nil
			},
			DeleteFunc: func(id int64) error {
				deletedIDs = append(deletedIDs, id)
				return nil
			},
		},
	)

	err := svc.Trash(1, vo.NewCloudPath("/shared"), 1)
	if err != nil {
		t.Fatalf("Trash: %v", err)
	}

	// Both shares should be deleted.
	if len(deletedIDs) != 2 {
		t.Fatalf("deleted %d shares, want 2", len(deletedIDs))
	}
	if deletedIDs[0] != 10 || deletedIDs[1] != 20 {
		t.Errorf("deleted IDs = %v, want [10, 20]", deletedIDs)
	}

	// Mounted RW share should trigger clone (EnsurePath called for mount user).
	if !ensurePathCalled {
		t.Error("EnsurePath was not called for RW mount clone")
	}
}

func TestTrashService_Trash_noSharesFound(t *testing.T) {
	folder := mock.NewTestNode(1, "/noshares", vo.NodeTypeFolder)

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return folder, nil, nil
			},
		},
		&mock.TrashRepositoryMock{
			InsertFunc: func(userID int64, n *entity.Node, descendants []entity.Node, deletedBy int64) error {
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{
			ListByOwnerPathPrefixFunc: func(ownerID int64, path vo.CloudPath) ([]entity.Share, error) {
				return nil, nil
			},
		},
	)

	err := svc.Trash(1, vo.NewCloudPath("/noshares"), 1)
	if err != nil {
		t.Fatalf("Trash: %v", err)
	}
}

func TestTrashService_Trash_shareListError(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)

	svc := NewTrashService(
		&mock.NodeRepositoryMock{
			GetWithDescendantsFunc: func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
				return folder, nil, nil
			},
		},
		&mock.TrashRepositoryMock{
			InsertFunc: func(userID int64, n *entity.Node, descendants []entity.Node, deletedBy int64) error {
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.ShareRepositoryMock{
			ListByOwnerPathPrefixFunc: func(ownerID int64, path vo.CloudPath) ([]entity.Share, error) {
				return nil, errors.New("database error")
			},
		},
	)

	// Trash should still succeed even if share lookup fails (best-effort cleanup).
	err := svc.Trash(1, vo.NewCloudPath("/shared"), 1)
	if err != nil {
		t.Fatalf("Trash should succeed despite share list error: %v", err)
	}
}
