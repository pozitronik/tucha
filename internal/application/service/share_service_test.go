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

func TestShareService_Share_idempotent_sameAccess(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")
	existing := mock.NewTestShare(1, "/shared", "user@example.com")
	existing.Access = vo.AccessReadOnly

	reinviteCalled := false
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return existing, nil
			},
			ReinviteFunc: func(id int64, access vo.AccessLevel) error {
				reinviteCalled = true
				return nil
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
		t.Fatalf("Share(idempotent same access): %v", err)
	}
	if share.ID != existing.ID {
		t.Error("should return existing share")
	}
	if reinviteCalled {
		t.Error("Reinvite should not be called when access is unchanged")
	}
}

func TestShareService_Share_updateAccess(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")
	existing := mock.NewTestShare(1, "/shared", "user@example.com")
	existing.Access = vo.AccessReadWrite

	var reinvitedAccess vo.AccessLevel
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return existing, nil
			},
			ReinviteFunc: func(id int64, access vo.AccessLevel) error {
				reinvitedAccess = access
				return nil
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
		t.Fatalf("Share(update access): %v", err)
	}
	if share.Access != vo.AccessReadOnly {
		t.Errorf("returned share Access = %q, want %q", share.Access, vo.AccessReadOnly)
	}
	if reinvitedAccess != vo.AccessReadOnly {
		t.Errorf("Reinvite called with access = %q, want %q", reinvitedAccess, vo.AccessReadOnly)
	}
}

func TestShareService_Share_reinviteRejected(t *testing.T) {
	folder := mock.NewTestNode(1, "/shared", vo.NodeTypeFolder)
	owner := mock.NewTestUser(1, "owner@example.com")
	existing := mock.NewTestShare(1, "/shared", "user@example.com")
	existing.Access = vo.AccessReadOnly
	existing.Status = vo.ShareRejected

	reinviteCalled := false
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByOwnerPathEmailFunc: func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
				return existing, nil
			},
			ReinviteFunc: func(id int64, access vo.AccessLevel) error {
				reinviteCalled = true
				return nil
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
		t.Fatalf("Share(reinvite rejected): %v", err)
	}
	if !reinviteCalled {
		t.Error("Reinvite should be called for rejected shares even with same access")
	}
	if share.Status != vo.SharePending {
		t.Errorf("returned share Status = %q, want %q", share.Status, vo.SharePending)
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
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return nil, nil // No existing node at mount point
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	err := svc.Mount(2, "shared-mount", share.InviteToken, vo.ConflictRename)
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

	err := svc.Mount(1, "mount", "unknown-token", vo.ConflictRename)
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

	err := svc.Mount(3, "mount", share.InviteToken, vo.ConflictRename)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Mount(wrong user) error = %v, want ErrForbidden", err)
	}
}

func TestShareService_Mount_conflictStrict(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	user := mock.NewTestUser(2, "user@example.com")
	existingNode := mock.NewTestNode(2, "/mount", vo.NodeTypeFolder)

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
		},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				return existingNode, nil // Mount point already exists
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	err := svc.Mount(2, "mount", share.InviteToken, vo.ConflictStrict)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("Mount(conflict strict) error = %v, want ErrAlreadyExists", err)
	}
}

func TestShareService_Mount_conflictRename(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	user := mock.NewTestUser(2, "user@example.com")
	existingNode := mock.NewTestNode(2, "/mount", vo.NodeTypeFolder)
	acceptedHome := ""

	callCount := 0
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
			AcceptFunc: func(inviteToken string, mountUserID int64, mountHome string) error {
				acceptedHome = mountHome
				return nil
			},
		},
		&mock.NodeRepositoryMock{
			GetFunc: func(userID int64, path vo.CloudPath) (*entity.Node, error) {
				callCount++
				// First call returns existing node, second returns nil (renamed path is free)
				if callCount == 1 {
					return existingNode, nil
				}
				return nil, nil
			},
		},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	err := svc.Mount(2, "mount", share.InviteToken, vo.ConflictRename)
	if err != nil {
		t.Fatalf("Mount(conflict rename): %v", err)
	}
	if acceptedHome != "/mount (1)" {
		t.Errorf("Mount(conflict rename) home = %q, want %q", acceptedHome, "/mount (1)")
	}
}

func TestShareService_Mount_normalizesLeadingSlash(t *testing.T) {
	share := mock.NewTestShare(1, "/shared", "user@example.com")
	user := mock.NewTestUser(2, "user@example.com")
	var acceptedHome string

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			GetByInviteTokenFunc: func(token string) (*entity.Share, error) {
				return share, nil
			},
			AcceptFunc: func(inviteToken string, mountUserID int64, mountHome string) error {
				acceptedHome = mountHome
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return user, nil },
		},
	)

	// Client sends mount name WITH leading slash -- must not produce "//shared".
	err := svc.Mount(2, "/shared", share.InviteToken, vo.ConflictRename)
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if acceptedHome != "/shared" {
		t.Errorf("mountHome = %q, want %q (no double slash)", acceptedHome, "/shared")
	}
}

