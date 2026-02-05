package service

import (
	"errors"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestAuthService_Validate_success(t *testing.T) {
	token := mock.NewTestToken(1, time.Now().Add(time.Hour))
	user := mock.NewTestUser(1, "user@example.com")

	svc := NewAuthService(
		&mock.TokenRepositoryMock{
			LookupAccessFunc: func(at string) (*entity.Token, error) {
				return token, nil
			},
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return user, nil
			},
		},
	)

	auth, err := svc.Validate(token.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if auth == nil {
		t.Fatal("Validate returned nil")
	}
	if auth.UserID != 1 {
		t.Errorf("UserID = %d, want 1", auth.UserID)
	}
	if auth.Email != "user@example.com" {
		t.Errorf("Email = %q", auth.Email)
	}
	if auth.CSRFToken != token.CSRFToken {
		t.Errorf("CSRFToken = %q, want %q", auth.CSRFToken, token.CSRFToken)
	}
}

func TestAuthService_Validate_emptyToken(t *testing.T) {
	svc := NewAuthService(&mock.TokenRepositoryMock{}, &mock.UserRepositoryMock{})
	auth, err := svc.Validate("")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if auth != nil {
		t.Error("Validate('') should return nil")
	}
}

func TestAuthService_Validate_unknownToken(t *testing.T) {
	svc := NewAuthService(
		&mock.TokenRepositoryMock{
			LookupAccessFunc: func(at string) (*entity.Token, error) {
				return nil, nil
			},
		},
		&mock.UserRepositoryMock{},
	)

	auth, err := svc.Validate("unknown")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if auth != nil {
		t.Error("Validate(unknown) should return nil")
	}
}

func TestAuthService_Validate_expiredToken(t *testing.T) {
	token := mock.NewTestToken(1, time.Now().Add(-time.Hour))
	deleteCalled := false
	deletedID := int64(0)

	svc := NewAuthService(
		&mock.TokenRepositoryMock{
			LookupAccessFunc: func(at string) (*entity.Token, error) {
				return token, nil
			},
			DeleteFunc: func(id int64) error {
				deleteCalled = true
				deletedID = id
				return nil
			},
		},
		&mock.UserRepositoryMock{},
	)

	auth, err := svc.Validate(token.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if auth != nil {
		t.Error("Validate(expired) should return nil")
	}
	if !deleteCalled {
		t.Error("Delete should be called for expired token")
	}
	if deletedID != token.ID {
		t.Errorf("Delete called with ID %d, want %d", deletedID, token.ID)
	}
}

func TestAuthService_Validate_userDeleted(t *testing.T) {
	token := mock.NewTestToken(1, time.Now().Add(time.Hour))

	svc := NewAuthService(
		&mock.TokenRepositoryMock{
			LookupAccessFunc: func(at string) (*entity.Token, error) {
				return token, nil
			},
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return nil, nil
			},
		},
	)

	auth, err := svc.Validate(token.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if auth != nil {
		t.Error("Validate(deleted user) should return nil")
	}
}

func TestAuthService_Validate_repoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := NewAuthService(
		&mock.TokenRepositoryMock{
			LookupAccessFunc: func(at string) (*entity.Token, error) {
				return nil, repoErr
			},
		},
		&mock.UserRepositoryMock{},
	)

	_, err := svc.Validate("token")
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}

func TestAuthService_ResolveUser_found(t *testing.T) {
	user := mock.NewTestUser(1, "user@example.com")
	svc := NewAuthService(
		&mock.TokenRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return user, nil
			},
		},
	)

	auth, err := svc.ResolveUser(1)
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if auth == nil {
		t.Fatal("ResolveUser returned nil")
	}
	if auth.Email != "user@example.com" {
		t.Errorf("Email = %q", auth.Email)
	}
}

func TestAuthService_ResolveUser_notFound(t *testing.T) {
	svc := NewAuthService(
		&mock.TokenRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return nil, nil
			},
		},
	)

	auth, err := svc.ResolveUser(999)
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if auth != nil {
		t.Error("ResolveUser(nonexistent) should return nil")
	}
}

func TestAuthService_ResolveUser_error(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewAuthService(
		&mock.TokenRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return nil, repoErr
			},
		},
	)

	_, err := svc.ResolveUser(1)
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
