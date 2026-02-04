package vo

import "testing"

func TestParseConflictMode(t *testing.T) {
	tests := []struct {
		input   string
		want    ConflictMode
		wantErr bool
	}{
		{"strict", ConflictStrict, false},
		{"rename", ConflictRename, false},
		{"replace", ConflictReplace, false},
		{"", ConflictRename, false},
		{"overwrite", "", true},
	}
	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			got, err := ParseConflictMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseConflictMode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseConflictMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConflictMode_String(t *testing.T) {
	if got := ConflictStrict.String(); got != "strict" {
		t.Errorf("ConflictStrict.String() = %q, want %q", got, "strict")
	}
	if got := ConflictRename.String(); got != "rename" {
		t.Errorf("ConflictRename.String() = %q, want %q", got, "rename")
	}
	if got := ConflictReplace.String(); got != "replace" {
		t.Errorf("ConflictReplace.String() = %q, want %q", got, "replace")
	}
}
