package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// FileVersionRepository implements repository.FileVersionRepository using SQLite.
type FileVersionRepository struct {
	db *sql.DB
}

// NewFileVersionRepository creates a FileVersionRepository from the given database connection.
func NewFileVersionRepository(db *DB) *FileVersionRepository {
	return &FileVersionRepository{db: db.Conn()}
}

// Insert records a new file version entry.
func (r *FileVersionRepository) Insert(version *entity.FileVersion) error {
	now := time.Now().Unix()
	_, err := r.db.Exec(
		`INSERT INTO file_versions (user_id, home, name, hash, size, rev, time) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		version.UserID, version.Home.String(), version.Name, version.Hash.String(), version.Size, version.Rev, now,
	)
	if err != nil {
		return fmt.Errorf("inserting file version: %w", err)
	}
	return nil
}

// ListByPath returns all version entries for the given user and path, ordered by time ascending.
func (r *FileVersionRepository) ListByPath(userID int64, path vo.CloudPath) ([]entity.FileVersion, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, home, name, hash, size, rev, time FROM file_versions WHERE user_id = ? AND home = ? ORDER BY time ASC`,
		userID, path.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("listing file versions: %w", err)
	}
	defer rows.Close()

	var versions []entity.FileVersion
	for rows.Next() {
		var v entity.FileVersion
		var homePath, hashStr string
		if err := rows.Scan(&v.ID, &v.UserID, &homePath, &v.Name, &hashStr, &v.Size, &v.Rev, &v.Time); err != nil {
			return nil, fmt.Errorf("scanning file version: %w", err)
		}
		v.Home = vo.NewCloudPath(homePath)
		v.Hash = vo.MustContentHash(hashStr)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
