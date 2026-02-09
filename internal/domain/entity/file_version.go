package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// FileVersion represents a single version entry in a file's history.
type FileVersion struct {
	ID     int64
	UserID int64
	Home   vo.CloudPath
	Name   string
	Hash   vo.ContentHash
	Size   int64
	Rev    int64
	Time   int64
}
