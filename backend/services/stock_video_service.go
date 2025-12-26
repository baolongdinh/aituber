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

// PrepareStockVideo searches, downloads, and loops video to match duration
func (sv *StockVideoService) PrepareStockVideo(keywords string, targetDuration float64, jobID string) (string, error) {
	// 1. Search for videos
	videoURL, err := sv.searchVideo(keywords)
	if err != nil {
		return "", fmt.Errorf("failed to search video: %w", err)
	}

	// 2. Download video
	rawVideoPath := filepath.Join(sv.tempDir, jobID, "stock", "raw_stock.mp4")
	if err := sv.downloadVideo(videoURL, rawVideoPath); err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	// 3. Loop video to match duration
	finalVideoPath := filepath.Join(sv.tempDir, jobID, "stock", "final_stock.mp4")
	if err := sv.loopVideoToDuration(rawVideoPath, finalVideoPath, targetDuration); err != nil {
		return "", fmt.Errorf("failed to loop video: %w", err)
	}

	return finalVideoPath, nil
}

// searchVideo searches Pexels for a video matching keywords
func (sv *StockVideoService) searchVideo(keywords string) (string, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", "10")
	params.Add("orientation", "landscape")
	params.Add("size", "medium") // Prefer HD

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", sv.apiKey)

	resp, err := sv.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pexels API returned status %d", resp.StatusCode)
	}

	var result PexelsVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Videos) == 0 {
		return "", fmt.Errorf("no videos found for keywords: %s", keywords)
	}

	// Pick a random video from results to vary content
	rand.Seed(time.Now().UnixNano())
	video := result.Videos[rand.Intn(len(result.Videos))]

	// Find best quality (HD 1920x1080 preferred)
	var bestLink string
	for _, file := range video.VideoFiles {
		if file.Quality == "hd" && file.Width == 1920 {
			bestLink = file.Link
			break
		}
	}

	// Fallback if no 1080p found
	if bestLink == "" && len(video.VideoFiles) > 0 {
		bestLink = video.VideoFiles[0].Link
	}

	if bestLink == "" {
		return "", fmt.Errorf("no valid video files found")
	}

	return bestLink, nil
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

	for i := 0; i < loops; i++ {
		file.WriteString(fmt.Sprintf("file '%s'\n", inputPath))
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
