package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func setupTrashHandlerAuth() (*service.AuthService, *entity.Token, *entity.User) {
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

func TestNewTrashHandler(t *testing.T) {
	authSvc, _, _ := setupTrashHandlerAuth()
	trashSvc := service.NewTrashService(
		&mock.NodeRepositoryMock{},
		&mock.TrashRepositoryMock{},
		&mock.ContentRepositoryMock{},
		&mock.ContentStorageMock{},
	)
	presenter := NewPresenter()

	handler := NewTrashHandler(authSvc, trashSvc, presenter)

	if handler == nil {
		t.Fatal("NewTrashHandler() returned nil")
	}
}

func TestTrashHandler_HandleTrashList(t *testing.T) {
	t.Run("returns 405 for non-GET methods", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/api/v2/trashbin?access_token=valid-token", nil)
			w := httptest.NewRecorder()

			handler.HandleTrashList(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleTrashList() with %s status = %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		}
	})

	t.Run("returns 403 for unauthenticated request", func(t *testing.T) {
		tokenRepo := &mock.TokenRepositoryMock{}
		userRepo := &mock.UserRepositoryMock{}
		authSvc := service.NewAuthService(tokenRepo, userRepo)
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodGet, "/api/v2/trashbin", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashList(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("HandleTrashList() status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("returns empty list when trash is empty", func(t *testing.T) {
		authSvc, _, testUser := setupTrashHandlerAuth()
		trashRepo := &mock.TrashRepositoryMock{
			ListFunc: func(userID int64) ([]entity.TrashItem, error) {
				return []entity.TrashItem{}, nil
			},
		}
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			trashRepo,
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodGet, "/api/v2/trashbin?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashList(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleTrashList() status = %d, want %d", w.Code, http.StatusOK)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Email != testUser.Email {
			t.Errorf("env.Email = %q, want %q", env.Email, testUser.Email)
		}

		body, _ := env.Body.(map[string]interface{})
		list, _ := body["list"].([]interface{})
		if len(list) != 0 {
			t.Errorf("list length = %d, want 0", len(list))
		}
	})

	t.Run("returns trash items", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashRepo := &mock.TrashRepositoryMock{
			ListFunc: func(userID int64) ([]entity.TrashItem, error) {
				return []entity.TrashItem{
					{
						ID:          1,
						UserID:      userID,
						Name:        "deleted.txt",
						Home:        vo.NewCloudPath("/deleted.txt"),
						Type:        vo.NodeTypeFile,
						DeletedAt:   time.Now().Unix(),
						DeletedFrom: "/",
					},
				}, nil
			},
		}
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			trashRepo,
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodGet, "/api/v2/trashbin?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashList(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleTrashList() status = %d, want %d", w.Code, http.StatusOK)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		body, _ := env.Body.(map[string]interface{})
		list, _ := body["list"].([]interface{})
		if len(list) != 1 {
			t.Errorf("list length = %d, want 1", len(list))
		}
	})
}

func TestTrashHandler_HandleTrashRestore(t *testing.T) {
	t.Run("returns 405 for non-POST methods", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodGet, "/api/v2/trashbin/restore?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashRestore(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleTrashRestore() status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("returns 400 for missing parameters", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		// Missing both path and restore_revision
		form := url.Values{}
		req := httptest.NewRequest(http.MethodPost, "/api/v2/trashbin/restore?access_token=valid-token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleTrashRestore(w, req)

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 400 {
			t.Errorf("env.Status = %d, want 400", env.Status)
		}
	})

	t.Run("returns 400 for invalid revision", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		form := url.Values{}
		form.Set("path", "/deleted.txt")
		form.Set("restore_revision", "not-a-number")
		req := httptest.NewRequest(http.MethodPost, "/api/v2/trashbin/restore?access_token=valid-token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleTrashRestore(w, req)

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 400 {
			t.Errorf("env.Status = %d, want 400", env.Status)
		}
	})

	t.Run("returns 404 for nonexistent item", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashRepo := &mock.TrashRepositoryMock{
			GetByPathAndRevFunc: func(userID int64, path vo.CloudPath, rev int64) (*entity.TrashItem, error) {
				return nil, nil // Not found
			},
		}
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			trashRepo,
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		form := url.Values{}
		form.Set("path", "/nonexistent.txt")
		form.Set("restore_revision", "123")
		req := httptest.NewRequest(http.MethodPost, "/api/v2/trashbin/restore?access_token=valid-token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.HandleTrashRestore(w, req)

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 404 {
			t.Errorf("env.Status = %d, want 404", env.Status)
		}
	})
}

func TestTrashHandler_HandleTrashEmpty(t *testing.T) {
	t.Run("returns 405 for non-POST methods", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			&mock.TrashRepositoryMock{},
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodGet, "/api/v2/trashbin/empty?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashEmpty(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleTrashEmpty() status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("empties trash successfully", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashRepo := &mock.TrashRepositoryMock{
			DeleteAllFunc: func(userID int64) ([]entity.TrashItem, error) {
				return []entity.TrashItem{}, nil
			},
		}
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			trashRepo,
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodPost, "/api/v2/trashbin/empty?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashEmpty(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleTrashEmpty() status = %d, want %d", w.Code, http.StatusOK)
		}

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 200 {
			t.Errorf("env.Status = %d, want 200", env.Status)
		}
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		authSvc, _, _ := setupTrashHandlerAuth()
		trashRepo := &mock.TrashRepositoryMock{
			DeleteAllFunc: func(userID int64) ([]entity.TrashItem, error) {
				return nil, errors.New("database error")
			},
		}
		trashSvc := service.NewTrashService(
			&mock.NodeRepositoryMock{},
			trashRepo,
			&mock.ContentRepositoryMock{},
			&mock.ContentStorageMock{},
		)
		handler := NewTrashHandler(authSvc, trashSvc, NewPresenter())

		req := httptest.NewRequest(http.MethodPost, "/api/v2/trashbin/empty?access_token=valid-token", nil)
		w := httptest.NewRecorder()

		handler.HandleTrashEmpty(w, req)

		var env Envelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if env.Status != 500 {
			t.Errorf("env.Status = %d, want 500", env.Status)
		}
	})
}
