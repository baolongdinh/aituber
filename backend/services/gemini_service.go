package services

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

	"aituber/models"
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

// GenerateYouTubeScript generates a long-form YouTube script from a topic
func (gs *GeminiService) GenerateYouTubeScript(topic string) ([]models.VideoSegment, error) {
	prompt := fmt.Sprintf(`Bạn là chuyên gia tạo content YouTube viral bằng tiếng Việt được 1 triệu view. Hãy viết kịch bản video YouTube về: "%s"

CẤU TRÚC BẮT BUỘC:
- Hook (0-15s): 1 câu hỏi cực mạnh hoặc 1 sự thật gây sốc, tạo tò mò ngay lập tức
- Problem Setup (15-45s): Đặt bối cảnh, tại sao người xem phải quan tâm
- Nội dung chính (3-6 phần, mỗi phần 60-90s): Mỗi phần một insight cụ thể, có dẫn chứng
- CTA cuối: Nhắc subscribe và goiý video liên quan

YÊU CẦU NỘI DUNG:
- Tổng script: 1000-1500 từ tiếng Việt.
- NHIP ĐIỆU CỰC NHANH (BẮT BUỘC): Mỗi "text" segment PHẢI CHỈ TỪ 10-15 từ (tương đương 2-3 giây đọc). 
- ĐA DẠNG BỐI CẢNH (QUY TẮC SỐ 1): Mỗi phân đoạn PHẢI mô tả một bối cảnh hình ảnh (visual_description) và từ khóa (pexels_search_query) HOÀN TOÀN MỚI, KHÁC BIỆT so với đoạn trước đó. Cấm dùng lại Subject cũ hay Action cũ.
- VISUAL HOOK (3S ĐẦU): Các segment thuộc phần Hook (0-5s) PHẢI có mô tả hình ảnh cực dồn dập, màu sắc rực rỡ hoặc bối cảnh gây shock (ví dụ: 'explosive colors', 'dramatic zoom', 'extreme close-up').
- NHẤT QUÁN PHONG CÁCH (STYLE): Hãy chọn 1 phong cách hình ảnh đại diện (ví dụ: 'Cinematic Movie', 'Cyberpunk Digital Art', 'Vintage Film') và áp dụng phong cách đó vào MỌI visual_description để video đồng nhất.
- Mảng JSON trả về phải có từ 30 đến 50 phần tử. KHÔNG ĐƯỢC LƯỜI BIẾNG.

BẮT BUỘC trả về JSON ARRAY (không kèm text gì khác):
[
  {
    "text": "Đoạn script tiếng Việt (khoảng 50-100 từ)...",
    "pexels_search_query": "person running fast stress city",
    "visual_description": "Cinematic close-up of a young man with a determined expression, sprinting through a crowded neon-lit futuristic city street at night, heavy rain falling, motion blur in the background, 4k ultra-realistic."
  }
]

QUY TẮC pexels_search_query (BẮT BUỘC):
1. Phải là tiếng Anh ngắn gọn (2-5 từ)
2. PHẢI có ĐỘNG TỪ mô tả chuyển động: running, falling, flying, exploding, rising, spinning, zooming
3. Mô tả hình ảnh B-roll trực quan, KHÔNG trừu tượng
4. Ví dụ TỐT: "money falling slow motion", "athlete running sunrise", "city timelapse traffic"
5. Ví dụ XẤU: "success", "teamwork", "growth" (quá chung chung)`, topic)

	result, err := gs.callGemini(prompt, 0.75, 8192)
	if err != nil {
		return nil, err
	}
	return gs.postProcessSegments(result), nil
}

