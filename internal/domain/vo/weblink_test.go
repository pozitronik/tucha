package vo

import (
	"regexp"
	"testing"
)

func TestGenerateWeblink_format(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}/[0-9a-f]{8}$`)
	link, err := GenerateWeblink()
	if err != nil {
		t.Fatalf("GenerateWeblink() error: %v", err)
	}
	if !pattern.MatchString(link) {
		t.Errorf("GenerateWeblink() = %q, does not match pattern %q", link, pattern.String())
	}
}

func TestGenerateWeblink_unique(t *testing.T) {
	a, err := GenerateWeblink()
	if err != nil {
		t.Fatalf("first GenerateWeblink() error: %v", err)
	}
	b, err := GenerateWeblink()
	if err != nil {
		t.Fatalf("second GenerateWeblink() error: %v", err)
	}
	if a == b {
		t.Errorf("two consecutive GenerateWeblink() calls produced identical results: %q", a)
	}
}

func TestGenerateWeblink_no_error(t *testing.T) {
	for i := 0; i < 10; i++ {
		if _, err := GenerateWeblink(); err != nil {
			t.Fatalf("GenerateWeblink() iteration %d error: %v", i, err)
		}
	}
}
