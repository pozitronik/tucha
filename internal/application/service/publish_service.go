package service

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/repository"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// PublishService handles publishing/unpublishing nodes via weblinks,
// listing published items, resolving weblinks, and cross-user cloning.
type PublishService struct {
	nodes    repository.NodeRepository
	contents repository.ContentRepository
}

// NewPublishService creates a new PublishService.
func NewPublishService(
	nodes repository.NodeRepository,
	contents repository.ContentRepository,
) *PublishService {
	return &PublishService{
		nodes:    nodes,
		contents: contents,
	}
}

// Publish assigns a public weblink to the node at the given path.
// If the node is already published, returns the existing weblink.
func (s *PublishService) Publish(userID int64, path vo.CloudPath) (string, error) {
	node, err := s.nodes.Get(userID, path)
	if err != nil {
		return "", err
	}
	if node == nil {
		return "", ErrNotFound
	}

	if node.Weblink != "" {
		return node.Weblink, nil
	}

	weblink, err := vo.GenerateWeblink()
	if err != nil {
		return "", err
	}

	if err := s.nodes.SetWeblink(userID, path, weblink); err != nil {
		return "", err
	}

	return weblink, nil
}

// Unpublish removes the public weblink from the node identified by its weblink value.
func (s *PublishService) Unpublish(userID int64, weblinkID string) error {
	node, err := s.nodes.GetByWeblink(weblinkID)
	if err != nil {
		return err
	}
	if node == nil {
		return ErrNotFound
	}
	if node.UserID != userID {
		return ErrForbidden
	}

	return s.nodes.SetWeblink(userID, node.Home, "")
}

// ListPublished returns all nodes with active weblinks for the given user.
func (s *PublishService) ListPublished(userID int64) ([]entity.Node, error) {
	return s.nodes.ListByWeblink(userID)
}

// ResolveWeblink finds a published node by its weblink identifier.
// If subPath is non-empty, resolves a descendant relative to the published folder.
func (s *PublishService) ResolveWeblink(weblinkID string, subPath string) (*entity.Node, error) {
	node, err := s.nodes.GetByWeblink(weblinkID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrNotFound
	}

	if subPath == "" || subPath == "/" {
		return node, nil
	}

	// Resolve subpath relative to the published node.
	target := node.Home.Join(subPath)
	child, err := s.nodes.Get(node.UserID, target)
	if err != nil {
		return nil, err
	}
	if child == nil {
		return nil, ErrNotFound
	}

	return child, nil
}

// Clone copies a published node into the caller's tree.
// Callers cannot clone their own published items.
func (s *PublishService) Clone(callerUserID int64, weblinkID string, targetFolder vo.CloudPath, conflict vo.ConflictMode) (*entity.Node, error) {
	source, err := s.nodes.GetByWeblink(weblinkID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, ErrNotFound
	}

	if source.UserID == callerUserID {
		return nil, ErrForbidden
	}

	if err := s.nodes.EnsurePath(callerUserID, targetFolder); err != nil {
		return nil, err
	}

	targetPath := targetFolder.Join(source.Name)

	// Handle conflict at target path.
	exists, err := s.nodes.Exists(callerUserID, targetPath)
	if err != nil {
		return nil, err
	}
	if exists {
		if conflict == vo.ConflictStrict {
			return nil, ErrAlreadyExists
		}
		if err := s.nodes.Delete(callerUserID, targetPath); err != nil {
			return nil, err
		}
	}

	if source.IsFile() {
		// Increment content ref count for the cloned file.
		if !source.Hash.IsZero() {
			if _, err := s.contents.Insert(source.Hash, source.Size); err != nil {
				return nil, err
			}
		}
		return s.nodes.CreateFile(callerUserID, targetPath, source.Hash, source.Size)
	}

	// For folders: create the folder, then recursively copy children.
	newFolder, err := s.nodes.CreateFolder(callerUserID, targetPath)
	if err != nil {
		return nil, err
	}

	if err := s.cloneChildren(source.UserID, source.Home, callerUserID, targetPath); err != nil {
		return nil, err
	}

	return newFolder, nil
}

// cloneChildren recursively copies children from one user's tree to another.
func (s *PublishService) cloneChildren(srcUserID int64, srcFolder vo.CloudPath, dstUserID int64, dstFolder vo.CloudPath) error {
	children, err := s.nodes.ListChildren(srcUserID, srcFolder, 0, 65535)
	if err != nil {
		return err
	}

	for _, child := range children {
		newHome := dstFolder.Join(child.Name)
		if child.IsFile() {
			if !child.Hash.IsZero() {
				if _, err := s.contents.Insert(child.Hash, child.Size); err != nil {
					return err
				}
			}
			if _, err := s.nodes.CreateFile(dstUserID, newHome, child.Hash, child.Size); err != nil {
				return err
			}
		} else {
			if _, err := s.nodes.CreateFolder(dstUserID, newHome); err != nil {
				return err
			}
			if err := s.cloneChildren(srcUserID, child.Home, dstUserID, newHome); err != nil {
				return err
			}
		}
	}
	return nil
}
