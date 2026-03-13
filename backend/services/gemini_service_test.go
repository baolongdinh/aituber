package services

import (
	"testing"
)

func TestExtractJSON(t *testing.T) {
	gs := &GeminiService{}

	tests := []struct {
		name     string
		input    string
		expected string // JSON string part we expect
	}{
		{
			name:     "Standard markdown blocks",
			input:    "Sure, here is your JSON:\n```json\n[{\"text\":\"Hello\"}]\n```\nHope this helps!",
			expected: "[{\"text\":\"Hello\"}]",
		},
		{
			name:     "JSON submerged in text without blocks",
			input:    "The data is [{\"text\":\"Direct JSON\"}] and that's it.",
			expected: "[{\"text\":\"Direct JSON\"}]",
		},
		{
			name:     "Multiple blocks (take first)",
			input:    "Part 1: ```json\n{\"a\":1}\n``` and Part 2: ```json\n{\"b\":2}\n```",
			expected: "{\"a\":1}",
		},
		{
			name:     "Object format instead of array",
			input:    "Here is an object: {\"status\":\"ok\"}",
			expected: "{\"status\":\"ok\"}",
		},
		{
			name:     "Messy whitespace and newlines",
			input:    " \n  \r\n [  \n {\"id\": 1 } \n ] \n ",
			expected: "[  \n {\"id\": 1 } \n ]",
		},
		{
			name:     "Nested structures (Array with inner arrays - real world case)",
			input:    "Outline generated: [\n  {\n    \"part_number\": 1,\n    \"key_points\": [\"point 1\", \"point 2\"]\n  }\n]",
			expected: "[\n  {\n    \"part_number\": 1,\n    \"key_points\": [\"point 1\", \"point 2\"]\n  }\n]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gs.extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON mismatch.\nInput: %q\nExpected: %q\nGot:      %q", tt.input, tt.expected, result)
			}
		})
	}
}
