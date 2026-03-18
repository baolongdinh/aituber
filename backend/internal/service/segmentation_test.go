package service

import (
	"testing"
)

func TestPostProcessSegments(t *testing.T) {
	gs := &GeminiService{}

	tests := []struct {
		name     string
		input    []VideoSegment
		expected []string // Check if these phrases are kept together
	}{
		{
			name: "Should not split short phrases without punctuation",
			input: []VideoSegment{
				{Text: "Chào anh em chúng ta cùng làm video viral nhé"},
			},
			expected: []string{"Chào anh em chúng ta cùng làm video viral nhé."},
		},
		{
			name: "Should split at punctuation even if short",
			input: []VideoSegment{
				{Text: "Chào các bạn, hôm nay tôi sẽ hướng dẫn các bạn làm video."},
			},
			expected: []string{
				"Chào các bạn.",
				"hôm nay tôi sẽ hướng dẫn các bạn làm video.",
			},
		},
		{
			name: "Should handle very long text by splitting at punctuation first",
			input: []VideoSegment{
				{Text: "Đây là một đoạn văn rất dài mà tôi muốn thử nghiệm để xem nó có cắt đúng chỗ không, ví dụ như là anh em chúng ta phải làm việc thật chăm chỉ để thành công rực rỡ trong tương lai không xa."},
			},
			expected: []string{
				"Đây là một đoạn văn rất dài mà tôi muốn thử nghiệm để xem nó có cắt đúng chỗ không.",
				"ví dụ như là anh em chúng ta phải làm việc thật chăm chỉ để thành công rực rỡ trong tương lai không xa.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gs.postProcessSegments(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d segments, got %d", len(tt.expected), len(result))
			}
			for i, res := range result {
				if i < len(tt.expected) && res.Text != tt.expected[i] {
					t.Errorf("segment %d: expected %q, got %q", i, tt.expected[i], res.Text)
				}
			}
		})
	}
}
