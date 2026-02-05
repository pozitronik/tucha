package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestAuthenticate(t *testing.T) {
	t.Run("returns authenticated user for valid token", func(t *testing.T) {
		testUser := mock.NewTestUser(1, "user@example.com")
		testToken := mock.NewTestToken(1, time.Now().Add(time.Hour))

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				if accessToken == "valid-token" {
					return testToken, nil
				}
				return nil, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				if id == testUser.ID {
					return testUser, nil
				}
				return nil, nil
			},
		}
		authSvc := service.NewAuthService(tokenRepo, userRepo)

		req := httptest.NewRequest(http.MethodGet, "/test?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		authed := authenticate(w, req, authSvc)

		if authed == nil {
			t.Fatal("authenticate() returned nil for valid token")
		}
		if authed.UserID != testUser.ID {
			t.Errorf("authed.UserID = %d, want %d", authed.UserID, testUser.ID)
		}
		if authed.Email != testUser.Email {
			t.Errorf("authed.Email = %q, want %q", authed.Email, testUser.Email)
		}

		// Response should not be written for successful auth
		if w.Code != http.StatusOK {
			t.Errorf("Response status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("returns nil and writes 403 for missing token", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		authed := authenticate(w, req, authSvc)

		if authed != nil {
			t.Error("authenticate() should return nil for missing token")
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Response status = %d, want %d", w.Code, http.StatusForbidden)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if env.Status != 403 {
			t.Errorf("env.Status = %d, want 403", env.Status)
		}
	})

	t.Run("returns nil and writes 403 for invalid token", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return nil, nil // Token not found
			},
		}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)

		req := httptest.NewRequest(http.MethodGet, "/test?access_token=invalid-token", nil)
		w := httptest.NewRecorder()

		authed := authenticate(w, req, authSvc)

		if authed != nil {
			t.Error("authenticate() should return nil for invalid token")
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Response status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns nil and writes 403 for expired token", func(t *testing.T) {
		expiredToken := mock.NewTestToken(1, time.Now().Add(-time.Hour)) // Expired

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return expiredToken, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)

		req := httptest.NewRequest(http.MethodGet, "/test?access_token=expired", nil)
		w := httptest.NewRecorder()

		authed := authenticate(w, req, authSvc)

		if authed != nil {
			t.Error("authenticate() should return nil for expired token")
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Response status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		testToken := mock.NewTestToken(1, time.Now().Add(time.Hour))

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return testToken, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return nil, nil // User not found
			},
		}
		authSvc := service.NewAuthService(tokenRepo, userRepo)

		req := httptest.NewRequest(http.MethodGet, "/test?access_token=token", nil)
		w := httptest.NewRecorder()

		authed := authenticate(w, req, authSvc)

		if authed != nil {
			t.Error("authenticate() should return nil when user not found")
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Response status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}
