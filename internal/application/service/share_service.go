package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

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

	// Check for existing share -- UNIQUE(owner_id, home, invited_email) allows at most one.
	existing, err := s.shares.GetByOwnerPathEmail(ownerID, home, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Update access level and re-invite if access changed or share was rejected.
		if existing.Access != access || existing.Status == vo.ShareRejected {
			if err := s.shares.Reinvite(existing.ID, access); err != nil {
				return nil, err
			}
			existing.Access = access
			if existing.Status == vo.ShareRejected {
				existing.Status = vo.SharePending
			}
		}
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
func (s *ShareService) Mount(userID int64, mountName string, inviteToken string, conflict vo.ConflictMode) error {
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

	// Normalize mount name: strip leading slash to avoid double-slash (e.g., "//dir").
	mountName = strings.TrimLeft(mountName, "/")

	// Check if mount point already exists and handle conflict
	mountHome := "/" + mountName
	mountPath := vo.NewCloudPath(mountHome)
	existing, err := s.nodes.Get(userID, mountPath)
	if err != nil {
		return err
	}

	if existing != nil {
		switch conflict {
		case vo.ConflictStrict:
			return ErrAlreadyExists
		case vo.ConflictRename:
			// Find unique name by appending suffix
			base := mountName
			for i := 1; existing != nil; i++ {
				mountName = fmt.Sprintf("%s (%d)", base, i)
				mountHome = "/" + mountName
				mountPath = vo.NewCloudPath(mountHome)
				existing, err = s.nodes.Get(userID, mountPath)
				if err != nil {
					return err
				}
			}
		}
	}

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

// ListMountedIn returns accepted shares whose mount point is a direct child of folderPath.
// For example, if folderPath is "/" and a share is mounted at "/SharedFolder",
// that share is included. A share mounted at "/a/b" would NOT be included for folderPath="/".
func (s *ShareService) ListMountedIn(userID int64, folderPath vo.CloudPath) ([]entity.Share, error) {
	all, err := s.shares.ListMountedByUser(userID)
	if err != nil {
		return nil, err
	}

	var result []entity.Share
	for _, share := range all {
		mountPath := vo.NewCloudPath(share.MountHome)
		if mountPath.Parent().String() == folderPath.String() {
			result = append(result, share)
		}
	}
	return result, nil
}

// MountResolution holds the result of resolving a path through a mounted share.
type MountResolution struct {
	Share     *entity.Share
	OwnerPath vo.CloudPath
}

// ResolveMount checks whether the given path falls under a mounted share for the user.
// If the path matches a mount point exactly or is a descendant, it returns the share
// and the corresponding path in the owner's tree.
// Returns nil if the path does not match any mount.
func (s *ShareService) ResolveMount(userID int64, path vo.CloudPath) (*MountResolution, error) {
	all, err := s.shares.ListMountedByUser(userID)
	if err != nil {
		return nil, err
	}

	pathStr := path.String()
	for i := range all {
		share := &all[i]
		// Normalize MountHome to avoid double-slash mismatches.
		mountHome := vo.NewCloudPath(share.MountHome).String()
		ownerHome := share.Home.String()

		if pathStr == mountHome {
			// Exact match: path IS the mount point
			return &MountResolution{Share: share, OwnerPath: share.Home}, nil
		}
		if strings.HasPrefix(pathStr, mountHome+"/") {
			// Descendant: replace mount prefix with owner prefix
			suffix := pathStr[len(mountHome):]
			resolved := vo.NewCloudPath(ownerHome + suffix)
			return &MountResolution{Share: share, OwnerPath: resolved}, nil
		}
	}
	return nil, nil
}

// generateInviteToken produces a random 32-character hex string for share tokens.
func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating invite token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
