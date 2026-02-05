package thumbnail

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage creates a temporary test image file and returns the file handle.
func createTestImage(t *testing.T, format string, width, height int) *os.File {
	t.Helper()

	tmpDir := t.TempDir()
	var filename string
	switch format {
	case "png":
		filename = filepath.Join(tmpDir, "test.png")
	case "jpeg":
		filename = filepath.Join(tmpDir, "test.jpg")
	default:
		t.Fatalf("Unsupported format: %s", format)
	}

	// Create a simple gradient image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Creating test image file: %v", err)
	}

	switch format {
	case "png":
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("Encoding PNG: %v", err)
		}
	case "jpeg":
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
			f.Close()
			t.Fatalf("Encoding JPEG: %v", err)
		}
	}

	// Seek back to beginning for reading
	_, _ = f.Seek(0, 0)
	return f
}

func TestNewGenerator(t *testing.T) {
	t.Run("creates generator with cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "cache")

		gen, err := NewGenerator(cacheDir)
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if gen == nil {
			t.Fatal("NewGenerator() returned nil")
		}

		// Verify cache directory was created
		info, err := os.Stat(cacheDir)
		if err != nil {
			t.Fatalf("Cache directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Cache path is not a directory")
		}
	})

	t.Run("creates generator without cache", func(t *testing.T) {
		gen, err := NewGenerator("")
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if gen == nil {
			t.Fatal("NewGenerator() returned nil")
		}
	})

	t.Run("creates nested cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "deep", "nested", "cache")

		gen, err := NewGenerator(cacheDir)
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if gen == nil {
			t.Fatal("NewGenerator() returned nil")
		}

		info, err := os.Stat(cacheDir)
		if err != nil {
			t.Fatalf("Nested cache directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Cache path is not a directory")
		}
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Run("generates JPEG thumbnail", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(filepath.Join(tmpDir, "cache"))
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		f := createTestImage(t, "jpeg", 800, 600)
		defer f.Close()

		preset := Preset{Width: 160, Height: 120}
		result, err := gen.Generate(f, "testhash_jpeg", preset)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if result == nil {
			t.Fatal("Generate() returned nil result")
		}
		if len(result.Data) == 0 {
			t.Error("Generate() returned empty data")
		}
		if result.ContentType != "image/jpeg" {
			t.Errorf("Generate() ContentType = %q, want %q", result.ContentType, "image/jpeg")
		}
	})

	t.Run("generates PNG thumbnail", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(filepath.Join(tmpDir, "cache"))
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		f := createTestImage(t, "png", 800, 600)
		defer f.Close()

		preset := Preset{Width: 160, Height: 120}
		result, err := gen.Generate(f, "testhash_png", preset)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if result == nil {
			t.Fatal("Generate() returned nil result")
		}
		if result.ContentType != "image/png" {
			t.Errorf("Generate() ContentType = %q, want %q", result.ContentType, "image/png")
		}
	})

	t.Run("returns original for IsOriginal preset", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(filepath.Join(tmpDir, "cache"))
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		f := createTestImage(t, "jpeg", 100, 100)
		stat, _ := f.Stat()
		originalSize := stat.Size()
		defer f.Close()

		preset := Preset{Width: 0, Height: 0} // Original size
		result, err := gen.Generate(f, "testhash_original", preset)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Result size should match original file size
		if int64(len(result.Data)) != originalSize {
			t.Errorf("Generate() returned %d bytes, want %d (original size)",
				len(result.Data), originalSize)
		}
	})

	t.Run("caches generated thumbnails", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "cache")
		gen, err := NewGenerator(cacheDir)
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		f1 := createTestImage(t, "jpeg", 400, 300)
		preset := Preset{Width: 160, Height: 120}

		result1, err := gen.Generate(f1, "cachehash", preset)
		f1.Close()
		if err != nil {
			t.Fatalf("First Generate() error = %v", err)
		}

		// Verify cache file exists
		cacheFile := filepath.Join(cacheDir, "cachehash_160x120")
		if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
			t.Error("Cache file was not created")
		}

		// Second generation should use cache
		f2 := createTestImage(t, "jpeg", 400, 300)
		result2, err := gen.Generate(f2, "cachehash", preset)
		f2.Close()
		if err != nil {
			t.Fatalf("Second Generate() error = %v", err)
		}

		if len(result1.Data) != len(result2.Data) {
			t.Error("Cached result differs from original")
		}
	})

	t.Run("works without cache", func(t *testing.T) {
		gen, err := NewGenerator("")
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		f := createTestImage(t, "jpeg", 400, 300)
		defer f.Close()

		preset := Preset{Width: 100, Height: 75}
		result, err := gen.Generate(f, "nocache", preset)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(result.Data) == 0 {
			t.Error("Generate() returned empty data")
		}
	})

	t.Run("returns error for invalid image data", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(filepath.Join(tmpDir, "cache"))
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		// Create a file with invalid image data
		invalidFile := filepath.Join(tmpDir, "invalid.jpg")
		if err := os.WriteFile(invalidFile, []byte("not an image"), 0644); err != nil {
			t.Fatalf("Writing invalid file: %v", err)
		}

		f, _ := os.Open(invalidFile)
		defer f.Close()

		preset := Preset{Width: 100, Height: 100}
		_, err = gen.Generate(f, "invalid", preset)
		if err == nil {
			t.Error("Generate() expected error for invalid image, got nil")
		}
	})
}

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
