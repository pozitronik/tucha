package port

import (
	"io"
	"os"

	"github.com/pozitronik/tucha/internal/domain/vo"
)

// ContentStorage provides content-addressable file storage on disk.
type ContentStorage interface {
	// Write stores data from the reader under the given hash.
	// Returns the number of bytes written.
	Write(hash vo.ContentHash, r io.Reader) (int64, error)

	// Open returns a file handle for the content identified by hash.
	// Returns os.ErrNotExist if the content does not exist.
	Open(hash vo.ContentHash) (*os.File, error)

	// Delete removes the content file for the given hash.
	// No error is returned if the file does not exist.
	Delete(hash vo.ContentHash) error

	// Exists checks whether content with the given hash exists on disk.
	Exists(hash vo.ContentHash) bool
}
