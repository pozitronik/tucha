package vo

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateWeblink produces a random weblink identifier in the format "{8hex}/{8hex}".
// Each segment is 8 hex characters (4 random bytes).
func GenerateWeblink() (string, error) {
	a := make([]byte, 4)
	b := make([]byte, 4)
	if _, err := rand.Read(a); err != nil {
		return "", fmt.Errorf("generating weblink first segment: %w", err)
	}
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating weblink second segment: %w", err)
	}
	return hex.EncodeToString(a) + "/" + hex.EncodeToString(b), nil
}
