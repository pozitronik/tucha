// Package mock provides callback-based test doubles for repository and port interfaces.
package mock

import (
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
)

// -- UserRepositoryMock --

// UserRepositoryMock is a test double for repository.UserRepository.
type UserRepositoryMock struct {
	UpsertFunc     func(email, password string, isAdmin bool, quotaBytes int64) (int64, error)
	CreateFunc     func(user *entity.User) (int64, error)
	GetByIDFunc    func(id int64) (*entity.User, error)
	GetByEmailFunc func(email string) (*entity.User, error)
	ListFunc       func() ([]entity.User, error)
	UpdateFunc     func(user *entity.User) error
	DeleteFunc     func(id int64) error
}

func (m *UserRepositoryMock) Upsert(email, password string, isAdmin bool, quotaBytes int64) (int64, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(email, password, isAdmin, quotaBytes)
	}
	return 0, nil
}

func (m *UserRepositoryMock) Create(user *entity.User) (int64, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return 0, nil
}

func (m *UserRepositoryMock) GetByID(id int64) (*entity.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *UserRepositoryMock) GetByEmail(email string) (*entity.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, nil
}

func (m *UserRepositoryMock) List() ([]entity.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

func (m *UserRepositoryMock) Update(user *entity.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
}

func (m *UserRepositoryMock) Delete(id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

// -- NodeRepositoryMock --

// NodeRepositoryMock is a test double for repository.NodeRepository.
type NodeRepositoryMock struct {
	GetFunc                func(userID int64, path vo.CloudPath) (*entity.Node, error)
	ListChildrenFunc       func(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error)
	CountChildrenFunc      func(userID int64, path vo.CloudPath) (int, int, error)
	CreateRootNodeFunc     func(userID int64) (*entity.Node, error)
	CreateFolderFunc       func(userID int64, path vo.CloudPath) (*entity.Node, error)
	CreateFileFunc         func(userID int64, path vo.CloudPath, hash vo.ContentHash, size int64) (*entity.Node, error)
	DeleteFunc             func(userID int64, path vo.CloudPath) error
	RenameFunc             func(userID int64, path vo.CloudPath, newName string) (*entity.Node, error)
	MoveFunc               func(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error)
	CopyFunc               func(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error)
	EnsurePathFunc         func(userID int64, path vo.CloudPath) error
	GetWithDescendantsFunc func(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error)
	TotalSizeFunc          func(userID int64) (int64, error)
	SetWeblinkFunc         func(userID int64, path vo.CloudPath, weblink string) error
	GetByWeblinkFunc       func(weblink string) (*entity.Node, error)
	ListByWeblinkFunc      func(userID int64) ([]entity.Node, error)
	ExistsFunc             func(userID int64, path vo.CloudPath) (bool, error)
}

func (m *NodeRepositoryMock) Get(userID int64, path vo.CloudPath) (*entity.Node, error) {
	if m.GetFunc != nil {
		return m.GetFunc(userID, path)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) ListChildren(userID int64, path vo.CloudPath, offset, limit int) ([]entity.Node, error) {
	if m.ListChildrenFunc != nil {
		return m.ListChildrenFunc(userID, path, offset, limit)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) CountChildren(userID int64, path vo.CloudPath) (int, int, error) {
	if m.CountChildrenFunc != nil {
		return m.CountChildrenFunc(userID, path)
	}
	return 0, 0, nil
}

func (m *NodeRepositoryMock) CreateRootNode(userID int64) (*entity.Node, error) {
	if m.CreateRootNodeFunc != nil {
		return m.CreateRootNodeFunc(userID)
	}
	return &entity.Node{ID: 1, UserID: userID, Name: "", Home: vo.NewCloudPath("/"), Type: vo.NodeTypeFolder}, nil
}

func (m *NodeRepositoryMock) CreateFolder(userID int64, path vo.CloudPath) (*entity.Node, error) {
	if m.CreateFolderFunc != nil {
		return m.CreateFolderFunc(userID, path)
	}
	return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFolder}, nil
}

func (m *NodeRepositoryMock) CreateFile(userID int64, path vo.CloudPath, hash vo.ContentHash, size int64) (*entity.Node, error) {
	if m.CreateFileFunc != nil {
		return m.CreateFileFunc(userID, path, hash, size)
	}
	return &entity.Node{ID: 1, UserID: userID, Home: path, Name: path.Name(), Type: vo.NodeTypeFile, Hash: hash, Size: size}, nil
}

func (m *NodeRepositoryMock) Delete(userID int64, path vo.CloudPath) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(userID, path)
	}
	return nil
}

func (m *NodeRepositoryMock) Rename(userID int64, path vo.CloudPath, newName string) (*entity.Node, error) {
	if m.RenameFunc != nil {
		return m.RenameFunc(userID, path, newName)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) Move(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	if m.MoveFunc != nil {
		return m.MoveFunc(userID, srcPath, targetFolder)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) Copy(userID int64, srcPath, targetFolder vo.CloudPath) (*entity.Node, error) {
	if m.CopyFunc != nil {
		return m.CopyFunc(userID, srcPath, targetFolder)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) EnsurePath(userID int64, path vo.CloudPath) error {
	if m.EnsurePathFunc != nil {
		return m.EnsurePathFunc(userID, path)
	}
	return nil
}

func (m *NodeRepositoryMock) GetWithDescendants(userID int64, path vo.CloudPath) (*entity.Node, []entity.Node, error) {
	if m.GetWithDescendantsFunc != nil {
		return m.GetWithDescendantsFunc(userID, path)
	}
	return nil, nil, nil
}

func (m *NodeRepositoryMock) TotalSize(userID int64) (int64, error) {
	if m.TotalSizeFunc != nil {
		return m.TotalSizeFunc(userID)
	}
	return 0, nil
}

func (m *NodeRepositoryMock) SetWeblink(userID int64, path vo.CloudPath, weblink string) error {
	if m.SetWeblinkFunc != nil {
		return m.SetWeblinkFunc(userID, path, weblink)
	}
	return nil
}

func (m *NodeRepositoryMock) GetByWeblink(weblink string) (*entity.Node, error) {
	if m.GetByWeblinkFunc != nil {
		return m.GetByWeblinkFunc(weblink)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) ListByWeblink(userID int64) ([]entity.Node, error) {
	if m.ListByWeblinkFunc != nil {
		return m.ListByWeblinkFunc(userID)
	}
	return nil, nil
}

func (m *NodeRepositoryMock) Exists(userID int64, path vo.CloudPath) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(userID, path)
	}
	return false, nil
}

// -- TokenRepositoryMock --

// TokenRepositoryMock is a test double for repository.TokenRepository.
type TokenRepositoryMock struct {
	CreateFunc       func(userID int64, ttlSeconds int) (*entity.Token, error)
	LookupAccessFunc func(accessToken string) (*entity.Token, error)
	DeleteFunc       func(id int64) error
}

func (m *TokenRepositoryMock) Create(userID int64, ttlSeconds int) (*entity.Token, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(userID, ttlSeconds)
	}
	return &entity.Token{ID: 1, UserID: userID, AccessToken: "test-access", RefreshToken: "test-refresh"}, nil
}

func (m *TokenRepositoryMock) LookupAccess(accessToken string) (*entity.Token, error) {
	if m.LookupAccessFunc != nil {
		return m.LookupAccessFunc(accessToken)
	}
	return nil, nil
}

func (m *TokenRepositoryMock) Delete(id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

// -- ContentRepositoryMock --

// ContentRepositoryMock is a test double for repository.ContentRepository.
type ContentRepositoryMock struct {
	ExistsFunc    func(hash vo.ContentHash) (bool, error)
	InsertFunc    func(hash vo.ContentHash, size int64) (bool, error)
	DecrementFunc func(hash vo.ContentHash) (bool, error)
}

func (m *ContentRepositoryMock) Exists(hash vo.ContentHash) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(hash)
	}
	return false, nil
}

func (m *ContentRepositoryMock) Insert(hash vo.ContentHash, size int64) (bool, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(hash, size)
	}
	return true, nil
}

