package httpapi

import (
	"testing"
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
