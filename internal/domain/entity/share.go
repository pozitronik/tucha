package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// Share represents a folder sharing invitation between users.
type Share struct {
	ID           int64
	OwnerID      int64
	Home         vo.CloudPath
	InvitedEmail string
	Access       vo.AccessLevel
	Status       vo.ShareStatus
	InviteToken  string
	MountHome    string
	MountUserID  *int64
	Created      int64
}

// IsPending returns true if the share invitation has not been acted upon.
func (s *Share) IsPending() bool {
	return s.Status == vo.SharePending
}

// IsAccepted returns true if the share invitation has been accepted.
func (s *Share) IsAccepted() bool {
	return s.Status == vo.ShareAccepted
}
