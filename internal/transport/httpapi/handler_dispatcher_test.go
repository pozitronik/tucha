package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestNewDispatchHandler(t *testing.T) {
	t.Run("trims trailing slash from external URL", func(t *testing.T) {
		handler := NewDispatchHandler(nil, "http://example.com/")
		if handler.externalURL != "http://example.com" {
			t.Errorf("externalURL = %q, want %q", handler.externalURL, "http://example.com")
		}
	})

	t.Run("preserves URL without trailing slash", func(t *testing.T) {
		handler := NewDispatchHandler(nil, "http://example.com")
		if handler.externalURL != "http://example.com" {
			t.Errorf("externalURL = %q, want %q", handler.externalURL, "http://example.com")
		}
	})
}

func TestDispatchHandler_HandleDispatcher(t *testing.T) {
	setupAuth := func() (*service.AuthService, *entity.Token, *entity.User) {
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
				return testUser, nil
			},
		}
		return service.NewAuthService(tokenRepo, userRepo), testToken, testUser
	}

	t.Run("returns shard URLs for authenticated user", func(t *testing.T) {
		authSvc, _, testUser := setupAuth()
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodPost, "/api/v2/dispatcher/?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleDispatcher(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleDispatcher() status = %d, want %d", w.Code, http.StatusOK)
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

		// Verify key shard URLs are present
		expectedShards := []string{"get", "upload", "thumbnails", "weblink_get", "weblink_video", "video"}
		for _, shard := range expectedShards {
			if _, ok := body[shard]; !ok {
				t.Errorf("body missing shard %q", shard)
			}
		}

		// Verify URL format
		getShard, _ := body["get"].([]interface{})
		if len(getShard) == 0 {
			t.Fatal("get shard is empty")
		}
		getURL := getShard[0].(map[string]interface{})["url"].(string)
		if getURL != "http://localhost:8080/get/" {
			t.Errorf("get shard URL = %q, want %q", getURL, "http://localhost:8080/get/")
		}
	})

	t.Run("returns 403 for unauthenticated request", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodPost, "/api/v2/dispatcher/", nil)
		w := httptest.NewRecorder()

		handler.HandleDispatcher(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleDispatcher() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})
}

func TestDispatchHandler_HandleOAuthDispatcher(t *testing.T) {
	setupAuth := func() *service.AuthService {
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
				return testUser, nil
			},
		}
		return service.NewAuthService(tokenRepo, userRepo)
	}

	t.Run("returns download shard for /d endpoint", func(t *testing.T) {
		authSvc := setupAuth()
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodGet, "/d?token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleOAuthDispatcher(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleOAuthDispatcher() status = %d, want %d", w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if !strings.Contains(body, "http://localhost:8080/get/") {
			t.Errorf("Response = %q, want to contain download URL", body)
		}
		if !strings.Contains(body, "127.0.0.1 1") {
			t.Errorf("Response = %q, want to contain '127.0.0.1 1'", body)
		}
	})

	t.Run("returns upload shard for /u endpoint", func(t *testing.T) {
		authSvc := setupAuth()
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodGet, "/u?token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleOAuthDispatcher(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleOAuthDispatcher() status = %d, want %d", w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if !strings.Contains(body, "http://localhost:8080/upload/") {
			t.Errorf("Response = %q, want to contain upload URL", body)
		}
	})

	t.Run("returns 401 for unauthenticated request", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodGet, "/d", nil)
		w := httptest.NewRecorder()

		handler.HandleOAuthDispatcher(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("HandleOAuthDispatcher() status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("returns 404 for unknown path", func(t *testing.T) {
		authSvc := setupAuth()
		handler := NewDispatchHandler(authSvc, "http://localhost:8080")

		req := httptest.NewRequest(http.MethodGet, "/x?token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleOAuthDispatcher(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleOAuthDispatcher() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
