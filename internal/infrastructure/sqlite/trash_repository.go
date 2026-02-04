package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// TrashRepository implements repository.TrashRepository using SQLite.
type TrashRepository struct {
	db *sql.DB
}

// NewTrashRepository creates a TrashRepository from the given database connection.
func NewTrashRepository(db *DB) *TrashRepository {
	return &TrashRepository{db: db.Conn()}
}

// Insert copies a node and its descendants into the trash table.
func (r *TrashRepository) Insert(userID int64, node *entity.Node, descendants []entity.Node, deletedBy int64) error {
	now := time.Now().Unix()
	deletedFrom := node.Home.Parent().String()

	// Insert the root node being trashed.
	if err := r.insertOne(userID, node, deletedFrom, deletedBy, now); err != nil {
		return err
	}

	// Insert all descendants.
	for i := range descendants {
		if err := r.insertOne(userID, &descendants[i], deletedFrom, deletedBy, now); err != nil {
			return err
		}
	}

	return nil
}

// insertOne inserts a single node into the trash table.
func (r *TrashRepository) insertOne(userID int64, node *entity.Node, deletedFrom string, deletedBy int64, deletedAt int64) error {
	var hashStr *string
	if !node.Hash.IsZero() {
		s := node.Hash.String()
		hashStr = &s
	}

	_, err := r.db.Exec(
		`INSERT INTO trash (user_id, name, home, node_type, size, hash, mtime, rev, grev, tree, deleted_at, deleted_from, deleted_by, created)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, node.Name, node.Home.String(), node.Type.String(),
		node.Size, hashStr, node.MTime, node.Rev, node.GRev, node.Tree,
		deletedAt, deletedFrom, deletedBy, node.Created,
	)
	if err != nil {
		return fmt.Errorf("inserting trash item: %w", err)
	}
	return nil
}

// List returns all trash items for a given user.
func (r *TrashRepository) List(userID int64) ([]entity.TrashItem, error) {
	rows, err := r.db.Query(
		`SELECT `+trashColumns+` FROM trash WHERE user_id = ? ORDER BY deleted_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing trash: %w", err)
	}
	defer rows.Close()

	var items []entity.TrashItem
	for rows.Next() {
		item, err := scanTrashItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning trash item: %w", err)
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// GetByPathAndRev finds a specific trash item by its original path and revision.
func (r *TrashRepository) GetByPathAndRev(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
	row := r.db.QueryRow(
		`SELECT `+trashColumns+` FROM trash WHERE user_id = ? AND home = ? AND rev = ?`,
		userID, path.String(), rev,
	)
	item, err := scanTrashItem(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting trash item: %w", err)
	}
	return item, nil
}

// Delete removes a single trash item by ID.
func (r *TrashRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM trash WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting trash item: %w", err)
	}
	return nil
}

// DeleteAll removes all trash items for a user and returns the deleted items
// so callers can clean up associated content.
func (r *TrashRepository) DeleteAll(userID int64) ([]entity.TrashItem, error) {
	// Fetch all items first so we can return them for content cleanup.
	items, err := r.List(userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec("DELETE FROM trash WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("emptying trash: %w", err)
	}

	return items, nil
}

// trashColumns is the standard column list for trash queries.
const trashColumns = `id, user_id, name, home, node_type, size, hash, mtime, rev, grev, tree, deleted_at, deleted_from, deleted_by, created`

// scanTrashItem scans a trash row into an entity.TrashItem.
func scanTrashItem(s interface{ Scan(...any) error }) (*entity.TrashItem, error) {
	var (
		item     entity.TrashItem
		hash     sql.NullString
		home     string
		nodeType string
	)

	err := s.Scan(
		&item.ID, &item.UserID, &item.Name, &home, &nodeType,
		&item.Size, &hash, &item.MTime, &item.Rev, &item.GRev, &item.Tree,
		&item.DeletedAt, &item.DeletedFrom, &item.DeletedBy, &item.Created,
	)
	if err != nil {
		return nil, err
	}

	item.Home = vo.NewCloudPath(home)
	item.Type, _ = vo.ParseNodeType(nodeType)
	if hash.Valid {
		item.Hash = vo.MustContentHash(hash.String)
	}

	return &item, nil
}
