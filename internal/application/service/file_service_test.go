package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func newFileServiceWithDefaults(
	nodes *mock.NodeRepositoryMock,
	contents *mock.ContentRepositoryMock,
	storage *mock.ContentStorageMock,
	users *mock.UserRepositoryMock,
) *FileService {
	quota := NewQuotaService(nodes, users)
	return NewFileService(nodes, contents, storage, quota)
}

func TestFileService_AddByHash_success(t *testing.T) {
	hash := mock.ValidHash()

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) { return false, nil },
		},
		&mock.ContentRepositoryMock{
			ExistsFunc: func(h vo.ContentHash) (bool, error) { return true, nil },
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1073741824}, nil
			},
		},
	)

	node, err := svc.AddByHash(1, vo.NewCloudPath("/file.txt"), hash, 100, vo.ConflictRename)
	if err != nil {
		t.Fatalf("AddByHash: %v", err)
	}
	if node == nil {
		t.Fatal("AddByHash returned nil")
	}
}

func TestFileService_AddByHash_contentNotFound(t *testing.T) {
	hash := mock.ValidHash()

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{
			ExistsFunc: func(h vo.ContentHash) (bool, error) { return false, nil },
		},
		&mock.ContentStorageMock{
			ExistsFunc: func(h vo.ContentHash) bool { return false },
		},
		&mock.UserRepositoryMock{},
	)

	_, err := svc.AddByHash(1, vo.NewCloudPath("/file.txt"), hash, 100, vo.ConflictRename)
	if !errors.Is(err, ErrContentNotFound) {
		t.Errorf("AddByHash(no content) error = %v, want ErrContentNotFound", err)
	}
}

func TestFileService_AddByHash_overQuota(t *testing.T) {
	hash := mock.ValidHash()

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) { return 900, nil },
		},
		&mock.ContentRepositoryMock{
			ExistsFunc: func(h vo.ContentHash) (bool, error) { return true, nil },
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1000}, nil
			},
		},
	)

	_, err := svc.AddByHash(1, vo.NewCloudPath("/big.bin"), hash, 200, vo.ConflictRename)
	if !errors.Is(err, ErrOverQuota) {
		t.Errorf("AddByHash(over quota) error = %v, want ErrOverQuota", err)
	}
}

func TestFileService_AddByHash_conflictStrict(t *testing.T) {
	hash := mock.ValidHash()

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			ExistsFunc:    func(userID int64, path vo.CloudPath) (bool, error) { return true, nil },
			TotalSizeFunc: func(userID int64) (int64, error) { return 0, nil },
		},
		&mock.ContentRepositoryMock{
			ExistsFunc: func(h vo.ContentHash) (bool, error) { return true, nil },
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1073741824}, nil
			},
		},
	)

	_, err := svc.AddByHash(1, vo.NewCloudPath("/existing.txt"), hash, 100, vo.ConflictStrict)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("AddByHash(strict conflict) error = %v, want ErrAlreadyExists", err)
	}
}

func TestFileService_AddByHash_conflictReplace(t *testing.T) {
	hash := mock.ValidHash()
	deleted := false

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			ExistsFunc: func(userID int64, path vo.CloudPath) (bool, error) { return true, nil },
			DeleteFunc: func(userID int64, path vo.CloudPath) error {
				deleted = true
				return nil
			},
			TotalSizeFunc: func(userID int64) (int64, error) { return 0, nil },
		},
		&mock.ContentRepositoryMock{
			ExistsFunc: func(h vo.ContentHash) (bool, error) { return true, nil },
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1073741824}, nil
			},
		},
	)

	_, err := svc.AddByHash(1, vo.NewCloudPath("/existing.txt"), hash, 100, vo.ConflictReplace)
	if err != nil {
		t.Fatalf("AddByHash(replace): %v", err)
	}
	if !deleted {
		t.Error("existing node was not deleted before replacement")
	}
}

func TestFileService_Remove_fileWithContent(t *testing.T) {
	hash := mock.ValidHash()
	node := mock.NewTestFileNode(1, "/file.txt", hash, 100)
	var decremented, diskDeleted bool

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return node, nil },
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
		&mock.UserRepositoryMock{},
	)

	err := svc.Remove(1, vo.NewCloudPath("/file.txt"))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !decremented {
		t.Error("content ref count was not decremented")
	}
	if !diskDeleted {
		t.Error("content was not deleted from disk")
	}
}

func TestFileService_Remove_fileWithoutContent(t *testing.T) {
	node := mock.NewTestNode(1, "/file.txt", vo.NodeTypeFile)
	// No hash set, so HasContent() = false.

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return node, nil },
		},
		&mock.ContentRepositoryMock{
			DecrementFunc: func(h vo.ContentHash) (bool, error) {
				t.Error("Decrement should not be called for file without content")
				return false, nil
			},
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Remove(1, vo.NewCloudPath("/file.txt"))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestFileService_Remove_nonexistent(t *testing.T) {
	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return nil, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Remove(1, vo.NewCloudPath("/nonexistent"))
	if err != nil {
		t.Fatalf("Remove(nonexistent) should succeed, got: %v", err)
	}
}

func TestFileService_Remove_folder(t *testing.T) {
	folder := mock.NewTestNode(1, "/folder", vo.NodeTypeFolder)

	svc := newFileServiceWithDefaults(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return folder, nil },
		},
		&mock.ContentRepositoryMock{
			DecrementFunc: func(h vo.ContentHash) (bool, error) {
				t.Error("Decrement should not be called for folder")
				return false, nil
			},
		},
		&mock.ContentStorageMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Remove(1, vo.NewCloudPath("/folder"))
	if err != nil {
		t.Fatalf("Remove(folder): %v", err)
	}
}
