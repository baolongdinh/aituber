package services

import (
	"bytes"
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
	prompt := fmt.Sprintf(`Bạn là một chuyên gia tạo content YouTube bằng tiếng Việt. Hãy viết một kịch bản video YouTube hoàn chỉnh về chủ đề: "%s"

YÊU CẦU:
- Ngôn ngữ: Tiếng Việt tự nhiên, dễ hiểu
- Cấu trúc: Hook mạnh → Intro → Nội dung chính (3-5 phần) → Kết luận + CTA
- Giọng điệu: Thân thiện, chuyên nghiệp, có chiều sâu
- QUAN TRỌNG: Chỉ viết lời thoại, KHÔNG có stage direction hay [music]

BẮT BUỘC TRẢ VỀ DUY NHẤT ĐỊNH DẠNG JSON ARRAY SAU (không kèm text nào khác ngoài array JSON này):
[
  {
    "text": "Đoạn lời thoại hook tạo sự tò mò (khoảng 30-50 từ)...",
    "pexels_search_query": "shocked person looking at phone"
  },
  {
    "text": "Đoạn lời thoại cho phần tiếp theo (khoảng 30-50 từ)...",
    "pexels_search_query": "time lapse futuristic city"
  }
]
Lưu ý: 
1. "pexels_search_query" phải là cụm từ khóa tiếng Anh RẤT NGẮN, miêu tả một HÀNH ĐỘNG/SỰ VẬT CỤ THỂ để tôi dùng tìm kiếm trên Pexels video (ví dụ: "sad man walking rain", "gold coins falling").
2. Mỗi "text" nên có độ dài khoảng 30-80 từ để video chuyển cảnh hợp lý, tổng script khoảng 800-1500 từ.`, topic)

	return gs.callGemini(prompt, 0.7, 4096)
}

// GenerateTikTokScript generates a short, viral TikTok script from a topic
func (gs *GeminiService) GenerateTikTokScript(topic string) ([]models.VideoSegment, error) {
	prompt := fmt.Sprintf(`Bạn là chuyên gia tạo content TikTok viral bằng tiếng Việt. Viết kịch bản TikTok về: "%s"

CẤU TRÚC:
HOOK (0-3s): 1 câu mạnh, tạo tò mò ngay lập tức
SETUP (3-15s): Đặt câu hỏi chưa trả lời
CONTENT (15-45s): Giá trị cốt lõi, ngắn gọn  
PAYOFF + CTA (45-60s): Kết thúc thỏa mãn + kêu gọi follow

BẮT BUỘC TRẢ VỀ DUY NHẤT ĐỊNH DẠNG JSON ARRAY SAU (không kèm text nào khác ngoài JSON):
[
  {
    "text": "Hook cực mạnh ở 1-2 câu đầu...",
    "pexels_search_query": "surprised face close up"
  },
  {
    "text": "Phần setup...",
    "pexels_search_query": "person thinking hard"
  }
]
Lưu ý: 
1. "pexels_search_query" phải là cụm từ khóa tiếng Anh cực hay, RẤT NGẮN để tìm B-roll trên Pexels (VD: "working hard fast", "dollar bills flying"). Tuyệt đối không để nguyên tiếng Việt.
2. Mỗi "text" dài khoảng 1-3 câu dài (20-40 từ) để nhịp video nhanh. Tổng độ dài script dưới 250 từ.`, topic)

	return gs.callGemini(prompt, 0.8, 2048)
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
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent?key=%s", apiKey)

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
