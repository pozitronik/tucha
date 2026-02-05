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

func TestNewCSRFHandler(t *testing.T) {
	tokenRepo := &mock.TokenRepositoryMock{}
	userRepo := &mock.UserRepositoryMock{}
	authSvc := service.NewAuthService(tokenRepo, userRepo)

	handler := NewCSRFHandler(authSvc)

	if handler == nil {
		t.Fatal("NewCSRFHandler() returned nil")
	}
}

func TestCSRFHandler_HandleCSRF(t *testing.T) {
	t.Run("returns CSRF token for authenticated user", func(t *testing.T) {
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
		handler := NewCSRFHandler(authSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/tokens/csrf?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleCSRF(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleCSRF() status = %d, want %d", w.Code, http.StatusOK)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 200 {
			t.Errorf("env.Status = %d, want 200", env.Status)
		}
		if env.Email != testUser.Email {
			t.Errorf("env.Email = %q, want %q", env.Email, testUser.Email)
		}

		body, ok := env.Body.(map[string]interface{})
		if !ok {
			t.Fatalf("env.Body is not a map")
		}
		csrfToken, ok := body["token"].(string)
		if !ok {
			t.Fatal("body[\"token\"] is not a string")
		}
		if csrfToken != testToken.CSRFToken {
			t.Errorf("CSRF token = %q, want %q", csrfToken, testToken.CSRFToken)
		}
	})

	t.Run("returns 403 for missing token", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		handler := NewCSRFHandler(authSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/tokens/csrf", nil)
		w := httptest.NewRecorder()

		handler.HandleCSRF(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleCSRF() status = %d, want %d", w.Code, http.StatusForbidden)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if env.Status != 403 {
			t.Errorf("env.Status = %d, want 403", env.Status)
		}
	})

	t.Run("returns 403 for invalid token", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return nil, nil // Token not found
			},
		}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		handler := NewCSRFHandler(authSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/tokens/csrf?access_token=invalid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleCSRF(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleCSRF() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns 403 for expired token", func(t *testing.T) {
		expiredToken := mock.NewTestToken(1, time.Now().Add(-time.Hour)) // Expired

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return expiredToken, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		handler := NewCSRFHandler(authSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/tokens/csrf?access_token=expired-token", nil)
		w := httptest.NewRecorder()

		handler.HandleCSRF(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleCSRF() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}
