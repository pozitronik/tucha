package contentstore

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/vo"
)

func validHash() vo.ContentHash {
	return vo.MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")
}

func TestDiskStore_WriteAndOpen(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	hash := validHash()
	data := []byte("hello world")

	n, err := store.Write(hash, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != int64(len(data)) {
		t.Errorf("Write returned %d bytes, want %d", n, len(data))
	}

	f, err := store.Open(hash)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("reading opened file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("Open returned %q, want %q", got, data)
	}
}

func TestDiskStore_shardDirs(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	hash := validHash()
	if _, err := store.Write(hash, bytes.NewReader([]byte("x"))); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify two-level shard directory structure.
	l1, l2 := hash.ShardPrefix()
	expectedDir := filepath.Join(dir, l1, l2)
	info, err := os.Stat(expectedDir)
	if err != nil {
		t.Fatalf("shard dir %q does not exist: %v", expectedDir, err)
	}
	if !info.IsDir() {
		t.Errorf("shard path %q is not a directory", expectedDir)
	}
}

func TestDiskStore_Open_nonexistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	_, err = store.Open(validHash())
	if !os.IsNotExist(err) {
		t.Errorf("Open nonexistent: got err %v, want os.ErrNotExist", err)
	}
}

func TestDiskStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	hash := validHash()
	if _, err := store.Write(hash, bytes.NewReader([]byte("data"))); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := store.Delete(hash); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = store.Open(hash)
	if !os.IsNotExist(err) {
		t.Errorf("Open after delete: got err %v, want os.ErrNotExist", err)
	}
}

func TestDiskStore_Delete_nonexistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	if err := store.Delete(validHash()); err != nil {
		t.Errorf("Delete nonexistent should return nil, got: %v", err)
	}
}

func TestDiskStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store, err := NewDiskStore(dir)
	if err != nil {
		t.Fatalf("NewDiskStore: %v", err)
	}

	hash := validHash()

	if store.Exists(hash) {
		t.Error("Exists before write = true, want false")
	}

	if _, err := store.Write(hash, bytes.NewReader([]byte("x"))); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !store.Exists(hash) {
		t.Error("Exists after write = false, want true")
	}
}
