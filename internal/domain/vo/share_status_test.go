package vo

import "testing"

func TestParseShareStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    ShareStatus
		wantErr bool
	}{
		{"pending", SharePending, false},
		{"accepted", ShareAccepted, false},
		{"rejected", ShareRejected, false},
		{"", "", true},
		{"cancelled", "", true},
	}
	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			got, err := ParseShareStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseShareStatus(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseShareStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestShareStatus_String(t *testing.T) {
	if got := SharePending.String(); got != "pending" {
		t.Errorf("SharePending.String() = %q, want %q", got, "pending")
	}
	if got := ShareAccepted.String(); got != "accepted" {
		t.Errorf("ShareAccepted.String() = %q, want %q", got, "accepted")
	}
	if got := ShareRejected.String(); got != "rejected" {
		t.Errorf("ShareRejected.String() = %q, want %q", got, "rejected")
	}
}
