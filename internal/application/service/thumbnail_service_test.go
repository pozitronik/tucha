package service

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/infrastructure/thumbnail"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

// createTestImageFile creates a test JPEG image and stores it for the mock.
// Returns the hash and sets up the mock to serve the file.
func createTestImageFile(t *testing.T) (vo.ContentHash, string) {
	t.Helper()

	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 150))
	for y := 0; y < 150; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / 200),
				G: uint8(y * 255 / 150),
				B: 100,
				A: 255,
			})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("Creating test image: %v", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Encoding JPEG: %v", err)
	}

	hash := mock.ValidHash()
	return hash, imgPath
}

func TestThumbnailService_Generate(t *testing.T) {
	t.Run("generates thumbnail for valid image file", func(t *testing.T) {
		tmpDir := t.TempDir()
		hash, imgPath := createTestImageFile(t)

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				if userID == 1 && path.String() == "/photos/image.jpg" {
					return mock.NewTestFileNode(1, "/photos/image.jpg", hash, 1024), nil
				}
				return nil, nil
			},
		}

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				if h == hash {
					return os.Open(imgPath)
				}
				return nil, os.ErrNotExist
			},
		}

		gen, err := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		svc := NewThumbnailService(nodeRepo, storageMock, gen)

		result, err := svc.Generate(1, vo.NewCloudPath("/photos/image.jpg"), "xw14")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if result == nil {
			t.Fatal("Generate() returned nil result")
		}
		if len(result.Data) == 0 {
			t.Error("Generate() returned empty data")
		}
		if result.ContentType != "image/jpeg" {
			t.Errorf("Generate() ContentType = %q, want %q", result.ContentType, "image/jpeg")
		}
	})

	t.Run("returns ErrNotFound for nonexistent file", func(t *testing.T) {
		tmpDir := t.TempDir()

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return nil, nil // File not found
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(nodeRepo, &mock.ContentStorageMock{}, gen)

		_, err := svc.Generate(1, vo.NewCloudPath("/nonexistent.jpg"), "xw14")
		if err != ErrNotFound {
			t.Errorf("Generate() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("returns ErrNotFound for folder", func(t *testing.T) {
		tmpDir := t.TempDir()

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return mock.NewTestNode(1, "/photos", vo.NodeTypeFolder), nil
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(nodeRepo, &mock.ContentStorageMock{}, gen)

		_, err := svc.Generate(1, vo.NewCloudPath("/photos"), "xw14")
		if err != ErrNotFound {
			t.Errorf("Generate() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("returns ErrNotFound for unsupported format", func(t *testing.T) {
		tmpDir := t.TempDir()

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return mock.NewTestFileNode(1, "/docs/file.pdf", mock.ValidHash(), 1024), nil
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(nodeRepo, &mock.ContentStorageMock{}, gen)

		_, err := svc.Generate(1, vo.NewCloudPath("/docs/file.pdf"), "xw14")
		if err != ErrNotFound {
			t.Errorf("Generate() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("uses default preset for unknown preset name", func(t *testing.T) {
		tmpDir := t.TempDir()
		hash, imgPath := createTestImageFile(t)

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return mock.NewTestFileNode(1, "/image.jpg", hash, 1024), nil
			},
		}

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return os.Open(imgPath)
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(nodeRepo, storageMock, gen)

		// Unknown preset should fall back to xw14
		result, err := svc.Generate(1, vo.NewCloudPath("/image.jpg"), "unknown_preset")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if result == nil || len(result.Data) == 0 {
			t.Error("Generate() should succeed with default preset")
		}
	})

	t.Run("returns error when storage fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		nodeRepo := &mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return mock.NewTestFileNode(1, "/image.jpg", mock.ValidHash(), 1024), nil
			},
		}

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return nil, errors.New("storage error")
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(nodeRepo, storageMock, gen)

		_, err := svc.Generate(1, vo.NewCloudPath("/image.jpg"), "xw14")
		if err == nil {
			t.Error("Generate() expected error for storage failure, got nil")
		}
	})
}

func TestThumbnailService_GenerateByHash(t *testing.T) {
	t.Run("generates thumbnail by hash", func(t *testing.T) {
		tmpDir := t.TempDir()
		hash, imgPath := createTestImageFile(t)

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				if h == hash {
					return os.Open(imgPath)
				}
				return nil, os.ErrNotExist
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(&mock.NodeRepositoryMock{}, storageMock, gen)

		result, err := svc.GenerateByHash(hash, "photo.jpg", "xw14")
		if err != nil {
			t.Fatalf("GenerateByHash() error = %v", err)
		}

		if result == nil {
			t.Fatal("GenerateByHash() returned nil result")
		}
		if len(result.Data) == 0 {
			t.Error("GenerateByHash() returned empty data")
		}
	})

	t.Run("returns ErrNotFound for unsupported format", func(t *testing.T) {
		tmpDir := t.TempDir()

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(&mock.NodeRepositoryMock{}, &mock.ContentStorageMock{}, gen)

		_, err := svc.GenerateByHash(mock.ValidHash(), "document.pdf", "xw14")
		if err != ErrNotFound {
			t.Errorf("GenerateByHash() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("uses default preset for unknown preset", func(t *testing.T) {
		tmpDir := t.TempDir()
		hash, imgPath := createTestImageFile(t)

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return os.Open(imgPath)
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(&mock.NodeRepositoryMock{}, storageMock, gen)

		result, err := svc.GenerateByHash(hash, "image.png", "nonexistent")
		if err != nil {
			t.Fatalf("GenerateByHash() error = %v", err)
		}
		if result == nil || len(result.Data) == 0 {
			t.Error("GenerateByHash() should succeed with default preset")
		}
	})

	t.Run("returns error when storage fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		storageMock := &mock.ContentStorageMock{
			OpenFunc: func(h vo.ContentHash) (*os.File, error) {
				return nil, os.ErrNotExist
			},
		}

		gen, _ := thumbnail.NewGenerator(filepath.Join(tmpDir, "thumbnails"))
		svc := NewThumbnailService(&mock.NodeRepositoryMock{}, storageMock, gen)

		_, err := svc.GenerateByHash(mock.ValidHash(), "image.jpg", "xw14")
		if err == nil {
			t.Error("GenerateByHash() expected error for missing content, got nil")
		}
	})
}