// GenerateTikTokScript generates a short, viral TikTok script from a topic
func (gs *GeminiService) GenerateTikTokScript(topic string) ([]models.VideoSegment, error) {
	prompt := fmt.Sprintf(`Bạn là chuyên gia Content Creator mảng TikTok/Shorts (hàng triệu view) với phong cách sâu sắc, lôi cuốn và kể chuyện (storytelling) cực đỉnh bằng tiếng Việt. Viết 1 kịch bản TikTok/Shorts có thời lượng tự nhiên rơi vào khoảng 1 phút 30 giây đến 1 phút 50 giây (1m30s - 1m50s) về: "%s"

CẤU TRÚC BẮT BUỘC:
- HOOK (0-5s): 1 câu mở đầu mang tính lật đổ nhận thức thông thường hoặc đánh trúng tim đen. Phải cực cháy!
- STORY/PROBLEM SETUP (5-20s): Đưa ra một câu chuyện ngắn hoặc đào sâu vào nỗi đau/vấn đề. Tạo sự đồng cảm mạnh mẽ.
- THE "MEAT" / INSIGHTS DUMPS (20-80s): 3-4 góc nhìn hoặc bài học thực chiến sâu sắc. Đừng nói những thứ chung chung ai cũng biết trên mạng. Phải có dẫn chứng, ví dụ thực tế hoặc logic thuyết phục.
- CLIMAX & PAYOFF (80-100s): Cú twist, bài học đọng lại hoặc kết luận thay đổi tư duy.
- CTA (100s+): Kêu gọi hành động tự nhiên, không gượng ép (VD: "Lưu video này lại để...").

- YÊU CẦU CHẤT LƯỢNG (PO REQUIREMENTS):
- Độ dài: Khoảng 300 - 450 từ (tiếng Việt). Đảm bảo thời lượng ~1m40s.
- NHIP ĐIỆU CỰC NHANH (BẮT BUỘC): Mỗi segment "text" PHẢI CHỈ TỪ 8-12 từ (~1.5 đến 2 giây đọc). 
- CẢNH MỚI HOÀN TOÀN: Tuyệt đối không để một bối cảnh hình ảnh lặp lại. Mỗi segment phải là một hình ảnh mới hoàn toàn.
- VISUAL HOOK (CỰC CHÁY): Trong 3-5 giây đầu tiên, hình ảnh phải cực kỳ kịch tính, màu sắc mạnh hoặc bối cảnh gây shock (VD: 'explosive colors', 'dramatic zoom').
- ĐỒNG NHẤT STYLE: Chọn duy nhất 1 Tone màu và Style nghệ thuật cho tất cả visual_description xuyên suốt video (VD: 'Cyberpunk', 'Cinematic Movie', 'Anime').
- PHÂN ĐOẠN DÀY ĐẶC: Mảng JSON trả về PHẢI BAO GỒM TỪ 20 ĐẾN 30 PHẦN TỬ.
- Tone: Cuốn hút, chân thật, ngôn ngữ đời thường sắc sảo. Như một chuyên gia đang tâm sự mỏng, không dùng từ sáo rỗng.

BẮT BUỘC trả về JSON ARRAY (không kèm text gì khác ngoài JSON):
[
  {
    "text": "Câu Hook hoặc một đoạn kịch bản ngắn...",
    "pexels_search_query": "shocked face close up slow motion",
    "visual_description": "Dramatic low-angle shot of a person dropping their phone in slow motion, eyes wide in disbelief, busy subway station background, high contrast lighting, cinematic aesthetic."
  }
]

QUY TẮC pexels_search_query (BẮT BUỘC):
1. Là Tiếng Anh, ngắn gọn 2-5 từ, TẬP TRUNG vào HÀNH ĐỘNG/CHUYỂN ĐỘNG hoặc BIỂU CẢM.
2. Phù hợp hoàn hảo với mood của đoạn text đó.
3. Ví dụ TỐT: "person stressed working late", "money falling slow motion", "city crowd walking fast", "brain exploding idea".
4. TUYỆT ĐỐI KHÔNG dùng từ trừu tượng kiểu "success", "mindset". Không dùng tiếng Việt.`, topic)

	// Temperature 0.8 to encourage creative, natural storytelling
	result, err := gs.callGemini(prompt, 0.8, 8192)
	if err != nil {
		return nil, err
	}
	return gs.postProcessSegments(result), nil
}

