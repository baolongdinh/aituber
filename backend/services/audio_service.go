package services

import (
	"aituber/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
}

// NewAudioService creates a new audio service
func NewAudioService(apiPool *utils.APIKeyPool, tempDir string, audioBitrate string, sampleRate int, crossfadeDuration float64) *AudioService {
	return &AudioService{
		apiPool: apiPool,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		tempDir:           tempDir,
		audioBitrate:      audioBitrate,
		sampleRate:        sampleRate,
		crossfadeDuration: crossfadeDuration,
	}
}

// FPTTTSRequest represents FPT.AI TTS API request
type FPTTTSRequest struct {
	Text   string  `json:"text"`
	Voice  string  `json:"voice"`
	Speed  float64 `json:"speed"`
	Format string  `json:"format"`
}

// FPTTTSResponse represents FPT.AI TTS API response
type FPTTTSResponse struct {
	Async   string `json:"async,omitempty"`
	Error   int    `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// GenerateAudioChunks generates audio for each text chunk
// Uses parallel processing with rate limiting
func (as *AudioService) GenerateAudioChunks(chunks []string, voice string, jobID string, maxConcurrent int) ([]string, error) {
	audioPaths := make([]string, len(chunks))
	errors := make([]error, len(chunks))

	// Create semaphore for rate limiting
	sem := make(chan struct{}, maxConcurrent)
	done := make(chan struct{})

	// Process chunks in parallel
	for i, chunk := range chunks {
		go func(index int, text string) {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			audioPath, err := as.generateSingleAudio(text, voice, jobID, index)
			if err != nil {
				errors[index] = err
			} else {
				audioPaths[index] = audioPath
			}

			if index == len(chunks)-1 {
				close(done)
			}
		}(i, chunk)
	}

	// Wait for all to complete
	<-done

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("failed to generate audio chunk %d: %w", i, err)
		}
	}

	return audioPaths, nil
}

// generateSingleAudio generates audio for a single text chunk with retry
func (as *AudioService) generateSingleAudio(text, voice, jobID string, index int) (string, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get API key from pool
		apiKey, err := as.apiPool.GetRandomKey()
		if err != nil {
			return "", fmt.Errorf("no available API keys: %w", err)
		}

		// Call TTS API
		audioData, err := as.callFPTTTS(text, voice, apiKey)
		if err != nil {
			// Mark key as failed
			as.apiPool.MarkFailed(apiKey, time.Duration(60)*time.Second)
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			continue
		}

		// Mark key as successful
		as.apiPool.MarkSuccess(apiKey)

		// Save audio to file
		audioPath := filepath.Join(as.tempDir, jobID, "audio", fmt.Sprintf("chunk_%03d.mp3", index))
		if err := as.saveAudioFile(audioData, audioPath); err != nil {
			return "", fmt.Errorf("failed to save audio: %w", err)
		}

		return audioPath, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// callFPTTTS calls FPT.AI TTS API
func (as *AudioService) callFPTTTS(text, voice, apiKey string) ([]byte, error) {
	// FPT.AI TTS API endpoint
	url := "https://api.fpt.ai/hmi/tts/v5"

	// Prepare request
	reqBody := FPTTTSRequest{
		Text:   text,
		Voice:  voice,
		Speed:  1.0,
		Format: "mp3",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	// Send request
	resp, err := as.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp FPTTTSResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return nil, fmt.Errorf("API error: %s (code: %d)", errResp.Message, errResp.Error)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Check if response is JSON (async) or audio data
	var apiResp FPTTTSResponse
	if json.Unmarshal(body, &apiResp) == nil && apiResp.Async != "" {
		// Async response - download from URL
		return as.downloadAudio(apiResp.Async)
	}

	// Direct audio response
	return body, nil
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
	if _, err := utils.CreateTempDir(filepath.Dir(dir), filepath.Base(dir)); err != nil {
		return err
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
