package services

import (
	"testing"
)

func TestSplitForAudio(t *testing.T) {
	tp := NewTextProcessor(100, 5.5) // Small chunk size for testing

	tests := []struct {
		name      string
		input     string
		minChunks int
		maxChunks int
	}{
		{
			name:      "Empty text",
			input:     "",
			minChunks: 0,
			maxChunks: 0,
		},
		{
			name:      "Short text",
			input:     "This is a short text.",
			minChunks: 1,
			maxChunks: 1,
		},
		{
			name: "Long text with sentences",
			input: "This is the first sentence. This is the second sentence. This is the third sentence. " +
				"This is the fourth sentence. This is the fifth sentence.",
			minChunks: 2,
			maxChunks: 4, // Allow flexible range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := tp.SplitForAudio(tt.input)
			if len(chunks) < tt.minChunks || len(chunks) > tt.maxChunks {
				t.Errorf("Expected %d-%d chunks, got %d", tt.minChunks, tt.maxChunks, len(chunks))
			}

			// Verify no chunk exceeds max size
			for i, chunk := range chunks {
				if len(chunk) > tp.AudioChunkSize+50 { // Allow overlap
					t.Errorf("Chunk %d exceeds max size: %d > %d", i, len(chunk), tp.AudioChunkSize)
				}
			}
		})
	}
}

func TestSplitForVideo(t *testing.T) {
	tp := NewTextProcessor(4500, 5.5)

	tests := []struct {
		name        string
		input       string
		minSegments int
		maxSegments int
	}{
		{
			name:        "Empty text",
			input:       "",
			minSegments: 0,
			maxSegments: 0,
		},
		{
			name:        "Short text",
			input:       "This is a test.",
			minSegments: 1,
			maxSegments: 1,
		},
		{
			name: "Medium text",
			input: "Đây là một bài kiểm tra. Chúng ta sẽ tạo video từ text này. " +
				"Video sẽ có audio được tạo từ TTS. Hệ thống sẽ chia text thành các phần nhỏ.",
			minSegments: 1,
			maxSegments: 5, // More flexible range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := tp.SplitForVideo(tt.input)

			if len(segments) < tt.minSegments || len(segments) > tt.maxSegments {
				t.Errorf("Expected %d-%d segments, got %d", tt.minSegments, tt.maxSegments, len(segments))
			}

			// Verify each segment has reasonable duration
			for i, seg := range segments {
				if seg.EstimatedDuration < 0 {
					t.Errorf("Segment %d has negative duration: %f", i, seg.EstimatedDuration)
				}
				if seg.Text == "" {
					t.Errorf("Segment %d has empty text", i)
				}
			}
		})
	}
}

func TestEstimateDuration(t *testing.T) {
	tp := NewTextProcessor(4500, 5.5)

	tests := []struct {
		name        string
		input       string
		minDuration float64
		maxDuration float64
	}{
		{
			name:        "Empty text",
			input:       "",
			minDuration: 0,
			maxDuration: 0,
		},
		{
			name:        "10 words",
			input:       "one two three four five six seven eight nine ten",
			minDuration: 3.0, // Should be around 4 seconds (10 words / 150 wpm * 60s * 1.1)
			maxDuration: 6.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tp.estimateDuration(tt.input)
			if duration < tt.minDuration || duration > tt.maxDuration {
				t.Errorf("Expected duration between %f and %f, got %f", tt.minDuration, tt.maxDuration, duration)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tp := NewTextProcessor(4500, 5.5)

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Empty", "", 0},
		{"Single word", "hello", 1},
		{"Multiple words", "hello world test", 3},
		{"With punctuation", "Hello, world! How are you?", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tp.countWords(tt.input)
			if count != tt.expected {
				t.Errorf("Expected %d words, got %d", tt.expected, count)
			}
		})
	}
}

func TestSplitIntoSentences(t *testing.T) {
	tp := NewTextProcessor(4500, 5.5)

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Single sentence",
			input:    "This is one sentence.",
			expected: 1,
		},
		{
			name:     "Multiple sentences",
			input:    "First sentence. Second sentence! Third sentence?",
			expected: 3,
		},
		{
			name:     "Vietnamese text",
			input:    "Đây là câu đầu tiên。 Đây là câu thứ hai！ Câu cuối？",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentences := tp.splitIntoSentences(tt.input)
			if len(sentences) != tt.expected {
				t.Errorf("Expected %d sentences, got %d", tt.expected, len(sentences))
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	tp := NewTextProcessor(4500, 5.5)

	text := "This is a test text. It has multiple sentences. We will analyze it."
	stats := tp.GetStats(text)

	// Verify stats structure
	if _, ok := stats["total_chars"]; !ok {
		t.Error("Missing total_chars in stats")
	}
	if _, ok := stats["total_words"]; !ok {
		t.Error("Missing total_words in stats")
	}
	if _, ok := stats["audio_chunks"]; !ok {
		t.Error("Missing audio_chunks in stats")
	}
	if _, ok := stats["video_segments"]; !ok {
		t.Error("Missing video_segments in stats")
	}
	if _, ok := stats["estimated_duration"]; !ok {
		t.Error("Missing estimated_duration in stats")
	}

	// Verify reasonable values
	if stats["total_chars"].(int) == 0 {
		t.Error("total_chars should not be 0")
	}
	if stats["total_words"].(int) == 0 {
		t.Error("total_words should not be 0")
	}
}
