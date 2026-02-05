package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestNewTokenHandler(t *testing.T) {
	tokenRepo := &mock.TokenRepositoryMock{}
	userRepo := &mock.UserRepositoryMock{}
	tokenSvc := service.NewTokenService(tokenRepo, userRepo)
	logger := &mock.LoggerMock{}

	handler := NewTokenHandler(tokenSvc, 3600, logger)

	if handler == nil {
		t.Fatal("NewTokenHandler() returned nil")
	}
	if handler.tokenTTLSeconds != 3600 {
		t.Errorf("tokenTTLSeconds = %d, want 3600", handler.tokenTTLSeconds)
	}
}

func TestTokenHandler_HandleToken(t *testing.T) {
	t.Run("returns 405 for non-POST methods", func(t *testing.T) {
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, &mock.UserRepositoryMock{})
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/token", nil)
			w := httptest.NewRecorder()

			handler.HandleToken(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleToken() with %s status = %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		}
	})

	t.Run("returns error for invalid client_id", func(t *testing.T) {
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, &mock.UserRepositoryMock{})
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		form := url.Values{}
		form.Set("client_id", "wrong-client")
		form.Set("grant_type", "password")
		form.Set("username", "user@example.com")
		form.Set("password", "pass")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleToken() status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp OAuthToken
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != "invalid_client" {
			t.Errorf("resp.Error = %q, want %q", resp.Error, "invalid_client")
		}
		if resp.ErrorCode != 2 {
			t.Errorf("resp.ErrorCode = %d, want 2", resp.ErrorCode)
		}
	})

	t.Run("returns error for unsupported grant_type", func(t *testing.T) {
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, &mock.UserRepositoryMock{})
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		form := url.Values{}
		form.Set("client_id", "cloud-win")
		form.Set("grant_type", "authorization_code")
		form.Set("username", "user@example.com")
		form.Set("password", "pass")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		var resp OAuthToken
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != "unsupported_grant_type" {
			t.Errorf("resp.Error = %q, want %q", resp.Error, "unsupported_grant_type")
		}
		if resp.ErrorCode != 3 {
			t.Errorf("resp.ErrorCode = %d, want 3", resp.ErrorCode)
		}
	})

	t.Run("returns error for invalid credentials", func(t *testing.T) {
		userRepo := &mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				return nil, nil // User not found
			},
		}
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, userRepo)
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		form := url.Values{}
		form.Set("client_id", "cloud-win")
		form.Set("grant_type", "password")
		form.Set("username", "nonexistent@example.com")
		form.Set("password", "wrong")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		var resp OAuthToken
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != "invalid_grant" {
			t.Errorf("resp.Error = %q, want %q", resp.Error, "invalid_grant")
		}
		if resp.ErrorCode != 4 {
			t.Errorf("resp.ErrorCode = %d, want 4", resp.ErrorCode)
		}
	})

	t.Run("returns token for valid credentials", func(t *testing.T) {
		testUser := &entity.User{
			ID:       1,
			Email:    "user@example.com",
			Password: "correctpassword",
		}
		testToken := &entity.Token{
			ID:           1,
			UserID:       1,
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		}

		userRepo := &mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				if email == testUser.Email {
					return testUser, nil
				}
				return nil, nil
			},
		}
		tokenRepo := &mock.TokenRepositoryMock{
			CreateFunc: func(userID int64, ttlSeconds int) (*entity.Token, error) {
				return testToken, nil
			},
		}
		tokenSvc := service.NewTokenService(tokenRepo, userRepo)
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		form := url.Values{}
		form.Set("client_id", "cloud-win")
		form.Set("grant_type", "password")
		form.Set("username", "user@example.com")
		form.Set("password", "correctpassword")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleToken() status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp OAuthToken
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != "" {
			t.Errorf("resp.Error = %q, want empty", resp.Error)
		}
		if resp.AccessToken != testToken.AccessToken {
			t.Errorf("resp.AccessToken = %q, want %q", resp.AccessToken, testToken.AccessToken)
		}
		if resp.RefreshToken != testToken.RefreshToken {
			t.Errorf("resp.RefreshToken = %q, want %q", resp.RefreshToken, testToken.RefreshToken)
		}
		if resp.ExpiresIn != 3600 {
			t.Errorf("resp.ExpiresIn = %d, want 3600", resp.ExpiresIn)
		}
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		testUser := &entity.User{
			ID:       1,
			Email:    "user@example.com",
			Password: "correctpassword",
		}

		userRepo := &mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				if email == testUser.Email {
					return testUser, nil
				}
				return nil, nil
			},
		}
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, userRepo)
		handler := NewTokenHandler(tokenSvc, 3600, &mock.LoggerMock{})

		form := url.Values{}
		form.Set("client_id", "cloud-win")
		form.Set("grant_type", "password")
		form.Set("username", "user@example.com")
		form.Set("password", "wrongpassword")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		var resp OAuthToken
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Error != "invalid_grant" {
			t.Errorf("resp.Error = %q, want %q", resp.Error, "invalid_grant")
		}
	})

	t.Run("logs authentication attempts", func(t *testing.T) {
		logger := &mock.LoggerMock{}
		userRepo := &mock.UserRepositoryMock{
			GetByEmailFunc: func(email string) (*entity.User, error) {
				return nil, nil
			},
		}
		tokenSvc := service.NewTokenService(&mock.TokenRepositoryMock{}, userRepo)
		handler := NewTokenHandler(tokenSvc, 3600, logger)

		form := url.Values{}
		form.Set("client_id", "cloud-win")
		form.Set("grant_type", "password")
		form.Set("username", "test@example.com")
		form.Set("password", "testpass")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleToken(w, req)

		// Verify logging occurred
		hasInfoLog := false
		hasWarnLog := false
		for _, entry := range logger.Captured {
			if entry.Level == "INFO" && strings.Contains(entry.Msg, "Auth attempt") {
				hasInfoLog = true
			}
			if entry.Level == "WARN" && strings.Contains(entry.Msg, "Auth failed") {
				hasWarnLog = true
			}
		}

		if !hasInfoLog {
			t.Error("Expected INFO log for auth attempt")
		}
		if !hasWarnLog {
			t.Error("Expected WARN log for auth failure")
		}
	})
}
