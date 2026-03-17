package service

import (
	"testing"
)

func TestGetCacheHash(t *testing.T) {
	sv := &StockVideoService{}

	tests := []struct {
		desc     string
		expected string // Expected MD5 hash
	}{
		{"A peaceful garden with blooming flowers", "d387a1d4cf1a2cd53de1fcdce30bad8d94fbf141f44248c0552673a9243f7b70"},
		{"trí tuệ nhân tạo", "e63fb701b58ee1e34f33559ca41b73fa919066004cc190716e71887e17c3309b"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		result := sv.getCacheHash(tt.desc)
		if result != tt.expected {
			t.Errorf("getCacheHash(%q) = %s; want %s", tt.desc, result, tt.expected)
		}
	}
}

func TestGetCacheHashConsistency(t *testing.T) {
	sv := &StockVideoService{}
	desc := "Consistency test for cache hashing"

	h1 := sv.getCacheHash(desc)
	h2 := sv.getCacheHash(desc)

	if h1 != h2 {
		t.Errorf("Hash is not consistent: %s vs %s", h1, h2)
	}
}
