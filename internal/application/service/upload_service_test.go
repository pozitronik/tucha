package service

import (
	"errors"
	"io"
	"testing"

	"tucha/internal/domain/vo"
	"tucha/internal/testutil/mock"
)

func TestUploadService_success(t *testing.T) {
	hash := mock.ValidHash()

	svc := NewUploadService(
		&mock.HasherMock{FixedHash: hash},
		&mock.ContentStorageMock{
			WriteFunc: func(h vo.ContentHash, r io.Reader) (int64, error) {
				return 5, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	got, err := svc.Upload([]byte("hello"))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if got.String() != hash.String() {
		t.Errorf("hash = %q, want %q", got.String(), hash.String())
	}
}

func TestUploadService_storageError(t *testing.T) {
	hash := mock.ValidHash()
	storageErr := errors.New("disk full")

	svc := NewUploadService(
		&mock.HasherMock{FixedHash: hash},
		&mock.ContentStorageMock{
			WriteFunc: func(h vo.ContentHash, r io.Reader) (int64, error) {
				return 0, storageErr
			},
		},
		&mock.ContentRepositoryMock{},
	)

	_, err := svc.Upload([]byte("data"))
	if !errors.Is(err, storageErr) {
		t.Errorf("error = %v, want %v", err, storageErr)
	}
}

func TestUploadService_dbError(t *testing.T) {
	hash := mock.ValidHash()
	dbErr := errors.New("insert failed")

	svc := NewUploadService(
		&mock.HasherMock{FixedHash: hash},
		&mock.ContentStorageMock{
			WriteFunc: func(h vo.ContentHash, r io.Reader) (int64, error) {
				return 5, nil
			},
		},
		&mock.ContentRepositoryMock{
			InsertFunc: func(h vo.ContentHash, size int64) (bool, error) {
				return false, dbErr
			},
		},
	)

	_, err := svc.Upload([]byte("data"))
	if !errors.Is(err, dbErr) {
		t.Errorf("error = %v, want %v", err, dbErr)
	}
}

func TestUploadService_emptyData(t *testing.T) {
	hash := mock.ValidHash()

	svc := NewUploadService(
		&mock.HasherMock{FixedHash: hash},
		&mock.ContentStorageMock{
			WriteFunc: func(h vo.ContentHash, r io.Reader) (int64, error) {
				return 0, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	got, err := svc.Upload([]byte{})
	if err != nil {
		t.Fatalf("Upload(empty): %v", err)
	}
	if got.IsZero() {
		t.Error("Upload(empty) returned zero hash")
	}
}
