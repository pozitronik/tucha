// Package hasher implements the mrCloud hash algorithm for file identification and deduplication.
//
// Algorithm:
//   - Files < 21 bytes: zero-pad content to 20 bytes, hex-encode (NOT SHA1).
//   - Files >= 21 bytes: SHA1("mrCloud" + content + decimal_size_string), uppercase hex.
package hasher

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"tucha/internal/domain/vo"
)

const (
	seed               = "mrCloud"
	smallFileThreshold = 21
	smallFileBuffer    = 20
)

// MrCloud implements port.Hasher using the mrCloud hash algorithm.
type MrCloud struct{}

// NewMrCloud creates a new MrCloud hasher.
func NewMrCloud() *MrCloud {
	return &MrCloud{}
}

// Compute calculates the mrCloud hash for the given data.
func (h *MrCloud) Compute(data []byte) vo.ContentHash {
	var raw string
	if len(data) < smallFileThreshold {
		raw = computeSmall(data)
	} else {
		raw = computeLarge(data)
	}
	return vo.MustContentHash(raw)
}

// ComputeReader calculates the mrCloud hash by streaming from a reader.
// The size parameter must be the total number of bytes.
func (h *MrCloud) ComputeReader(r io.Reader, size int64) (vo.ContentHash, error) {
	if size < smallFileThreshold {
		data, err := io.ReadAll(r)
		if err != nil {
			return vo.ContentHash{}, fmt.Errorf("reading small file: %w", err)
		}
		return vo.MustContentHash(computeSmall(data)), nil
	}

	sha := sha1.New()

	if _, err := sha.Write([]byte(seed)); err != nil {
		return vo.ContentHash{}, fmt.Errorf("writing seed: %w", err)
	}

	if _, err := io.Copy(sha, r); err != nil {
		return vo.ContentHash{}, fmt.Errorf("streaming content: %w", err)
	}

	sizeStr := fmt.Sprintf("%d", size)
	if _, err := sha.Write([]byte(sizeStr)); err != nil {
		return vo.ContentHash{}, fmt.Errorf("writing size suffix: %w", err)
	}

	raw := strings.ToUpper(hex.EncodeToString(sha.Sum(nil)))
	return vo.MustContentHash(raw), nil
}

// computeSmall handles files shorter than 21 bytes: zero-pad to 20 bytes, hex-encode.
func computeSmall(data []byte) string {
	buf := make([]byte, smallFileBuffer)
	copy(buf, data)
	return strings.ToUpper(hex.EncodeToString(buf))
}

// computeLarge handles files of 21+ bytes: SHA1(seed + content + size_string).
func computeLarge(data []byte) string {
	h := sha1.New()
	h.Write([]byte(seed))
	h.Write(data)
	h.Write([]byte(fmt.Sprintf("%d", len(data))))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}
