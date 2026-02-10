package sqlite

import (
	"testing"

	"github.com/pozitronik/tucha/internal/application/port"
	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/infrastructure/contentstore"
)

// TestShareLifecycle_TrashDeletesShares exercises the full flow:
// owner creates folder, shares it, then trashes the folder.
// After trash, the share record must be gone.
func TestShareLifecycle_TrashDeletesShares(t *testing.T) {
	db := openTestDB(t)
	userRepo := NewUserRepository(db)
	nodeRepo := NewNodeRepository(db)
	contentRepo := NewContentRepository(db)
	trashRepo := NewTrashRepository(db)
	shareRepo := NewShareRepository(db)

	// We need a ContentStorage (disk store) for TrashService.
	store, err := contentstore.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatalf("creating disk store: %v", err)
	}

	// Create owner and invited user.
	ownerID, err := userRepo.Create(&entity.User{Email: "owner@test.com", Password: "pass"})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	invitedID, err := userRepo.Create(&entity.User{Email: "invited@test.com", Password: "pass"})
	if err != nil {
		t.Fatalf("create invited: %v", err)
	}

	// Create root and folder for owner.
	if _, err := nodeRepo.CreateRootNode(ownerID); err != nil {
		t.Fatalf("create root: %v", err)
	}
	folderPath := vo.NewCloudPath("/test_dir")
	if _, err := nodeRepo.CreateFolder(ownerID, folderPath); err != nil {
		t.Fatalf("create folder: %v", err)
	}

	// Create share.
	shareObj := &entity.Share{
		OwnerID:      ownerID,
		Home:         folderPath,
		InvitedEmail: "invited@test.com",
		Access:       vo.AccessReadOnly,
		Status:       vo.SharePending,
		InviteToken:  "test-token-123",
	}
	shareID, err := shareRepo.Create(shareObj)
	if err != nil {
		t.Fatalf("create share: %v", err)
	}

	// Verify share exists.
	shares, err := shareRepo.ListByOwnerPath(ownerID, folderPath)
	if err != nil {
		t.Fatalf("list shares: %v", err)
	}
	if len(shares) != 1 {
		t.Fatalf("expected 1 share, got %d", len(shares))
	}
	if shares[0].ID != shareID {
		t.Fatalf("share ID = %d, want %d", shares[0].ID, shareID)
	}

	// Verify ListByOwnerPathPrefix finds it.
	prefixShares, err := shareRepo.ListByOwnerPathPrefix(ownerID, folderPath)
	if err != nil {
		t.Fatalf("ListByOwnerPathPrefix: %v", err)
	}
	if len(prefixShares) != 1 {
		t.Fatalf("ListByOwnerPathPrefix returned %d shares, want 1", len(prefixShares))
	}

	// Now trash the folder via TrashService.
	trashSvc := service.NewTrashService(nodeRepo, trashRepo, contentRepo, store, shareRepo)
	if err := trashSvc.Trash(ownerID, folderPath, ownerID); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	// Verify the share was deleted.
	afterShares, err := shareRepo.ListByOwnerPath(ownerID, folderPath)
	if err != nil {
		t.Fatalf("list shares after trash: %v", err)
	}
	if len(afterShares) != 0 {
		t.Errorf("expected 0 shares after trash, got %d", len(afterShares))
	}

	// Verify incoming invite is gone too.
	incoming, err := shareRepo.ListIncoming("invited@test.com")
	if err != nil {
		t.Fatalf("ListIncoming: %v", err)
	}
	if len(incoming) != 0 {
		t.Errorf("expected 0 incoming after trash, got %d", len(incoming))
	}

	// Verify owner can create a new folder with the same name -- no stale shares.
	if _, err := nodeRepo.CreateFolder(ownerID, folderPath); err != nil {
		t.Fatalf("recreate folder: %v", err)
	}
	freshShares, err := shareRepo.ListByOwnerPath(ownerID, folderPath)
	if err != nil {
		t.Fatalf("list shares on recreated folder: %v", err)
	}
	if len(freshShares) != 0 {
		t.Errorf("recreated folder has %d stale shares, want 0", len(freshShares))
	}

	_ = invitedID // used only for user creation FK
}

