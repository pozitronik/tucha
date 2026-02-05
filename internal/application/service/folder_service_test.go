package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestFolderService_Get_found(t *testing.T) {
	node := mock.NewTestNode(1, "/docs", vo.NodeTypeFolder)

	svc := NewFolderService(&mock.NodeRepositoryMock{
		GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
			return node, nil
		},
	})

	got, err := svc.Get(1, vo.NewCloudPath("/docs"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.Name != "docs" {
		t.Errorf("Name = %q, want %q", got.Name, "docs")
	}
}

func TestFolderService_Get_notFound(t *testing.T) {
	svc := NewFolderService(&mock.NodeRepositoryMock{
		GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
			return nil, nil
		},
	})

	got, err := svc.Get(1, vo.NewCloudPath("/nonexistent"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Error("Get should return nil for nonexistent path")
	}
}

func TestFolderService_ListChildren(t *testing.T) {
	svc := NewFolderService(&mock.NodeRepositoryMock{
		ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
			return []entity.Node{
				{Name: "a.txt", Type: vo.NodeTypeFile},
				{Name: "sub", Type: vo.NodeTypeFolder},
			}, nil
		},
	})

	children, err := svc.ListChildren(1, vo.NewCloudPath("/"), 0, 100)
	if err != nil {
		t.Fatalf("ListChildren: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("len = %d, want 2", len(children))
	}
}

func TestFolderService_CountChildren(t *testing.T) {
	svc := NewFolderService(&mock.NodeRepositoryMock{
		CountChildrenFunc: func(userID int64, path vo.CloudPath) (int, int, error) {
			return 3, 7, nil
		},
	})

	folders, files, err := svc.CountChildren(1, vo.NewCloudPath("/"))
	if err != nil {
		t.Fatalf("CountChildren: %v", err)
	}
	if folders != 3 || files != 7 {
		t.Errorf("CountChildren = (%d, %d), want (3, 7)", folders, files)
	}
}

func TestFolderService_CreateFolder_success(t *testing.T) {
	svc := NewFolderService(&mock.NodeRepositoryMock{
		ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) {
			return false, nil
		},
		EnsurePathFunc: func(userID int64, path vo.CloudPath) error {
			return nil
		},
		CreateFolderFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
			return mock.NewTestNode(userID, path.String(), vo.NodeTypeFolder), nil
		},
	})

	node, err := svc.CreateFolder(1, vo.NewCloudPath("/new-folder"))
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if node == nil {
		t.Fatal("CreateFolder returned nil")
	}
}

func TestFolderService_CreateFolder_alreadyExists(t *testing.T) {
	svc := NewFolderService(&mock.NodeRepositoryMock{
		ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) {
			return true, nil
		},
	})

	_, err := svc.CreateFolder(1, vo.NewCloudPath("/existing"))
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("CreateFolder(existing) error = %v, want ErrAlreadyExists", err)
	}
}
