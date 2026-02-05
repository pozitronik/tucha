package entity

import (
	"github.com/pozitronik/tucha/internal/domain/vo"
	"testing"
)

func TestShare_IsPending(t *testing.T) {
	tests := []struct {
		status vo.ShareStatus
		want   bool
	}{
		{vo.SharePending, true},
		{vo.ShareAccepted, false},
		{vo.ShareRejected, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			s := &Share{Status: tt.status}
			if got := s.IsPending(); got != tt.want {
				t.Errorf("IsPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShare_IsAccepted(t *testing.T) {
	tests := []struct {
		status vo.ShareStatus
		want   bool
	}{
		{vo.SharePending, false},
		{vo.ShareAccepted, true},
		{vo.ShareRejected, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			s := &Share{Status: tt.status}
			if got := s.IsAccepted(); got != tt.want {
				t.Errorf("IsAccepted() = %v, want %v", got, tt.want)
			}
		})
	}
}
