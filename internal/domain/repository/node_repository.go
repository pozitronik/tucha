// Package repository defines domain repository interfaces.
// Implementations live in the infrastructure layer.
package repository

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// NodeRepository persists and retrieves filesystem nodes.
type NodeRepository interface {
	// Get retrieves a single node by user ID and cloud path.
	// Returns nil, nil if not found.
	Get(userID int64, path vo.CloudPath) (*entity.Node, error)

	// ListChildren returns the children of the given folder, with offset/limit pagination.
	ListChildren(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error)

	// CountChildren returns the count of folders and files that are direct children.
	CountChildren(userID int64, path vo.CloudPath) (folders, files int, err error)

	// CreateRootNode creates the root folder node "/" with no parent.
	// Called when a new user is created to initialize their storage tree.
	CreateRootNode(userID int64) (*entity.Node, error)

	// CreateFolder creates a new folder node at the given path.
	CreateFolder(userID int64, path vo.CloudPath) (*entity.Node, error)

	// CreateFile creates a new file node at the given path with the specified hash and size.
	CreateFile(userID int64, path vo.CloudPath, hash vo.ContentHash, size int64) (*entity.Node, error)

	// Delete removes a node at the given path.
	// For folders, cascading delete is handled by the database constraints.
	Delete(userID int64, path vo.CloudPath) error

	// Rename changes the name of a node, updating its path and all descendant paths.
	Rename(userID int64, path vo.CloudPath, newName string) (*entity.Node, error)

	// Move moves a node from srcPath into the target folder.
	Move(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error)

	// Copy duplicates a node (and its children for folders) from srcPath into the target folder.
	Copy(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error)

	// EnsurePath creates all intermediate folders for the given path, skipping
	// any that already exist. Analogous to "mkdir -p".
	EnsurePath(userID int64, path vo.CloudPath) error

	// GetWithDescendants retrieves a node and all its descendants (for folders).
	// For files, the descendants slice is empty.
	// Returns nil, nil, nil if the node is not found.
	GetWithDescendants(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error)

	// TotalSize returns the total size of all file nodes belonging to the given user.
	TotalSize(userID int64) (int64, error)

	// SetWeblink assigns a weblink identifier to the node at the given path.
	SetWeblink(userID int64, path vo.CloudPath, weblink string) error

	// GetByWeblink retrieves a node by its weblink identifier (across all users).
	// Returns nil, nil if not found.
	GetByWeblink(weblink string) (*entity.Node, error)

	// ListByWeblink returns all nodes with a non-empty weblink for the given user.
	ListByWeblink(userID int64) ([]entity.Node, error)

	// Exists checks whether a node exists at the given path for the user.
	Exists(userID int64, path vo.CloudPath) (bool, error)
}
