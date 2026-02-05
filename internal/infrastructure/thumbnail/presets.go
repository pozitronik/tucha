// Package thumbnail provides image thumbnail generation and caching.
package thumbnail

// Preset defines thumbnail dimensions.
type Preset struct {
	Width  int
	Height int
}

// Presets maps preset names to dimensions as defined in API specification.
var Presets = map[string]Preset{
	"xw11": {Width: 26, Height: 26},
	"xw27": {Width: 28, Height: 38},
	"xw22": {Width: 36, Height: 24},
	"xw12": {Width: 52, Height: 35},
	"xw28": {Width: 64, Height: 43},
	"xw23": {Width: 72, Height: 48},
	"xw14": {Width: 160, Height: 107},
	"xw17": {Width: 160, Height: 120},
	"xw10": {Width: 160, Height: 120},
	"xw29": {Width: 150, Height: 150},
	"xw24": {Width: 168, Height: 112},
	"xw20": {Width: 170, Height: 113},
	"xw15": {Width: 206, Height: 137},
	"xw19": {Width: 206, Height: 206},
	"xw26": {Width: 270, Height: 365},
	"xw18": {Width: 305, Height: 230},
	"xw13": {Width: 320, Height: 213},
	"xw16": {Width: 320, Height: 240},
	"xw25": {Width: 336, Height: 224},
	"xw21": {Width: 340, Height: 226},
	"xw2":  {Width: 1000, Height: 667},
	"xw1":  {Width: 0, Height: 0}, // Original size
}

// GetPreset returns the preset for the given name, or nil if not found.
func GetPreset(name string) *Preset {
	if p, ok := Presets[name]; ok {
		return &p
	}
	return nil
}

// IsOriginal returns true if the preset represents original size.
func (p Preset) IsOriginal() bool {
	return p.Width == 0 && p.Height == 0
}
