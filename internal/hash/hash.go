// Package hash implements the mrCloud hash algorithm
// for file identification and deduplication.
//
// Algorithm:
//   - Files < 21 bytes: zero-pad content to 20 bytes, hex-encode (NOT SHA1).
//   - Files >= 21 bytes: SHA1("mrCloud" + content + decimal_size_string), uppercase hex.
package hash

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	// seed is prepended to file content before hashing (files >= 21 bytes).
	seed = "mrCloud"

	// smallFileThreshold is the size below which files are zero-padded instead of hashed.
	smallFileThreshold = 21

	// smallFileBuffer is the size of the zero-padded buffer for small files.
	smallFileBuffer = 20
)

// Compute calculates the mrCloud hash for the given data.
// Returns a 40-character uppercase hex string.
func Compute(data []byte) string {
	if len(data) < smallFileThreshold {
		return computeSmall(data)
	}
	return computeLarge(data)
}

// ComputeReader calculates the mrCloud hash by streaming from a reader.
// The size parameter must be the total number of bytes that will be read.
// For files < 21 bytes, the entire content is read into memory.
// Returns a 40-character uppercase hex string.
func ComputeReader(r io.Reader, size int64) (string, error) {
	if size < smallFileThreshold {
		data, err := io.ReadAll(r)
		if err != nil {
			return "", fmt.Errorf("reading small file: %w", err)
		}
		return computeSmall(data), nil
	}

	h := sha1.New()

	// Write seed.
	if _, err := h.Write([]byte(seed)); err != nil {
		return "", fmt.Errorf("writing seed: %w", err)
	}

	// Stream content.
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("streaming content: %w", err)
	}

	// Write size suffix.
	sizeStr := fmt.Sprintf("%d", size)
	if _, err := h.Write([]byte(sizeStr)); err != nil {
		return "", fmt.Errorf("writing size suffix: %w", err)
	}

	return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
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
