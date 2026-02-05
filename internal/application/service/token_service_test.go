package service

import (
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestTokenService_Create(t *testing.T) {
	svc := NewTokenService(
		&mock.TokenRepositoryMock{
			CreateFunc: func(userID int64, ttlSeconds int) (*entity.Token, error) {
				return &entity.Token{ID: 1, UserID: userID, AccessToken: "at"}, nil
			},
		},
		&mock.UserRepositoryMock{},
	)

	tok, err := svc.Create(42, 3600)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tok.UserID != 42 {
		t.Errorf("UserID = %d, want 42", tok.UserID)
	}
}

func TestTokenService_Authenticate_success(t *testing.T) {
	user := mock.NewTestUser(1, "user@example.com")
	user.Password = "correct"

	svc := NewTokenService(
		&mock.TokenRepositoryMock{
			CreateFunc: func(userID int64, ttlSeconds int) (*entity.Token, error) {
				return &entity.Token{ID: 1, UserID: userID, AccessToken: "new-at"}, nil
			},
		},
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				if email == "user@example.com" {
					return user, nil
				}
				return nil, nil
			},
		},
	)

	tok, err := svc.Authenticate("user@example.com", "correct", 3600)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if tok.AccessToken != "new-at" {
		t.Errorf("AccessToken = %q", tok.AccessToken)
	}
}

func TestTokenService_Authenticate_wrongPassword(t *testing.T) {
	user := mock.NewTestUser(1, "user@example.com")
	user.Password = "correct"

	svc := NewTokenService(
		&mock.TokenRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				return user, nil
			},
		},
	)

	_, err := svc.Authenticate("user@example.com", "wrong", 3600)
	if err != ErrNotFound {
		t.Errorf("Authenticate(wrong password) error = %v, want ErrNotFound", err)
	}
}

func TestTokenService_Authenticate_unknownEmail(t *testing.T) {
	svc := NewTokenService(
		&mock.TokenRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				return nil, nil
			},
		},
	)

	_, err := svc.Authenticate("unknown@example.com", "any", 3600)
	if err != ErrNotFound {
		t.Errorf("Authenticate(unknown email) error = %v, want ErrNotFound", err)
	}
}
