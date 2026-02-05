package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

// Generator creates and caches image thumbnails.
type Generator struct {
	cacheDir string
	mu       sync.RWMutex
}

// NewGenerator creates a new thumbnail generator with the specified cache directory.
// If cacheDir is empty, thumbnails are generated on-the-fly without caching.
func NewGenerator(cacheDir string) (*Generator, error) {
	if cacheDir != "" {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("creating thumbnail cache dir: %w", err)
		}
	}
	return &Generator{cacheDir: cacheDir}, nil
}

// Result holds the generated thumbnail data.
type Result struct {
	Data        []byte
	ContentType string
}

// Generate creates a thumbnail for the given image file at the specified preset size.
// Returns the thumbnail data and content type.
func (g *Generator) Generate(srcFile *os.File, hash string, preset Preset) (*Result, error) {
	// Check cache first
	if g.cacheDir != "" {
		if cached := g.loadFromCache(hash, preset); cached != nil {
			return cached, nil
		}
	}

	// Read and decode the source image
	srcData, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, fmt.Errorf("reading source file: %w", err)
	}

	// Reset file position for potential reuse
	_, _ = srcFile.Seek(0, io.SeekStart)

	// Detect format and decode
	reader := bytes.NewReader(srcData)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	// For original size preset, just return the original data
	if preset.IsOriginal() {
		return &Result{
			Data:        srcData,
			ContentType: formatToContentType(format),
		}, nil
	}

	// Calculate target dimensions maintaining aspect ratio
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	targetWidth, targetHeight := fitWithinBounds(srcWidth, srcHeight, preset.Width, preset.Height)

	// Create scaled image
	scaled := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, srcBounds, draw.Over, nil)

	// Encode the result
	var buf bytes.Buffer
	var contentType string

	switch format {
	case "png":
		err = png.Encode(&buf, scaled)
		contentType = "image/png"
	case "gif":
		err = gif.Encode(&buf, scaled, nil)
		contentType = "image/gif"
	default:
		// Default to JPEG for unknown formats
		err = jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: 85})
		contentType = "image/jpeg"
	}

	if err != nil {
		return nil, fmt.Errorf("encoding thumbnail: %w", err)
	}

	result := &Result{
		Data:        buf.Bytes(),
		ContentType: contentType,
	}

	// Cache the result
	if g.cacheDir != "" {
		g.saveToCache(hash, preset, result)
	}

	return result, nil
}

// IsSupportedFormat checks if the file extension indicates a supported image format.
func IsSupportedFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

// fitWithinBounds calculates dimensions that fit within maxWidth x maxHeight
// while maintaining aspect ratio.
func fitWithinBounds(srcWidth, srcHeight, maxWidth, maxHeight int) (int, int) {
	if srcWidth <= maxWidth && srcHeight <= maxHeight {
		return srcWidth, srcHeight
	}

	widthRatio := float64(maxWidth) / float64(srcWidth)
	heightRatio := float64(maxHeight) / float64(srcHeight)

	ratio := widthRatio
	if heightRatio < widthRatio {
		ratio = heightRatio
	}

	targetWidth := int(float64(srcWidth) * ratio)
	targetHeight := int(float64(srcHeight) * ratio)

	// Ensure minimum dimensions
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	return targetWidth, targetHeight
}

func formatToContentType(format string) string {
	switch format {
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func (g *Generator) cacheKey(hash string, preset Preset) string {
	return fmt.Sprintf("%s_%dx%d", hash, preset.Width, preset.Height)
}

func (g *Generator) cachePath(key string) string {
	return filepath.Join(g.cacheDir, key)
}

func (g *Generator) loadFromCache(hash string, preset Preset) *Result {
	g.mu.RLock()
	defer g.mu.RUnlock()

	key := g.cacheKey(hash, preset)
	path := g.cachePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Detect content type from cached data
	contentType := "image/jpeg"
	if len(data) > 8 {
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			contentType = "image/png"
		} else if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			contentType = "image/gif"
		}
	}

	return &Result{
		Data:        data,
		ContentType: contentType,
	}
}

func (g *Generator) saveToCache(hash string, preset Preset, result *Result) {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := g.cacheKey(hash, preset)
	path := g.cachePath(key)

	// Write atomically via temp file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, result.Data, 0644); err != nil {
		return
	}
	_ = os.Rename(tmpPath, path)
}
