package services

import (
	"testing"
)

func TestGetCacheHash(t *testing.T) {
	sv := &StockVideoService{}

	tests := []struct {
		desc     string
		expected string // Expected MD5 hash
	}{
		{"A peaceful garden with blooming flowers", "2dbec66f09c36e4ad0c36c7346919a14"},
		{"trí tuệ nhân tạo", "8136bb7aedaf9621425f067b01531445"},
		{"", "empty"},
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
