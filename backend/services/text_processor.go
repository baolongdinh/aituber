package services

import (
	"aituber/models"
	"strings"
	"unicode"
)

// vietnameseStopWords contains common Vietnamese function words to ignore
var vietnameseStopWords = map[string]bool{
	// pronouns / particles
	"tôi": true, "bạn": true, "chúng": true, "ta": true, "họ": true, "mình": true,
	"anh": true, "chị": true, "em": true, "ông": true, "bà": true, "cô": true,
	// prepositions / conjunctions
	"và": true, "hoặc": true, "nhưng": true, "vì": true, "nên": true, "thì": true,
	"mà": true, "để": true, "của": true, "với": true, "trong": true, "trên": true,
	"dưới": true, "ngoài": true, "sau": true, "trước": true, "theo": true,
	"tại": true, "ở": true, "từ": true, "đến": true, "về": true, "qua": true,
	// verbs (common auxiliary)
	"là": true, "được": true, "có": true, "sẽ": true, "đã": true, "đang": true,
	"làm": true, "cho": true, "đây": true, "đó": true, "này": true, "kia": true,
	// quantifiers
	"một": true, "các": true, "những": true, "mọi": true, "nhiều": true, "ít": true,
	"rất": true, "cũng": true, "còn": true, "lại": true, "nữa": true, "thêm": true,
	// common adjectives that are too generic
	"tốt": true, "xấu": true, "lớn": true, "nhỏ": true, "mới": true, "cũ": true,
	"cao": true, "thấp": true, "nhanh": true, "chậm": true,
}

// englishStopWords contains common English stop words to ignore
var englishStopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
	"through": true, "during": true, "is": true, "are": true, "was": true, "were": true,
	"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
	"should": true, "may": true, "might": true, "can": true, "this": true, "that": true,
	"these": true, "those": true, "i": true, "you": true, "he": true, "she": true,
	"we": true, "they": true, "it": true, "its": true, "not": true, "also": true,
	"so": true, "as": true, "if": true, "then": true, "than": true, "when": true,
	"very": true, "just": true, "more": true, "most": true, "such": true,
}

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
		// Priority: ; : , - — .
		punctuations := []string{";", ":", ",", " - ", " — ", "."}
		bestPuncIdx := -1

		// Helper to find punctuation in a range
		findPunc := func(start, end int) int {
			localBestIdx := -1
			for _, punc := range punctuations {
				// Find LAST occurrence of this punctuation within range
				// Extract substring to search in
				if start >= end {
					continue
				}
				searchArea := remaining[start:end]

				if idx := strings.LastIndex(searchArea, punc); idx != -1 {
					// absolute index = start + idx + length of punctuation
					actualIdx := start + idx + len(punc)

					// Keep punctuation with the preceding chunk usually, or split after it
					if actualIdx > localBestIdx {
						localBestIdx = actualIdx
					}
				}
			}
			return localBestIdx
		}

		// First pass: Try preferred range [limit/3, limit]
		limitIdx := limit
		if limitIdx > len(remaining) {
			limitIdx = len(remaining)
		}
		bestPuncIdx = findPunc(searchStart, limitIdx)

		// Second pass: If no punctuation found in preferred range, try [0, limit/3]
		// This prevents "hard splits" when punctuation is only at the start
		if bestPuncIdx == -1 {
			bestPuncIdx = findPunc(0, searchStart)
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

// ExtractKeywordsFromText extracts meaningful keywords from a text segment for use as a Pexels search query.
// It strips common Vietnamese and English stop words and returns up to 5 significant words.
// An optional styleHint (e.g. "cinematic nature") is appended to the result.
func (tp *TextProcessor) ExtractKeywordsFromText(text, styleHint string) string {
	if text == "" {
		if styleHint != "" {
			return styleHint
		}
		return "abstract"
	}

	// Lowercase for processing
	lower := strings.ToLower(text)

	// Remove punctuation, keep letters/digits/spaces
	var cleaned strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			cleaned.WriteRune(r)
		} else {
			cleaned.WriteRune(' ')
		}
	}

	words := strings.Fields(cleaned.String())

	// Filter stop words and very short tokens
	var meaningful []string
	seen := make(map[string]bool)
	for _, w := range words {
		if len([]rune(w)) < 2 {
			continue
		}
		if vietnameseStopWords[w] || englishStopWords[w] {
			continue
		}
		if seen[w] {
			continue // deduplicate
		}
		seen[w] = true
		meaningful = append(meaningful, w)
		if len(meaningful) >= 5 {
			break
		}
	}

	if len(meaningful) == 0 {
		if styleHint != "" {
			return styleHint
		}
		return "abstract"
	}

	result := strings.Join(meaningful, " ")
	if styleHint != "" {
		result += " " + styleHint
	}
	return result
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
