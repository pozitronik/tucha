package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestUserService_Create_success(t *testing.T) {
	var createdUser *entity.User
	var rootNodeCreated bool

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) { return nil, nil },
			CreateFunc: func(user *entity.User) (int64, error) {
				createdUser = user
				return 1, nil
			},
		},
		&mock.NodeRepositoryMock{
			CreateRootNodeFunc: func(userID int64) (*entity.Node, error) {
				rootNodeCreated = true
				return &entity.Node{ID: 1}, nil
			},
		},
		1073741824,
	)

	user, err := svc.Create("new@example.com", "pass", false, 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("ID = %d, want 1", user.ID)
	}
	if !rootNodeCreated {
		t.Error("root node was not created")
	}
	if createdUser.QuotaBytes != 1073741824 {
		t.Errorf("default quota = %d, want 1073741824", createdUser.QuotaBytes)
	}
}

func TestUserService_Create_duplicate(t *testing.T) {
	existing := mock.NewTestUser(1, "dup@example.com")

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) { return existing, nil },
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	_, err := svc.Create("dup@example.com", "pass", false, 0)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("Create(duplicate) error = %v, want ErrAlreadyExists", err)
	}
}

func TestUserService_Create_customQuota(t *testing.T) {
	var capturedQuota int64

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) { return nil, nil },
			CreateFunc: func(user *entity.User) (int64, error) {
				capturedQuota = user.QuotaBytes
				return 1, nil
			},
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	_, err := svc.Create("user@example.com", "pass", false, 5000)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if capturedQuota != 5000 {
		t.Errorf("custom quota = %d, want 5000", capturedQuota)
	}
}

func TestUserService_Create_defaultQuota(t *testing.T) {
	var capturedQuota int64

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) { return nil, nil },
			CreateFunc: func(user *entity.User) (int64, error) {
				capturedQuota = user.QuotaBytes
				return 1, nil
			},
		},
		&mock.NodeRepositoryMock{},
		999,
	)

	// quotaBytes <= 0 should use default.
	_, err := svc.Create("user@example.com", "pass", false, 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if capturedQuota != 999 {
		t.Errorf("default quota = %d, want 999", capturedQuota)
	}
}

func TestUserService_Update_partialFields(t *testing.T) {
	existing := &entity.User{
		ID:         1,
		Email:      "old@example.com",
		Password:   "oldpass",
		IsAdmin:    false,
		QuotaBytes: 1000,
	}

	var updated *entity.User
	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return existing, nil },
			UpdateFunc: func(user *entity.User) error {
				updated = user
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	// Only change password and admin status; email and quota preserved.
	err := svc.Update(&entity.User{ID: 1, Password: "newpass", IsAdmin: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Email != "old@example.com" {
		t.Errorf("Email = %q, want preserved", updated.Email)
	}
	if updated.Password != "newpass" {
		t.Errorf("Password = %q, want %q", updated.Password, "newpass")
	}
	if !updated.IsAdmin {
		t.Error("IsAdmin = false, want true")
	}
	if updated.QuotaBytes != 1000 {
		t.Errorf("QuotaBytes = %d, want preserved 1000", updated.QuotaBytes)
	}
}

func TestUserService_Update_notFound(t *testing.T) {
	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return nil, nil },
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	err := svc.Update(&entity.User{ID: 999})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Update(not found) error = %v, want ErrNotFound", err)
	}
}

func TestUserService_Delete_success(t *testing.T) {
	existing := mock.NewTestUser(1, "user@example.com")
	deleted := false

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return existing, nil },
			DeleteFunc: func(id int64) error {
				deleted = true
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	err := svc.Delete(1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("Delete was not called on repo")
	}
}

func TestUserService_Delete_notFound(t *testing.T) {
	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return nil, nil },
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	err := svc.Delete(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete(not found) error = %v, want ErrNotFound", err)
	}
}

func TestUserService_Update_isAdminAlwaysApplied(t *testing.T) {
	existing := &entity.User{
		ID:         1,
		Email:      "u@example.com",
		Password:   "p",
		IsAdmin:    true,
		QuotaBytes: 1000,
	}
	var updated *entity.User

	svc := NewUserService(
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return existing, nil },
			UpdateFunc: func(user *entity.User) error {
				updated = user
				return nil
			},
		},
		&mock.NodeRepositoryMock{},
		1073741824,
	)

	// Sending IsAdmin=false should override existing true.
	err := svc.Update(&entity.User{ID: 1, IsAdmin: false})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.IsAdmin {
		t.Error("IsAdmin should be overridden to false")
	}
}
