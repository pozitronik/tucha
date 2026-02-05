package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// ShareService handles folder sharing invitations: creating, listing, mounting,
// unmounting, and rejecting shares.
type ShareService struct {
	shares   repository.ShareRepository
	nodes    repository.NodeRepository
	contents repository.ContentRepository
	users    repository.UserRepository
}

// NewShareService creates a new ShareService.
func NewShareService(
	shares repository.ShareRepository,
	nodes repository.NodeRepository,
	contents repository.ContentRepository,
	users repository.UserRepository,
) *ShareService {
	return &ShareService{
		shares:   shares,
		nodes:    nodes,
		contents: contents,
		users:    users,
	}
}

// Share creates a folder sharing invitation.
// Cannot share with self (owner email must differ from invited email).
func (s *ShareService) Share(ownerID int64, home vo.CloudPath, email string, access vo.AccessLevel) (*entity.Share, error) {
	// Verify the folder exists.
	node, err := s.nodes.Get(ownerID, home)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrNotFound
	}

	// Cannot share with self.
	owner, err := s.users.GetByID(ownerID)
	if err != nil {
		return nil, err
	}
	if owner != nil && owner.Email == email {
		return nil, ErrForbidden
	}

	// Check for existing share.
	existing, err := s.shares.GetByOwnerPathEmail(ownerID, home, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	token, err := generateInviteToken()
	if err != nil {
		return nil, err
	}

	share := &entity.Share{
		OwnerID:      ownerID,
		Home:         home,
		InvitedEmail: email,
		Access:       access,
		Status:       vo.SharePending,
		InviteToken:  token,
	}

	id, err := s.shares.Create(share)
	if err != nil {
		return nil, err
	}
	share.ID = id

	return share, nil
}

// Unshare removes a folder sharing invitation.
func (s *ShareService) Unshare(ownerID int64, home vo.CloudPath, email string) error {
	share, err := s.shares.GetByOwnerPathEmail(ownerID, home, email)
	if err != nil {
		return err
	}
	if share == nil {
		return ErrNotFound
	}

	return s.shares.Delete(share.ID)
}

// GetShareInfo returns all share members for a given folder.
func (s *ShareService) GetShareInfo(ownerID int64, home vo.CloudPath) ([]entity.Share, error) {
	return s.shares.ListByOwnerPath(ownerID, home)
}

// ListIncoming returns all pending share invitations for the given email.
func (s *ShareService) ListIncoming(email string) ([]entity.Share, error) {
	return s.shares.ListIncoming(email)
}

// Mount accepts a share invitation and records the mount point.
func (s *ShareService) Mount(userID int64, mountName string, inviteToken string) error {
	share, err := s.shares.GetByInviteToken(inviteToken)
	if err != nil {
		return err
	}
	if share == nil {
		return ErrNotFound
	}

	// Verify the caller's email matches the invited email.
	user, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil || user.Email != share.InvitedEmail {
		return ErrForbidden
	}

	mountHome := "/" + mountName
	return s.shares.Accept(inviteToken, userID, mountHome)
}

// Unmount removes a mount point. If cloneCopy is true, copies the shared content
// into the user's own tree before unmounting.
func (s *ShareService) Unmount(userID int64, mountHome string, cloneCopy bool) error {
	share, err := s.shares.Unmount(userID, mountHome)
	if err != nil {
		return err
	}
	if share == nil {
		return ErrNotFound
	}

	if cloneCopy {
		dstPath := vo.NewCloudPath(mountHome)
		if err := s.nodes.EnsurePath(userID, dstPath); err != nil {
			return err
		}
		if err := cloneTree(s.nodes, s.contents, share.OwnerID, share.Home, userID, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// Reject rejects a share invitation.
func (s *ShareService) Reject(userID int64, inviteToken string) error {
	share, err := s.shares.GetByInviteToken(inviteToken)
	if err != nil {
		return err
	}
	if share == nil {
		return ErrNotFound
	}

	// Verify the caller's email matches the invited email.
	user, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil || user.Email != share.InvitedEmail {
		return ErrForbidden
	}

	return s.shares.Reject(inviteToken)
}

// generateInviteToken produces a random 32-character hex string for share tokens.
func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating invite token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
