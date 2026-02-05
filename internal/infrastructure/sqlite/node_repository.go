package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// NodeRepository implements repository.NodeRepository using SQLite.
type NodeRepository struct {
	db *sql.DB
}

// NewNodeRepository creates a NodeRepository from the given database connection.
func NewNodeRepository(db *DB) *NodeRepository {
	return &NodeRepository{db: db.Conn()}
}

// Get retrieves a single node by user ID and cloud path.
// Returns nil, nil if not found.
func (r *NodeRepository) Get(userID int64, path vo.CloudPath) (*entity.Node, error) {
	row := r.db.QueryRow(
		`SELECT `+nodeColumns+` FROM nodes WHERE user_id = ? AND home = ?`,
		userID, path.String(),
	)
	n, err := scanNode(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}
	return n, nil
}

// ListChildren returns the children of the given folder, with offset/limit pagination.
func (r *NodeRepository) ListChildren(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
	parent, err := r.Get(userID, path)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", path)
	}

	rows, err := r.db.Query(
		`SELECT `+nodeColumns+` FROM nodes WHERE user_id = ? AND parent_id = ?
		 ORDER BY node_type DESC, name ASC
		 LIMIT ? OFFSET ?`,
		userID, parent.ID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("listing children: %w", err)
	}
	defer rows.Close()

	var nodes []entity.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning node: %w", err)
		}
		nodes = append(nodes, *n)
	}
	return nodes, rows.Err()
}

// CountChildren returns the count of folders and files that are direct children.
func (r *NodeRepository) CountChildren(userID int64, path vo.CloudPath) (folders, files int, err error) {
	parent, err := r.Get(userID, path)
	if err != nil {
		return 0, 0, err
	}
	if parent == nil {
		return 0, 0, fmt.Errorf("parent folder not found: %s", path)
	}

	err = r.db.QueryRow(
		`SELECT
			COALESCE(SUM(CASE WHEN node_type='folder' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN node_type='file' THEN 1 ELSE 0 END), 0)
		 FROM nodes WHERE user_id = ? AND parent_id = ?`,
		userID, parent.ID,
	).Scan(&folders, &files)
	if err != nil {
		return 0, 0, fmt.Errorf("counting children: %w", err)
	}
	return folders, files, nil
}

// CreateRootNode creates the root folder node "/" with no parent.
func (r *NodeRepository) CreateRootNode(userID int64) (*entity.Node, error) {
	root := vo.NewCloudPath("/")
	now := time.Now().Unix()
	res, err := r.db.Exec(
		`INSERT INTO nodes (user_id, parent_id, name, home, node_type, size, mtime, created)
		 VALUES (?, NULL, '', '/', 'folder', 0, ?, ?)`,
		userID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating root node: %w", err)
	}
	id, _ := res.LastInsertId()
	return &entity.Node{
		ID:      id,
		UserID:  userID,
		Name:    "",
		Home:    root,
		Type:    vo.NodeTypeFolder,
		MTime:   now,
		Rev:     1,
		GRev:    1,
		Created: now,
	}, nil
}

// CreateFolder creates a new folder node at the given path.
func (r *NodeRepository) CreateFolder(userID int64, path vo.CloudPath) (*entity.Node, error) {
	name := path.Name()
	parentPath := path.Parent()

	parent, err := r.Get(userID, parentPath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", parentPath)
	}

	now := time.Now().Unix()
	res, err := r.db.Exec(
		`INSERT INTO nodes (user_id, parent_id, name, home, node_type, size, mtime, rev, grev, tree, created)
		 VALUES (?, ?, ?, ?, 'folder', 0, ?, 1, 1, '', ?)`,
		userID, parent.ID, name, path.String(), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating folder: %w", err)
	}

	id, _ := res.LastInsertId()
	return &entity.Node{
		ID:       id,
		UserID:   userID,
		ParentID: &parent.ID,
		Name:     name,
		Home:     path,
		Type:     vo.NodeTypeFolder,
		Size:     0,
		MTime:    now,
		Rev:      1,
		GRev:     1,
		Created:  now,
	}, nil
}

