// Package port defines application-layer interfaces for infrastructure concerns.
// Implementations live in the infrastructure layer.
package port

import (
	"io"

	"tucha/internal/domain/vo"
)

// Hasher computes content hashes for file deduplication.
type Hasher interface {
	// Compute calculates the hash for the given data.
	Compute(data []byte) vo.ContentHash

	// ComputeReader calculates the hash by streaming from a reader.
	// The size parameter must be the total number of bytes.
	ComputeReader(r io.Reader, size int64) (vo.ContentHash, error)
}