func (m *ContentRepositoryMock) Decrement(hash vo.ContentHash) (bool, error) {
	if m.DecrementFunc != nil {
		return m.DecrementFunc(hash)
	}
	return false, nil
}

// -- TrashRepositoryMock --

// TrashRepositoryMock is a test double for repository.TrashRepository.
type TrashRepositoryMock struct {
	InsertFunc          func(userID int64, node *entity.Node, descendants []entity.Node, deletedBy int64) error
	ListFunc            func(userID int64) ([]entity.TrashItem, error)
	GetByPathAndRevFunc func(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error)
	DeleteFunc          func(id int64) error
	DeleteAllFunc       func(userID int64) ([]entity.TrashItem, error)
}

func (m *TrashRepositoryMock) Insert(userID int64, node *entity.Node, descendants []entity.Node, deletedBy int64) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(userID, node, descendants, deletedBy)
	}
	return nil
}

func (m *TrashRepositoryMock) List(userID int64) ([]entity.TrashItem, error) {
	if m.ListFunc != nil {
		return m.ListFunc(userID)
	}
	return nil, nil
}

func (m *TrashRepositoryMock) GetByPathAndRev(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
	if m.GetByPathAndRevFunc != nil {
		return m.GetByPathAndRevFunc(userID, path, rev)
	}
	return nil, nil
}

