package thumbnail

import (
	"testing"
)

func TestFitWithinBounds(t *testing.T) {
	tests := []struct {
		name       string
		srcW, srcH int
		maxW, maxH int
		wantW      int
		wantH      int
	}{
		// Image smaller than bounds - no change
		{name: "smaller image", srcW: 100, srcH: 100, maxW: 200, maxH: 200, wantW: 100, wantH: 100},

		// Exact fit
		{name: "exact fit", srcW: 200, srcH: 200, maxW: 200, maxH: 200, wantW: 200, wantH: 200},

		// Wide image - width constrained
		{name: "wide image", srcW: 1000, srcH: 500, maxW: 200, maxH: 200, wantW: 200, wantH: 100},

		// Tall image - height constrained
		{name: "tall image", srcW: 500, srcH: 1000, maxW: 200, maxH: 200, wantW: 100, wantH: 200},

		// Square image larger than bounds
		{name: "large square", srcW: 400, srcH: 400, maxW: 200, maxH: 200, wantW: 200, wantH: 200},

		// Rectangular bounds
		{name: "rectangular bounds", srcW: 800, srcH: 600, maxW: 160, maxH: 107, wantW: 142, wantH: 107},

		// Very wide image
		{name: "very wide", srcW: 2000, srcH: 100, maxW: 200, maxH: 200, wantW: 200, wantH: 10},

		// Very tall image
		{name: "very tall", srcW: 100, srcH: 2000, maxW: 200, maxH: 200, wantW: 10, wantH: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := fitWithinBounds(tt.srcW, tt.srcH, tt.maxW, tt.maxH)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("fitWithinBounds(%d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.srcW, tt.srcH, tt.maxW, tt.maxH, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestFormatToContentType(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"png", "image/png"},
		{"gif", "image/gif"},
		{"jpeg", "image/jpeg"},
		{"unknown", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		got := formatToContentType(tt.format)
		if got != tt.want {
			t.Errorf("formatToContentType(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.JPG", true},
		{"image.JPEG", true},
		{"image.png", true},
		{"image.PNG", true},
		{"image.gif", true},
		{"image.GIF", true},
		{"document.pdf", false},
		{"video.mp4", false},
		{"file.txt", false},
		{"image.webp", false}, // Not supported yet
		{"noextension", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsSupportedFormat(tt.filename)
		if got != tt.want {
			t.Errorf("IsSupportedFormat(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}
