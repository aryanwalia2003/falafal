package report

import (
	"fmt"
	"strconv"
	"strings"
)

// HumanSize formats a byte count as a human-readable string (KiB/MiB/GiB/...).
func HumanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := "KMGTPE"
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), units[exp])
}

var sizeUnits = map[string]int64{
	"B": 1,
	"":  1,

	"KB": 1000, "MB": 1000 * 1000, "GB": 1000 * 1000 * 1000, "TB": 1000 * 1000 * 1000 * 1000,

	"KIB": 1024, "MIB": 1024 * 1024, "GIB": 1024 * 1024 * 1024, "TIB": 1024 * 1024 * 1024 * 1024,
	"K": 1024, "M": 1024 * 1024, "G": 1024 * 1024 * 1024, "T": 1024 * 1024 * 1024 * 1024,
}

// ParseSize parses a human-friendly size like "100MB", "1.5GiB", "512k", or a
// bare byte count, returning the size in bytes. It is the inverse of HumanSize,
// used for flags like --min-size/--max-size.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("invalid size %q: no numeric value", s)
	}
	numPart, unitPart := s[:i], strings.ToUpper(strings.TrimSpace(s[i:]))

	n, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	mult, ok := sizeUnits[unitPart]
	if !ok {
		return 0, fmt.Errorf("invalid size %q: unknown unit %q", s, unitPart)
	}
	return int64(n * float64(mult)), nil
}