func TestShareService_ResolveMount_doubleSlashMountHome(t *testing.T) {
	uid := int64(2)
	// Simulate a share stored with malformed MountHome (double leading slash).
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "//SharedDocs",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	// Path uses normalized single-slash form.
	res, err := svc.ResolveMount(uid, vo.NewCloudPath("/SharedDocs/file.txt"))
	if err != nil {
		t.Fatalf("ResolveMount: %v", err)
	}
	if res == nil {
		t.Fatal("ResolveMount returned nil for double-slash MountHome")
	}
	if res.OwnerPath.String() != "/Documents/file.txt" {
		t.Errorf("OwnerPath = %q, want %q", res.OwnerPath.String(), "/Documents/file.txt")
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

func TestShareService_ListMountedIn_rootLevel(t *testing.T) {
	uid := int64(2)
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "/SharedDocs",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	shares, err := svc.ListMountedIn(uid, vo.NewCloudPath("/"))
	if err != nil {
		t.Fatalf("ListMountedIn: %v", err)
	}
	if len(shares) != 1 {
		t.Fatalf("len = %d, want 1", len(shares))
	}
	if shares[0].MountHome != "/SharedDocs" {
		t.Errorf("MountHome = %q, want %q", shares[0].MountHome, "/SharedDocs")
	}
}

func TestShareService_ListMountedIn_nested(t *testing.T) {
	uid := int64(2)
	// Mount at /a/b -- parent is /a, not /
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "/a/b",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	// Should NOT appear in root listing.
	shares, err := svc.ListMountedIn(uid, vo.NewCloudPath("/"))
	if err != nil {
		t.Fatalf("ListMountedIn: %v", err)
	}
	if len(shares) != 0 {
		t.Errorf("root: len = %d, want 0", len(shares))
	}

	// Should appear in /a listing.
	shares, err = svc.ListMountedIn(uid, vo.NewCloudPath("/a"))
	if err != nil {
		t.Fatalf("ListMountedIn: %v", err)
	}
	if len(shares) != 1 {
		t.Errorf("/a: len = %d, want 1", len(shares))
	}
}

func TestShareService_ListMountedIn_empty(t *testing.T) {
	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return nil, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	shares, err := svc.ListMountedIn(1, vo.NewCloudPath("/"))
	if err != nil {
		t.Fatalf("ListMountedIn: %v", err)
	}
	if len(shares) != 0 {
		t.Errorf("len = %d, want 0", len(shares))
	}
}

func TestShareService_ResolveMount_exactMatch(t *testing.T) {
	uid := int64(2)
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "/SharedDocs",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	res, err := svc.ResolveMount(uid, vo.NewCloudPath("/SharedDocs"))
	if err != nil {
		t.Fatalf("ResolveMount: %v", err)
	}
	if res == nil {
		t.Fatal("ResolveMount returned nil")
	}
	if res.OwnerPath.String() != "/Documents" {
		t.Errorf("OwnerPath = %q, want %q", res.OwnerPath.String(), "/Documents")
	}
	if res.Share.ID != 1 {
		t.Errorf("Share.ID = %d, want 1", res.Share.ID)
	}
}

func TestShareService_ResolveMount_subpath(t *testing.T) {
	uid := int64(2)
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "/SharedDocs",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	res, err := svc.ResolveMount(uid, vo.NewCloudPath("/SharedDocs/sub/file.txt"))
	if err != nil {
		t.Fatalf("ResolveMount: %v", err)
	}
	if res == nil {
		t.Fatal("ResolveMount returned nil")
	}
	if res.OwnerPath.String() != "/Documents/sub/file.txt" {
		t.Errorf("OwnerPath = %q, want %q", res.OwnerPath.String(), "/Documents/sub/file.txt")
	}
}

func TestShareService_ResolveMount_noMatch(t *testing.T) {
	uid := int64(2)
	mountedShare := entity.Share{
		ID:        1,
		OwnerID:   1,
		Home:      vo.NewCloudPath("/Documents"),
		MountHome: "/SharedDocs",
		Status:    vo.ShareAccepted,
	}
	mountedShare.MountUserID = &uid

	svc := NewShareService(
		&mock.ShareRepositoryMock{
			ListMountedByUserFunc: func(userID int64) ([]entity.Share, error) {
				return []entity.Share{mountedShare}, nil
			},
		},
		&mock.NodeRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.UserRepositoryMock{},
	)

	res, err := svc.ResolveMount(uid, vo.NewCloudPath("/OtherFolder"))
	if err != nil {
		t.Fatalf("ResolveMount: %v", err)
	}
	if res != nil {
		t.Errorf("ResolveMount returned non-nil for non-matching path: %+v", res)
	}
}
