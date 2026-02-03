package vo

import "fmt"

// ConflictMode determines how to handle file path conflicts.
type ConflictMode string

const (
	// ConflictStrict returns an error if the target path already exists.
	ConflictStrict ConflictMode = "strict"
	// ConflictRename replaces the existing file at the target path.
	// Named "rename" per the protocol, but the actual behavior is replacement.
	ConflictRename ConflictMode = "rename"
	// ConflictReplace explicitly replaces the existing file.
	ConflictReplace ConflictMode = "replace"
)

// ParseConflictMode converts a raw string to a ConflictMode.
// Empty string defaults to ConflictRename (protocol default behavior).
func ParseConflictMode(raw string) (ConflictMode, error) {
	switch raw {
	case "strict":
		return ConflictStrict, nil
	case "rename", "":
		return ConflictRename, nil
	case "replace":
		return ConflictReplace, nil
	default:
		return "", fmt.Errorf("unknown conflict mode: %q", raw)
	}
}

// String returns the string representation of the conflict mode.
func (m ConflictMode) String() string {
	return string(m)
}
