package sqlite

import (
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("opening test DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestFileVersionRepository_InsertAndList(t *testing.T) {
	db := openTestDB(t)
	userRepo := NewUserRepository(db)
	repo := NewFileVersionRepository(db)

	// Create user so FK constraint is satisfied.
	userID, err := userRepo.Create(&entity.User{
		Email:    "test@example.com",
		Password: "pass",
	})
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	hash1, _ := vo.NewContentHash("0000000000000000000000000000000000000001")
	hash2, _ := vo.NewContentHash("0000000000000000000000000000000000000002")
	path := vo.NewCloudPath("/test/file.bin")

	// Insert first version.
	err = repo.Insert(&entity.FileVersion{
		UserID: userID,
		Home:   path,
		Name:   "file.bin",
		Hash:   hash1,
		Size:   1024,
		Rev:    1,
	})
	if err != nil {
		t.Fatalf("Insert version 1: %v", err)
	}

	// Insert second version with same user/path but different hash.
	err = repo.Insert(&entity.FileVersion{
		UserID: userID,
		Home:   path,
		Name:   "file.bin",
		Hash:   hash2,
		Size:   2048,
		Rev:    1,
	})
	if err != nil {
		t.Fatalf("Insert version 2: %v", err)
	}

	// List versions.
	versions, err := repo.ListByPath(userID, path)
	if err != nil {
		t.Fatalf("ListByPath: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("len(versions) = %d, want 2", len(versions))
	}

	if versions[0].Size != 1024 {
		t.Errorf("versions[0].Size = %d, want 1024", versions[0].Size)
	}
	if versions[1].Size != 2048 {
		t.Errorf("versions[1].Size = %d, want 2048", versions[1].Size)
	}
}

// TestFileVersionRepository_FullAddByHashFlow tests the full scenario:
// 1. Create user, 2. Create root node, 3. Add content,
// 4. AddByHash twice, 5. Verify 2 version entries.
func TestFileVersionRepository_FullAddByHashFlow(t *testing.T) {
	db := openTestDB(t)

	userRepo := NewUserRepository(db)
	nodeRepo := NewNodeRepository(db)
	contentRepo := NewContentRepository(db)
	versionRepo := NewFileVersionRepository(db)

	// Create user.
	userID, err := userRepo.Create(&entity.User{
		Email:    "test@example.com",
		Password: "pass",
	})
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	// Create root node.
	_, err = nodeRepo.CreateRootNode(userID)
	if err != nil {
		t.Fatalf("CreateRootNode: %v", err)
	}

	hash1, _ := vo.NewContentHash("AAAA000000000000000000000000000000000001")
	hash2, _ := vo.NewContentHash("BBBB000000000000000000000000000000000002")
	path := vo.NewCloudPath("/test/file.bin")

	// Register content so AddByHash's Exists check passes.
	contentRepo.Insert(hash1, 1024)
	contentRepo.Insert(hash2, 2048)

	// === First AddByHash flow ===
	// 1. EnsurePath for parent
	err = nodeRepo.EnsurePath(userID, path.Parent())
	if err != nil {
		t.Fatalf("EnsurePath (1st): %v", err)
	}

	// 2. CreateFile
	node1, err := nodeRepo.CreateFile(userID, path, hash1, 1024)
	if err != nil {
		t.Fatalf("CreateFile (1st): %v", err)
	}

	// 3. Record version
	err = versionRepo.Insert(&entity.FileVersion{
		UserID: userID,
		Home:   node1.Home,
		Name:   node1.Name,
		Hash:   node1.Hash,
		Size:   node1.Size,
		Rev:    node1.Rev,
	})
	if err != nil {
		t.Fatalf("Insert version 1: %v", err)
	}

	// === Second AddByHash flow (ConflictReplace) ===
	// 1. Delete existing node
	err = nodeRepo.Delete(userID, path)
	if err != nil {
		t.Fatalf("Delete existing node: %v", err)
	}

	// 2. EnsurePath for parent (should still exist)
	err = nodeRepo.EnsurePath(userID, path.Parent())
	if err != nil {
		t.Fatalf("EnsurePath (2nd): %v", err)
	}

	// 3. CreateFile with new hash
	node2, err := nodeRepo.CreateFile(userID, path, hash2, 2048)
	if err != nil {
		t.Fatalf("CreateFile (2nd): %v", err)
	}

	// 4. Record version
	err = versionRepo.Insert(&entity.FileVersion{
		UserID: userID,
		Home:   node2.Home,
		Name:   node2.Name,
		Hash:   node2.Hash,
		Size:   node2.Size,
		Rev:    node2.Rev,
	})
	if err != nil {
		t.Fatalf("Insert version 2: %v", err)
	}

	// === Verify ===
	versions, err := versionRepo.ListByPath(userID, path)
	if err != nil {
		t.Fatalf("ListByPath: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("len(versions) = %d, want 2", len(versions))
	}

	// Version 1
	if versions[0].Hash != hash1 {
		t.Errorf("versions[0].Hash = %v, want %v", versions[0].Hash, hash1)
	}
	if versions[0].Size != 1024 {
		t.Errorf("versions[0].Size = %d, want 1024", versions[0].Size)
	}

	// Version 2
	if versions[1].Hash != hash2 {
		t.Errorf("versions[1].Hash = %v, want %v", versions[1].Hash, hash2)
	}
	if versions[1].Size != 2048 {
		t.Errorf("versions[1].Size = %d, want 2048", versions[1].Size)
	}
}
