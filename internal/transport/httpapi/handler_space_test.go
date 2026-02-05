package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestNewSpaceHandler(t *testing.T) {
	tokenRepo := &mock.TokenRepositoryMock{}
	userRepo := &mock.UserRepositoryMock{}
	nodeRepo := &mock.NodeRepositoryMock{}
	authSvc := service.NewAuthService(tokenRepo, userRepo)
	quotaSvc := service.NewQuotaService(nodeRepo, userRepo)

	handler := NewSpaceHandler(authSvc, quotaSvc)

	if handler == nil {
		t.Fatal("NewSpaceHandler() returned nil")
	}
}

func TestSpaceHandler_HandleSpace(t *testing.T) {
	t.Run("returns space info for authenticated user", func(t *testing.T) {
		testUser := &entity.User{
			ID:         1,
			Email:      "user@example.com",
			QuotaBytes: 10737418240, // 10GB
		}
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
		nodeRepo := &mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) {
				return 1073741824, nil // 1GB used
			},
		}

		authSvc := service.NewAuthService(tokenRepo, userRepo)
		quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
		handler := NewSpaceHandler(authSvc, quotaSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/user/space?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleSpace(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleSpace() status = %d, want %d", w.Code, http.StatusOK)
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

		bytesTotal, _ := body["bytes_total"].(float64)
		bytesUsed, _ := body["bytes_used"].(float64)
		overquota, _ := body["overquota"].(bool)

		if int64(bytesTotal) != testUser.QuotaBytes {
			t.Errorf("bytes_total = %v, want %d", bytesTotal, testUser.QuotaBytes)
		}
		if int64(bytesUsed) != 1073741824 {
			t.Errorf("bytes_used = %v, want %d", bytesUsed, 1073741824)
		}
		if overquota {
			t.Error("overquota = true, want false")
		}
	})

	t.Run("returns overquota when usage exceeds quota", func(t *testing.T) {
		testUser := &entity.User{
			ID:         1,
			Email:      "user@example.com",
			QuotaBytes: 1073741824, // 1GB
		}
		testToken := mock.NewTestToken(1, time.Now().Add(time.Hour))

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return testToken, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return testUser, nil
			},
		}
		nodeRepo := &mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) {
				return 2147483648, nil // 2GB used (over 1GB quota)
			},
		}

		authSvc := service.NewAuthService(tokenRepo, userRepo)
		quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
		handler := NewSpaceHandler(authSvc, quotaSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/user/space?access_token=token", nil)
		w := httptest.NewRecorder()

		handler.HandleSpace(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleSpace() status = %d, want %d", w.Code, http.StatusOK)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		body, _ := env.Body.(map[string]interface{})
		overquota, _ := body["overquota"].(bool)

		if !overquota {
			t.Error("overquota = false, want true")
		}
	})

	t.Run("returns 403 for unauthenticated request", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		nodeRepo := &mock.NodeRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
		handler := NewSpaceHandler(authSvc, quotaSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/user/space", nil)
		w := httptest.NewRecorder()

		handler.HandleSpace(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleSpace() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns 500 on quota service error", func(t *testing.T) {
		testUser := mock.NewTestUser(1, "user@example.com")
		testToken := mock.NewTestToken(1, time.Now().Add(time.Hour))

		tokenRepo := &mock.TokenRepositoryMock{
			LookupAccessFunc: func(accessToken string) (*entity.Token, error) {
				return testToken, nil
			},
		}
		userRepo := &mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return testUser, nil
			},
		}
		nodeRepo := &mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) {
				return 0, errors.New("database error")
			},
		}

		authSvc := service.NewAuthService(tokenRepo, userRepo)
		quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
		handler := NewSpaceHandler(authSvc, quotaSvc)

		req := httptest.NewRequest(http.MethodGet, "/api/v2/user/space?access_token=token", nil)
		w := httptest.NewRecorder()

		handler.HandleSpace(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("HandleSpace() status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}
