package services

import (
	"aituber/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AudioService handles text-to-speech and audio processing
type AudioService struct {
	apiPool           *utils.APIKeyPool
	httpClient        *http.Client
	tempDir           string
	audioBitrate      string
	sampleRate        int
	crossfadeDuration float64
	rateLimiter       <-chan time.Time
}

// NewAudioService creates a new audio service
func NewAudioService(apiPool *utils.APIKeyPool, tempDir string, audioBitrate string, sampleRate int, crossfadeDuration float64) *AudioService {
	// Create rate limiter (1 request every 500ms = 2 RPS)
	// This prevents hitting FPT.AI rate limits
	limiter := time.Tick(500 * time.Millisecond)

	return &AudioService{
		apiPool: apiPool,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		tempDir:           tempDir,
		audioBitrate:      audioBitrate,
		sampleRate:        sampleRate,
		crossfadeDuration: crossfadeDuration,
		rateLimiter:       limiter,
	}
}

// FPTTTSResponse represents FPT.AI TTS API response
type FPTTTSResponse struct {
	Async     string `json:"async,omitempty"`
	Error     int    `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// GenerateAudioChunks generates audio for each text chunk
// Uses parallel processing with rate limiting
func (as *AudioService) GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error) {
	audioPaths := make([]string, len(chunks))
	errors := make([]error, len(chunks))

	log.Printf("[AudioService] Starting audio generation for %d chunks (Concurrency: %d)", len(chunks), maxConcurrent)

	// Create semaphore for rate limiting
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// Process chunks in parallel
	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, text string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			audioPath, err := as.generateSingleAudio(text, voice, speed, jobID, index)
			if err != nil {
				errors[index] = err
			} else {
				audioPaths[index] = audioPath
			}
		}(i, chunk)
	}

	// Wait for all to complete
	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio chunk %d: %w", i, err)
		}
	}

	return audioPaths, nil
}

// generateSingleAudio generates audio for a single text chunk with retry
func (as *AudioService) generateSingleAudio(text, voice string, speed float64, jobID string, index int) (string, error) {
	maxRetries := 3
	var lastErr error

	log.Printf("[Chunk %d] Calling TTS - TEXT: %s ", index, text)
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get API key from pool
		apiKey, err := as.apiPool.GetRandomKey()
		if err != nil {
			return "", fmt.Errorf("no available API keys: %w", err)
		}

		// Call TTS API - this returns async URL or direct audio
		log.Printf("[Chunk %d] Calling TTS API (attempt %d/%d)", index, attempt+1, maxRetries)
		asyncURL, apiErr := as.callFPTTTSAsync(text, voice, speed, apiKey)
		if apiErr != nil {
			// API call failed - blacklist the key
			log.Printf("[Chunk %d] API call failed: %v", index, apiErr)
			as.apiPool.MarkFailed(apiKey, time.Duration(60)*time.Second)
			lastErr = apiErr
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// API call succeeded - mark key as successful
		log.Printf("[Chunk %d] API call successful, async URL: %s", index, asyncURL)
		as.apiPool.MarkSuccess(apiKey)

		// Now download the audio with retry (file may not be ready yet)
		log.Printf("[Chunk %d] Starting download with retry...", index)
		audioData, downloadErr := as.downloadAudioWithRetry(asyncURL, index)
		if downloadErr != nil {
			// Download failed even after retries
			// We will retry the entire process (get new key -> call API -> download)
			log.Printf("[Chunk %d] Download failed after all retries: %v. Retrying API call (Attempt %d/%d)...", index, downloadErr, attempt+1, maxRetries)
			lastErr = downloadErr
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[Chunk %d] Download successful, size: %d bytes", index, len(audioData))

		// Save audio to file
		audioPath := filepath.Join(as.tempDir, jobID, "audio", fmt.Sprintf("chunk_%03d.mp3", index))
		if err := as.saveAudioFile(audioData, audioPath); err != nil {
			return "", fmt.Errorf("failed to save audio: %w", err)
		}

		return audioPath, nil
	}

	return "", fmt.Errorf("failed after %d retries. Last error: %v", maxRetries, lastErr)
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

	// Wait a bit before returning to give FPT time to register the job
	time.Sleep(2 * time.Second)

	return apiResp.Async, nil
}

// downloadAudioWithRetry downloads audio with retry logic
// FPT.AI files need 5s-2min processing time, so we retry until successful
func (as *AudioService) downloadAudioWithRetry(url string, chunkIndex int) ([]byte, error) {
	maxRetries := 10                 // 100 retries
	retryInterval := 5 * time.Second // 5 seconds between retries
	// Total time: 25 * 5s = 125s = 2 minutes

	log.Printf("[Chunk %d] Starting download retry loop (max %d retries, %v interval)", chunkIndex, maxRetries, retryInterval)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (except first attempt)
			if attempt%10 == 0 {
				// Log every 10th retry to avoid spam
				log.Printf("[Chunk %d] Retry attempt %d/%d...", chunkIndex, attempt, maxRetries)
			}
			time.Sleep(retryInterval)
		}

		data, err := as.downloadAudio(url)
		if err == nil {
			// Success!
			log.Printf("[Chunk %d] Download successful on attempt %d", chunkIndex, attempt+1)
			return data, nil
		}

		// Failed, record error and retry
		lastErr = err
		if attempt == 0 {
			// Log first failure (file likely not ready yet)
			log.Printf("[Chunk %d] First download attempt failed (expected - file processing): %v", chunkIndex, err)
		}
	}

	log.Printf("[Chunk %d] All %d retry attempts exhausted", chunkIndex, maxRetries)
	return nil, fmt.Errorf("failed to download after %d retries (8 minutes): %w", maxRetries, lastErr)
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
