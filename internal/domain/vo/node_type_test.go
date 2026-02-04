package vo

import "testing"

func TestParseNodeType(t *testing.T) {
	tests := []struct {
		input   string
		want    NodeType
		wantErr bool
	}{
		{"file", NodeTypeFile, false},
		{"folder", NodeTypeFolder, false},
		{"", "", true},
		{"directory", "", true},
		{"FILE", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseNodeType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseNodeType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseNodeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNodeType_IsFile(t *testing.T) {
	if !NodeTypeFile.IsFile() {
		t.Error("NodeTypeFile.IsFile() = false, want true")
	}
	if NodeTypeFolder.IsFile() {
		t.Error("NodeTypeFolder.IsFile() = true, want false")
	}
}

func TestNodeType_IsFolder(t *testing.T) {
	if !NodeTypeFolder.IsFolder() {
		t.Error("NodeTypeFolder.IsFolder() = false, want true")
	}
	if NodeTypeFile.IsFolder() {
		t.Error("NodeTypeFile.IsFolder() = true, want false")
	}
}

func TestNodeType_String(t *testing.T) {
	if got := NodeTypeFile.String(); got != "file" {
		t.Errorf("NodeTypeFile.String() = %q, want %q", got, "file")
	}
	if got := NodeTypeFolder.String(); got != "folder" {
		t.Errorf("NodeTypeFolder.String() = %q, want %q", got, "folder")
	}
}
