package service

import (
	"github.com/pozitronik/tucha/internal/domain/repository"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// cloneTree recursively copies a folder tree from one user to another,
// incrementing content reference counts for each file.
// The destination folder must already exist.
func cloneTree(
	nodes repository.NodeRepository,
	contents repository.ContentRepository,
	srcUserID int64, srcFolder vo.CloudPath,
	dstUserID int64, dstFolder vo.CloudPath,
) error {
	children, err := nodes.ListChildren(srcUserID, srcFolder, 0, 65535)
	if err != nil {
		return err
	}

	for _, child := range children {
		newPath := dstFolder.Join(child.Name)
		if child.IsFile() {
			if !child.Hash.IsZero() {
				if _, err := contents.Insert(child.Hash, child.Size); err != nil {
					return err
				}
			}
			if _, err := nodes.CreateFile(dstUserID, newPath, child.Hash, child.Size); err != nil {
				return err
			}
		} else {
			if _, err := nodes.CreateFolder(dstUserID, newPath); err != nil {
				return err
			}
			if err := cloneTree(nodes, contents, srcUserID, child.Home, dstUserID, newPath); err != nil {
				return err
			}
		}
	}
	return nil
}
