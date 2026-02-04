package vo

import "testing"

func TestNewContentHash_valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"uppercase 40-hex", "C172C6E2FF47284FF33F348FEA7EECE532F6C051", "C172C6E2FF47284FF33F348FEA7EECE532F6C051"},
		{"lowercase converted", "c172c6e2ff47284ff33f348fea7eece532f6c051", "C172C6E2FF47284FF33F348FEA7EECE532F6C051"},
		{"mixed case", "c172C6e2ff47284FF33f348FEA7eece532f6c051", "C172C6E2FF47284FF33F348FEA7EECE532F6C051"},
		{"all zeros", "0000000000000000000000000000000000000000", "0000000000000000000000000000000000000000"},
		{"all F", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewContentHash(tt.input)
			if err != nil {
				t.Fatalf("NewContentHash(%q) unexpected error: %v", tt.input, err)
			}
			if got := h.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewContentHash_invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"too short", "C172C6E2FF47284FF33F348FEA7EECE532F6C0"},
		{"too long", "C172C6E2FF47284FF33F348FEA7EECE532F6C05100"},
		{"invalid chars", "G172C6E2FF47284FF33F348FEA7EECE532F6C051"},
		{"spaces", "C172C6E2FF47284FF33F 48FEA7EECE532F6C051"},
		{"special chars", "C172C6E2FF47284FF33F-48FEA7EECE532F6C051"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContentHash(tt.input)
			if err == nil {
				t.Errorf("NewContentHash(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestMustContentHash_panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustContentHash with invalid input did not panic")
		}
	}()
	MustContentHash("invalid")
}

func TestMustContentHash_valid(t *testing.T) {
	h := MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")
	if h.String() != "C172C6E2FF47284FF33F348FEA7EECE532F6C051" {
		t.Errorf("unexpected hash: %q", h.String())
	}
}

func TestContentHash_ShardPrefix(t *testing.T) {
	h := MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051")
	l1, l2 := h.ShardPrefix()
	if l1 != "C1" {
		t.Errorf("shard l1 = %q, want %q", l1, "C1")
	}
	if l2 != "72" {
		t.Errorf("shard l2 = %q, want %q", l2, "72")
	}
}

func TestContentHash_IsZero(t *testing.T) {
	tests := []struct {
		name string
		hash ContentHash
		want bool
	}{
		{"zero value", ContentHash{}, true},
		{"valid hash", MustContentHash("C172C6E2FF47284FF33F348FEA7EECE532F6C051"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hash.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
