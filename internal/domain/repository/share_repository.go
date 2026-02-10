package repository

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// ShareRepository persists and retrieves folder sharing invitations.
type ShareRepository interface {
	// Create inserts a new share invitation.
	Create(share *entity.Share) (int64, error)

	// GetByOwnerPathEmail finds a share by owner, path, and invited email.
	// Returns nil, nil if not found.
	GetByOwnerPathEmail(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error)

	// GetByInviteToken finds a share by its invite token.
	// Returns nil, nil if not found.
	GetByInviteToken(token string) (*entity.Share, error)

	// Delete removes a share by ID.
	Delete(id int64) error

	// ListByOwnerPath returns all shares for a given owner and path.
	ListByOwnerPath(ownerID int64, home vo.CloudPath) ([]entity.Share, error)

	// ListByOwnerPathPrefix returns all shares where the owner matches and the
	// shared home path equals the given path or is a descendant of it.
	ListByOwnerPathPrefix(ownerID int64, path vo.CloudPath) ([]entity.Share, error)

	// ListIncoming returns all pending and accepted shares where the invited email matches.
	ListIncoming(email string) ([]entity.Share, error)

	// Reinvite updates the access level of an existing share. If the share was
	// rejected, its status is reset to pending so the recipient sees it again.
	Reinvite(id int64, access vo.AccessLevel) error

	// Accept transitions a share to "accepted" and records the mount details.
	Accept(inviteToken string, mountUserID int64, mountHome string) error

	// Reject transitions a share to "rejected".
	Reject(inviteToken string) error

	// Unmount clears mount details and transitions the share back to "pending".
	// Returns the share before clearing so callers can access mount info.
	Unmount(userID int64, mountHome string) (*entity.Share, error)

	// GetByMountPath finds a share by mount user and mount path.
	// Returns nil, nil if not found.
	GetByMountPath(userID int64, mountHome string) (*entity.Share, error)

	// ListMountedByUser returns all accepted shares mounted by the given user.
	ListMountedByUser(userID int64) ([]entity.Share, error)
}
