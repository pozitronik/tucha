package sqlite

import (
	"database/sql"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// scanNode scans a node row into an entity.Node.
// The column order must match: id, user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created.
func scanNode(s interface{ Scan(...any) error }) (*entity.Node, error) {
	var (
		n        entity.Node
		parentID sql.NullInt64
		hash     sql.NullString
		home     string
		nodeType string
	)

	err := s.Scan(
		&n.ID, &n.UserID, &parentID,
		&n.Name, &home, &nodeType,
		&n.Size, &hash,
		&n.MTime, &n.Rev, &n.GRev, &n.Tree, &n.Created,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		n.ParentID = &parentID.Int64
	}
	n.Home = vo.NewCloudPath(home)
	n.Type, _ = vo.ParseNodeType(nodeType)
	if hash.Valid {
		n.Hash = vo.MustContentHash(hash.String)
	}

	return &n, nil
}

// nodeColumns is the standard column list for node queries.
const nodeColumns = `id, user_id, parent_id, name, home, node_type, size, hash, mtime, rev, grev, tree, created`
