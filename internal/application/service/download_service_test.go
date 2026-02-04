package service

import (
	"errors"
	"os"
	"testing"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
	"tucha/internal/testutil/mock"
)

func TestDownloadService_Resolve_success(t *testing.T) {
	hash := mock.ValidHash()
	node := mock.NewTestFileNode(1, "/file.txt", hash, 100)

	// Create a temporary file to return from Open.
	tmp, err := os.CreateTemp(t.TempDir(), "dl")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()

	svc := NewDownloadService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return node, nil
			},
		},
		&mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return tmp, nil
			},
		},
	)

	result, err := svc.Resolve(1, vo.NewCloudPath("/file.txt"))
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if result == nil {
		t.Fatal("Resolve returned nil")
	}
	if result.Node.Name != "file.txt" {
		t.Errorf("Node.Name = %q", result.Node.Name)
	}
}

func TestDownloadService_Resolve_notFound(t *testing.T) {
	svc := NewDownloadService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return nil, nil
			},
		},
		&mock.ContentStorageMock{},
	)

	_, err := svc.Resolve(1, vo.NewCloudPath("/missing"))
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(not found) error = %v, want ErrNotFound", err)
	}
}

func TestDownloadService_Resolve_folder(t *testing.T) {
	folder := mock.NewTestNode(1, "/folder", vo.NodeTypeFolder)

	svc := NewDownloadService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return folder, nil
			},
		},
		&mock.ContentStorageMock{},
	)

	_, err := svc.Resolve(1, vo.NewCloudPath("/folder"))
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(folder) error = %v, want ErrNotFound", err)
	}
}

func TestDownloadService_ResolveByNode_success(t *testing.T) {
	hash := mock.ValidHash()
	node := mock.NewTestFileNode(1, "/file.txt", hash, 100)

	tmp, err := os.CreateTemp(t.TempDir(), "dl")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()

	svc := NewDownloadService(
		&mock.NodeRepositoryMock{},
		&mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return tmp, nil
			},
		},
	)

	result, err := svc.ResolveByNode(node)
	if err != nil {
		t.Fatalf("ResolveByNode: %v", err)
	}
	if result.File == nil {
		t.Error("File is nil")
	}
}

func TestDownloadService_ResolveByNode_folder(t *testing.T) {
	folder := mock.NewTestNode(1, "/folder", vo.NodeTypeFolder)

	svc := NewDownloadService(
		&mock.NodeRepositoryMock{},
		&mock.ContentStorageMock{},
	)

	_, err := svc.ResolveByNode(folder)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ResolveByNode(folder) error = %v, want ErrNotFound", err)
	}
}
