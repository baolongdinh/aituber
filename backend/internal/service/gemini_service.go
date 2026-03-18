package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// GeminiService generates video scripts using Google Gemini API
type GeminiService struct {
	apiKeys    []string
	keyIndex   uint64
	httpClient *http.Client
}

// NewGeminiService creates a new Gemini service with round-robin key rotation
func NewGeminiService(apiKeys []string) *GeminiService {
	return &GeminiService{
		apiKeys: apiKeys,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HasKeys returns true if at least one API key is configured
func (gs *GeminiService) HasKeys() bool {
	return len(gs.apiKeys) > 0
}

// getNextKey returns the next API key in round-robin fashion
func (gs *GeminiService) getNextKey() (string, error) {
	if len(gs.apiKeys) == 0 {
		return "", fmt.Errorf("no Gemini API keys configured")
	}
	idx := atomic.AddUint64(&gs.keyIndex, 1) % uint64(len(gs.apiKeys))
	return gs.apiKeys[idx], nil
}

// geminiRequest is the request body for Gemini API
type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig geminiGenConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

// geminiResponse is the response from Gemini API
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (gs *GeminiService) GenerateYouTubeScript(topic string) (*GeneratedScript, error) {
	prompt := fmt.Sprintf(`You are an expert Vietnamese YouTube scriptwriter AND visual director. Your task is to write a complete video script about: "%s"

PHASE 1 — GENRE DETECTION & STYLE BIBLE:
First, identify the genre of this topic: (Documentary/News, Science/Tech, History, Fashion/Lifestyle, Nature, etc.)
Then define the VISUAL STYLE BIBLE for this genre:
- SUBJECT/NARRATOR: One consistent visual anchor (e.g., for documentary: "Middle Eastern cityscape at dusk", for science: "Modern lab setting with specific equipment", for news: "War journalist in tactical vest")
- VISUAL GRAMMAR: Camera style (drone, handheld, static), color grade (warm, cold, desaturated), film quality (8K sharp, gritty grain, cinematic)
- MOOD: Atmosphere that fits the topic (tense, inspiring, mysterious, urgent)

PHASE 2 — WRITE THE SCRIPT:
Structure: Hook → Problem → Main content → CTA
Length: 1000–1500 Vietnamese words. Each text segment: 10–15 words.
Total: 30–50 JSON segments.

CRITICAL RULE FOR visual_description:
Each visual_description MUST ILLUSTRATE EXACTLY what the "text" is saying at that moment.
ASK YOURSELF: "If someone reads this text, what specific image would make a great B-roll for it?"

Example mapping (topic: Middle East conflict):
- text: "Hàng triệu người dân phải rời bỏ nhà cửa vì chiến tranh"
  ✅ visual_description: "Low-angle cinematic shot, Syrian refugee camp at dawn, rows of white canvas tents stretching to horizon, exhausted families with bundled belongings walking along dusty road, children clutching parents, pale morning light, desaturated color grade, 8K documentary style."
  ❌ BAD: "A man walks past a cracked wall" (unrelated to mass displacement)

- text: "Dầu mỏ — thứ vàng đen khiến cả thế giới nhòm ngó Trung Đông"
  ✅ visual_description: "Extreme wide shot, vast oil field at sunset in Saudi Arabia desert, dozens of pump jacks rhythmically nodding, orange-red sky reflecting off black crude pools, thick industrial pipes, shallow depth of field, cinematic 8K."

FORMULA for EVERY visual_description:
"[Camera angle + movement], [Location/Setting that fits the TEXT content], [Specific objects/people/action that ILLUSTRATE the text], [Lighting that matches mood], [Texture details], [Film style]."

ADDITIONAL RULES:
- visual_description: ENGLISH ONLY, min 60 words.
- pexels_search_query: 4–6 English keywords matching the visual.
- text: Vietnamese only.
- **CRITICAL**: Always end "text" segments at a logical punctuation mark (.,!?). Never split a meaningful phrase (like "chúng tôi, đi làm, học tập") across segments.
- Keep consistent visual grammar (same film style, same color grade) across segments.
- NEVER generate a visual that contradicts or ignores the text content.
- Return ONLY a JSON object with "title" (a catchy Vietnamese title) and "segments" array.

{
  "title": "Tiêu đề video cực hay và gây tò mò",
  "segments": [
    {
      "text": "Vietnamese narration (10–15 words)...",
      "pexels_search_query": "english keywords matching visual",
      "visual_description": "Cinematic description that ILLUSTRATES the text content (min 60 words, English)."
    }
  ]
}`, topic)

	result, err := gs.callGemini(prompt, 0.65, 8192)
	if err != nil {
		return nil, err
	}
	result.Segments = gs.postProcessSegments(result.Segments)
	return result, nil
}

func (gs *GeminiService) GenerateTikTokScript(topic string) (*GeneratedScript, error) {
	prompt := fmt.Sprintf(`You are a viral Vietnamese TikTok director and storyteller. Write a complete 1m30s–2m TikTok script in Vietnamese about: "%s"

DURATION REQUIREMENTS (NON-NEGOTIABLE — you MUST meet these):
- MINIMUM 25 JSON segments. Target is 28-30 segments. COUNT before you output.
- Each "text" segment: 10-18 Vietnamese words. NO shorter.
- **CRITICAL**: Every segment MUST end with a punctuation mark (.,!?) to ensure natural TTS pauses. 
- Never split a meaningful phrase (like "anh em", "gia đình") between segments.
- Total words across ALL text segments combined: MINIMUM 280 words.
- If you are about to stop before 25 segments, detect this and ADD MORE content.

SCRIPT STRUCTURE (expand each section, do not rush):
- Segments 1-3: Hook. A shocking statistic or controversial question that grabs attention.
- Segments 4-8: Set the scene. Historical context, geography, key players involved.
- Segments 9-14: Cause and effect. Why did this happen? Who benefits? Show the chain of events.
- Segments 15-20: Evidence and turning points. Specific events, dates, consequences.
- Segments 21-25: The hidden angle. What most people don't know. The unexpected twist.
- Segments 26-30: Emotional conclusion. What does this mean for us? CTA to follow.

VISUAL RULE: Each visual_description must ILLUSTRATE EXACTLY what the text says.
Examples for a Middle East documentary topic:
- text: "Hàng chục quốc gia cung cấp vũ khi cho cac phe phai khac nhau"
  visual: "Wide aerial shot, dusty Syrian border crossing, military convoys with various national markings crossing checkpoint, armed soldiers inspecting cargo trucks, harsh afternoon desert light, documentary drone footage style."
- text: "Dau mo bien Trung Dong thanh chiec banh ngot ma ca the gioi muon giang tay"
  visual: "Extreme wide shot, sprawling Saudi oil field at golden hour, hundreds of pump jacks moving rhythmically, thick black pipelines stretching to horizon, orange sky reflecting off crude oil surface, cinematic 8K."

FORMULA: "[Camera], [Specific subject from TEXT], [Action mirroring narration], [Lighting], [Film style]."

CONSISTENCY: Use same color grade and film style in ALL visual_descriptions.

SELF-CHECK: Before returning, count your segments. If less than 25, add more script content.

Return ONLY a valid JSON object. No markdown, no explanation:
{
  "title": "Tiêu đề video cực viral",
  "segments": [
    {
      "text": "Cau tieng Viet 10-16 tu...",
      "pexels_search_query": "english keywords",
      "visual_description": "Cinematic B-roll mirroring text, English, min 50 words."
    }
  ]
}`, topic)

	result, err := gs.callGemini(prompt, 0.70, 8192)
	if err != nil {
		return nil, err
	}
	result.Segments = gs.postProcessSegments(result.Segments)
	return result, nil
}

// callGemini calls the Gemini API and parses response into GeneratedScript
func (gs *GeminiService) callGemini(prompt string, temperature float64, maxTokens int) (*GeneratedScript, error) {
	if !gs.HasKeys() {
		return nil, fmt.Errorf("no Gemini API keys configured")
	}

	maxRetries := 8 // Support up to 8 sequential attempts
	baseDelay := 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		apiKey, err := gs.getNextKey()
		if err != nil {
			return nil, err
		}

		result, err := gs.callWithKey(apiKey, prompt, temperature, maxTokens)
		if err == nil {
			return result, nil
		}

		// Calculate exponential backoff: 2s, 4s, 8s, 16s, 32s, 60s, 60s...
		delay := baseDelay * time.Duration(1<<uint(attempt))
		if delay > 60*time.Second {
			delay = 60 * time.Second
		}

		log.Printf("[Gemini] Attempt %d/%d failed: %v", attempt+1, maxRetries, err)
		log.Printf("[Gemini] Backing off for %v before next attempt...", delay)
		lastErr = err
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("all %d Gemini attempts failed. Last error: %w", maxRetries, lastErr)
}

// postProcessSegments cuts Gemini's standard-length segments into smaller "fast paced" sub-segments (~10-25 words)
// and copies the VisualPrompt to each new piece. This simulates the pacing without straining the LLM.
// It prioritizes splitting by punctuation (comma, period) to maintain natural audio flow.
func (gs *GeminiService) postProcessSegments(raw []VideoSegment) []VideoSegment {
	var final []VideoSegment

	for _, seg := range raw {
		text := strings.TrimSpace(seg.Text)
		if text == "" {
			continue
		}

		words := strings.Fields(text)
		// Intelligent segment splitting for long texts or those with punctuation
		var chunks []string
		var currentChunk strings.Builder
		wordCount := 0
		maxWords := 25 // Target chunk size

		for i, word := range words {
			currentChunk.WriteString(word)
			currentChunk.WriteString(" ")
			wordCount++

			// Check for punctuation at the end of the word
			hasPunc := false
			if len(word) > 0 {
				lastChar := word[len(word)-1]
				// Recognize standard sentence/clause boundaries
				hasPunc = lastChar == '.' || lastChar == ',' || lastChar == '!' || lastChar == '?' || lastChar == ';' || lastChar == ':'
			}

			isLastWord := i == len(words)-1

			// Split conditions:
			// 1. We hit a punctuation mark AND we have a decent amount of text (>= 6 words to avoid tiny fragments)
			// 2. We are WAY over the limit (maxWords * 1.5) and MUST split at a space to prevent FPT AI issues
			// 3. It's the end of the text
			if (hasPunc && wordCount >= 3) || wordCount >= (maxWords+10) || isLastWord {
				chunkText := strings.TrimSpace(currentChunk.String())
				if chunkText != "" {
					chunks = append(chunks, chunkText)
				}
				currentChunk.Reset()
				wordCount = 0
			}
		}

		for _, chunkText := range chunks {
			chunkText = strings.TrimRight(chunkText, ".,!?;: ")
			if chunkText == "" {
				continue
			}

			finalSeg := VideoSegment{
				Text:              chunkText + ".", // Add period so TTS pauses properly
				VisualPrompt:      seg.VisualPrompt,
				VisualDescription: seg.VisualDescription,
			}
			final = append(final, finalSeg)
		}
	}

	return final
}

// callWithKey calls Gemini and returns parsed GeneratedScript
func (gs *GeminiService) callWithKey(apiKey, prompt string, temperature float64, maxTokens int) (*GeneratedScript, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite-preview:generateContent?key=%s", apiKey)

	// Note: We use system instructions implicitly in the prompt, or we can use the response_mime_type feature in Gemini API but raw text is fine when formatted.
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxTokens,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(body, &gemResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w. Body: %s", err, string(body))
	}

	if gemResp.Error != nil {
		return nil, fmt.Errorf("gemini API error %d: %s", gemResp.Error.Code, gemResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(body))
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response. Body: %s", string(body))
	}

	text := gemResp.Candidates[0].Content.Parts[0].Text
	if text == "" {
		return nil, fmt.Errorf("gemini returned empty text")
	}

	// Robust JSON extraction
	text = gs.extractJSON(text)

	var script GeneratedScript
	if err := json.Unmarshal([]byte(text), &script); err != nil {
		// Attempt to parse as array of segments for backward compatibility or LLM laziness
		var segments []VideoSegment
		if err2 := json.Unmarshal([]byte(text), &segments); err2 == nil {
			log.Printf("[Gemini] LLM returned array instead of object, wrapping it")
			return &GeneratedScript{Segments: segments}, nil
		}
		return nil, fmt.Errorf("failed to parse JSON script. Error: %w. Raw text: %s", err, text)
	}

	if len(script.Segments) == 0 {
		return nil, fmt.Errorf("parsed JSON script has no segments")
	}

	log.Printf("[Gemini] Generated JSON script with %d segments and title: %q", len(script.Segments), script.Title)
	return &script, nil
}

// GenerateImageForKeyword generates a stock-style cinematic image using gemini-2.5-flash-image.
// Returns raw PNG bytes. Used as fallback when Pexels is unavailable.
// orientation: "portrait" (9:16 for TikTok) or "landscape" (16:9 for YouTube).
// visualDesc: optional cinematic scene description from the video script (preferred over keyword when non-empty).
func (gs *GeminiService) GenerateImageForKeyword(keyword, visualDesc, orientation string) ([]byte, error) {
	if !gs.HasKeys() {
		return nil, fmt.Errorf("no Gemini API keys configured")
	}

	apiKey, err := gs.getNextKey()
	if err != nil {
		return nil, err
	}

	// Map orientation to supported aspect ratio
	aspectRatio := "16:9"
	if orientation == "portrait" {
		aspectRatio = "9:16"
	}

	// Build image prompt: prefer rich visual_description from script; fall back to short keyword.
	var imagePrompt string
	if visualDesc != "" {
		// visualDesc is already a detailed cinematic description – just enforce aspect ratio and quality constraints.
		imagePrompt = fmt.Sprintf(
			"%s "+
				"Aspect ratio %s. Photorealistic, high resolution, no text, no watermarks.",
			visualDesc, aspectRatio,
		)
	} else {
		// Fallback: craft a generic cinematic prompt from the short keyword.
		imagePrompt = fmt.Sprintf(
			"Professional cinematic B-roll stock photo: %s. "+
				"Dramatic lighting, shallow depth of field, high resolution, "+
				"photorealistic, no text, no watermarks, no people faces. "+
				"Aspect ratio %s. Suitable for a documentary or news video segment.",
			keyword, aspectRatio,
		)
	}

	// gemini-2.5-flash-image uses the standard generateContent endpoint
	// with responseModalities set to IMAGE
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image-preview:generateContent?key=%s",
		apiKey,
	)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": imagePrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"IMAGE"},
			"temperature":        1.0,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create image request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("image generation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini-2.5-flash-image returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse standard Gemini response – image is in candidates[0].content.parts[0].inlineData
	var gemResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"` // base64
					} `json:"inlineData,omitempty"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &gemResp); err != nil {
		return nil, fmt.Errorf("failed to parse image response: %w. Body: %s", err, string(body))
	}

	if len(gemResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini-2.5-flash-image returned no candidates. Body: %s", string(body))
	}

	for _, part := range gemResp.Candidates[0].Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != "" {
			imgBytes, decErr := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if decErr != nil {
				return nil, fmt.Errorf("failed to decode base64 image: %w", decErr)
			}
			log.Printf("[Gemini/Image] Generated %s fallback image for %q (%d bytes, mime: %s)",
				orientation, keyword, len(imgBytes), part.InlineData.MimeType)
			return imgBytes, nil
		}
	}

	return nil, fmt.Errorf("gemini-2.5-flash-image returned no image data in parts. Body: %s", string(body))
}

// ---------- Series Video Generation ----------

// GenerateSeriesOutline generates a structured outline for a multi-part series.
// Returns ordered list of SeriesPartOutline (one per part).
func (gs *GeminiService) GenerateSeriesOutline(topic, platform string, numParts int) ([]SeriesPartOutline, error) {
	if !gs.HasKeys() {
		return nil, fmt.Errorf("no Gemini API keys configured")
	}

	contentType := "TikTok (1-2 phút, ngắn, viral)"
	if platform == "youtube" {
		contentType = "YouTube (5-10 phút, sâu hơn)"
	}

	prompt := fmt.Sprintf(`Bạn là chuyên gia chiến lược content series viral bằng tiếng Việt.

Chủ đề series: "%s"
Platform: %s
Số tập: %d

Hãy tạo outline cho %d tập, đảm bảo:
1. Mỗi tập có thể xem ĐỘC LẬP (không cần xem tập trước)
2. Toàn series có mạch logic tăng dần: từ vấn đề → nguyên nhân → giải pháp → nâng cao → tổng kết
3. Mỗi tập có góc nhìn RIÊNG BIỆT, không trùng lặp
4. Tập 1 phải có hook mạnh để kéo người vào series
5. Tập cuối phải có cảm giác "hoàn chỉnh"

BẮT BUỘC trả về JSON ARRAY (không có text nào khác):
[
  {
    "part_number": 1,
    "title": "Tiêu đề tập (ngắn gọn, gây tò mò)",
    "summary": "Tóm tắt nội dung 1-2 câu",
    "key_points": ["Điểm chính 1", "Điểm chính 2", "Điểm chính 3"]
  }
]`, topic, contentType, numParts, numParts)

	rawText, err := gs.callGeminiRaw(prompt, 0.7, 4096)
	if err != nil {
		return nil, fmt.Errorf("series outline generation failed: %w", err)
	}

	var outlines []SeriesPartOutline
	if err := json.Unmarshal([]byte(rawText), &outlines); err != nil {
		return nil, fmt.Errorf("failed to parse series outline JSON: %w. Raw: %s", err, rawText)
	}
	if len(outlines) == 0 {
		return nil, fmt.Errorf("series outline is empty")
	}

	log.Printf("[Gemini] Generated series outline: %d parts for topic: %q", len(outlines), topic)
	return outlines, nil
}

// GenerateSeriesPartScript generates a full video script for a single part of a series.
// `outlines` is the full series outline for context. `partIndex` is 0-based.
func (gs *GeminiService) GenerateSeriesPartScript(topic, platform string, outlines []SeriesPartOutline, partIndex int) (*GeneratedScript, error) {
	if partIndex < 0 || partIndex >= len(outlines) {
		return nil, fmt.Errorf("partIndex %d out of range (total %d)", partIndex, len(outlines))
	}

	part := outlines[partIndex]
	totalParts := len(outlines)

	// Build neighboring context
	prevTitle := ""
	if partIndex > 0 {
		prevTitle = outlines[partIndex-1].Title
	}
	nextTitle := ""
	if partIndex < totalParts-1 {
		nextTitle = outlines[partIndex+1].Title
	}

	// Build full series context summary for Gemini
	var seriesCtx strings.Builder
	seriesCtx.WriteString("TOÀN BỘ SERIES:\n")
	for i, o := range outlines {
		marker := ""
		if i == partIndex {
			marker = " ← TẬP NÀY"
		}
		seriesCtx.WriteString(fmt.Sprintf("  Tập %d: %s%s\n", o.PartNumber, o.Title, marker))
	}

	isFirstPart := partIndex == 0
	isLastPart := partIndex == totalParts-1

	hookRule := ""
	if isFirstPart {
		hookRule = "- Đây là TẬP ĐẦU TIÊN: hook phải cực mạnh, giới thiệu nhẹ về series (VD: \"Đây là bí mật số 1 trong chuỗi X điều...\")"
	} else if isLastPart {
		hookRule = fmt.Sprintf("- Đây là TẬP CUỐI: hook thừa nhận đây là phần cuối, kết thúc bằng bài học tổng quát. Tập trước là: \"%s\"", prevTitle)
	} else {
		hookRule = fmt.Sprintf("- Đây là tập %d/%d, tập trước: \"%s\", tập sau: \"%s\". Hook KHÔNG được spoil tập sau, KHÔNG cần nhắc tập trước trực tiếp",
			partIndex+1, totalParts, prevTitle, nextTitle)
	}

	var platformRule string
	if platform == "tiktok" {
		platformRule = `- Độ dài: 300-450 từ (TikTok ~1m30-1m50s). Mỗi segment 8-12 từ. Khoảng 20-30 segments.
- ĐA DẠNG THỊ GIÁC: Mỗi segment BẮT BUỘC có bối cảnh hình ảnh mới hoàn toàn. Cấm lặp lại. 
- VISUAL HOOK: Ép Hook cực mạnh, dồn dập trong 5s đầu.
- STYLE: Đồng nhất một phong cách nghệ thuật xuyên suốt.
- Tone: nhanh, sắc bén, storytelling, câu ngắn xen lẫn câu dài`
	} else {
		platformRule = `- Độ dài: 1000-1500 từ (YouTube 5-8 phút). Mỗi segment 10-15 từ. Khoảng 40-60 segments.
- ĐA DẠNG THỊ GIÁC: Tách segment mới ngay khi bối cảnh hình ảnh thay đổi. Cấm dùng 1 visual cho nhiều segment lời thoại.
- STYLE: Duy trì tính thẩm mỹ nhất quán qua MỌI phân đoạn.
- Tone: thẳng thắn, sâu sắc, có dẫn chứng cụ thể`
	}

	prompt := fmt.Sprintf(`Bạn là chuyên gia Content Creator series viral tiếng Việt.

%s

CHỦ ĐỀ SERIES: "%s"
TẬP NÀY: Tập %d/%d – "%s"
TÓM TẮT: %s
ĐIỂM CHÍNH CẦN COVER: %s

LUẬT BẮT BUỘC:
%s
%s
- TUYỆT ĐỐI KHÔNG kết thúc bằng "Xem tập tiếp theo để biết..." hay bất kỳ cliffhanger nào
- Mỗi tập phải có kết luận TỰ HOÀN CHỈNH
- Không sáo rỗng, không văn phong AI cứng nhắc

BẮT BUỘC trả về JSON OBJECT (không có text nào khác):
{
  "title": "Tiêu đề tập cực hay (Vietnamese)",
  "segments": [
    {
      "text": "Đoạn script tiếng Việt ngắn (8-15 từ)...",
      "pexels_search_query": "english short action keywords",
      "visual_description": "Consistent cinematic description (UNIVERSAL GOLD STANDARD): [Consistent Subject + Physics-based Material Details] + [Detailed Action] + [Lighting/Environment] + [Ultra-sharp 8k details]."
    }
  ]
}

QUY TẮC NHẤT QUÁN:
1. LUÔN CHỌN CHỦ THỂ CỐ ĐỊNH: Nhân vật hoặc vật thể chính phải có mô tả thuộc tính vật lý cụ thể để giữ consistency.
2. SIÊU CHI TIẾT: Visual_description phải đạt chuẩn chuyên nghiệp, mô tả rõ chất liệu (kim loại, vải, khói, bụi,...) và ánh sáng thực tế.`,
		seriesCtx.String(),
		topic,
		partIndex+1, totalParts, part.Title,
		part.Summary,
		strings.Join(part.KeyPoints, ", "),
		hookRule,
		platformRule,
	)

	script, err := gs.callGemini(prompt, 0.8, 8192)
	if err != nil {
		return nil, fmt.Errorf("series part %d script failed: %w", partIndex+1, err)
	}

	script.Segments = gs.postProcessSegments(script.Segments)

	log.Printf("[Gemini) Generated script for series part %d/%d (%d sub-segments)", partIndex+1, totalParts, len(script.Segments))
	return script, nil
}

// callGeminiRaw calls Gemini and returns the raw text response (no JSON parsing).
func (gs *GeminiService) callGeminiRaw(prompt string, temperature float64, maxTokens int) (string, error) {
	maxRetries := 5
	baseDelay := 2 * time.Second
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		apiKey, err := gs.getNextKey()
		if err != nil {
			return "", err
		}

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite-preview:generateContent?key=%s", apiKey)
		reqBody := geminiRequest{
			Contents:         []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
			GenerationConfig: geminiGenConfig{Temperature: temperature, MaxOutputTokens: maxTokens},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := gs.httpClient.Do(req)
		if err != nil {
			lastErr = err
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}
			time.Sleep(delay)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var gemResp geminiResponse
		if err := json.Unmarshal(body, &gemResp); err != nil {
			lastErr = fmt.Errorf("parse error: %w", err)
			continue
		}
		if gemResp.Error != nil {
			lastErr = fmt.Errorf("API error %d: %s", gemResp.Error.Code, gemResp.Error.Message)
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response")
			continue
		}

		text := gemResp.Candidates[0].Content.Parts[0].Text
		text = gs.extractJSON(text)
		return text, nil
	}

	return "", fmt.Errorf("callGeminiRaw failed after %d retries: %w", maxRetries, lastErr)
}

// extractJSON finds the first complete JSON block [...] or {...} in a string.
// It uses bracket balancing to support nested structures without being fooled by
// additional text or multiple JSON blocks.
func (gs *GeminiService) extractJSON(text string) string {
	startArray := strings.Index(text, "[")
	startObj := strings.Index(text, "{")

	start := -1
	var open, close byte

	if startArray != -1 && (startObj == -1 || startArray < startObj) {
		start = startArray
		open = '['
		close = ']'
	} else if startObj != -1 {
		start = startObj
		open = '{'
		close = '}'
	}

	if start == -1 {
		return text
	}

	count := 0
	for i := start; i < len(text); i++ {
		if text[i] == open {
			count++
		} else if text[i] == close {
			count--
			if count == 0 {
				return text[start : i+1]
			}
		}
	}

	// If we couldn't find a matching closing bracket,
	// return from start to end as a fallback.
	return text[start:]
}
