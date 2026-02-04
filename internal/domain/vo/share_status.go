package vo

import "fmt"

// ShareStatus represents the state of a folder share invitation.
type ShareStatus string

const (
	// SharePending indicates the invite has been sent but not yet acted upon.
	SharePending ShareStatus = "pending"
	// ShareAccepted indicates the invite has been accepted (mounted).
	ShareAccepted ShareStatus = "accepted"
	// ShareRejected indicates the invite has been rejected.
	ShareRejected ShareStatus = "rejected"
)

// ParseShareStatus converts a raw string to a ShareStatus.
// Returns an error for unknown values.
func ParseShareStatus(raw string) (ShareStatus, error) {
	switch raw {
	case "pending":
		return SharePending, nil
	case "accepted":
		return ShareAccepted, nil
	case "rejected":
		return ShareRejected, nil
	default:
		return "", fmt.Errorf("unknown share status: %q", raw)
	}
}

// String returns the string representation of the share status.
func (s ShareStatus) String() string {
	return string(s)
}
