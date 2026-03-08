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
	videoURLs, err := sv.searchMultipleVideos(keywords, targetDuration, "landscape")
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
// orientation: "landscape" (YouTube, 1920x1080) or "portrait" (TikTok, 1080x1920)
func (sv *StockVideoService) PrepareSegmentVideo(keywords string, audioDuration float64, jobID string, segIndex int, orientation string) (string, error) {
	if orientation == "" {
		orientation = "landscape"
	}

	segDir := filepath.Join(sv.tempDir, jobID, "stock", fmt.Sprintf("seg_%03d", segIndex))
	if err := os.MkdirAll(segDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create segment dir: %w", err)
	}

	fmt.Printf("[SegVideo %d] Searching Pexels for: %q (need %.2fs, orientation: %s)\n", segIndex, keywords, audioDuration, orientation)

	// 1. Search Pexels – fetch up to 15 candidates per query
	videoInfos, err := sv.searchVideoInfos(keywords, 15, orientation)
	if err != nil || len(videoInfos) == 0 {
		// Fallback: try generic keyword with same orientation
		fmt.Printf("[SegVideo %d] Primary search failed (%v), trying fallback\n", segIndex, err)
		fallbackKw := "abstract nature"
		if orientation == "portrait" {
			fallbackKw = "abstract vertical"
		}
		videoInfos, err = sv.searchVideoInfos(fallbackKw, 15, orientation)
		if err != nil || len(videoInfos) == 0 {
			// Last resort: try landscape orientation as fallback for portrait
			fmt.Printf("[SegVideo %d] Fallback also failed, trying landscape fallback\n", segIndex)
			videoInfos, err = sv.searchVideoInfos("abstract nature", 15, "landscape")
			if err != nil {
				return "", fmt.Errorf("pexels search failed even with fallback: %w", err)
			}
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

	// 4. Normalize + trim to exact audioDuration — resolution depends on platform orientation
	trimmedPath := filepath.Join(segDir, "segment.mp4")
	var vfFilter string
	if orientation == "portrait" {
		// TikTok: 1080x1920 (9:16 portrait)
		vfFilter = "scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30,format=yuv420p"
	} else {
		// YouTube: 1920x1080 (16:9 landscape)
		vfFilter = "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30,format=yuv420p"
	}

	if err := utils.RunFFmpegCommand([]string{
		"-i", concatPath,
		"-t", fmt.Sprintf("%.3f", audioDuration),
		"-vf", vfFilter,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-an", // no audio track – audio comes from TTS
		"-y", trimmedPath,
	}); err != nil {
		return "", fmt.Errorf("segment %d normalize+trim failed: %w", segIndex, err)
	}

	fmt.Printf("[SegVideo %d] Ready: %s (%.2fs, %s)\n", segIndex, trimmedPath, audioDuration, orientation)
	return trimmedPath, nil
}

// videoInfo holds just the URL + duration of a Pexels video file match
type videoInfo struct {
	Link     string
	Duration int
}

// searchVideoInfos searches Pexels and returns ordered list of (link, duration) for the best-quality files.
// orientation: "landscape", "portrait", or "square"
func (sv *StockVideoService) searchVideoInfos(keywords string, perPage int, orientation string) ([]videoInfo, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", fmt.Sprintf("%d", perPage))
	params.Add("orientation", orientation)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", sv.apiKey)

	var resp *http.Response
	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		resp, err = sv.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("pexels API rate limited (429)")
			time.Sleep(3 * time.Second) // Extra backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("pexels API returned status %d", resp.StatusCode)
			continue
		}

		// Success
		break
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pexels search failed after %d retries: %v", maxRetries, lastErr)
	}
	defer resp.Body.Close()

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
			if orientation == "portrait" {
				// For portrait: prefer 1080x1920 or tall videos
				ar := 0.0
				if file.Width > 0 {
					ar = float64(file.Height) / float64(file.Width)
				}
				isPortrait916 := ar > 1.77 && ar < 1.79
				if file.Width == 1080 && file.Height == 1920 {
					score = 10000
				} else if isPortrait916 && file.Height >= 1280 {
					score = 5000
				} else if isPortrait916 {
					score = 1000
				} else if file.Quality == "hd" {
					score = 500
				} else {
					score = 1
				}
				score += file.Height // taller = better for portrait
			} else {
				// For landscape: prefer 1920x1080
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
			}
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
func (sv *StockVideoService) searchMultipleVideos(keywords string, targetDuration float64, orientation string) ([]string, error) {
	baseURL := "https://api.pexels.com/videos/search"
	params := url.Values{}
	params.Add("query", keywords)
	params.Add("per_page", "100") // Get more results to filter
	params.Add("orientation", orientation)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", sv.apiKey)

	var resp *http.Response
	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		resp, err = sv.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("pexels API rate limited (429)")
			time.Sleep(3 * time.Second) // Extra backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("pexels API returned status %d", resp.StatusCode)
			continue
		}

		// Success
		break
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pexels search failed after %d retries: %v", maxRetries, lastErr)
	}
	defer resp.Body.Close()

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
			var bestLink string
			var bestScore int

			for _, file := range video.VideoFiles {
				currentScore := 0
				var aspectRatio float64

				if orientation == "portrait" {
					if file.Width > 0 {
						aspectRatio = float64(file.Height) / float64(file.Width)
					}
					isPortrait916 := aspectRatio > 1.77 && aspectRatio < 1.78
					if file.Width == 1080 && file.Height == 1920 {
						currentScore = 10000
					} else if isPortrait916 && file.Height >= 1280 {
						currentScore = 5000
					} else if isPortrait916 {
						currentScore = 1000
					} else if file.Quality == "hd" {
						currentScore = 500
					} else {
						currentScore = 1
					}
					currentScore += file.Height
				} else {
					if file.Height > 0 {
						aspectRatio = float64(file.Width) / float64(file.Height)
					}
					is16_9 := aspectRatio > 1.77 && aspectRatio < 1.78
					if file.Width == 1920 && file.Height == 1080 {
						currentScore = 10000
					} else if is16_9 && file.Width >= 1280 {
						currentScore = 5000
					} else if is16_9 {
						currentScore = 1000
					} else if file.Quality == "hd" {
						currentScore = 500
					} else {
						currentScore = 1
					}
					currentScore += file.Width
				}

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
		return nil, fmt.Errorf("no short videos (5-35s) found for keywords: %s", keywords)
	}

	var selectedURLs []string
	var totalDuration float64

	for _, video := range shortVideos {
		selectedURLs = append(selectedURLs, video.Link)
		totalDuration += float64(video.Duration)

		if totalDuration >= targetDuration {
			break
		}
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
	safeTargetDuration := targetDuration + 5.0

	if currentEffective < safeTargetDuration {
		fmt.Printf("[Stock Video] Effective duration (%.1fs) < target (%.1fs), looping videos...\n", currentEffective, safeTargetDuration)

		// Seed random for variety
		rand.Seed(time.Now().UnixNano())

		// Keep adding random videos until we have enough duration
		for currentEffective < safeTargetDuration {
			randomIdx := rand.Intn(len(inputPaths))
			finalInputPaths = append(finalInputPaths, inputPaths[randomIdx])

			duration, _ := utils.GetVideoDuration(inputPaths[randomIdx])
			currentRawDuration += duration
			currentCount++

			currentEffective = currentRawDuration - float64(currentCount-1)*transitionDuration

			if len(finalInputPaths) > 100 {
				break
			}
		}
		fmt.Printf("[Stock Video] Extended to %d video segments (effective ~%.1fs)\n", len(finalInputPaths), currentEffective)
	}

	// Use FFmpeg's MergeVideosWithTransition utility
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

	// Trim to target duration + 2s buffer
	return utils.TrimVideo(mergedPath, outputPath, targetDuration+2.0)
}
