package entity

import "testing"

func TestContent_IsUnreferenced(t *testing.T) {
	tests := []struct {
		name     string
		refCount int64
		want     bool
	}{
		{"zero refs", 0, true},
		{"negative refs", -1, true},
		{"one ref", 1, false},
		{"many refs", 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{RefCount: tt.refCount}
			if got := c.IsUnreferenced(); got != tt.want {
				t.Errorf("IsUnreferenced() = %v, want %v (refCount=%d)", got, tt.want, tt.refCount)
			}
		})
	}
}
