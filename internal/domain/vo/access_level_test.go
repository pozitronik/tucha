package vo

import "testing"

func TestParseAccessLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    AccessLevel
		wantErr bool
	}{
		{"read_only", AccessReadOnly, false},
		{"r", AccessReadOnly, false},
		{"read_write", AccessReadWrite, false},
		{"rw", AccessReadWrite, false},
		{"", "", true},
		{"readonly", "", true},
		{"write", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAccessLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAccessLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseAccessLevel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAccessLevel_String(t *testing.T) {
	if got := AccessReadOnly.String(); got != "read_only" {
		t.Errorf("AccessReadOnly.String() = %q, want %q", got, "read_only")
	}
	if got := AccessReadWrite.String(); got != "read_write" {
		t.Errorf("AccessReadWrite.String() = %q, want %q", got, "read_write")
	}
}

func TestAccessLevel_APIString(t *testing.T) {
	tests := []struct {
		level AccessLevel
		want  string
	}{
		{AccessReadOnly, "r"},
		{AccessReadWrite, "rw"},
		{AccessLevel("unknown"), "unknown"},
	}
	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			got := tt.level.APIString()
			if got != tt.want {
				t.Errorf("APIString() = %q, want %q", got, tt.want)
			}
		})
	}
}
