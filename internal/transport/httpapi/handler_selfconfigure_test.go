package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pozitronik/tucha/internal/config"
)

func TestNewSelfConfigureHandler(t *testing.T) {
	endpoints := config.EndpointsConfig{
		API:        "https://api.example.com",
		OAuth:      "https://oauth.example.com",
		Dispatcher: "https://dispatcher.example.com",
		Upload:     "https://upload.example.com",
		Download:   "https://download.example.com",
	}

	handler := NewSelfConfigureHandler(endpoints)

	if handler == nil {
		t.Fatal("NewSelfConfigureHandler() returned nil")
	}
	if handler.endpoints.API != endpoints.API {
		t.Errorf("handler.endpoints.API = %q, want %q", handler.endpoints.API, endpoints.API)
	}
}

func TestSelfConfigureHandler_HandleSelfConfigure(t *testing.T) {
	endpoints := config.EndpointsConfig{
		API:        "https://api.example.com",
		OAuth:      "https://oauth.example.com",
		Dispatcher: "https://dispatcher.example.com",
		Upload:     "https://upload.example.com",
		Download:   "https://download.example.com",
	}
	handler := NewSelfConfigureHandler(endpoints)

	t.Run("returns config for root path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.HandleSelfConfigure(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleSelfConfigure() status = %d, want %d", w.Code, http.StatusOK)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
		}

		var resp selfConfigureResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.API != endpoints.API {
			t.Errorf("resp.API = %q, want %q", resp.API, endpoints.API)
		}
		if resp.OAuth != endpoints.OAuth {
			t.Errorf("resp.OAuth = %q, want %q", resp.OAuth, endpoints.OAuth)
		}
		if resp.Dispatcher != endpoints.Dispatcher {
			t.Errorf("resp.Dispatcher = %q, want %q", resp.Dispatcher, endpoints.Dispatcher)
		}
		if resp.Upload != endpoints.Upload {
			t.Errorf("resp.Upload = %q, want %q", resp.Upload, endpoints.Upload)
		}
		if resp.Download != endpoints.Download {
			t.Errorf("resp.Download = %q, want %q", resp.Download, endpoints.Download)
		}
	})

	t.Run("returns 404 for non-root paths", func(t *testing.T) {
		paths := []string{"/other", "/api", "/foo/bar", "//"}
		for _, path := range paths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.HandleSelfConfigure(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("HandleSelfConfigure(%q) status = %d, want %d", path, w.Code, http.StatusNotFound)
			}
		}
	})

	t.Run("returns 405 for non-GET methods", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			handler.HandleSelfConfigure(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleSelfConfigure() with %s status = %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		}
	})
}
