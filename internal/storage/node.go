package storage

import (
	"database/sql"
	"fmt"
	"path"
	"strings"
	"time"

	"tucha/internal/model"
)

// NodeStore handles filesystem node CRUD operations.
type NodeStore struct {
	db *sql.DB
}

// NewNodeStore creates a new NodeStore using the given database connection.
func NewNodeStore(db *DB) *NodeStore {
	return &NodeStore{db: db.Conn()}
}

// Get retrieves a single node by user ID and cloud path.
// Returns nil if not found.
func (s *NodeStore) Get(userID int64, homePath string) (*model.Node, error) {
	n := &model.Node{}
	var parentID sql.NullInt64
	var hash sql.NullString
	err := s.db.QueryRow(
		`SELECT id, user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created
		 FROM nodes WHERE user_id = ? AND home = ?`,
		userID, homePath,
	).Scan(&n.ID, &n.UserID, &parentID, &n.Name, &n.Home, &n.Type, &n.Size, &hash, &n.MTime, &n.Rev, &n.GRev, &n.Tree, &n.Created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}
	if parentID.Valid {
		n.ParentID = &parentID.Int64
	}
	if hash.Valid {
		n.Hash = hash.String
	}
	return n, nil
}

// ListChildren returns the children of the given folder path, with offset/limit pagination.
func (s *NodeStore) ListChildren(userID int64, homePath string, offset, limit int) ([]model.Node, error) {
	// Find parent node ID.
	parent, err := s.Get(userID, homePath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", homePath)
	}

	rows, err := s.db.Query(
		`SELECT id, user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created
		 FROM nodes WHERE user_id = ? AND parent_id = ?
		 ORDER BY node_type DESC, name ASC
		 LIMIT ? OFFSET ?`,
		userID, parent.ID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("listing children: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		var parentID sql.NullInt64
		var hash sql.NullString
		if err := rows.Scan(&n.ID, &n.UserID, &parentID, &n.Name, &n.Home, &n.Type, &n.Size, &hash, &n.MTime, &n.Rev, &n.GRev, &n.Tree, &n.Created); err != nil {
			return nil, fmt.Errorf("scanning node: %w", err)
		}
		if parentID.Valid {
			n.ParentID = &parentID.Int64
		}
		if hash.Valid {
			n.Hash = hash.String
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

// CountChildren returns the number of files and folders that are direct children
// of the given folder.
func (s *NodeStore) CountChildren(userID int64, homePath string) (folders, files int, err error) {
	parent, err := s.Get(userID, homePath)
	if err != nil {
		return 0, 0, err
	}
	if parent == nil {
		return 0, 0, fmt.Errorf("parent folder not found: %s", homePath)
	}

	err = s.db.QueryRow(
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

// CreateFolder creates a new folder node at the given path.
// Returns the created node.
func (s *NodeStore) CreateFolder(userID int64, homePath string) (*model.Node, error) {
	homePath = normalizePath(homePath)
	name := path.Base(homePath)
	parentPath := path.Dir(homePath)

	parent, err := s.Get(userID, parentPath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", parentPath)
	}

	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO nodes (user_id, parent_id, name, home, node_type, size, mtime, rev, grev, tree, created)
		 VALUES (?, ?, ?, ?, 'folder', 0, ?, 1, 1, '', ?)`,
		userID, parent.ID, name, homePath, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating folder: %w", err)
	}

	id, _ := res.LastInsertId()
	return &model.Node{
		ID:       id,
		UserID:   userID,
		ParentID: &parent.ID,
		Name:     name,
		Home:     homePath,
		Type:     "folder",
		Size:     0,
		MTime:    now,
		Rev:      1,
		GRev:     1,
		Created:  now,
	}, nil
}

// CreateFile creates a new file node at the given path with the specified hash and size.
// Returns the created node.
func (s *NodeStore) CreateFile(userID int64, homePath, hash string, size int64) (*model.Node, error) {
	homePath = normalizePath(homePath)
	name := path.Base(homePath)
	parentPath := path.Dir(homePath)

	parent, err := s.Get(userID, parentPath)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent folder not found: %s", parentPath)
	}

	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO nodes (user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created)
		 VALUES (?, ?, ?, ?, 'file', ?, ?, ?, 1, 1, '', ?)`,
		userID, parent.ID, name, homePath, size, hash, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	id, _ := res.LastInsertId()
	return &model.Node{
		ID:       id,
		UserID:   userID,
		ParentID: &parent.ID,
		Name:     name,
		Home:     homePath,
		Type:     "file",
		Size:     size,
		Hash:     hash,
		MTime:    now,
		Rev:      1,
		GRev:     1,
		Created:  now,
	}, nil
}

// Delete removes a node (file or folder) at the given path.
// For folders, cascading delete is handled by the ON DELETE CASCADE constraint.
func (s *NodeStore) Delete(userID int64, homePath string) error {
	homePath = normalizePath(homePath)
	_, err := s.db.Exec(
		"DELETE FROM nodes WHERE user_id = ? AND home = ?",
		userID, homePath,
	)
	if err != nil {
		return fmt.Errorf("deleting node: %w", err)
	}
	return nil
}

// Rename changes the name of a node at the given path.
// Returns the updated node.
func (s *NodeStore) Rename(userID int64, homePath, newName string) (*model.Node, error) {
	homePath = normalizePath(homePath)
	parentPath := path.Dir(homePath)
	newHome := path.Join(parentPath, newName)

	now := time.Now().Unix()

	// Update the node itself.
	_, err := s.db.Exec(
		`UPDATE nodes SET name = ?, home = ?, mtime = ? WHERE user_id = ? AND home = ?`,
		newName, newHome, now, userID, homePath,
	)
	if err != nil {
		return nil, fmt.Errorf("renaming node: %w", err)
	}

	// Update children paths if it was a folder.
	if err := s.updateChildPaths(userID, homePath, newHome); err != nil {
		return nil, err
	}

	return s.Get(userID, newHome)
}

// Move moves a node from srcPath to the target folder.
// Returns the updated node.
func (s *NodeStore) Move(userID int64, srcPath, targetFolder string) (*model.Node, error) {
	srcPath = normalizePath(srcPath)
	targetFolder = normalizePath(targetFolder)
	name := path.Base(srcPath)
	newHome := path.Join(targetFolder, name)

	// Find new parent.
	newParent, err := s.Get(userID, targetFolder)
	if err != nil {
		return nil, err
	}
	if newParent == nil {
		return nil, fmt.Errorf("target folder not found: %s", targetFolder)
	}

	now := time.Now().Unix()
	_, err = s.db.Exec(
		`UPDATE nodes SET parent_id = ?, home = ?, mtime = ? WHERE user_id = ? AND home = ?`,
		newParent.ID, newHome, now, userID, srcPath,
	)
	if err != nil {
		return nil, fmt.Errorf("moving node: %w", err)
	}

	// Update children paths if it was a folder.
	if err := s.updateChildPaths(userID, srcPath, newHome); err != nil {
		return nil, err
	}

	return s.Get(userID, newHome)
}

// Copy duplicates a node (and its children for folders) from srcPath to targetFolder.
// Returns the new node.
func (s *NodeStore) Copy(userID int64, srcPath, targetFolder string) (*model.Node, error) {
	srcPath = normalizePath(srcPath)
	targetFolder = normalizePath(targetFolder)

	src, err := s.Get(userID, srcPath)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("source not found: %s", srcPath)
	}

	name := path.Base(srcPath)
	newHome := path.Join(targetFolder, name)

	if src.Type == "file" {
		return s.CreateFile(userID, newHome, src.Hash, src.Size)
	}

	// Copy folder: create the folder, then recursively copy children.
	newFolder, err := s.CreateFolder(userID, newHome)
	if err != nil {
		return nil, err
	}

	if err := s.copyChildren(userID, srcPath, newHome); err != nil {
		return nil, err
	}

	return newFolder, nil
}

// TotalSize returns the total size of all file nodes belonging to the given user.
func (s *NodeStore) TotalSize(userID int64) (int64, error) {
	var total sql.NullInt64
	err := s.db.QueryRow(
		"SELECT COALESCE(SUM(size), 0) FROM nodes WHERE user_id = ? AND node_type = 'file'",
		userID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("calculating total size: %w", err)
	}
	return total.Int64, nil
}

// Exists checks whether a node exists at the given path for the user.
func (s *NodeStore) Exists(userID int64, homePath string) (bool, error) {
	homePath = normalizePath(homePath)
	var exists bool
	err := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM nodes WHERE user_id = ? AND home = ?)",
		userID, homePath,
	).Scan(&exists)
	return exists, err
}

// updateChildPaths recursively updates the home path of all descendants
// when a parent is renamed or moved.
func (s *NodeStore) updateChildPaths(userID int64, oldPrefix, newPrefix string) error {
	// Update all descendants whose home starts with oldPrefix + "/".
	_, err := s.db.Exec(
		`UPDATE nodes SET home = ? || SUBSTR(home, ?) WHERE user_id = ? AND home LIKE ? AND home != ?`,
		newPrefix, len(oldPrefix)+1, userID, oldPrefix+"/%", oldPrefix,
	)
	if err != nil {
		return fmt.Errorf("updating child paths: %w", err)
	}
	return nil
}

// copyChildren recursively copies all children from srcFolder to dstFolder.
func (s *NodeStore) copyChildren(userID int64, srcFolder, dstFolder string) error {
	children, err := s.ListChildren(userID, srcFolder, 0, 65535)
	if err != nil {
		return err
	}

	for _, child := range children {
		newHome := path.Join(dstFolder, child.Name)
		if child.Type == "file" {
			if _, err := s.CreateFile(userID, newHome, child.Hash, child.Size); err != nil {
				return err
			}
		} else {
			if _, err := s.CreateFolder(userID, newHome); err != nil {
				return err
			}
			if err := s.copyChildren(userID, child.Home, newHome); err != nil {
				return err
			}
		}
	}
	return nil
}

// normalizePath cleans the path and ensures it starts with "/".
func normalizePath(p string) string {
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/"
	}
	p = path.Clean(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
