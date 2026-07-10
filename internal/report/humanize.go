package report

import "fmt"

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
