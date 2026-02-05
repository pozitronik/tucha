package cli

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		pattern string
		want    bool
	}{
		// Exact match
		{name: "exact match", s: "user@example.com", pattern: "user@example.com", want: true},
		{name: "exact no match", s: "user@example.com", pattern: "other@example.com", want: false},

		// Wildcard at start
		{name: "wildcard start match", s: "user@example.com", pattern: "*@example.com", want: true},
		{name: "wildcard start no match", s: "user@other.com", pattern: "*@example.com", want: false},

		// Wildcard at end
		{name: "wildcard end match", s: "user@example.com", pattern: "user@*", want: true},
		{name: "wildcard end no match", s: "admin@example.com", pattern: "user@*", want: false},

		// Wildcard in middle
		{name: "wildcard middle match", s: "user@example.com", pattern: "user@*.com", want: true},
		{name: "wildcard middle no match", s: "user@example.org", pattern: "user@*.com", want: false},

		// Multiple wildcards
		{name: "multiple wildcards match", s: "user@example.com", pattern: "*@*.*", want: true},
		{name: "multiple wildcards specific", s: "admin@mail.example.com", pattern: "*@mail.*", want: true},

		// Only wildcard
		{name: "only wildcard", s: "anything", pattern: "*", want: true},

		// Case insensitivity
		{name: "case insensitive", s: "User@Example.COM", pattern: "user@example.com", want: true},
		{name: "case insensitive wildcard", s: "USER@EXAMPLE.COM", pattern: "*@example.com", want: true},

		// Edge cases
		{name: "empty pattern no wildcard", s: "user@example.com", pattern: "", want: false},
		{name: "empty string with wildcard", s: "", pattern: "*", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.s, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}
