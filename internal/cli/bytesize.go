// Package cli provides command-line interface parsing and command execution.
package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// Binary unit multipliers (1024-based).
const (
	_          = iota
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

// ParseByteSize parses a human-readable byte size string into bytes.
// Supports: B, KB, MB, GB, TB (case-insensitive, binary units 1024-based).
// Examples: "16GB" -> 17179869184, "512MB" -> 536870912, "1024" -> 1024
func ParseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	s = strings.ToUpper(s)

	// Try to find unit suffix
	var multiplier float64 = 1
	var numStr string

	switch {
	case strings.HasSuffix(s, "TB"):
		multiplier = TB
		numStr = strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "GB"):
		multiplier = GB
		numStr = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = MB
		numStr = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = KB
		numStr = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		multiplier = 1
		numStr = strings.TrimSuffix(s, "B")
	default:
		// No unit suffix, treat as bytes
		numStr = s
	}

	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("missing numeric value in size string")
	}

	// Parse as float to support decimals like "1.5GB"
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value %q: %w", numStr, err)
	}

	if num < 0 {
		return 0, fmt.Errorf("negative size not allowed: %v", num)
	}

	return int64(num * multiplier), nil
}

// FormatByteSize formats bytes as a human-readable string.
// Examples: 17179869184 -> "16.0 GB", 536870912 -> "512.0 MB"
func FormatByteSize(bytes int64) string {
	if bytes < 0 {
		return fmt.Sprintf("%d B", bytes)
	}

	b := float64(bytes)

	switch {
	case b >= TB:
		return fmt.Sprintf("%.1f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.1f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", b/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
