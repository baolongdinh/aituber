package utils

import (
	"fmt"
	"math"
)

// FormatSRTTimestamp formats seconds to SRT timestamp format (HH:MM:SS,mmm)
func FormatSRTTimestamp(seconds float64) string {
	d := int(seconds)
	ms := int(math.Round((seconds - float64(d)) * 1000))

	h := d / 3600
	m := (d % 3600) / 60
	s := d % 60

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
