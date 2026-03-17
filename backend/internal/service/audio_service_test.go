package service

import (
	"testing"
)

func TestMapToElevenLabsVoice(t *testing.T) {
	as := &AudioService{}

	tests := []struct {
		voiceName string
		expected  string // ElevenLabs ID
	}{
		{"minhquang", "ipTvfDXAg1zowfF1rv9w"},                              // Male
		{"giahuy", "ipTvfDXAg1zowfF1rv9w"},                                 // Male
		{"leminh", "Si3s1VCb7dLbeqH57kiC"},                                 // Female (fallback)
		{"random_long_id_already_exists", "random_long_id_already_exists"}, // Pass-through
	}

	for _, tt := range tests {
		result := as.mapToElevenLabsVoice(tt.voiceName)
		if result != tt.expected {
			t.Errorf("mapToElevenLabsVoice(%s) = %s; want %s", tt.voiceName, result, tt.expected)
		}
	}
}
