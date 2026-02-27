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
	"sync"
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
			Timeout: 10 * time.Minute,
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

	// 2. Download all videos in parallel
	var videoPaths []string
	var mutex sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrency to 5

	fmt.Printf("[Stock Video] Downloading %d videos in parallel...\n", len(videoURLs))

	for i, videoURL := range videoURLs {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			videoPath := filepath.Join(sv.tempDir, jobID, "stock", fmt.Sprintf("segment_%d.mp4", index))
			fmt.Printf("[Stock Video] Downloading video %d/%d...\n", index+1, len(videoURLs))

			if err := sv.downloadVideo(url, videoPath); err != nil {
				fmt.Printf("[Stock Video] Failed to download video %d: %v (Skipping)\n", index, err)
				return
			}

			mutex.Lock()
			videoPaths = append(videoPaths, videoPath)
			mutex.Unlock()
		}(i, videoURL)
	}

	wg.Wait()

	if len(videoPaths) == 0 {
		return "", fmt.Errorf("failed to download any videos")
	}

	// 3. Merge videos with transitions
	fmt.Printf("[Stock Video] Merging %d videos with transitions...\n", len(videoPaths))
	finalVideoPath := filepath.Join(sv.tempDir, jobID, "stock", "final_stock.mp4")
	if err := sv.mergeVideosWithTransition(videoPaths, finalVideoPath, targetDuration); err != nil {
		return "", fmt.Errorf("failed to merge videos: %w", err)
	}

	return finalVideoPath, nil
}

// PrepareSegmentVideo fetches stock video for a SINGLE audio segment (by index).
// It searches Pexels with `keywords`, downloads videos one-by-one until their
// combined duration exceeds `audioDuration`, concatenates them, and trims to
// exactly `audioDuration`. Returns the path to the ready segment video.
func (sv *StockVideoService) PrepareSegmentVideo(keywords string, audioDuration float64, jobID string, segIndex int) (string, error) {
	segDir := filepath.Join(sv.tempDir, jobID, "stock", fmt.Sprintf("seg_%03d", segIndex))
	if err := os.MkdirAll(segDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create segment dir: %w", err)
	}

	fmt.Printf("[SegVideo %d] Searching Pexels for: %q (need %.2fs)\n", segIndex, keywords, audioDuration)

	// 1. Search Pexels – fetch up to 15 candidates per query
	videoInfos, err := sv.searchVideoInfos(keywords, 15)
	if err != nil || len(videoInfos) == 0 {
		// Fallback: try generic "abstract" when keyword-specific search fails
		fmt.Printf("[SegVideo %d] Primary search failed (%v), trying fallback 'abstract'\n", segIndex, err)
		videoInfos, err = sv.searchVideoInfos("abstract nature", 15)
		if err != nil {
			return "", fmt.Errorf("pexels search failed even with fallback: %w", err)
		}
	}

	// 2. Greedily download videos until we have enough duration
	var downloadedPaths []string
	var totalDuration float64
	downloadIdx := 0

	for totalDuration < audioDuration+0.5 && downloadIdx < len(videoInfos) {
		info := videoInfos[downloadIdx]
		downloadIdx++

		dlPath := filepath.Join(segDir, fmt.Sprintf("raw_%02d.mp4", downloadIdx))
		fmt.Printf("[SegVideo %d] Downloading video %d (%.0fs clip)...\n", segIndex, downloadIdx, float64(info.Duration))

		if err := sv.downloadVideo(info.Link, dlPath); err != nil {
			fmt.Printf("[SegVideo %d] Download failed, skipping: %v\n", segIndex, err)
			continue
		}
		downloadedPaths = append(downloadedPaths, dlPath)
		totalDuration += float64(info.Duration)
	}

	if len(downloadedPaths) == 0 {
		return "", fmt.Errorf("segment %d: failed to download any video for keywords %q", segIndex, keywords)
	}

	// 3. If only one clip and it's long enough, just trim it directly
	var concatPath string
	if len(downloadedPaths) == 1 {
		concatPath = downloadedPaths[0]
	} else {
		// 3a. Build concat list and join without re-encoding (fast)
		listPath := filepath.Join(segDir, "concat_list.txt")
		f, err := os.Create(listPath)
		if err != nil {
			return "", err
		}
		for _, p := range downloadedPaths {
			absP, _ := filepath.Abs(p)
			f.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(absP)))
		}
		f.Close()

		concatPath = filepath.Join(segDir, "concat.mp4")
		if err := utils.RunFFmpegCommand([]string{
			"-f", "concat",
			"-safe", "0",
			"-i", listPath,
			"-c", "copy",
			"-y", concatPath,
		}); err != nil {
			return "", fmt.Errorf("segment %d concat failed: %w", segIndex, err)
		}
	}

	// 4. Normalize + trim to exact audioDuration
	trimmedPath := filepath.Join(segDir, "segment.mp4")
	if err := utils.RunFFmpegCommand([]string{
		"-i", concatPath,
		"-t", fmt.Sprintf("%.3f", audioDuration),
		"-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30,format=yuv420p",
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-an", // no audio track – audio comes from TTS
		"-y", trimmedPath,
	}); err != nil {
		return "", fmt.Errorf("segment %d normalize+trim failed: %w", segIndex, err)
	}

	fmt.Printf("[SegVideo %d] Ready: %s (%.2fs)\n", segIndex, trimmedPath, audioDuration)
	return trimmedPath, nil
}

