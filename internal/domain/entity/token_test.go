package entity

import (
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{"future expiry", now + 3600, false},
		{"past expiry", now - 3600, true},
		// Token uses >, so exact boundary (expiresAt == now) means not expired.
		{"exact boundary", now, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &Token{ExpiresAt: tt.expiresAt}
			if got := tok.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v (expiresAt=%d, now~=%d)", got, tt.want, tt.expiresAt, now)
			}
		})
	}
}
