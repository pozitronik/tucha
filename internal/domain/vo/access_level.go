package vo

import "fmt"

// AccessLevel represents the permission level of a folder share.
type AccessLevel string

const (
	// AccessReadOnly grants read-only access to the shared folder.
	AccessReadOnly AccessLevel = "read_only"
	// AccessReadWrite grants read and write access to the shared folder.
	AccessReadWrite AccessLevel = "read_write"
)

// ParseAccessLevel converts a raw string to an AccessLevel.
// Accepts both internal ("read_only"/"read_write") and API short forms ("r"/"rw").
func ParseAccessLevel(raw string) (AccessLevel, error) {
	switch raw {
	case "read_only", "r":
		return AccessReadOnly, nil
	case "read_write", "rw":
		return AccessReadWrite, nil
	default:
		return "", fmt.Errorf("unknown access level: %q", raw)
	}
}

// String returns the internal string representation of the access level.
func (a AccessLevel) String() string {
	return string(a)
}

// APIString returns the API representation ("read_only" or "read_write").
func (a AccessLevel) APIString() string {
	return string(a)
}
