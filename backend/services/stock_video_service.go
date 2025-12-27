package services

import (
	"aituber/utils"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// StockVideoService handles stock video searching and downloading
type StockVideoService struct {
	apiKey     string
	httpClient *http.Client
	tempDir    string
}

// NewStockVideoService creates a new stock video service
func NewStockVideoService(apiKey, tempDir string) *StockVideoService {
	return &StockVideoService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		tempDir: tempDir,
	}
}

// PexelsVideoResponse represents Pexels API response
type PexelsVideoResponse struct {
	Videos []struct {
		ID         int `json:"id"`
		Width      int `json:"width"`
		Height     int `json:"height"`
		Duration   int `json:"duration"`
		VideoFiles []struct {
			ID       int    `json:"id"`
			Quality  string `json:"quality"` // hd, sd, uhd
			FileType string `json:"file_type"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Link     string `json:"link"`
		} `json:"video_files"`
	} `json:"videos"`
}

// PrepareStockVideo searches, downloads multiple short videos, and merges them to match duration
func (sv *StockVideoService) PrepareStockVideo(keywords string, targetDuration float64, jobID string) (string, error) {
	// 1. Search for multiple short videos (5-10s)
	videoURLs, err := sv.searchMultipleVideos(keywords, targetDuration)
	if err != nil {
		return "", fmt.Errorf("failed to search videos: %w", err)
	}

	fmt.Printf("[Stock Video] Found %d short videos for keywords: %s\n", len(videoURLs), keywords)

	// 2. Download all videos
	var videoPaths []string
	for i, videoURL := range videoURLs {
		fmt.Printf("[Stock Video] Downloading video %d/%d...\n", i+1, len(videoURLs))
		videoPath := filepath.Join(sv.tempDir, jobID, "stock", fmt.Sprintf("segment_%d.mp4", i))
		if err := sv.downloadVideo(videoURL, videoPath); err != nil {
			return "", fmt.Errorf("failed to download video %d: %w", i, err)
		}
		videoPaths = append(videoPaths, videoPath)
	}

	// 3. Merge videos with transitions
	fmt.Printf("[Stock Video] Merging %d videos with transitions...\n", len(videoPaths))
	finalVideoPath := filepath.Join(sv.tempDir, jobID, "stock", "final_stock.mp4")
	if err := sv.mergeVideosWithTransition(videoPaths, finalVideoPath, targetDuration); err != nil {
		return "", fmt.Errorf("failed to merge videos: %w", err)
	}

	return finalVideoPath, nil
}

// searchMultipleVideos searches Pexels for multiple short videos (5-10s) matching keywords
func (sv *StockVideoService) searchMultipleVideos(keywords string, targetDuration float64) ([]string, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", "100") // Get more results to filter
	params.Add("orientation", "landscape")
	params.Add("size", "medium")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", sv.apiKey)

	resp, err := sv.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pexels API returned status %d", resp.StatusCode)
	}

	var result PexelsVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Videos) == 0 {
		return nil, fmt.Errorf("no videos found for keywords: %s", keywords)
	}

	// Filter videos by duration (5-10 seconds preferred)
	var shortVideos []struct {
		Duration int
		Link     string
	}

	for _, video := range result.Videos {
		// Only accept videos between 5-15 seconds (flexible range)
		if video.Duration >= 5 && video.Duration <= 15 {
			// Find best quality HD link
			var bestLink string
			for _, file := range video.VideoFiles {
				if file.Quality == "hd" && file.Width == 1920 {
					bestLink = file.Link
					break
				}
			}
			// Fallback to first available
			if bestLink == "" && len(video.VideoFiles) > 0 {
				bestLink = video.VideoFiles[0].Link
			}

			if bestLink != "" {
				shortVideos = append(shortVideos, struct {
					Duration int
					Link     string
				}{video.Duration, bestLink})
			}
		}
	}

	if len(shortVideos) == 0 {
		return nil, fmt.Errorf("no short videos (5-15s) found for keywords: %s", keywords)
	}

	// Calculate how many videos we need to cover target duration
	var selectedURLs []string
	var totalDuration float64

	// Pick videos in order (most relevant first, not random)
	for _, video := range shortVideos {
		selectedURLs = append(selectedURLs, video.Link)
		totalDuration += float64(video.Duration)

		// Stop when we have enough duration (+ buffer)
		if totalDuration >= targetDuration {
			break
		}

		// Limit to max 10 videos to avoid too many downloads
		if len(selectedURLs) >= 10 {
			break
		}
	}

	if len(selectedURLs) == 0 {
		return nil, fmt.Errorf("failed to select videos")
	}

	return selectedURLs, nil
}

// downloadVideo downloads file from URL
func (sv *StockVideoService) downloadVideo(url, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	resp, err := sv.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// loopVideoToDuration loops video until it exceeds target duration, then trims
func (sv *StockVideoService) loopVideoToDuration(inputPath, outputPath string, targetDuration float64) error {
	// Get input duration
	duration, err := utils.GetVideoDuration(inputPath)
	if err != nil {
		return err
	}

	// Calculate how many loops needed
	loops := int(targetDuration/duration) + 1

	// Create loop list file
	listPath := filepath.Join(filepath.Dir(outputPath), "loop_list.txt")
	file, err := os.Create(listPath)
	if err != nil {
		return err
	}

	// Convert to absolute path to avoid path duplication issues
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	for i := 0; i < loops; i++ {
		// Use forward slashes for FFmpeg compatibility
		ffmpegPath := filepath.ToSlash(absInputPath)
		file.WriteString(fmt.Sprintf("file '%s'\n", ffmpegPath))
	}
	file.Close()

	// Concatenate (loop)
	loopedPath := filepath.Join(filepath.Dir(outputPath), "looped_temp.mp4")
	err = utils.RunFFmpegCommand([]string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		"-y", loopedPath,
	})
	if err != nil {
		return fmt.Errorf("concat failed: %w", err)
	}

	// Trim to exact duration
	return utils.TrimVideo(loopedPath, outputPath, targetDuration)
}

// mergeVideosWithTransition merges multiple videos with transitions and trims to target duration
func (sv *StockVideoService) mergeVideosWithTransition(inputPaths []string, outputPath string, targetDuration float64) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("no input videos to merge")
	}

	// If only one video, loop it to match duration
	if len(inputPaths) == 1 {
		return sv.loopVideoToDuration(inputPaths[0], outputPath, targetDuration)
	}

	// Calculate total duration of downloaded videos
	var totalDuration float64
	for _, path := range inputPaths {
		duration, err := utils.GetVideoDuration(path)
		if err != nil {
			return fmt.Errorf("failed to get duration of %s: %w", path, err)
		}
		totalDuration += duration
	}

	// If total duration is less than target, loop videos to fill the gap
	finalInputPaths := inputPaths
	if totalDuration < targetDuration {
		fmt.Printf("[Stock Video] Total duration (%.1fs) < target (%.1fs), looping videos...\n", totalDuration, targetDuration)

		// Seed random for variety
		rand.Seed(time.Now().UnixNano())

		// Keep adding random videos until we have enough duration
		currentDuration := totalDuration
		for currentDuration < targetDuration*1.2 { // Add 20% buffer
			// Pick a truly random video from the downloaded ones
			randomIdx := rand.Intn(len(inputPaths))
			finalInputPaths = append(finalInputPaths, inputPaths[randomIdx])

			duration, _ := utils.GetVideoDuration(inputPaths[randomIdx])
			currentDuration += duration

			// Limit to avoid infinite loops
			if len(finalInputPaths) > 50 {
				break
			}
		}
		fmt.Printf("[Stock Video] Extended to %d video segments (total ~%.1fs)\n", len(finalInputPaths), currentDuration)
	}

	// Use FFmpeg's MergeVideosWithTransition utility
	// This merges with fade transitions
	mergedPath := filepath.Join(filepath.Dir(outputPath), "merged_temp.mp4")

	err := utils.MergeVideosWithTransition(
		finalInputPaths,
		mergedPath,
		1.0,         // 1 second transition
		30,          // 30 fps
		"1920x1080", // Resolution
	)
	if err != nil {
		return fmt.Errorf("failed to merge videos: %w", err)
	}

	// Trim to exact target duration
	return utils.TrimVideo(mergedPath, outputPath, targetDuration)
}
