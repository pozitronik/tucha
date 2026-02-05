package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestPublishService_Publish_new(t *testing.T) {
	node := mock.NewTestNode(1, "/docs", vo.NodeTypeFolder)
	// No weblink yet.

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return node, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	weblink, err := svc.Publish(1, vo.NewCloudPath("/docs"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if weblink == "" {
		t.Error("Publish returned empty weblink")
	}
}

func TestPublishService_Publish_existing(t *testing.T) {
	node := mock.NewTestNode(1, "/docs", vo.NodeTypeFolder)
	node.Weblink = "existing/weblink"

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return node, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	weblink, err := svc.Publish(1, vo.NewCloudPath("/docs"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if weblink != "existing/weblink" {
		t.Errorf("weblink = %q, want %q", weblink, "existing/weblink")
	}
}

func TestPublishService_Publish_notFound(t *testing.T) {
	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	_, err := svc.Publish(1, vo.NewCloudPath("/missing"))
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Publish(not found) error = %v, want ErrNotFound", err)
	}
}

func TestPublishService_Unpublish_success(t *testing.T) {
	node := &entity.Node{
		UserID:  1,
		Home:    vo.NewCloudPath("/docs"),
		Weblink: "abc/def",
	}
	var cleared bool

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return node, nil
			},
			SetWeblinkFunc: func(userID int64, path vo.CloudPath, weblink string) error {
				if weblink == "" {
					cleared = true
				}
				return nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	err := svc.Unpublish(1, "abc/def")
	if err != nil {
		t.Fatalf("Unpublish: %v", err)
	}
	if !cleared {
		t.Error("weblink was not cleared")
	}
}

func TestPublishService_Unpublish_notFound(t *testing.T) {
	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	err := svc.Unpublish(1, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Unpublish(not found) error = %v, want ErrNotFound", err)
	}
}

func TestPublishService_Unpublish_wrongUser(t *testing.T) {
	node := &entity.Node{
		UserID:  2,
		Home:    vo.NewCloudPath("/docs"),
		Weblink: "abc/def",
	}

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return node, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	err := svc.Unpublish(1, "abc/def")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Unpublish(wrong user) error = %v, want ErrForbidden", err)
	}
}

func TestPublishService_ListPublished(t *testing.T) {
	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			ListByWeblinkFunc: func(userID int64) ([]entity.Node, error) {
				return []entity.Node{{Name: "a"}}, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	nodes, err := svc.ListPublished(1)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("len = %d, want 1", len(nodes))
	}
}

func TestPublishService_ResolveWeblink_root(t *testing.T) {
	node := mock.NewTestNode(1, "/docs", vo.NodeTypeFolder)
	node.Weblink = "abc/def"

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return node, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	got, err := svc.ResolveWeblink("abc/def", "")
	if err != nil {
		t.Fatalf("ResolveWeblink: %v", err)
	}
	if got.Name != "docs" {
		t.Errorf("Name = %q, want %q", got.Name, "docs")
	}
}

func TestPublishService_ResolveWeblink_subpath(t *testing.T) {
	parent := mock.NewTestNode(1, "/docs", vo.NodeTypeFolder)
	parent.Weblink = "abc/def"

	child := mock.NewTestFileNode(1, "/docs/readme.txt", mock.ValidHash(), 50)

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return parent, nil
			},
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				if path.String() == "/docs/readme.txt" {
					return child, nil
				}
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
	)

	got, err := svc.ResolveWeblink("abc/def", "readme.txt")
	if err != nil {
		t.Fatalf("ResolveWeblink(subpath): %v", err)
	}
	if got.Name != "readme.txt" {
		t.Errorf("Name = %q, want %q", got.Name, "readme.txt")
	}
}

func TestPublishService_ResolveWeblink_notFound(t *testing.T) {
	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) { return nil, nil },
		},
		&mock.ContentRepositoryMock{},
	)

	_, err := svc.ResolveWeblink("missing", "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ResolveWeblink(not found) error = %v, want ErrNotFound", err)
	}
}

func TestPublishService_Clone_success(t *testing.T) {
	source := mock.NewTestFileNode(2, "/shared/file.txt", mock.ValidHash(), 100)
	source.Weblink = "abc/def"

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) { return source, nil },
			ExistsFunc:       func(userID int64, path vo.CloudPath) (bool, error) { return false, nil },
		},
		&mock.ContentRepositoryMock{},
	)

	node, err := svc.Clone(1, "abc/def", vo.NewCloudPath("/my-files"), vo.ConflictRename)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if node == nil {
		t.Fatal("Clone returned nil")
	}
}

func TestPublishService_Clone_self(t *testing.T) {
	source := mock.NewTestFileNode(1, "/my/file.txt", mock.ValidHash(), 100)
	source.Weblink = "abc/def"

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) { return source, nil },
		},
		&mock.ContentRepositoryMock{},
	)

	_, err := svc.Clone(1, "abc/def", vo.NewCloudPath("/dest"), vo.ConflictRename)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Clone(self) error = %v, want ErrForbidden", err)
	}
}

func TestPublishService_Clone_conflict(t *testing.T) {
	source := mock.NewTestFileNode(2, "/shared/file.txt", mock.ValidHash(), 100)
	source.Weblink = "abc/def"

	svc := NewPublishService(
		&mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) { return source, nil },
			ExistsFunc:       func(userID int64, path vo.CloudPath) (bool, error) { return true, nil },
		},
		&mock.ContentRepositoryMock{},
	)

	_, err := svc.Clone(1, "abc/def", vo.NewCloudPath("/my-files"), vo.ConflictStrict)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("Clone(strict conflict) error = %v, want ErrAlreadyExists", err)
	}
}
