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
- Giọng điệu: Thẳng thắn, gần gũi như người bạn thông thái, KHÔNG rập khuôn.
- ELEVENLABS AUDIO TAGS: Hãy sử dụng các tag cảm xúc trong "text" để tăng tính hấp dẫn (VD: [laughs], [laughs harder], [whispers], [excited], [sighs], [sarcastic], [curious], [crying]). Đặt tag ở đầu hoặc giữa câu phù hợp với ngữ cảnh.
- Mỗi "text" segment phải TỰ KHÉP: có thể đứng độc lập mà không cần context trước.
- Tránh dùng cụm sáo rỗng như "Thật thú vị", "Hãy cùng tìm hiểu".
- Tổng script: 1000-1500 từ tiếng Việt.

BẮT BUỘC trả về JSON ARRAY (không kèm text gì khác):
[
  {
    "text": "Đoạn script tiếng Việt (100-150 từ)...",
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

	return gs.callGemini(prompt, 0.75, 4096)
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

YÊU CẦU CHẤT LƯỢNG (PO REQUIREMENTS):
- Độ dài: Khoảng 300 - 450 từ (tiếng Việt). Đảm bảo thời lượng đọc tốn khoảng ~1m40s.
- ELEVENLABS AUDIO TAGS: BẮT BUỘC sử dụng các tag cảm xúc như [laughs], [whispers], [excited], [sighs], [sarcastic], [curious], [mischievously] vào trong "text" để tăng tính viral và storytelling.
- NO CLIFFHANGERS: TUYỆT ĐỐI KHÔNG làm nội dung dở dang kiểu "Đón xem Phần 2", "Follow để xem tiếp". Video phải có một kết luận TRỌN VẸN.
- Giọng điệu (Tone): Cuốn hút, chân thật, như một chuyên gia đang ngồi tâm sự mỏng với người xem. Không dùng từ ngữ sáo rỗng hay văn phong rập khuôn của AI. Hãy dùng ngôn ngữ đời thường, sắc sảo.
- Có nhịp điệu (Pacing): Câu ngắn xen lẫn câu dài để tạo nhịp điệu khi đọc.
- Mỗi "text" segment (đoạn) phải dài khoảng 50-80 từ. Số lượng segment nên rơi vào khoảng 5 đến 8 segments để nội dung sâu sắc hơn.

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
	return gs.callGemini(prompt, 0.8, 4096)
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
		if err != nil {
			// Calculate exponential backoff: 2s, 4s, 8s, 16s, 32s, 60s, 60s...
			delay := baseDelay * time.Duration(1<<uint(attempt))
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}

			log.Printf("[Gemini] Attempt %d/%d failed: %v", attempt+1, maxRetries, err)
			log.Printf("[Gemini] Backing off for %v before next attempt...", delay)
			lastErr = err
			time.Sleep(delay)
			continue
		}
		return result, nil
	}

	return nil, fmt.Errorf("all %d Gemini attempts failed. Last error: %w", maxRetries, lastErr)
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

	// Clean markdown block if present
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

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
		platformRule = `- Độ dài: 300-450 từ (TikTok ~1m30-1m50s). Mỗi segment 50-80 từ. Khoảng 5-8 segments.
- Tone: nhanh, sắc bén, storytelling, câu ngắn xen lẫn câu dài`
	} else {
		platformRule = `- Độ dài: 1000-1500 từ (YouTube 5-8 phút). Mỗi segment 100-150 từ. Khoảng 8-12 segments.
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
- ELEVENLABS AUDIO TAGS: Sử dụng các tag cảm xúc [laughs], [whispers], [excited], [sighs], [sarcastic] để làm script sống động hơn.
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

	segments, err := gs.callGemini(prompt, 0.8, 4096)
	if err != nil {
		return nil, fmt.Errorf("series part %d script failed: %w", partIndex+1, err)
	}

	log.Printf("[Gemini] Generated script for series part %d/%d (%d segments)", partIndex+1, totalParts, len(segments))
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

		text := strings.TrimSpace(gemResp.Candidates[0].Content.Parts[0].Text)
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
		return text, nil
	}

	return "", fmt.Errorf("callGeminiRaw failed after %d retries: %w", maxRetries, lastErr)
}
