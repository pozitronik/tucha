package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestShareService_Share_success(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return folder, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return owner, nil },
		},
	)

	share, err := svc.Share(1, vo.NewCloudPath("/shared"), "invitee@example.com", vo.AccessReadOnly)
	if err != nil {
		t.Fatalf("Share: %v", err)
	}
	if share == nil {
		t.Fatal("Share returned nil")
	}
	if share.InviteToken == "" {
		t.Error("InviteToken is empty")
	}
}

func TestShareService_Share_self(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return folder, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return owner, nil },
		},
	)

	_, err := svc.Share(1, vo.NewCloudPath("/shared"), "owner@example.com", vo.AccessReadOnly)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Share(self) error = %v, want ErrForbidden", err)
	}
}

func TestShareService_Share_folderNotFound(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return nil, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	_, err := svc.Share(1, vo.NewCloudPath("/missing"), "user@example.com", vo.AccessReadOnly)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Share(folder not found) error = %v, want ErrNotFound", err)
	}
}

func TestShareService_Share_idempotent(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")
	existing := mock.NewTestShare(1, "/shared", "user@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return existing, nil
			},
		},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) { return folder, nil },
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return owner, nil },
		},
	)

	share, err := svc.Share(1, vo.NewCloudPath("/shared"), "user@example.com", vo.AccessReadOnly)
	if err != nil {
		t.Fatalf("Share(idempotent): %v", err)
	}
	if share.ID != existing.ID {
		t.Error("should return existing share")
	}
}

func TestShareService_Unshare_success(t *testing.T) {
	existing := mock.NewTestShare(1, "/shared", "user@example.com")
	deleted := false

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return existing, nil
			},
			DeleteFunc: func(id int64) error {
				deleted = true
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Unshare(1, vo.NewCloudPath("/shared"), "user@example.com")
	if err != nil {
		t.Fatalf("Unshare: %v", err)
	}
	if !deleted {
		t.Error("share was not deleted")
	}
}

func TestShareService_Unshare_notFound(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return nil, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Unshare(1, vo.NewCloudPath("/shared"), "user@example.com")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Unshare(not found) error = %v, want ErrNotFound", err)
	}
}

func TestShareService_GetShareInfo(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListByOwnerPathFunc: func(ownerID int64, home vo.CloudPath) ([]entity.Share, error) {
				return []entity.Share{{ID: 1}}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	shares, err := svc.GetShareInfo(1, vo.NewCloudPath("/shared"))
	if err != nil {
		t.Fatalf("GetShareInfo: %v", err)
	}
	if len(shares) != 1 {
		t.Errorf("len = %d, want 1", len(shares))
	}
}

func TestShareService_ListIncoming(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListIncomingFunc: func(email string) ([]entity.Share, error) {
				return []entity.Share{{ID: 1}, {ID: 2}}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	shares, err := svc.ListIncoming("user@example.com")
	if err != nil {
		t.Fatalf("ListIncoming: %v", err)
	}
	if len(shares) != 2 {
		t.Errorf("len = %d, want 2", len(shares))
	}
}

func TestShareService_Mount_success(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	user := mock.NewTestUser(2, "user@example.com")
	accepted := false

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
			AcceptFunc: func(inviteToken string, mountUserID int64, mountHome string) error {
				accepted = true
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	err := svc.Mount(2, "shared-mount", share.InviteToken)
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if !accepted {
		t.Error("share was not accepted")
	}
}

func TestShareService_Mount_notFound(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) { return nil, nil },
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Mount(1, "mount", "unknown-token")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Mount(not found) error = %v, want ErrNotFound", err)
	}
}

func TestShareService_Mount_wrongUser(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	wrongUser := mock.NewTestUser(3, "other@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return wrongUser, nil },
		},
	)

	err := svc.Mount(3, "mount", share.InviteToken)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Mount(wrong user) error = %v, want ErrForbidden", err)
	}
}

func TestShareService_Unmount_success(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			UnmountFunc: func(userID int64, mountHome string) (*entity.Share, error) {
				return share, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Unmount(2, "/mount", false)
	if err != nil {
		t.Fatalf("Unmount: %v", err)
	}
}

func TestShareService_Unmount_withClone(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	share.OwnerID = 1

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			UnmountFunc: func(userID int64, mountHome string) (*entity.Share, error) {
				return share, nil
			},
		},
		&mock.NodeRepositoryMock{
			ListChildrenFunc: func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	err := svc.Unmount(2, "/mount", true)
	if err != nil {
		t.Fatalf("Unmount(clone): %v", err)
	}
}

func TestShareService_Reject_success(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	user := mock.NewTestUser(2, "user@example.com")
	rejected := false

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
			RejectFunc: func(inviteToken string) error {
				rejected = true
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	err := svc.Reject(2, share.InviteToken)
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if !rejected {
		t.Error("share was not rejected")
	}
}

func TestShareService_Reject_wrongUser(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	wrongUser := mock.NewTestUser(3, "other@example.com")

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return wrongUser, nil },
		},
	)

	err := svc.Reject(3, share.InviteToken)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Reject(wrong user) error = %v, want ErrForbidden", err)
	}
}
