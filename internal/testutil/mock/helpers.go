package mock

import (
	"time"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// ValidHash returns a deterministic valid ContentHash for testing.
func ValidHash() vo.ContentHash {
	return vo.MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")
}

// NewTestUser creates a minimal User for testing.
func NewTestUser(id int64, email string) *entity.User {
	return &entity.User{
		ID:         id,
		Email:      email,
		Password:   "password",
		QuotaBytes: 1073741824,
		Created:    time.Now().Unix(),
	}
}

// NewTestNode creates a minimal Node for testing.
func NewTestNode(userID int64, path string, nodeType vo.NodeType) *entity.Node {
	cp := vo.NewCloudPath(path)
	return &entity.Node{
		ID:     1,
		UserID: userID,
		Name:   cp.Name(),
		Home:   cp,
		Type:   nodeType,
	}
}

// NewTestFileNode creates a file Node with hash and size for testing.
func NewTestFileNode(userID int64, path string, hash vo.ContentHash, size int64) *entity.Node {
	cp := vo.NewCloudPath(path)
	return &entity.Node{
		ID:     1,
		UserID: userID,
		Name:   cp.Name(),
		Home:   cp,
		Type:   vo.NodeTypeFile,
		Hash:   hash,
		Size:   size,
	}
}

// NewTestShare creates a minimal Share for testing.
func NewTestShare(ownerID int64, path string, email string) *entity.Share {
	return &entity.Share{
		ID:           1,
		OwnerID:      ownerID,
		Home:         vo.NewCloudPath(path),
		InvitedEmail: email,
		Access:       vo.AccessReadOnly,
		Status:       vo.SharePending,
		InviteToken:  "test-invite-token",
	}
}

// NewTestTrashItem creates a minimal TrashItem for testing.
func NewTestTrashItem(userID int64, path string, nodeType vo.NodeType) *entity.TrashItem {
	cp := vo.NewCloudPath(path)
	return &entity.TrashItem{
		ID:     1,
		UserID: userID,
		Name:   cp.Name(),
		Home:   cp,
		Type:   nodeType,
	}
}

// NewTestToken creates a minimal Token for testing.
func NewTestToken(userID int64, expiresAt time.Time) *entity.Token {
	return &entity.Token{
		ID:           1,
		UserID:       userID,
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		CSRFToken:    "csrf-token-789",
		ExpiresAt:    expiresAt.Unix(),
		Created:      time.Now().Unix(),
	}
}
