package hasher

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/vo"
)

func TestMrCloud_Compute_small(t *testing.T) {
	h := NewMrCloud()

	tests := []struct {
		name string
		data []byte
		want string
	}{
		// Empty data: 20 zero bytes hex-encoded = 40 zeros.
		{"empty", []byte{}, "0000000000000000000000000000000000000000"},
		// Single byte 0x41 ('A') followed by 19 zero bytes.
		{"one byte", []byte{0x41}, "4100000000000000000000000000000000000000"},
		// Exactly 20 bytes: no zero-padding needed, just hex-encode.
		{"exactly 20 bytes", bytes.Repeat([]byte{0xFF}, 20), "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"[:40]},
		// 10 bytes: 10 data bytes + 10 zero bytes.
		{"10 bytes", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "0102030405060708090A00000000000000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := h.Compute(tt.data)
			if got.String() != tt.want {
				t.Errorf("Compute(%v) = %q, want %q", tt.data, got.String(), tt.want)
			}
		})
	}
}

func TestMrCloud_Compute_large(t *testing.T) {
	h := NewMrCloud()

	// Exactly 21 bytes: switches to SHA1 path.
	data21 := bytes.Repeat([]byte{0xAB}, 21)
	got := h.Compute(data21)

	// The result must be a valid ContentHash.
	if got.IsZero() {
		t.Fatal("Compute(21 bytes) returned zero hash")
	}
	if len(got.String()) != 40 {
		t.Errorf("Compute(21 bytes) length = %d, want 40", len(got.String()))
	}

	// Large file should produce deterministic results.
	data := bytes.Repeat([]byte("hello world"), 100)
	a := h.Compute(data)
	b := h.Compute(data)
	if a.String() != b.String() {
		t.Errorf("non-deterministic: %q != %q", a.String(), b.String())
	}
}

func TestMrCloud_Compute_boundary(t *testing.T) {
	h := NewMrCloud()

	// 20 bytes: small path (< 21).
	data20 := bytes.Repeat([]byte{0xBB}, 20)
	small := h.Compute(data20)

	// 21 bytes: large path (>= 21).
	data21 := bytes.Repeat([]byte{0xBB}, 21)
	large := h.Compute(data21)

	// They must differ because different algorithms are used.
	if small.String() == large.String() {
		t.Error("20-byte and 21-byte data produced same hash, but should use different algorithms")
	}
}

func TestMrCloud_ComputeReader_equivalence(t *testing.T) {
	h := NewMrCloud()

	tests := []struct {
		name string
		data []byte
	}{
		{"small", []byte("hello")},
		{"boundary", bytes.Repeat([]byte{0xCC}, 21)},
		{"large", bytes.Repeat([]byte("test data"), 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromBytes := h.Compute(tt.data)
			fromReader, err := h.ComputeReader(bytes.NewReader(tt.data), int64(len(tt.data)))
			if err != nil {
				t.Fatalf("ComputeReader error: %v", err)
			}
			if fromBytes.String() != fromReader.String() {
				t.Errorf("Compute=%q, ComputeReader=%q", fromBytes.String(), fromReader.String())
			}
		})
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestMrCloud_ComputeReader_readerError(t *testing.T) {
	h := NewMrCloud()

	// Small file path error.
	_, err := h.ComputeReader(errReader{}, 5)
	if err == nil {
		t.Error("ComputeReader with failing reader (small) should return error")
	}

	// Large file path error.
	_, err = h.ComputeReader(errReader{}, 100)
	if err == nil {
		t.Error("ComputeReader with failing reader (large) should return error")
	}
}

func TestMrCloud_Compute_outputFormat(t *testing.T) {
	h := NewMrCloud()
	data := []byte("some content for validation")
	got := h.Compute(data)

	// Verify it produces a valid ContentHash (40 uppercase hex).
	_, err := vo.NewContentHash(got.String())
	if err != nil {
		t.Errorf("Compute output is not a valid ContentHash: %v", err)
	}
	if got.String() != strings.ToUpper(got.String()) {
		t.Errorf("output not uppercase: %q", got.String())
	}
}