func (m *TrashRepositoryMock) Delete(id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *TrashRepositoryMock) DeleteAll(userID int64) ([]entity.TrashItem, error) {
	if m.DeleteAllFunc != nil {
		return m.DeleteAllFunc(userID)
	}
	return nil, nil
}

// -- FileVersionRepositoryMock --

// FileVersionRepositoryMock is a test double for repository.FileVersionRepository.
type FileVersionRepositoryMock struct {
	InsertFunc     func(version *entity.FileVersion) error
	ListByPathFunc func(userID int64, path vo.CloudPath) ([]entity.FileVersion, error)
}

func (m *FileVersionRepositoryMock) Insert(version *entity.FileVersion) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(version)
	}
	return nil
}

func (m *FileVersionRepositoryMock) ListByPath(userID int64, path vo.CloudPath) ([]entity.FileVersion, error) {
	if m.ListByPathFunc != nil {
		return m.ListByPathFunc(userID, path)
	}
	return nil, nil
}

// -- ShareRepositoryMock --

// ShareRepositoryMock is a test double for repository.ShareRepository.
type ShareRepositoryMock struct {
	CreateFunc                func(share *entity.Share) (int64, error)
	GetByOwnerPathEmailFunc   func(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error)
	GetByInviteTokenFunc      func(token string) (*entity.Share, error)
	DeleteFunc                func(id int64) error
	ListByOwnerPathFunc       func(ownerID int64, home vo.CloudPath) ([]entity.Share, error)
	ListByOwnerPathPrefixFunc func(ownerID int64, path vo.CloudPath) ([]entity.Share, error)
	ListIncomingFunc          func(email string) ([]entity.Share, error)
	AcceptFunc                func(inviteToken string, mountUserID int64, mountHome string) error
	RejectFunc                func(inviteToken string) error
	UnmountFunc               func(userID int64, mountHome string) (*entity.Share, error)
	GetByMountPathFunc        func(userID int64, mountHome string) (*entity.Share, error)
	ListMountedByUserFunc     func(userID int64) ([]entity.Share, error)
}

func (m *ShareRepositoryMock) Create(share *entity.Share) (int64, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(share)
	}
	return 1, nil
}

func (m *ShareRepositoryMock) GetByOwnerPathEmail(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
	if m.GetByOwnerPathEmailFunc != nil {
		return m.GetByOwnerPathEmailFunc(ownerID, home, email)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) GetByInviteToken(token string) (*entity.Share, error) {
	if m.GetByInviteTokenFunc != nil {
		return m.GetByInviteTokenFunc(token)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) Delete(id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *ShareRepositoryMock) ListByOwnerPath(ownerID int64, home vo.CloudPath) ([]entity.Share, error) {
	if m.ListByOwnerPathFunc != nil {
		return m.ListByOwnerPathFunc(ownerID, home)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) ListByOwnerPathPrefix(ownerID int64, path vo.CloudPath) ([]entity.Share, error) {
	if m.ListByOwnerPathPrefixFunc != nil {
		return m.ListByOwnerPathPrefixFunc(ownerID, path)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) ListIncoming(email string) ([]entity.Share, error) {
	if m.ListIncomingFunc != nil {
		return m.ListIncomingFunc(email)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) Accept(inviteToken string, mountUserID int64, mountHome string) error {
	if m.AcceptFunc != nil {
		return m.AcceptFunc(inviteToken, mountUserID, mountHome)
	}
	return nil
}

func (m *ShareRepositoryMock) Reject(inviteToken string) error {
	if m.RejectFunc != nil {
		return m.RejectFunc(inviteToken)
	}
	return nil
}

func (m *ShareRepositoryMock) Unmount(userID int64, mountHome string) (*entity.Share, error) {
	if m.UnmountFunc != nil {
		return m.UnmountFunc(userID, mountHome)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) GetByMountPath(userID int64, mountHome string) (*entity.Share, error) {
	if m.GetByMountPathFunc != nil {
		return m.GetByMountPathFunc(userID, mountHome)
	}
	return nil, nil
}

func (m *ShareRepositoryMock) ListMountedByUser(userID int64) ([]entity.Share, error) {
	if m.ListMountedByUserFunc != nil {
		return m.ListMountedByUserFunc(userID)
	}
	return nil, nil
}