// TestShareLifecycle_TrashClonesRWMount exercises: owner trashes a folder
// that has an accepted RW mount. The mount user should get a cloned copy.
func TestShareLifecycle_TrashClonesRWMount(t *testing.T) {
	db := openTestDB(t)
	userRepo := NewUserRepository(db)
	nodeRepo := NewNodeRepository(db)
	contentRepo := NewContentRepository(db)
	trashRepo := NewTrashRepository(db)
	shareRepo := NewShareRepository(db)

	store, err := contentstore.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatalf("creating disk store: %v", err)
	}

	ownerID, _ := userRepo.Create(&entity.User{Email: "owner@test.com", Password: "pass"})
	mountID, _ := userRepo.Create(&entity.User{Email: "mount@test.com", Password: "pass"})

	// Owner's tree.
	if _, err := nodeRepo.CreateRootNode(ownerID); err != nil {
		t.Fatalf("create owner root: %v", err)
	}
	if _, err := nodeRepo.CreateFolder(ownerID, vo.NewCloudPath("/shared")); err != nil {
		t.Fatalf("create owner folder: %v", err)
	}

	// Mount user's tree.
	if _, err := nodeRepo.CreateRootNode(mountID); err != nil {
		t.Fatalf("create mount root: %v", err)
	}

	// Create share and accept it (Accept stores mount details; Create does not).
	shareObj := &entity.Share{
		OwnerID:      ownerID,
		Home:         vo.NewCloudPath("/shared"),
		InvitedEmail: "mount@test.com",
		Access:       vo.AccessReadWrite,
		Status:       vo.SharePending,
		InviteToken:  "rw-token",
	}
	if _, err := shareRepo.Create(shareObj); err != nil {
		t.Fatalf("create share: %v", err)
	}
	if err := shareRepo.Accept("rw-token", mountID, "/MountedShared"); err != nil {
		t.Fatalf("accept share: %v", err)
	}

	// Trash owner's folder.
	trashSvc := service.NewTrashService(nodeRepo, trashRepo, contentRepo, store, shareRepo)
	if err := trashSvc.Trash(ownerID, vo.NewCloudPath("/shared"), ownerID); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	// Share should be deleted.
	remaining, _ := shareRepo.ListByOwnerPath(ownerID, vo.NewCloudPath("/shared"))
	if len(remaining) != 0 {
		t.Errorf("expected 0 shares, got %d", len(remaining))
	}

	// Mount user should have a cloned folder at the mount path.
	mountFolder, err := nodeRepo.Get(mountID, vo.NewCloudPath("/MountedShared"))
	if err != nil {
		t.Fatalf("Get mount folder: %v", err)
	}
	if mountFolder == nil {
		t.Error("expected cloned folder at /MountedShared for mount user, got nil")
	}
}

// TestShareLifecycle_ResolveMountAfterAccept verifies that ResolveMount
// finds the mount and correctly resolves subpaths after Accept.
func TestShareLifecycle_ResolveMountAfterAccept(t *testing.T) {
	db := openTestDB(t)
	userRepo := NewUserRepository(db)
	nodeRepo := NewNodeRepository(db)
	contentRepo := NewContentRepository(db)
	shareRepo := NewShareRepository(db)

	ownerID, _ := userRepo.Create(&entity.User{Email: "owner@test.com", Password: "pass"})
	mountID, _ := userRepo.Create(&entity.User{Email: "mount@test.com", Password: "pass"})

	if _, err := nodeRepo.CreateRootNode(ownerID); err != nil {
		t.Fatalf("create owner root: %v", err)
	}
	if _, err := nodeRepo.CreateFolder(ownerID, vo.NewCloudPath("/shared_dir")); err != nil {
		t.Fatalf("create owner folder: %v", err)
	}
	if _, err := nodeRepo.CreateRootNode(mountID); err != nil {
		t.Fatalf("create mount root: %v", err)
	}

	// Create and accept share.
	shareObj := &entity.Share{
		OwnerID:      ownerID,
		Home:         vo.NewCloudPath("/shared_dir"),
		InvitedEmail: "mount@test.com",
		Access:       vo.AccessReadOnly,
		Status:       vo.SharePending,
		InviteToken:  "resolve-token",
	}
	if _, err := shareRepo.Create(shareObj); err != nil {
		t.Fatalf("create share: %v", err)
	}

	// Use ShareService.Mount (not raw Accept) to test the full flow.
	shareSvc := service.NewShareService(shareRepo, nodeRepo, contentRepo, userRepo)
	if err := shareSvc.Mount(mountID, "shared_dir", "resolve-token", vo.ConflictRename); err != nil {
		t.Fatalf("Mount: %v", err)
	}

	// Verify the share record is accepted.
	mounted, err := shareRepo.ListMountedByUser(mountID)
	if err != nil {
		t.Fatalf("ListMountedByUser: %v", err)
	}
	if len(mounted) != 1 {
		t.Fatalf("expected 1 mounted share, got %d", len(mounted))
	}
	t.Logf("MountHome = %q, MountUserID = %v, Status = %q, Access = %q",
		mounted[0].MountHome, mounted[0].MountUserID, mounted[0].Status, mounted[0].Access)

	// Now test ResolveMount for a file path under the mount.
	resolution, err := shareSvc.ResolveMount(mountID, vo.NewCloudPath("/shared_dir/file.txt"))
	if err != nil {
		t.Fatalf("ResolveMount error: %v", err)
	}
	if resolution == nil {
		t.Fatal("ResolveMount returned nil -- mount not found")
	}
	if resolution.Share.Access != vo.AccessReadOnly {
		t.Errorf("Access = %q, want read_only", resolution.Share.Access)
	}
	if resolution.OwnerPath.String() != "/shared_dir/file.txt" {
		t.Errorf("OwnerPath = %q, want %q", resolution.OwnerPath.String(), "/shared_dir/file.txt")
	}
}

// contentStoreAdapter satisfies port.ContentStorage using contentstore.DiskStore.
var _ port.ContentStorage = (*contentstore.DiskStore)(nil)
