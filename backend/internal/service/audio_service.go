package service

import (
	"aituber/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AudioService handles text-to-speech and audio processing
type AudioService struct {
	apiPool           *utils.APIKeyPool
	elevenLabsAPIKey  string
	httpClient        *http.Client
	tempDir           string
	audioBitrate      string
	sampleRate        int
	crossfadeDuration float64
	rateLimiter       <-chan time.Time
	ttsCachePath      string // Path to TTS URL cache JSON file
}

// NewAudioService creates a new audio service
func NewAudioService(apiPool *utils.APIKeyPool, elevenLabsKey string, tempDir string, audioBitrate string, sampleRate int, crossfadeDuration float64) *AudioService {
	limiter := time.Tick(5000 * time.Millisecond)

	cachePath := filepath.Join(tempDir, "tts_url_cache.json")

	return &AudioService{
		apiPool:          apiPool,
		elevenLabsAPIKey: elevenLabsKey,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		tempDir:           tempDir,
		audioBitrate:      audioBitrate,
		sampleRate:        sampleRate,
		crossfadeDuration: crossfadeDuration,
		rateLimiter:       limiter,
		ttsCachePath:      cachePath,
	}
}

// ttsCacheKey builds the cache lookup key
func ttsCacheKey(text, voice string, speed float64) string {
	return fmt.Sprintf("%s|%s|%.1f", text, voice, speed)
}

// ttsCacheMu protects concurrent writes to the JSON cache file
var ttsCacheMu sync.Mutex

// ttsReadCache reads the JSON cache and returns the URL for the given key, or empty string
func (as *AudioService) ttsReadCache(key string) string {
	ttsCache, err := os.ReadFile(as.ttsCachePath)
	if err != nil {
		return ""
	}
	var cache map[string]string
	if json.Unmarshal(ttsCache, &cache) != nil {
		return ""
	}
	return cache[key]
}

// ttsWriteCache persists a key→url pair to the JSON cache file
func (as *AudioService) ttsWriteCache(key, url string) {
	ttsCacheMu.Lock()
	defer ttsCacheMu.Unlock()

	cache := make(map[string]string)
	if existing, err := os.ReadFile(as.ttsCachePath); err == nil {
		_ = json.Unmarshal(existing, &cache)
	}
	cache[key] = url

	if data, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(as.ttsCachePath), 0755)
		_ = os.WriteFile(as.ttsCachePath, data, 0644)
	}
}

