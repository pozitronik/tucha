package repository

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// FileVersionRepository persists and retrieves file version history entries.
type FileVersionRepository interface {
	// Insert records a new version entry.
	Insert(version *entity.FileVersion) error

	// ListByPath returns all version entries for the given user and path, ordered by time ascending.
	ListByPath(userID int64, path vo.CloudPath) ([]entity.FileVersion, error)
}