// videoInfo holds just the URL + duration of a Pexels video file match
type videoInfo struct {
	Link     string
	Duration int
}

// searchVideoInfos searches Pexels and returns ordered list of (link, duration) for the best-quality files.
func (sv *StockVideoService) searchVideoInfos(keywords string, perPage int) ([]videoInfo, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", fmt.Sprintf("%d", perPage))
	params.Add("orientation", "landscape")

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

	var infos []videoInfo
	for _, video := range result.Videos {
		if video.Duration < 3 || video.Duration > 60 {
			continue
		}
		bestLink, bestScore := "", 0
		for _, file := range video.VideoFiles {
			score := 0
			ar := 0.0
			if file.Height > 0 {
				ar = float64(file.Width) / float64(file.Height)
			}
			is169 := ar > 1.77 && ar < 1.79
			if file.Width == 1920 && file.Height == 1080 {
				score = 10000
			} else if is169 && file.Width >= 1280 {
				score = 5000
			} else if is169 {
				score = 1000
			} else if file.Quality == "hd" {
				score = 500
			} else {
				score = 1
			}
			score += file.Width
			if score > bestScore {
				bestScore = score
				bestLink = file.Link
			}
		}
		if bestLink != "" {
			infos = append(infos, videoInfo{Link: bestLink, Duration: video.Duration})
		}
	}
	return infos, nil
}

// searchMultipleVideos searches Pexels for multiple short videos (5-10s) matching keywords
func (sv *StockVideoService) searchMultipleVideos(keywords string, targetDuration float64) ([]string, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", "100") // Get more results to filter
	params.Add("orientation", "landscape")

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
		if video.Duration >= 5 && video.Duration <= 35 {
			// Find best quality link (Prioritize 1080p > 16:9 > HD)
			var bestLink string
			var bestScore int

			for _, file := range video.VideoFiles {
				currentScore := 0

				// Calculate aspect ratio
				var aspectRatio float64
				if file.Height > 0 {
					aspectRatio = float64(file.Width) / float64(file.Height)
				}

				// Check for 16:9 (approx 1.77)
				is16_9 := aspectRatio > 1.77 && aspectRatio < 1.78

				if file.Width == 1920 && file.Height == 1080 {
					currentScore = 10000 // Perfect 1080p match
				} else if is16_9 && file.Width >= 1280 {
					currentScore = 5000 // 720p+ 16:9
				} else if is16_9 {
					currentScore = 1000 // Any 16:9
				} else if file.Quality == "hd" {
					currentScore = 500 // Non-16:9 HD
				} else {
					currentScore = 1 // Fallback
				}

				// Add width to score to prefer higher resolution among same category
				currentScore += file.Width

				if currentScore > bestScore {
					bestScore = currentScore
					bestLink = file.Link
				}
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

		// Limit to max 100 videos to avoid too many downloads
		if len(selectedURLs) >= 100 {
			break
		}
	}

	if len(selectedURLs) == 0 {
		return nil, fmt.Errorf("failed to select videos")
	}

	return selectedURLs, nil
}

// downloadVideo downloads file from URL with retry
func (sv *StockVideoService) downloadVideo(url, path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("[Stock Video] Retrying download (attempt %d/%d)...\n", attempt+1, maxRetries)
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		// Create request with context is better, but client timeout handles it globally
		resp, err := sv.httpClient.Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		file, err := os.Create(path)
		if err != nil {
			resp.Body.Close()
			return err
		}

		_, err = io.Copy(file, resp.Body)
		resp.Body.Close()
		file.Close()

		if err != nil {
			lastErr = err
			continue
		}

		return nil // Success
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
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

	// If effective duration (considering transitions) is less than target, loop videos to fill the gap
	finalInputPaths := inputPaths
	const transitionDuration = 1.0 // Matches the hardcoded value below

	// Effective duration = TotalRawDuration - (Count-1)*TransitionDuration
	currentRawDuration := totalDuration
	currentCount := len(finalInputPaths)
	currentEffective := currentRawDuration - float64(currentCount-1)*transitionDuration

	// Add 5 seconds buffer to target duration to ensure video is always longer than audio
	// efficiently handling potential FFmpeg timing mismatches or decoding delays
	safeTargetDuration := targetDuration + 5.0

	if currentEffective < safeTargetDuration {
		fmt.Printf("[Stock Video] Effective duration (%.1fs) < target (%.1fs), looping videos...\n", currentEffective, safeTargetDuration)

		// Seed random for variety
		rand.Seed(time.Now().UnixNano())

		// Keep adding random videos until we have enough duration
		for currentEffective < safeTargetDuration {
			// Pick a truly random video from the downloaded ones
			randomIdx := rand.Intn(len(inputPaths))
			finalInputPaths = append(finalInputPaths, inputPaths[randomIdx])

			duration, _ := utils.GetVideoDuration(inputPaths[randomIdx])
			currentRawDuration += duration
			currentCount++

			// Recalculate effective duration
			currentEffective = currentRawDuration - float64(currentCount-1)*transitionDuration

			// Limit to avoid infinite loops
			if len(finalInputPaths) > 100 {
				break
			}
		}
		fmt.Printf("[Stock Video] Extended to %d video segments (effective ~%.1fs)\n", len(finalInputPaths), currentEffective)
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

	// Trim to target duration + 2s buffer (let -shortest in next step handle the exact cut)
	// This prevents video being slightly shorter than audio due to frame boundaries
	return utils.TrimVideo(mergedPath, outputPath, targetDuration+2.0)
}
