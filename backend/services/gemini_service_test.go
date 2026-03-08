package services

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONCleanAndUnmarshal(t *testing.T) {
	rawInputs := []string{
		"```json\n[{\"text\":\"Hello\",\"pexels_search_query\":\"world\"}]\n```",
		"[{\"text\":\"Hello\",\"pexels_search_query\":\"world\"}]",
		"  \n[{\"text\":\"Hello\",\"pexels_search_query\":\"world\"}]\n  ",
	}

	for i, input := range rawInputs {
		t.Run("Input "+string(rune(i+'A')), func(t *testing.T) {
			clean := strings.TrimSpace(input)
			clean = strings.TrimPrefix(clean, "```json")
			clean = strings.TrimPrefix(clean, "```")
			clean = strings.TrimSuffix(clean, "```")
			clean = strings.TrimSpace(clean)

			var segments []map[string]interface{}
			err := json.Unmarshal([]byte(clean), &segments)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if len(segments) != 1 {
				t.Fatalf("Expected 1 segment, got %d", len(segments))
			}
			if segments[0]["text"] != "Hello" {
				t.Errorf("Expected 'Hello', got '%v'", segments[0]["text"])
			}
			if segments[0]["pexels_search_query"] != "world" {
				t.Errorf("Expected 'world', got '%v'", segments[0]["pexels_search_query"])
			}
		})
	}
}
