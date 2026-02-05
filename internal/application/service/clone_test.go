package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestCloneTree(t *testing.T) {
	t.Run("clones empty folder", func(t *testing.T) {
		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return []entity.Node{}, nil
			},
		}
		contentRepo := &mock.ContentRepositoryMock{}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != nil {
			t.Errorf("cloneTree() error = %v, want nil", err)
		}
	})

	t.Run("clones folder with files", func(t *testing.T) {
		hash := mock.ValidHash()
		filesCreated := make(map[string]bool)
		contentsInserted := make(map[string]bool)

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				if userID == 1 && path.String() == "/src" {
					return []entity.Node{
						{Name: "file1.txt", Home: vo.NewCloudPath("/src/file1.txt"), Type: vo.NodeTypeFile, Hash: hash, Size: 100},
						{Name: "file2.txt", Home: vo.NewCloudPath("/src/file2.txt"), Type: vo.NodeTypeFile, Hash: hash, Size: 200},
					}, nil
				}
				return []entity.Node{}, nil
			},
			CreateFileFunc: func(userID int64, path vo.CloudPath, h vo.ContentHash, size int64) (*entity.Node, error) {
				filesCreated[path.String()] = true
				return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFile, Hash: h, Size: size}, nil
			},
		}

		contentRepo := &mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				contentsInserted[h.String()] = true
				return true, nil
			},
		}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != nil {
			t.Fatalf("cloneTree() error = %v", err)
		}

		// Verify files were created in destination
		if !filesCreated["/dst/file1.txt"] {
			t.Error("cloneTree() did not create file1.txt")
		}
		if !filesCreated["/dst/file2.txt"] {
			t.Error("cloneTree() did not create file2.txt")
		}

		// Verify content reference was incremented
		if !contentsInserted[hash.String()] {
			t.Error("cloneTree() did not increment content reference")
		}
	})

	t.Run("clones nested folder structure", func(t *testing.T) {
		hash := mock.ValidHash()
		foldersCreated := make(map[string]bool)
		filesCreated := make(map[string]bool)

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				if userID != 1 {
					return []entity.Node{}, nil
				}
				switch path.String() {
				case "/src":
					return []entity.Node{
						{Name: "subdir", Home: vo.NewCloudPath("/src/subdir"), Type: vo.NodeTypeFolder},
						{Name: "root.txt", Home: vo.NewCloudPath("/src/root.txt"), Type: vo.NodeTypeFile, Hash: hash, Size: 50},
					}, nil
				case "/src/subdir":
					return []entity.Node{
						{Name: "nested.txt", Home: vo.NewCloudPath("/src/subdir/nested.txt"), Type: vo.NodeTypeFile, Hash: hash, Size: 100},
					}, nil
				}
				return []entity.Node{}, nil
			},
			CreateFolderFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				foldersCreated[path.String()] = true
				return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFolder}, nil
			},
			CreateFileFunc: func(userID int64, path vo.CloudPath, h vo.ContentHash, size int64) (*entity.Node, error) {
				filesCreated[path.String()] = true
				return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFile}, nil
			},
		}

		contentRepo := &mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				return true, nil
			},
		}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != nil {
			t.Fatalf("cloneTree() error = %v", err)
		}

		// Verify folder structure
		if !foldersCreated["/dst/subdir"] {
			t.Error("cloneTree() did not create subdir folder")
		}

		// Verify files
		if !filesCreated["/dst/root.txt"] {
			t.Error("cloneTree() did not create root.txt")
		}
		if !filesCreated["/dst/subdir/nested.txt"] {
			t.Error("cloneTree() did not create nested.txt")
		}
	})

	t.Run("skips content increment for zero hash", func(t *testing.T) {
		insertCalled := false

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				if path.String() == "/src" {
					return []entity.Node{
						{Name: "empty.txt", Home: vo.NewCloudPath("/src/empty.txt"), Type: vo.NodeTypeFile, Hash: vo.ContentHash{}, Size: 0},
					}, nil
				}
				return []entity.Node{}, nil
			},
			CreateFileFunc: func(userID int64, path vo.CloudPath, h vo.ContentHash, size int64) (*entity.Node, error) {
				return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFile}, nil
			},
		}

		contentRepo := &mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				insertCalled = true
				return true, nil
			},
		}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != nil {
			t.Fatalf("cloneTree() error = %v", err)
		}

		if insertCalled {
			t.Error("cloneTree() should not insert content for zero hash")
		}
	})

	t.Run("returns error when ListChildren fails", func(t *testing.T) {
		expectedErr := errors.New("list error")

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return nil, expectedErr
			},
		}

		err := cloneTree(nodeRepo, &mock.ContentRepositoryMock{}, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != expectedErr {
			t.Errorf("cloneTree() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("returns error when content insert fails", func(t *testing.T) {
		expectedErr := errors.New("insert error")

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return []entity.Node{
					{Name: "file.txt", Home: vo.NewCloudPath("/src/file.txt"), Type: vo.NodeTypeFile, Hash: mock.ValidHash(), Size: 100},
				}, nil
			},
		}

		contentRepo := &mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				return false, expectedErr
			},
		}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != expectedErr {
			t.Errorf("cloneTree() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("returns error when CreateFile fails", func(t *testing.T) {
		expectedErr := errors.New("create file error")

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return []entity.Node{
					{Name: "file.txt", Home: vo.NewCloudPath("/src/file.txt"), Type: vo.NodeTypeFile, Hash: mock.ValidHash(), Size: 100},
				}, nil
			},
			CreateFileFunc: func(userID int64, path vo.CloudPath, h vo.ContentHash, size int64) (*entity.Node, error) {
				return nil, expectedErr
			},
		}

		contentRepo := &mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				return true, nil
			},
		}

		err := cloneTree(nodeRepo, contentRepo, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != expectedErr {
			t.Errorf("cloneTree() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("returns error when CreateFolder fails", func(t *testing.T) {
		expectedErr := errors.New("create folder error")

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return []entity.Node{
					{Name: "subdir", Home: vo.NewCloudPath("/src/subdir"), Type: vo.NodeTypeFolder},
				}, nil
			},
			CreateFolderFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return nil, expectedErr
			},
		}

		err := cloneTree(nodeRepo, &mock.ContentRepositoryMock{}, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != expectedErr {
			t.Errorf("cloneTree() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("returns error when nested clone fails", func(t *testing.T) {
		expectedErr := errors.New("nested error")
		callCount := 0

		nodeRepo := &mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				callCount++
				if callCount == 1 {
					return []entity.Node{
						{Name: "subdir", Home: vo.NewCloudPath("/src/subdir"), Type: vo.NodeTypeFolder},
					}, nil
				}
				// Second call (for nested folder) fails
				return nil, expectedErr
			},
			CreateFolderFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return &entity.Node{ID: 1, UserID: userID, Home: path, Type: vo.NodeTypeFolder}, nil
			},
		}

		err := cloneTree(nodeRepo, &mock.ContentRepositoryMock{}, 1, vo.NewCloudPath("/src"), 2, vo.NewCloudPath("/dst"))
		if err != expectedErr {
			t.Errorf("cloneTree() error = %v, want %v", err, expectedErr)
		}
	})
}
