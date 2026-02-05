package httpapi

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/domain/vo"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestIsVideoFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		// Video formats
		{"video.mp4", true},
		{"video.MP4", true},
		{"video.mkv", true},
		{"video.MKV", true},
		{"video.avi", true},
		{"video.mov", true},
		{"video.wmv", true},
		{"video.flv", true},
		{"video.webm", true},
		{"video.m4v", true},
		{"video.mpeg", true},
		{"video.mpg", true},

		// Non-video formats
		{"image.jpg", false},
		{"image.png", false},
		{"document.pdf", false},
		{"audio.mp3", false},
		{"file.txt", false},
		{"noextension", false},
		{"", false},

		// Edge cases
		{".mp4", true},
		{"path/to/video.mp4", true},
		{"file.name.mp4", true},
		{"VIDEO.AVI", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := isVideoFormat(tt.filename)
			if got != tt.want {
				t.Errorf("isVideoFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestVideoHandler_HandleHLSPlaylist(t *testing.T) {
	t.Run("returns 405 for non-GET requests", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodPost, "/video/0p/test.m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("returns 404 for missing 0p prefix", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodGet, "/video/other/test.m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 404 for missing .m3u8 suffix", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodGet, "/video/0p/test", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 400 for invalid base64", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodGet, "/video/0p/!!!invalid!!.m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns 404 for nonexistent weblink", func(t *testing.T) {
		nodeRepo := &mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return nil, nil // Not found
			},
		}
		contentRepo := &mock.ContentRepositoryMock{}
		publishSvc := service.NewPublishService(nodeRepo, contentRepo)

		handler := NewVideoHandler(publishSvc, nil, "http://localhost")

		weblinkID := "seg1/seg2"
		b64Weblink := base64.StdEncoding.EncodeToString([]byte(weblinkID))

		req := httptest.NewRequest(http.MethodGet, "/video/0p/"+b64Weblink+".m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 404 for non-video file", func(t *testing.T) {
		nodeRepo := &mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return &entity.Node{
					ID:      1,
					UserID:  1,
					Name:    "image.jpg",
					Home:    vo.NewCloudPath("/image.jpg"),
					Type:    vo.NodeTypeFile,
					Weblink: weblink,
				}, nil
			},
		}
		contentRepo := &mock.ContentRepositoryMock{}
		publishSvc := service.NewPublishService(nodeRepo, contentRepo)

		handler := NewVideoHandler(publishSvc, nil, "http://localhost")

		weblinkID := "seg1/seg2"
		b64Weblink := base64.StdEncoding.EncodeToString([]byte(weblinkID))

		req := httptest.NewRequest(http.MethodGet, "/video/0p/"+b64Weblink+".m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns valid HLS playlist for video file", func(t *testing.T) {
		nodeRepo := &mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return &entity.Node{
					ID:      1,
					UserID:  1,
					Name:    "video.mp4",
					Home:    vo.NewCloudPath("/video.mp4"),
					Type:    vo.NodeTypeFile,
					Weblink: weblink,
				}, nil
			},
		}
		contentRepo := &mock.ContentRepositoryMock{}
		publishSvc := service.NewPublishService(nodeRepo, contentRepo)

		handler := NewVideoHandler(publishSvc, nil, "http://localhost:8080")

		weblinkID := "seg1/seg2"
		b64Weblink := base64.StdEncoding.EncodeToString([]byte(weblinkID))

		req := httptest.NewRequest(http.MethodGet, "/video/0p/"+b64Weblink+".m3u8", nil)
		w := httptest.NewRecorder()

		handler.HandleHLSPlaylist(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleHLSPlaylist() status = %d, want %d", w.Code, http.StatusOK)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/vnd.apple.mpegurl" {
			t.Errorf("HandleHLSPlaylist() Content-Type = %q, want %q", contentType, "application/vnd.apple.mpegurl")
		}

		body := w.Body.String()
		if !strings.HasPrefix(body, "#EXTM3U") {
			t.Error("HandleHLSPlaylist() response should start with #EXTM3U")
		}
		if !strings.Contains(body, "#EXT-X-ENDLIST") {
			t.Error("HandleHLSPlaylist() response should contain #EXT-X-ENDLIST")
		}
		if !strings.Contains(body, "http://localhost:8080/public/"+weblinkID) {
			t.Error("HandleHLSPlaylist() response should contain video URL")
		}
	})
}

func TestVideoHandler_HandleVideoStream(t *testing.T) {
	t.Run("returns 405 for non-GET requests", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodPost, "/video/seg1/seg2", nil)
		w := httptest.NewRecorder()

		handler.HandleVideoStream(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleVideoStream() status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("routes 0p requests to HLS playlist handler", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		// This should be routed to HLS handler which returns 404 for missing .m3u8
		req := httptest.NewRequest(http.MethodGet, "/video/0p/something", nil)
		w := httptest.NewRecorder()

		handler.HandleVideoStream(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleVideoStream() for 0p path status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 404 for invalid path format", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost")

		tests := []string{
			"/video/",
			"/video/single",
			"/video//empty",
		}

		for _, path := range tests {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.HandleVideoStream(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("HandleVideoStream(%q) status = %d, want %d", path, w.Code, http.StatusNotFound)
			}
		}
	})

	t.Run("returns 404 for nonexistent weblink", func(t *testing.T) {
		nodeRepo := &mock.NodeRepositoryMock{
			GetByWeblinkFunc: func(weblink string) (*entity.Node, error) {
				return nil, nil
			},
		}
		contentRepo := &mock.ContentRepositoryMock{}
		publishSvc := service.NewPublishService(nodeRepo, contentRepo)

		handler := NewVideoHandler(publishSvc, nil, "http://localhost")

		req := httptest.NewRequest(http.MethodGet, "/video/seg1/seg2", nil)
		w := httptest.NewRecorder()

		handler.HandleVideoStream(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleVideoStream() status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestNewVideoHandler(t *testing.T) {
	t.Run("trims trailing slash from external URL", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost:8080/")

		if handler.externalURL != "http://localhost:8080" {
			t.Errorf("NewVideoHandler() externalURL = %q, want %q", handler.externalURL, "http://localhost:8080")
		}
	})

	t.Run("preserves URL without trailing slash", func(t *testing.T) {
		handler := NewVideoHandler(nil, nil, "http://localhost:8080")

		if handler.externalURL != "http://localhost:8080" {
			t.Errorf("NewVideoHandler() externalURL = %q, want %q", handler.externalURL, "http://localhost:8080")
		}
	})
}