// CreateFile creates a new file node at the given path with the specified hash and size.
func (r *NodeRepository) CreateFile(userID int64, path vo.CloudPath, hash vo.ContentHash, size int64) (*entity.Node, error) {
	name := path.Name()
	parentPath := path.Parent()

	parent, err := r.Get(userID, parentPath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", parentPath)
	}

	now := time.Now().Unix()
	res, err := r.db.Exec(
		`INSERT INTO nodes (user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created)
		 VALUES (?, ?, ?, ?, 'file', ?, ?, ?, 1, 1, '', ?)`,
		userID, parent.ID, name, path.String(), size, hash.String(), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	id, _ := res.LastInsertId()
	return &entity.Node{
		ID:       id,
		UserID:   userID,
		ParentID: &parent.ID,
		Name:     name,
		Home:     path,
		Type:     vo.NodeTypeFile,
		Size:     size,
		Hash:     hash,
		MTime:    now,
		Rev:      1,
		GRev:     1,
		Created:  now,
	}, nil
}

// Delete removes a node at the given path.
func (r *NodeRepository) Delete(userID int64, path vo.CloudPath) error {
	_, err := r.db.Exec(
		"DELETE FROM nodes WHERE user_id = ? AND home = ?",
		userID, path.String(),
	)
	if err != nil {
		return fmt.Errorf("deleting node: %w", err)
	}
	return nil
}

// Rename changes the name of a node, updating its path and all descendant paths.
func (r *NodeRepository) Rename(userID int64, path vo.CloudPath, newName string) (*entity.Node, error) {
	parentPath := path.Parent()
	newHome := parentPath.Join(newName)
	now := time.Now().Unix()

	_, err := r.db.Exec(
		`UPDATE nodes SET name = ?, home = ?, mtime = ? WHERE user_id = ? AND home = ?`,
		newName, newHome.String(), now, userID, path.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("renaming node: %w", err)
	}

	if err := r.updateChildPaths(userID, path, newHome); err != nil {
		return nil, err
	}

	return r.Get(userID, newHome)
}

// Move moves a node from srcPath into the target folder.
func (r *NodeRepository) Move(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	name := srcPath.Name()
	newHome := targetFolder.Join(name)

	newParent, err := r.Get(userID, targetFolder)
	if err != nil {
		return nil, err
	}
	if newParent == nil {
		return nil, fmt.Errorf("target folder not found: %s", targetFolder)
	}

	now := time.Now().Unix()
	_, err = r.db.Exec(
		`UPDATE nodes SET parent_id = ?, home = ?, mtime = ? WHERE user_id = ? AND home = ?`,
		newParent.ID, newHome.String(), now, userID, srcPath.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("moving node: %w", err)
	}

	if err := r.updateChildPaths(userID, srcPath, newHome); err != nil {
		return nil, err
	}

	return r.Get(userID, newHome)
}

// Copy duplicates a node (and its children for folders) from srcPath into the target folder.
func (r *NodeRepository) Copy(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	src, err := r.Get(userID, srcPath)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("source not found: %s", srcPath)
	}

	name := srcPath.Name()
	newHome := targetFolder.Join(name)

	if src.IsFile() {
		return r.CreateFile(userID, newHome, src.Hash, src.Size)
	}

	newFolder, err := r.CreateFolder(userID, newHome)
	if err != nil {
		return nil, err
	}

	if err := r.copyChildren(userID, srcPath, newHome); err != nil {
		return nil, err
	}

	return newFolder, nil
}

// GetWithDescendants retrieves a node and all its descendants (for folders).
// For files, the descendants slice is empty.
func (r *NodeRepository) GetWithDescendants(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
	node, err := r.Get(userID, path)
	if err != nil {
		return nil, nil, err
	}
	if node == nil {
		return nil, nil, nil
	}

	if node.IsFile() {
		return node, nil, nil
	}

	rows, err := r.db.Query(
		`SELECT `+nodeColumns+` FROM nodes WHERE user_id = ? AND home LIKE ?`,
		userID, path.String()+"/%",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("listing descendants: %w", err)
	}
	defer rows.Close()

	var descendants []entity.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("scanning descendant: %w", err)
		}
		descendants = append(descendants, *n)
	}
	return node, descendants, rows.Err()
}

