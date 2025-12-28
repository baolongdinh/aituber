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
	MaxSubtitleLength    int     // Default: 100 chars
}

// NewTextProcessor creates a new text processor
func NewTextProcessor(audioChunkSize int, videoSegmentDuration float64) *TextProcessor {
	return &TextProcessor{
		AudioChunkSize:       audioChunkSize,
		VideoSegmentDuration: videoSegmentDuration,
		AvgWordsPerMinute:    150.0, // Vietnamese average reading speed
		MaxSubtitleLength:    100,
	}
}

// SplitForAudio splits text into chunks suitable for TTS
// - Maximum characters per chunk defined by AudioChunkSize
// - Splits strictly at sentence boundaries where possible
// - Uses smart splitting for long sentences (punctuation > phrases)
func (tp *TextProcessor) SplitForAudio(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	if len(text) <= tp.AudioChunkSize {
		return []string{text}
	}

	chunks := []string{}
	sentences := tp.splitIntoSentences(text)

	currentChunk := ""

	for _, sentence := range sentences {
		// Clean sentence
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Calculate potential length if we add this sentence
		// Add 1 for space if currentChunk is not empty
		potentialLen := len(currentChunk) + len(sentence)
		if currentChunk != "" {
			potentialLen++
		}

		if potentialLen <= tp.AudioChunkSize {
			// Add to current chunk
			if currentChunk != "" {
				currentChunk += " " + sentence
			} else {
				currentChunk = sentence
			}
		} else {
			// Current chunk full, save it
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}

			// Start new chunk with current sentence
			// If single sentence is too long, we must split it intelligently
			if len(sentence) > tp.AudioChunkSize {
				smartChunks := tp.smartSplit(sentence, tp.AudioChunkSize)
				chunks = append(chunks, smartChunks...)
				currentChunk = ""
			} else {
				currentChunk = sentence
			}
		}
	}

	// Add final chunk
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// SplitForSubtitles splits text into chunks where each chunk is one subtitle line and one audio file.
// Prioritizes readability and sentence boundaries.
func (tp *TextProcessor) SplitForSubtitles(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	chunks := []string{}
	sentences := tp.splitIntoSentences(text)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) <= tp.MaxSubtitleLength {
			chunks = append(chunks, sentence)
			continue
		}

		// Sentence too long, split by clauses (comma, semicolon)
		subChunks := tp.splitByClauses(sentence, tp.MaxSubtitleLength)
		chunks = append(chunks, subChunks...)
	}

	return chunks
}

// splitByClauses splits a long sentence by punctuation (comma, semicolon) or words if needed
func (tp *TextProcessor) splitByClauses(text string, limit int) []string {
	chunks := []string{}

	// Split by comma first
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ';'
	})

	currentMsg := ""
	for i, part := range parts {
		part = strings.TrimSpace(part)

		// Add comma back if it's not the last part (approximation)
		suffix := ""
		if i < len(parts)-1 {
			suffix = ","
		}

		if len(currentMsg)+len(part)+len(suffix)+1 <= limit {
			if currentMsg != "" {
				currentMsg += " " + part + suffix
			} else {
				currentMsg = part + suffix
			}
		} else {
			if currentMsg != "" {
				chunks = append(chunks, currentMsg)
			}

			// Check if the part itself is too long
			if len(part+suffix) > limit {
				// Split using smartSplit (handling words and other punctuation)
				// We use smartSplit instead of splitLongText for better results
				smartChunks := tp.smartSplit(part+suffix, limit)
				chunks = append(chunks, smartChunks...)
				currentMsg = ""
			} else {
				currentMsg = part + suffix
			}
		}
	}

	if currentMsg != "" {
		chunks = append(chunks, currentMsg)
	}

	return chunks
}

// smartSplit splits a long text intelligently based on punctuation priorities
func (tp *TextProcessor) smartSplit(text string, limit int) []string {
	var chunks []string
	remaining := text

	for len(remaining) > limit {
		// Find the best split point within the limit
		// We look backwards from the limit to find the first suitable split point
		splitIdx := -1

		// Search range: we want to find a split point roughly between limit/3 and limit
		// to avoid creating too many tiny chunks, but priority is validity (< limit)
		searchStart := limit / 3
		if searchStart < 0 {
			searchStart = 0
		}

		// 1. Try splitting at major punctuation (comma, semicolon, colon, etc.)
		// Priority: ; : , - —
		punctuations := []string{";", ":", ",", " - ", " — "}
		bestPuncIdx := -1

		for _, punc := range punctuations {
			// Find LAST occurrence of this punctuation within limit
			// We restrict search to the "end region" of the allowed chunk to maximize chunk size
			limitIdx := limit
			if limitIdx > len(remaining) {
				limitIdx = len(remaining)
			}

			// Extract substring to search in
			searchArea := remaining[searchStart:limitIdx]

			if idx := strings.LastIndex(searchArea, punc); idx != -1 {
				// absolute index = searchStart + idx + length of punctuation (to keep punctuation with the first part)
				actualIdx := searchStart + idx + len(punc)

				// Keep punctuation with the preceding chunk usually, or split after it
				if actualIdx > bestPuncIdx {
					bestPuncIdx = actualIdx
				}
			}
		}

		if bestPuncIdx != -1 {
			splitIdx = bestPuncIdx
		} else {
			// 2. Fallback: Split at logical phrase boundaries (spaces)
			// Find last space before limit
			limitIdx := limit
			if limitIdx > len(remaining) {
				limitIdx = len(remaining)
			}

			lastSpace := strings.LastIndex(remaining[:limitIdx], " ")
			if lastSpace != -1 {
				splitIdx = lastSpace
			} else {
				// 3. Last Resort: Hard split at limit
				splitIdx = limit
			}
		}

		// Perform the split
		chunk := strings.TrimSpace(remaining[:splitIdx])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		remaining = strings.TrimSpace(remaining[splitIdx:])
	}

	// Append the rest
	if remaining != "" {
		chunks = append(chunks, remaining)
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
