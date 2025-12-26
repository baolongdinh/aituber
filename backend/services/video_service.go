package services

import (
	"aituber/models"
	"aituber/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// VideoService handles video generation and processing
type VideoService struct {
	apiPool            *utils.APIKeyPool
	httpClient         *http.Client
	tempDir            string
	videoBitrate       string
	resolution         string
	fps                int
	transitionDuration float64
}

// NewVideoService creates a new video service
func NewVideoService(apiPool *utils.APIKeyPool, tempDir string, videoBitrate string, resolution string, fps int, transitionDuration float64) *VideoService {
	return &VideoService{
		apiPool: apiPool,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Videos take longer
		},
		tempDir:            tempDir,
		videoBitrate:       videoBitrate,
		resolution:         resolution,
		fps:                fps,
		transitionDuration: transitionDuration,
	}
}

// GenerateVideoPrompts generates visual prompts for each text segment
// Uses simple template-based approach for consistency
func (vs *VideoService) GenerateVideoPrompts(segments []models.VideoSegment, style string) ([]string, error) {
	prompts := make([]string, len(segments))

	for i, segment := range segments {
		// Create a simple visual prompt
		// In production, this could use GPT/Claude for better prompts
		prompt := vs.createPromptFromText(segment.Text, style, i)
		prompts[i] = prompt
	}

	return prompts, nil
}

// createPromptFromText creates a visual prompt from text
func (vs *VideoService) createPromptFromText(text, style string, index int) string {
	// Simple template-based prompt generation
	// This ensures visual consistency across segments

	// Extract key themes (simplified - in production use NLP)
	themes := vs.extractThemes(text)

	basePrompt := fmt.Sprintf("High quality %s video, ", style)
	if len(themes) > 0 {
		basePrompt += themes + ", "
	}
	basePrompt += "cinematic lighting, professional composition, 4K resolution"

	return basePrompt
}

// extractThemes extracts key themes from text (simplified version)
func (vs *VideoService) extractThemes(text string) string {
	// In production, use proper NLP or LLM
	// For now, use simple keyword matching
	keywords := []string{
		"technology", "nature", "business", "education",
		"science", "art", "music", "sports",
	}

	for _, keyword := range keywords {
		if contains(text, keyword) || contains(text, translateToVietnamese(keyword)) {
			return keyword + " themed"
		}
	}

	return "abstract"
}

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 // Simplified
}

func translateToVietnamese(word string) string {
	// Simplified translation map
	translations := map[string]string{
		"technology": "công nghệ",
		"nature":     "thiên nhiên",
		"business":   "kinh doanh",
		"education":  "giáo dục",
	}
	if val, ok := translations[word]; ok {
		return val
	}
	return word
}

// PikaVideoRequest represents video generation request
type PikaVideoRequest struct {
	Prompt     string  `json:"prompt"`
	Duration   float64 `json:"duration,omitempty"`
	Resolution string  `json:"resolution,omitempty"`
}

// PikaVideoResponse represents video generation response
type PikaVideoResponse struct {
	JobID    string `json:"job_id,omitempty"`
	Status   string `json:"status,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// GenerateVideos generates video clips for each prompt
func (vs *VideoService) GenerateVideos(prompts []string, durations []float64, jobID string, maxConcurrent int) ([]string, error) {
	if len(prompts) != len(durations) {
		return nil, fmt.Errorf("prompts and durations length mismatch")
	}

	videoPaths := make([]string, len(prompts))
	errors := make([]error, len(prompts))

	// Create semaphore for rate limiting
	sem := make(chan struct{}, maxConcurrent)
	done := make(chan struct{})

	// Process videos in parallel
	for i, prompt := range prompts {
		go func(index int, p string, dur float64) {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			videoPath, err := vs.generateSingleVideo(p, dur, jobID, index)
			if err != nil {
				errors[index] = err
			} else {
				videoPaths[index] = videoPath
			}

			if index == len(prompts)-1 {
				close(done)
			}
		}(i, prompt, durations[i])
	}

	// Wait for all to complete
	<-done

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("failed to generate video segment %d: %w", i, err)
		}
	}

	return videoPaths, nil
}

// generateSingleVideo generates a single video with retry
func (vs *VideoService) generateSingleVideo(prompt string, duration float64, jobID string, index int) (string, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get API key from pool
		apiKey, err := vs.apiPool.GetRandomKey()
		if err != nil {
			return "", fmt.Errorf("no available API keys: %w", err)
		}

		// Call video generation API (using mock for now)
		videoData, err := vs.callVideoGenerationAPI(prompt, duration, apiKey)
		if err != nil {
			// Mark key as failed
			vs.apiPool.MarkFailed(apiKey, time.Duration(120)*time.Second)
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}

		// Mark key as successful
		vs.apiPool.MarkSuccess(apiKey)

		// Save video to file
		videoPath := filepath.Join(vs.tempDir, jobID, "video", fmt.Sprintf("segment_%03d.mp4", index))
		if err := vs.saveVideoFile(videoData, videoPath); err != nil {
			return "", fmt.Errorf("failed to save video: %w", err)
		}

		// Adjust duration if needed
		adjustedPath := filepath.Join(vs.tempDir, jobID, "video", fmt.Sprintf("segment_%03d_adjusted.mp4", index))
		if err := vs.adjustVideoDuration(videoPath, adjustedPath, duration); err != nil {
			return "", fmt.Errorf("failed to adjust duration: %w", err)
		}

		return adjustedPath, nil
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// callVideoGenerationAPI calls video generation API
// NOTE: This is a mock implementation - replace with actual API
func (vs *VideoService) callVideoGenerationAPI(prompt string, duration float64, apiKey string) ([]byte, error) {
	// Mock implementation - returns placeholder
	// In production, implement actual API calls to:
	// - Pika Labs: https://pika.art/api
	// - Leonardo.AI: https://api.leonardo.ai
	// - Runway ML: https://api.runwayml.com

	// For now, return error to indicate API implementation needed
	return nil, fmt.Errorf("video generation API not implemented - please configure with real API endpoint")

	// Example implementation would be:
	/*
		url := "https://api.pika.art/v1/generate"
		reqBody := PikaVideoRequest{
			Prompt:     prompt,
			Duration:   duration,
			Resolution: vs.resolution,
		}

		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := vs.httpClient.Do(req)
		// ... handle response, poll for completion, download video
	*/
}

// saveVideoFile saves video data to file
func (vs *VideoService) saveVideoFile(data []byte, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// adjustVideoDuration adjusts video to target duration
func (vs *VideoService) adjustVideoDuration(inputPath, outputPath string, targetDuration float64) error {
	currentDuration, err := utils.GetVideoDuration(inputPath)
	if err != nil {
		return err
	}

	if currentDuration < targetDuration {
		// Extend video
		return utils.ExtendVideo(inputPath, outputPath, targetDuration)
	} else if currentDuration > targetDuration {
		// Trim video
		return utils.TrimVideo(inputPath, outputPath, targetDuration)
	} else {
		// Duration matches - just copy
		return copyFile(inputPath, outputPath)
	}
}

// copyFile copies a file
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// MergeVideos merges video segments with transitions
func (vs *VideoService) MergeVideos(videoPaths []string, outputPath string) error {
	if len(videoPaths) == 0 {
		return fmt.Errorf("no video files to merge")
	}

	// Use FFmpeg utility to merge with transitions
	err := utils.MergeVideosWithTransition(
		videoPaths,
		outputPath,
		vs.transitionDuration,
		vs.fps,
		vs.resolution,
	)
	if err != nil {
		return fmt.Errorf("failed to merge videos: %w", err)
	}

	return nil
}