// callGemini calls the Gemini API and parses response into JSON segment array
func (gs *GeminiService) callGemini(prompt string, temperature float64, maxTokens int) ([]models.VideoSegment, error) {
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

// postProcessSegments cuts Gemini's standard-length segments into smaller "fast paced" sub-segments (~10-15 words)
// and copies the VisualPrompt to each new piece. This simulates the pacing without straining the LLM.
// It prioritizes splitting by punctuation (comma, period) to maintain grammatical flow,
// falling back to word count if a clause is too long.
func (gs *GeminiService) postProcessSegments(raw []models.VideoSegment) []models.VideoSegment {
	var final []models.VideoSegment

	for _, seg := range raw {
		text := strings.TrimSpace(seg.Text)
		if text == "" {
			continue
		}

		// Chia đoạn văn bản một cách thông minh: dựa vào dấu câu trước
		var chunks []string
		var currentChunk strings.Builder
		wordCount := 0

		words := strings.Fields(text)
		maxWords := 15 // Kích thước tối đa cho 1 mẩu để không quá dài

		for i, word := range words {
			currentChunk.WriteString(word)
			currentChunk.WriteString(" ")
			wordCount++

			// Kiểm tra từ cuối cùng có gắn dấu chấm, phẩy...
			hasPunc := false
			if len(word) > 0 {
				lastChar := word[len(word)-1]
				hasPunc = lastChar == '.' || lastChar == ',' || lastChar == '!' || lastChar == '?' || lastChar == ';' || lastChar == ':'
			}

			// Điều kiện ngắt thành 1 segment con:
			// 1. Chứa dấu câu, VÀ vế này đủ dài (>= 3 chữ) để tránh các câu cụt lủn (ví dụ: "Vâng,")
			// 2. HOẶC chứa đủ lượng token giới hạn (tránh vế dài lê thê)
			// 3. HOẶC là từ cuối cùng của segment
			isLastWord := i == len(words)-1

			if (hasPunc && wordCount >= 3) || wordCount >= maxWords || isLastWord {
				chunkText := strings.TrimSpace(currentChunk.String())
				if chunkText != "" {
					chunks = append(chunks, chunkText)
				}
				currentChunk.Reset()
				wordCount = 0
			}
		}

		// Tạo segments
		for _, chunkText := range chunks {
			// Loại bỏ các dấu câu cuối dòng cũ để ép thêm vào duy nhất đúng một dấu '.' cho TTS ngắt
			chunkText = strings.TrimRight(chunkText, ".,!?;: ")
			if chunkText == "" {
				continue
			}

			finalSeg := models.VideoSegment{
				Text:              chunkText + ".", // Add period so TTS pauses properly
				EstimatedDuration: seg.EstimatedDuration,
				VisualPrompt:      seg.VisualPrompt,
				VisualDescription: seg.VisualDescription,
			}
			final = append(final, finalSeg)
		}
	}

	return final
}

// callWithKey calls Gemini and returns parsed json
func (gs *GeminiService) callWithKey(apiKey, prompt string, temperature float64, maxTokens int) ([]models.VideoSegment, error) {
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

	var segments []models.VideoSegment
	if err := json.Unmarshal([]byte(text), &segments); err != nil {
		return nil, fmt.Errorf("failed to parse JSON script. Error: %w. Raw text: %s", err, text)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("parsed JSON script is empty array")
	}

	log.Printf("[Gemini] Generated JSON script with %d segments", len(segments))
	return segments, nil
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
func (gs *GeminiService) GenerateSeriesOutline(topic, platform string, numParts int) ([]models.SeriesPartOutline, error) {
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

	var outlines []models.SeriesPartOutline
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
func (gs *GeminiService) GenerateSeriesPartScript(topic, platform string, outlines []models.SeriesPartOutline, partIndex int) ([]models.VideoSegment, error) {
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

BẮT BUỘC trả về JSON ARRAY (không có text nào khác):
[
  {
    "text": "Đoạn script tiếng Việt...",
    "pexels_search_query": "english short action keywords",
    "visual_description": "Cinematic 4k detailed description of the scene with character actions, lighting, and camera angle in English."
  }
]

QUY TẮC pexels_search_query:
1. Tiếng Anh, 2-5 từ, có ĐỘNG TỪ/CHUYỂN ĐỘNG
2. Ví dụ TỐT: "money falling slow motion", "person stressed working late"
3. KHÔNG dùng từ trừu tượng như "success", "mindset"`,
		seriesCtx.String(),
		topic,
		partIndex+1, totalParts, part.Title,
		part.Summary,
		strings.Join(part.KeyPoints, ", "),
		hookRule,
		platformRule,
	)

	segments, err := gs.callGemini(prompt, 0.8, 8192)
	if err != nil {
		return nil, fmt.Errorf("series part %d script failed: %w", partIndex+1, err)
	}

	segments = gs.postProcessSegments(segments)

	log.Printf("[Gemini] Generated script for series part %d/%d (%d sub-segments)", partIndex+1, totalParts, len(segments))
	return segments, nil
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
