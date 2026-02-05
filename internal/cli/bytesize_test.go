package cli

import (
	"testing"
)

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		// Basic units
		{name: "bytes numeric", input: "1024", want: 1024},
		{name: "bytes with B", input: "1024B", want: 1024},
		{name: "kilobytes", input: "1KB", want: 1024},
		{name: "megabytes", input: "1MB", want: 1024 * 1024},
		{name: "gigabytes", input: "1GB", want: 1024 * 1024 * 1024},
		{name: "terabytes", input: "1TB", want: 1024 * 1024 * 1024 * 1024},

		// Common values
		{name: "16GB", input: "16GB", want: 17179869184},
		{name: "512MB", input: "512MB", want: 536870912},
		{name: "8GB", input: "8GB", want: 8589934592},

		// Case insensitivity
		{name: "lowercase gb", input: "16gb", want: 17179869184},
		{name: "mixed case Gb", input: "16Gb", want: 17179869184},
		{name: "lowercase mb", input: "512mb", want: 536870912},

		// Decimal values
		{name: "decimal GB", input: "1.5GB", want: 1610612736},
		{name: "decimal MB", input: "2.5MB", want: 2621440},

		// Whitespace
		{name: "with spaces", input: "  16GB  ", want: 17179869184},
		{name: "space before unit", input: "16 GB", want: 17179869184},

		// Zero
		{name: "zero bytes", input: "0", want: 0},
		{name: "zero GB", input: "0GB", want: 0},

		// Errors
		{name: "empty string", input: "", wantErr: true},
		{name: "only spaces", input: "   ", wantErr: true},
		{name: "invalid number", input: "abc", wantErr: true},
		{name: "invalid unit", input: "16XB", wantErr: true},
		{name: "negative value", input: "-16GB", wantErr: true},
		{name: "only unit", input: "GB", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseByteSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseByteSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseByteSize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatByteSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{name: "zero", bytes: 0, want: "0 B"},
		{name: "small bytes", bytes: 512, want: "512 B"},
		{name: "exactly 1KB", bytes: 1024, want: "1.0 KB"},
		{name: "kilobytes", bytes: 2048, want: "2.0 KB"},
		{name: "exactly 1MB", bytes: 1024 * 1024, want: "1.0 MB"},
		{name: "megabytes", bytes: 536870912, want: "512.0 MB"},
		{name: "exactly 1GB", bytes: 1024 * 1024 * 1024, want: "1.0 GB"},
		{name: "gigabytes", bytes: 17179869184, want: "16.0 GB"},
		{name: "terabytes", bytes: 1099511627776, want: "1.0 TB"},
		{name: "decimal KB", bytes: 1536, want: "1.5 KB"},
		{name: "decimal GB", bytes: 1610612736, want: "1.5 GB"},
		{name: "negative", bytes: -100, want: "-100 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatByteSize(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatByteSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestByteSizeRoundTrip(t *testing.T) {
	// Test that common values can be parsed and formatted correctly
	inputs := []string{"1KB", "1MB", "1GB", "1TB", "16GB", "512MB"}

	for _, input := range inputs {
		parsed, err := ParseByteSize(input)
		if err != nil {
			t.Errorf("ParseByteSize(%q) failed: %v", input, err)
			continue
		}

		formatted := FormatByteSize(parsed)
		reparsed, err := ParseByteSize(formatted)
		if err != nil {
			t.Errorf("ParseByteSize(%q) failed on round-trip: %v", formatted, err)
			continue
		}

		if parsed != reparsed {
			t.Errorf("Round-trip failed: %q -> %d -> %q -> %d", input, parsed, formatted, reparsed)
		}
	}
}
