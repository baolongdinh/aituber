package services

import (
	"aituber/models"
	"strings"
	"unicode"
)

// TextProcessor handles text segmentation for audio and video
type TextProcessor struct {
	AudioChunkSize       int
	VideoSegmentDuration float64
	AvgWordsPerMinute    float64 // Default: 150 words per minute
}

// NewTextProcessor creates a new text processor
func NewTextProcessor(audioChunkSize int, videoSegmentDuration float64) *TextProcessor {
	return &TextProcessor{
		AudioChunkSize:       audioChunkSize,
		VideoSegmentDuration: videoSegmentDuration,
		AvgWordsPerMinute:    150.0, // Vietnamese average reading speed
	}
}

// SplitForAudio splits text into chunks suitable for TTS
// - Maximum 4500 characters per chunk
// - Prefers splitting at sentence boundaries (. ! ?)
// - Doesn't split in the middle of words
// - Overlaps 50 characters between chunks to maintain context
func (tp *TextProcessor) SplitForAudio(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	if len(text) <= tp.AudioChunkSize {
		return []string{text}
	}

	chunks := []string{}
	currentPos := 0
	textLen := len(text)

	for currentPos < textLen {
		// Determine chunk end position
		endPos := currentPos + tp.AudioChunkSize
		if endPos >= textLen {
			// Last chunk
			chunks = append(chunks, text[currentPos:])
			break
		}

		// Find sentence boundary near endPos
		splitPos := tp.findSentenceBoundary(text, currentPos, endPos)
		if splitPos == -1 {
			// No sentence boundary found, find word boundary
			splitPos = tp.findWordBoundary(text, endPos)
		}

		// Extract chunk
		chunk := strings.TrimSpace(text[currentPos:splitPos])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move to next position with overlap
		overlap := 50
		if splitPos-overlap > currentPos {
			currentPos = splitPos - overlap
		} else {
			currentPos = splitPos
		}
	}

	return chunks
}

// SplitForVideo splits text into segments based on estimated reading duration
// Each segment should be approximately 5-6 seconds when spoken
func (tp *TextProcessor) SplitForVideo(text string) []models.VideoSegment {
	text = strings.TrimSpace(text)
	if text == "" {
		return []models.VideoSegment{}
	}

	segments := []models.VideoSegment{}

	// Split into sentences first
	sentences := tp.splitIntoSentences(text)

	currentSegment := ""
	currentDuration := 0.0

	for _, sentence := range sentences {
		sentenceDuration := tp.estimateDuration(sentence)

		// Check if adding this sentence exceeds target duration
		if currentDuration > 0 && currentDuration+sentenceDuration > tp.VideoSegmentDuration {
			// Save current segment
			if currentSegment != "" {
				segments = append(segments, models.VideoSegment{
					Text:              strings.TrimSpace(currentSegment),
					EstimatedDuration: currentDuration,
					VisualPrompt:      "", // Will be generated later
				})
			}
			// Start new segment
			currentSegment = sentence
			currentDuration = sentenceDuration
		} else {
			// Add to current segment
			if currentSegment != "" {
				currentSegment += " " + sentence
			} else {
				currentSegment = sentence
			}
			currentDuration += sentenceDuration
		}
	}

	// Add final segment
	if currentSegment != "" {
		segments = append(segments, models.VideoSegment{
			Text:              strings.TrimSpace(currentSegment),
			EstimatedDuration: currentDuration,
			VisualPrompt:      "",
		})
	}

	return segments
}

// estimateDuration estimates how long it takes to speak the text
// Based on average words per minute (150 words/min for Vietnamese)
func (tp *TextProcessor) estimateDuration(text string) float64 {
	wordCount := tp.countWords(text)
	if wordCount == 0 {
		return 0.0
	}

	// Calculate base duration
	durationMinutes := float64(wordCount) / tp.AvgWordsPerMinute
	durationSeconds := durationMinutes * 60.0

	// Add 10% buffer for natural pauses
	return durationSeconds * 1.1
}

// countWords counts the number of words in text
func (tp *TextProcessor) countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// splitIntoSentences splits text into individual sentences
func (tp *TextProcessor) splitIntoSentences(text string) []string {
	sentences := []string{}
	current := ""

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current += string(runes[i])

		// Check for sentence ending
		if tp.isSentenceEnding(runes[i]) {
			// Look ahead to avoid splitting on abbreviations
			if i+1 < len(runes) && unicode.IsSpace(runes[i+1]) {
				sentence := strings.TrimSpace(current)
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				current = ""
			}
		}
	}

	// Add remaining text
	if current != "" {
		sentence := strings.TrimSpace(current)
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// isSentenceEnding checks if character is a sentence ending
func (tp *TextProcessor) isSentenceEnding(r rune) bool {
	return r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？'
}

// findSentenceBoundary finds the nearest sentence boundary in range
func (tp *TextProcessor) findSentenceBoundary(text string, start, preferredEnd int) int {
	// Search backward from preferredEnd to find sentence ending
	for i := preferredEnd; i > start; i-- {
		if i < len(text) && tp.isSentenceEnding(rune(text[i])) {
			// Found sentence ending, include it
			return i + 1
		}
	}

	// No sentence boundary found
	return -1
}

// findWordBoundary finds the nearest word boundary (space) before position
func (tp *TextProcessor) findWordBoundary(text string, pos int) int {
	// Search backward from pos to find space
	for i := pos; i > 0; i-- {
		if unicode.IsSpace(rune(text[i])) {
			return i
		}
	}

	// Fallback to original position
	return pos
}

// GetStats returns statistics about text processing
func (tp *TextProcessor) GetStats(text string) map[string]interface{} {
	audioChunks := tp.SplitForAudio(text)
	videoSegments := tp.SplitForVideo(text)

	totalVideoDuration := 0.0
	for _, seg := range videoSegments {
		totalVideoDuration += seg.EstimatedDuration
	}

	return map[string]interface{}{
		"total_chars":          len(text),
		"total_words":          tp.countWords(text),
		"audio_chunks":         len(audioChunks),
		"video_segments":       len(videoSegments),
		"estimated_duration":   totalVideoDuration,
		"avg_segment_duration": totalVideoDuration / float64(len(videoSegments)),
	}
}
