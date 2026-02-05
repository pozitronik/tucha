package thumbnail

import (
	"testing"
)

func TestGetPreset(t *testing.T) {
	tests := []struct {
		name       string
		presetName string
		wantNil    bool
		wantWidth  int
		wantHeight int
	}{
		{name: "xw11", presetName: "xw11", wantWidth: 26, wantHeight: 26},
		{name: "xw14", presetName: "xw14", wantWidth: 160, wantHeight: 107},
		{name: "xw1 original", presetName: "xw1", wantWidth: 0, wantHeight: 0},
		{name: "xw2 large", presetName: "xw2", wantWidth: 1000, wantHeight: 667},
		{name: "unknown", presetName: "unknown", wantNil: true},
		{name: "empty", presetName: "", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := GetPreset(tt.presetName)
			if tt.wantNil {
				if p != nil {
					t.Errorf("GetPreset(%q) = %+v, want nil", tt.presetName, p)
				}
				return
			}
			if p == nil {
				t.Fatalf("GetPreset(%q) = nil, want non-nil", tt.presetName)
			}
			if p.Width != tt.wantWidth || p.Height != tt.wantHeight {
				t.Errorf("GetPreset(%q) = {%d, %d}, want {%d, %d}",
					tt.presetName, p.Width, p.Height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestPresetIsOriginal(t *testing.T) {
	original := GetPreset("xw1")
	if original == nil || !original.IsOriginal() {
		t.Error("xw1 should be original size preset")
	}

	scaled := GetPreset("xw14")
	if scaled == nil || scaled.IsOriginal() {
		t.Error("xw14 should not be original size preset")
	}
}

func TestAllPresetsExist(t *testing.T) {
	// All presets from API spec should exist
	expectedPresets := []string{
		"xw11", "xw27", "xw22", "xw12", "xw28", "xw23", "xw14", "xw17",
		"xw10", "xw29", "xw24", "xw20", "xw15", "xw19", "xw26", "xw18",
		"xw13", "xw16", "xw25", "xw21", "xw2", "xw1",
	}

	for _, name := range expectedPresets {
		if GetPreset(name) == nil {
			t.Errorf("Preset %q should exist", name)
		}
	}
}