// FPTTTSResponse represents FPT.AI TTS API response
type FPTTTSResponse struct {
	Async     string `json:"async,omitempty"`
	Error     int    `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ElevenLabsTTSWithTimestampsResponse represents ElevenLabs TTS API response with timestamps
type ElevenLabsTTSWithTimestampsResponse struct {
	Audio     []byte `json:"audio"`
	Alignment struct {
		Chars            []string `json:"chars"`
		CharStartTimesMs []int    `json:"char_start_times_ms"`
		CharEndTimesMs   []int    `json:"char_end_times_ms"`
	} `json:"alignment"`
}

// GenerateAudioChunks generates audio for each text chunk (FPT.AI flow)
func (as *AudioService) GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error) {
	audioPaths := make([]string, len(chunks))
	errors := make([]error, len(chunks))

	log.Printf("[AudioService] Starting chunked audio generation (FPT) for %d chunks", len(chunks))

	// Create semaphore
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, text string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Force FPT fallback logic by passing provider context if needed,
			// but here we just call the old robust segment flow.
			audioPath, err := as.generateSingleAudioFPT(text, voice, speed, jobID, index)
			if err == nil {
				audioPath, err = as.postProcessAudio(audioPath, jobID, index)
			}
			if err != nil {
				errors[index] = err
			} else {
				audioPaths[index] = audioPath
			}
		}(i, chunk)
	}

	wg.Wait()
	return audioPaths, nil
}

// GenerateSingleAudio generates a single audio chunk
func (as *AudioService) GenerateSingleAudio(text, voice string, speed float64, jobID string, index int) (string, error) {
	audioPath, err := as.generateSingleAudioFPT(text, voice, speed, jobID, index)
	if err != nil {
		return "", err
	}
	return as.postProcessAudio(audioPath, jobID, index)
}

// GenerateAudioFullScript generates TTS for the entire script at once (ElevenLabs flow)
// It then splits the audio into segments based on word alignments.
func (as *AudioService) GenerateAudioFullScript(segments []VideoSegment, voice string, jobID string) ([]string, error) {
	if as.elevenLabsAPIKey == "" || as.elevenLabsAPIKey == "placeholder" {
		return nil, fmt.Errorf("ElevenLabs API Key is missing")
	}

	log.Printf("[AudioService] Starting Full-Script TTS with ElevenLabs for %d segments", len(segments))

	// 1. Join all text segments
	var fullContent strings.Builder
	for i, seg := range segments {
		fullContent.WriteString(seg.Text)
		if i < len(segments)-1 {
			fullContent.WriteString(" ") // Add space between segments for more natural flow
		}
	}

	// 2. Map Voice ID
	actualVoiceID := as.mapToElevenLabsVoice(voice)

	// 3. Call ElevenLabs with timestamps
	log.Printf("[AudioService] Calling ElevenLabs with timestamps for voice: %s", actualVoiceID)
	audioData, alignment, err := as.callElevenLabsTTSWithTimestamps(fullContent.String(), actualVoiceID)
	if err != nil {
		return nil, fmt.Errorf("ElevenLabs full script failed: %w", err)
	}

	// 4. Save the master audio file
	masterPath := filepath.Join(as.tempDir, jobID, "audio", "master_full.mp3")
	if err := as.saveAudioFile(audioData, masterPath); err != nil {
		return nil, err
	}

	// 5. Calculate split points for each segment
	// We need to find the timestamp where each segment ends by matching strings.
	audioPaths := make([]string, len(segments))
	var lastEnd float64 = 0.0

	// Words and their end times from alignment
	// Simpler heuristic: ElevenLabs "with-timestamps" returns character-level alignment.
	// We count characters in each segment text to find the split timestamp.
	var charIndexOffset int = 0
	for i, seg := range segments {
		segLen := len(seg.Text)
		targetCharIndex := charIndexOffset + segLen - 1
		if targetCharIndex >= len(alignment.CharEndTimesMs) {
			targetCharIndex = len(alignment.CharEndTimesMs) - 1
		}

		endMs := alignment.CharEndTimesMs[targetCharIndex]
		endSec := float64(endMs) / 1000.0

		// Extract segment
		segmentPath := filepath.Join(as.tempDir, jobID, "audio", fmt.Sprintf("chunk_%03d.mp3", i))
		duration := endSec - lastEnd
		if duration <= 0 {
			duration = 0.1 // Minimum
		}

		err := utils.ExtractAudioSegment(masterPath, lastEnd, duration, segmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to split audio for segment %d: %w", i, err)
		}

		// Post-process (silence removal)
		pacedPath, _ := as.postProcessAudio(segmentPath, jobID, i)
		audioPaths[i] = pacedPath

		lastEnd = endSec
		charIndexOffset += segLen + 1 // +1 for the space we added
	}

	return audioPaths, nil
}

// mapToElevenLabsVoice maps FPT voices or takes long ID
func (as *AudioService) mapToElevenLabsVoice(voiceID string) string {
	const (
		elevenMaleID   = "ipTvfDXAg1zowfF1rv9w"
		elevenFemaleID = "Si3s1VCb7dLbeqH57kiC"
	)
	if len(voiceID) >= 10 {
		return voiceID
	}
	isMale := false
	maleVoices := []string{"minhquang", "giahuy", "vandoan", "manhduc"}
	for _, mv := range maleVoices {
		if voiceID == mv {
			isMale = true
			break
		}
	}
	if isMale {
		return elevenMaleID
	}
	return elevenFemaleID
}

// generateSingleAudioFPT calls FPT.AI TTS and polls for the result.
// Cache strategy: before calling FPT, check local JSON cache. If a URL exists,
// try to download it. If download succeeds, reuse it. Otherwise call FPT as normal
// and save the new URL to cache. No RAM caching — always reads/writes JSON file.
func (as *AudioService) generateSingleAudioFPT(text, voice string, speed float64, jobID string, index int) (string, error) {
	audioPath := filepath.Join(as.tempDir, jobID, "audio", fmt.Sprintf("chunk_%03d.mp3", index))
	cacheKey := ttsCacheKey(text, voice, speed)

	// --- CACHE LOOKUP ---
	if cachedURL := as.ttsReadCache(cacheKey); cachedURL != "" {
		log.Printf("[Chunk %d] TTS cache HIT for key (%.30s...). Trying URL: %s", index, cacheKey, cachedURL)
		if data, err := as.downloadAudio(cachedURL); err == nil {
			if saveErr := as.saveAudioFile(data, audioPath); saveErr == nil {
				log.Printf("[Chunk %d] TTS cache served successfully.", index)
				// No need to return postProcessAudio here – caller does it
				return audioPath, nil
			}
		} else {
			log.Printf("[Chunk %d] Cached URL no longer valid (%v). Falling back to FPT call.", index, err)
		}
	}

	// --- NORMAL FPT FLOW ---
	maxAPIRetries := 36
	var lastErr error
	var asyncURLs []string

	for attempt := 0; attempt < maxAPIRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[Chunk %d] Re-requesting FPT.AI TTS (Attempt %d/%d)", index, attempt+1, maxAPIRetries)
		}

		apiKey, err := as.apiPool.GetRandomKey()
		if err != nil {
			return "", fmt.Errorf("no available FPT API keys: %w", err)
		}

		asyncURL, apiErr := as.callFPTTTSAsync(text, voice, speed, apiKey)
		if apiErr != nil {
			log.Printf("[Chunk %d] FPT API call failed: %v", index, apiErr)
			as.apiPool.MarkFailed(apiKey, 15*time.Second)
			lastErr = apiErr
			time.Sleep(3 * time.Second)
			continue
		}
		as.apiPool.MarkSuccess(apiKey)

		// Save URL to cache immediately so future runs can try it
		as.ttsWriteCache(cacheKey, asyncURL)
		log.Printf("[Chunk %d] TTS URL cached: %s", index, asyncURL)

		asyncURLs = append(asyncURLs, asyncURL)

		audioData, downloadErr := as.pollForAudioDownloadList(asyncURLs, index)
		if downloadErr != nil {
			log.Printf("[Chunk %d] Poll exhausted for %d URLs, will re-request TTS: %v", index, len(asyncURLs), downloadErr)
			lastErr = downloadErr
			continue
		}

		if err := as.saveAudioFile(audioData, audioPath); err != nil {
			return "", err
		}
		return as.postProcessAudio(audioPath, jobID, index)
	}
	return "", fmt.Errorf("FPT failed after %d API attempts, last error: %v", maxAPIRetries, lastErr)
}

// callElevenLabsTTSWithTimestamps calls ElevenLabs API and returns audio + alignment
func (as *AudioService) callElevenLabsTTSWithTimestamps(text, voiceID string) ([]byte, ElevenLabsTTSWithTimestampsResponse_Alignment, error) {
	// The endpoint for timestamps is slightly different and requires a streaming output format
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s/stream/with-timestamps", voiceID)

	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, ElevenLabsTTSWithTimestampsResponse_Alignment{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", as.elevenLabsAPIKey)

	resp, err := as.httpClient.Do(req)
	if err != nil {
		return nil, ElevenLabsTTSWithTimestampsResponse_Alignment{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, ElevenLabsTTSWithTimestampsResponse_Alignment{}, fmt.Errorf("ElevenLabs API returned %d: %s", resp.StatusCode, string(body))
	}

	// The "with-timestamps" response is a JSON stream where each line/chunk contains audio and alignment.
	// Since we are calling the non-streaming REST wrapper as a block, we need to assemble it.
	// Actually, for REST it returns a combined JSON object.
	var fullAudio []byte
	var finalAlignment ElevenLabsTTSWithTimestampsResponse_Alignment

	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk struct {
			AudioBase64 string                                        `json:"audio_base64"`
			Alignment   ElevenLabsTTSWithTimestampsResponse_Alignment `json:"alignment"`
		}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		if chunk.AudioBase64 != "" {
			audio, _ := base64.StdEncoding.DecodeString(chunk.AudioBase64)
			fullAudio = append(fullAudio, audio...)
		}

		if len(chunk.Alignment.Chars) > 0 {
			finalAlignment.Chars = append(finalAlignment.Chars, chunk.Alignment.Chars...)
			finalAlignment.CharStartTimesMs = append(finalAlignment.CharStartTimesMs, chunk.Alignment.CharStartTimesMs...)
			finalAlignment.CharEndTimesMs = append(finalAlignment.CharEndTimesMs, chunk.Alignment.CharEndTimesMs...)
		}
	}

	return fullAudio, finalAlignment, nil
}

// ElevenLabsTTSWithTimestampsResponse_Alignment helper struct
type ElevenLabsTTSWithTimestampsResponse_Alignment struct {
	Chars            []string `json:"chars"`
	CharStartTimesMs []int    `json:"char_start_times_ms"`
	CharEndTimesMs   []int    `json:"char_end_times_ms"`
}

// callElevenLabsTTS calls ElevenLabs Text-to-Speech API (Legacy/Simple fallback)
func (as *AudioService) callElevenLabsTTS(text, voiceID string) ([]byte, error) {
	// Male: ipTvfDXAg1zowfF1rv9w
	// Female: Si3s1VCb7dLbeqH57kiC
	const (
		elevenMaleID   = "ipTvfDXAg1zowfF1rv9w"
		elevenFemaleID = "Si3s1VCb7dLbeqH57kiC"
	)

	actualVoiceID := voiceID
	// If it's a long ID, it's already an ElevenLabs ID
	if len(actualVoiceID) < 10 {
		// It's an FPT voice ID, map it to ElevenLabs by gender
		isMale := false
		maleVoices := []string{"minhquang", "giahuy", "vandoan", "manhduc"}
		for _, mv := range maleVoices {
			if actualVoiceID == mv {
				isMale = true
				break
			}
		}

		if isMale {
			actualVoiceID = elevenMaleID
		} else {
			actualVoiceID = elevenFemaleID
		}
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", actualVoiceID)

	// ElevenLabs settings for v3
	payload := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_multilingual_v2", // Multilingual v2 is super stable for VN
		"voice_settings": map[string]interface{}{
			"stability":         0.5,
			"similarity_boost":  0.75,
			"style":             0.0,
			"use_speaker_boost": true,
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", as.elevenLabsAPIKey)

	resp, err := as.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// postProcessAudio handles silence removal and path management
func (as *AudioService) postProcessAudio(audioPath, jobID string, index int) (string, error) {
	pacedPath := filepath.Join(as.tempDir, jobID, "audio", fmt.Sprintf("chunk_paced_%03d.mp3", index))
	if err := utils.RemoveAudioSilence(audioPath, pacedPath); err == nil {
		os.Remove(audioPath)
		return pacedPath, nil
	}
	log.Printf("[Chunk %d] Silence removal failed (using original)", index)
	return audioPath, nil
}

// callFPTTTSAsync calls FPT.AI TTS API and returns the async URL
func (as *AudioService) callFPTTTSAsync(text, voice string, speed float64, apiKey string) (string, error) {
	// Wait for rate limiter
	<-as.rateLimiter

	// FPT.AI TTS API endpoint
	url := "https://api.fpt.ai/hmi/tts/v5"

	// Create HTTP request with plain text body
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(text))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (voice and speed must be in headers, not JSON body)
	req.Header.Set("api-key", apiKey)
	req.Header.Set("voice", voice)
	req.Header.Set("speed", fmt.Sprintf("%.1f", speed))

	// Send request
	resp, err := as.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp FPTTTSResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return "", fmt.Errorf("API error: %s (code: %d)", errResp.Message, errResp.Error)
		}
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response to get async URL
	var apiResp FPTTTSResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w. Body: %s", err, string(body))
	}

	if apiResp.Error != 0 {
		return "", fmt.Errorf("API error: %s (code: %d)", apiResp.Message, apiResp.Error)
	}

	if apiResp.Async == "" {
		return "", fmt.Errorf("no async URL in response. Body: %s", string(body))
	}

	log.Printf("[TTS API] Received async URL: %s (request_id: %s)", apiResp.Async, apiResp.RequestID)

	// Đợi một khoảng ngắn để FPT tạo file. Thay vì 5s cứng ngắc, chờ 3s là đủ cho chunk nhỏ.
	time.Sleep(3 * time.Second)

	return apiResp.Async, nil
}

// pollForAudioDownloadList polls a list of FPT.AI generated audio URLs.
// Quy định theo ý tưởng mới: Tổng thời gian chờ tối đa khoảng 60s.
// Nó lặp qua tất cả URLs trong danh sách, nếu bất kỳ URL nào trả về data thành công thì thoát và lấy kết quả đó.
func (as *AudioService) pollForAudioDownloadList(urls []string, chunkIndex int) ([]byte, error) {
	maxAttempts := 15
	pollInterval := 4 * time.Second // 15 attempts * 4s = ~60s tổng thời gian chờ timeout
	var lastErr error

	for i := 1; i <= maxAttempts; i++ {
		var any404 bool

		for _, url := range urls {
			data, err := as.downloadAudio(url)
			if err == nil {
				log.Printf("[Chunk %d] Audio ready after %d poll attempt(s) from one of the URLs", chunkIndex, i)
				return data, nil
			}

			lastErr = err
			if strings.Contains(err.Error(), "404") {
				any404 = true
			}
		}

		if any404 {
			log.Printf("[Chunk %d] Audio not ready (404) for %d URLs, waiting 4s (attempt %d/%d, max ~60s)", chunkIndex, len(urls), i, maxAttempts)
		} else {
			log.Printf("[Chunk %d] Download error: %v, waiting 4s (attempt %d/%d, max ~60s)", chunkIndex, lastErr, i, maxAttempts)
		}

		// Giữ nguyên 4s cho mỗi lần thử để rải đều trong 60s
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("all %d URLs still 404 or err after ~60s wait (poll exhausted): %w", len(urls), lastErr)
}

// downloadAudio downloads audio from URL
func (as *AudioService) downloadAudio(url string) ([]byte, error) {
	resp, err := as.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return data, nil
}

// saveAudioFile saves audio data to file
func (as *AudioService) saveAudioFile(data []byte, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// MergeAudioFiles merges audio files with crossfade
func (as *AudioService) MergeAudioFiles(audioPaths []string, outputPath string) error {
	if len(audioPaths) == 0 {
		return fmt.Errorf("no audio files to merge")
	}

	// Use FFmpeg utility to merge with crossfade
	err := utils.MergeAudioWithCrossfade(
		audioPaths,
		outputPath,
		as.crossfadeDuration,
		as.audioBitrate,
	)
	if err != nil {
		return fmt.Errorf("failed to merge audio: %w", err)
	}

	return nil
}
