package utils

import (
	"testing"
)

func TestFormatSRTTimestamp(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0.0, "00:00:00,000"},
		{1.0, "00:00:01,000"},
		{61.52, "00:01:01,520"},
		{3661.123, "01:01:01,123"},
		{7200.0, "02:00:00,000"},
	}

	for _, tt := range tests {
		result := FormatSRTTimestamp(tt.seconds)
		if result != tt.expected {
			t.Errorf("FormatSRTTimestamp(%.3f) = %s; want %s", tt.seconds, result, tt.expected)
		}
	}
}