// TotalSize returns the total size of all file nodes belonging to the given user.
func (r *NodeRepository) TotalSize(userID int64) (int64, error) {
	var total sql.NullInt64
	err := r.db.QueryRow(
		"SELECT COALESCE(SUM(size), 0) FROM nodes WHERE user_id = ? AND node_type = 'file'",
		userID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("calculating total size: %w", err)
	}
	return total.Int64, nil
}

// SetWeblink assigns a weblink identifier to the node at the given path.
func (r *NodeRepository) SetWeblink(userID int64, path vo.CloudPath, weblink string) error {
	var weblinkVal *string
	if weblink != "" {
		weblinkVal = &weblink
	}
	_, err := r.db.Exec(
		`UPDATE nodes SET weblink = ? WHERE user_id = ? AND home = ?`,
		weblinkVal, userID, path.String(),
	)
	if err != nil {
		return fmt.Errorf("setting weblink: %w", err)
	}
	return nil
}

// GetByWeblink retrieves a node by its weblink identifier (across all users).
func (r *NodeRepository) GetByWeblink(weblink string) (*entity.Node, error) {
	row := r.db.QueryRow(
		`SELECT `+nodeColumns+` FROM nodes WHERE weblink = ?`,
		weblink,
	)
	n, err := scanNode(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting node by weblink: %w", err)
	}
	return n, nil
}

// ListByWeblink returns all nodes with a non-empty weblink for the given user.
func (r *NodeRepository) ListByWeblink(userID int64) ([]entity.Node, error) {
	rows, err := r.db.Query(
		`SELECT `+nodeColumns+` FROM nodes WHERE user_id = ? AND weblink IS NOT NULL ORDER BY name ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing published nodes: %w", err)
	}
	defer rows.Close()

	var nodes []entity.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning published node: %w", err)
		}
		nodes = append(nodes, *n)
	}
	return nodes, rows.Err()
}

// Exists checks whether a node exists at the given path for the user.
func (r *NodeRepository) Exists(userID int64, path vo.CloudPath) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM nodes WHERE user_id = ? AND home = ?)",
		userID, path.String(),
	).Scan(&exists)
	return exists, err
}

// EnsurePath creates all intermediate folders for the given path, skipping
// any that already exist. Iterates from shallowest to deepest ancestor.
func (r *NodeRepository) EnsurePath(userID int64, path vo.CloudPath) error {
	if path.IsRoot() {
		return nil
	}

	// Collect ancestors from target up to root (exclusive).
	var segments []vo.CloudPath
	current := path
	for !current.IsRoot() {
		segments = append(segments, current)
		current = current.Parent()
	}

	// Reverse: shallowest first so parents exist before children.
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	for _, seg := range segments {
		exists, err := r.Exists(userID, seg)
		if err != nil {
			return fmt.Errorf("checking path %s: %w", seg, err)
		}
		if !exists {
			if _, err := r.CreateFolder(userID, seg); err != nil {
				return fmt.Errorf("creating intermediate folder %s: %w", seg, err)
			}
		}
	}

	return nil
}

// updateChildPaths updates the home path of all descendants when a parent is renamed or moved.
func (r *NodeRepository) updateChildPaths(userID int64, oldPrefix, newPrefix vo.CloudPath) error {
	_, err := r.db.Exec(
		`UPDATE nodes SET home = ? || SUBSTR(home, ?) WHERE user_id = ? AND home LIKE ? AND home != ?`,
		newPrefix.String(), len(oldPrefix.String())+1, userID, oldPrefix.String()+"/%", oldPrefix.String(),
	)
	if err != nil {
		return fmt.Errorf("updating child paths: %w", err)
	}
	return nil
}

// copyChildren recursively copies all children from srcFolder to dstFolder.
func (r *NodeRepository) copyChildren(userID int64, srcFolder, dstFolder vo.CloudPath) error {
	children, err := r.ListChildren(userID, srcFolder, 0, 65535)
	if err != nil {
		return err
	}

	for _, child := range children {
		newHome := dstFolder.Join(child.Name)
		if child.IsFile() {
			if _, err := r.CreateFile(userID, newHome, child.Hash, child.Size); err != nil {
				return err
			}
		} else {
			if _, err := r.CreateFolder(userID, newHome); err != nil {
				return err
			}
			if err := r.copyChildren(userID, child.Home, newHome); err != nil {
				return err
			}
		}
	}
	return nil
}
