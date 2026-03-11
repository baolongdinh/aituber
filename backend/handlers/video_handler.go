package handlers

import (
	"aituber/config"
	"aituber/models"
	"aituber/services"
	"aituber/utils"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VideoHandler handles video generation requests
type VideoHandler struct {
	cfg               *config.Config
	textProcessor     *services.TextProcessor
	audioService      *services.AudioService
	videoService      *services.VideoService
	stockVideoService *services.StockVideoService
	composerService   *services.ComposerService
	geminiService     *services.GeminiService

	// In-memory job tracking
	jobs    map[string]*models.JobStatus
	jobsMux sync.RWMutex
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(cfg *config.Config) *VideoHandler {
	// Create API key pools
	ttsPool := utils.NewAPIKeyPool(cfg.TTSAPIKeys)

	var videoPool *utils.APIKeyPool
	if len(cfg.VideoAPIKeys) > 0 {
		videoPool = utils.NewAPIKeyPool(cfg.VideoAPIKeys)
	} else {
		videoPool = utils.NewAPIKeyPool([]string{"placeholder"})
	}

	// Initialize services
	textProcessor := services.NewTextProcessor(cfg.AudioChunkSize, cfg.VideoSegmentDuration)

	audioService := services.NewAudioService(
		ttsPool,
		cfg.ElevenLabsAPIKey,
		cfg.TempDir,
		cfg.AudioBitrate,
		cfg.AudioSampleRate,
		cfg.AudioCrossfadeDuration,
	)

	videoService := services.NewVideoService(
		videoPool,
		cfg.TempDir,
		cfg.VideoBitrate,
		cfg.VideoResolution,
		cfg.VideoFPS,
		cfg.VideoTransitionDuration,
	)

	geminiService := services.NewGeminiService(cfg.GeminiAPIKeys)
	hfService := services.NewHuggingFaceService(cfg.HuggingFaceTokens)
	stockVideoService := services.NewStockVideoService(cfg.PexelsAPIKey, cfg.TempDir, cfg.CacheDir, geminiService, hfService)
	composerService := services.NewComposerService(cfg.VideoBitrate)

	return &VideoHandler{
		cfg:               cfg,
		textProcessor:     textProcessor,
		audioService:      audioService,
		videoService:      videoService,
		stockVideoService: stockVideoService,
		composerService:   composerService,
		geminiService:     geminiService,
		jobs:              make(map[string]*models.JobStatus),
	}
}

// Generate handles POST /api/generate
func (h *VideoHandler) Generate(c *gin.Context) {
	var req models.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate platform
	if req.Platform != "youtube" && req.Platform != "tiktok" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform must be 'youtube' or 'tiktok'"})
		return
	}

	// Validate topic
	if req.Topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic is required"})
		return
	}

	// If no pre-written script, we need Gemini to generate one
	if req.Script == "" && !h.geminiService.HasKeys() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No GEMINI_API_KEYS configured — cannot auto-generate script. Please provide a pre-written script or add GEMINI_API_KEYS to .env"})
		return
	}

	// Set default speaking speed if not provided
	if req.SpeakingSpeed == 0 {
		if req.Platform == "tiktok" {
			req.SpeakingSpeed = 1.2 // TikTok: faster pacing = more engaging
		} else {
			req.SpeakingSpeed = 1.0 // YouTube: normal balanced pace
		}
	}
	// Validate speaking speed range
	if req.SpeakingSpeed < 0.5 || req.SpeakingSpeed > 2.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Speaking speed must be between 0.5 and 2.0"})
		return
	}

	// Auto-generate ContentName from topic if not provided
	if req.ContentName == "" {
		req.ContentName = slugify(req.Topic)
	} else {
		req.ContentName = slugify(req.ContentName)
	}
	// Append short timestamp to avoid collisions
	req.ContentName = fmt.Sprintf("%s-%s", req.ContentName, time.Now().Format("0102-1504"))

	// Generate job ID
	jobID := uuid.New().String()

	// Create job status
	job := &models.JobStatus{
		JobID:       jobID,
		Platform:    req.Platform,
		ContentName: req.ContentName,
		Status:      "processing",
		Progress:    0,
		CurrentStep: "Initializing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	h.jobsMux.Lock()
	h.jobs[jobID] = job
	h.jobsMux.Unlock()

	// Start background processing
	go h.processVideoGeneration(jobID, req)

	// Return job ID immediately
	c.JSON(http.StatusOK, models.GenerateResponse{
		JobID:  jobID,
		Status: "processing",
	})
}

// GetStatus handles GET /api/status/:job_id
func (h *VideoHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	h.jobsMux.RLock()
	job, exists := h.jobs[jobID]
	h.jobsMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Build response
	resp := models.StatusResponse{
		Status:      job.Status,
		Progress:    job.Progress,
		CurrentStep: job.CurrentStep,
	}

	if job.Status == "completed" && job.VideoPath != "" {
		videoURL := fmt.Sprintf("/api/download/%s", jobID)
		resp.VideoURL = &videoURL
	}

	if job.Status == "completed" && job.SavedPath != "" {
		resp.SavedPath = &job.SavedPath
	}

	if job.Error != nil {
		errMsg := job.Error.Error()
		resp.Error = &errMsg
	}

	c.JSON(http.StatusOK, resp)
}

// DownloadSubtitle handles GET /api/download-subtitle/:job_id
func (h *VideoHandler) DownloadSubtitle(c *gin.Context) {
	jobID := c.Param("job_id")

	h.jobsMux.RLock()
	job, exists := h.jobs[jobID]
	h.jobsMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job not completed yet"})
		return
	}

	srtPath := filepath.Join(h.cfg.TempDir, jobID, "output", "subtitles.srt")

	if _, err := os.Stat(srtPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtitle file not found"})
		return
	}

	c.Header("Content-Type", "application/x-subrip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=subtitles_%s.srt", jobID))
	c.File(srtPath)
}

// Download handles GET /api/download/:job_id
func (h *VideoHandler) Download(c *gin.Context) {
	jobID := c.Param("job_id")

	h.jobsMux.RLock()
	job, exists := h.jobs[jobID]
	h.jobsMux.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if job.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job not completed yet"})
		return
	}

	if job.VideoPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video file not found"})
		return
	}

	// Stream video file
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=video_%s.mp4", jobID))
	c.File(job.VideoPath)

	// Schedule cleanup after download (1 hour)
	go utils.ScheduleCleanup(h.cfg.TempDir, jobID, 1*time.Hour)
}

// registerJob creates and stores a new job status directly. Used by SeriesHandler.
func (h *VideoHandler) registerJob(jobID string, req models.GenerateRequest) *models.JobStatus {
	// Auto-generate ContentName from topic if not provided
	contentName := req.ContentName
	if contentName == "" {
		contentName = slugify(req.Topic)
	} else {
		contentName = slugify(contentName)
	}

	job := &models.JobStatus{
		JobID:       jobID,
		Platform:    req.Platform,
		ContentName: contentName,
		Status:      "processing",
		Progress:    0,
		CurrentStep: "Initializing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	h.jobsMux.Lock()
	h.jobs[jobID] = job
	h.jobsMux.Unlock()

	return job
}

// processVideoGeneration processes video generation in background
func (h *VideoHandler) processVideoGeneration(jobID string, req models.GenerateRequest) {
	// helper function to update status
	updateStatus := func(step string, progress int) {
		h.jobsMux.Lock()
		if job, exists := h.jobs[jobID]; exists {
			job.CurrentStep = step
			job.Progress = progress
			job.UpdatedAt = time.Now()
		}
		h.jobsMux.Unlock()
		log.Printf("[Job %s] %s (%d%%)", jobID, step, progress)
	}

	updateStatus("Initializing job", 1)

	// ── Step 0: Generate script with Gemini (if not pre-provided) ──
	var segments []models.VideoSegment
	script := req.Script
	if script == "" {
		updateStatus("Generating script with Gemini AI", 8)
		var genErr error
		if req.Platform == "tiktok" {
			segments, genErr = h.geminiService.GenerateTikTokScript(req.Topic)
		} else {
			segments, genErr = h.geminiService.GenerateYouTubeScript(req.Topic)
		}
		if genErr != nil {
			h.markJobFailed(jobID, fmt.Errorf("Gemini script generation failed: %w", genErr))
			return
		}
		log.Printf("[Job %s] Generated script (%d segments) for topic: %q", jobID, len(segments), req.Topic)
	} else {
		// Legacy support for direct script input (fallback to simple extraction)
		if len(script) > h.cfg.MaxTextLength {
			script = script[:h.cfg.MaxTextLength]
			log.Printf("[Job %s] Script truncated to %d chars", jobID, h.cfg.MaxTextLength)
		}
		chunks := h.textProcessor.SplitForSubtitles(script)
		for _, chunk := range chunks {
			segments = append(segments, models.VideoSegment{
				Text:         chunk,
				VisualPrompt: h.textProcessor.ExtractKeywordsFromText(chunk, req.StockKeywords),
			})
		}
		log.Printf("[Job %s] Created %d segments from direct script text", jobID, len(segments))
	}

	h.processVideoGenerationWithSegments(jobID, req, segments, nil)
}

// processVideoGenerationWithSegments runs the video pipeline given an already generated script/segment list.
// If job is provided, it updates that struct, otherwise it fetches from jobs map.
func (h *VideoHandler) processVideoGenerationWithSegments(jobID string, req models.GenerateRequest, segments []models.VideoSegment, job *models.JobStatus) {
	// Helper function to update status
	updateStatus := func(step string, progress int) {
		h.jobsMux.Lock()
		if job != nil {
			job.CurrentStep = step
			job.Progress = progress
			job.UpdatedAt = time.Now()
		} else if j, exists := h.jobs[jobID]; exists {
			j.CurrentStep = step
			j.Progress = progress
			j.UpdatedAt = time.Now()
		}
		h.jobsMux.Unlock()
		log.Printf("[Job %s] %s (%d%%)", jobID, step, progress)
	}

	updateStatus("Creating temporary directories", 3)

	// Create temp directories
	tempDir, err := utils.CreateTempDir(h.cfg.TempDir, jobID)
	if err != nil {
		h.markJobFailed(jobID, fmt.Errorf("failed to create temp dir: %w", err))
		return
	}

	// Determine platform orientation
	orientation := "landscape"
	if req.Platform == "tiktok" {
		orientation = "portrait"
	}

	// Step 1: Extract text items into flattened array
	updateStatus("Preparing text for audio generation", 12)
	var audioTexts []string
	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) != "" {
			audioTexts = append(audioTexts, seg.Text)
		}
	}

	if len(audioTexts) == 0 {
		h.markJobFailed(jobID, fmt.Errorf("no valid script segments extracted to process"))
		return
	}

	// Step 2: Generate audio
	var audioPaths []string
	if req.TTSProvider == "elevenlabs" {
		updateStatus("Generating full-script audio with ElevenLabs", 20)
		audioPaths, err = h.audioService.GenerateAudioFullScript(segments, req.Voice, jobID)
	} else {
		// --- FPT.AI Optimized Consolidate Flow ---
		type ttsGroup struct {
			text     string
			segIdxs  []int
			charLens []int
		}
		var ttsGroups []ttsGroup
		var currentGroup ttsGroup
		var currentChars int

		for i, seg := range segments {
			txt := strings.TrimSpace(seg.Text)
			if txt == "" {
				continue
			}

			// If adding this segment exceeds chunk size, start a new group
			if currentChars > 0 && currentChars+len(txt) > h.cfg.AudioChunkSize {
				ttsGroups = append(ttsGroups, currentGroup)
				currentGroup = ttsGroup{}
				currentChars = 0
			}

			if currentGroup.text != "" {
				currentGroup.text += " . " // Standard pause separator
			}
			currentGroup.text += txt
			currentGroup.segIdxs = append(currentGroup.segIdxs, i)
			currentGroup.charLens = append(currentGroup.charLens, len(txt))
			currentChars += len(txt)
		}
		if currentChars > 0 {
			ttsGroups = append(ttsGroups, currentGroup)
		}

		var groupTexts []string
		for _, g := range ttsGroups {
			groupTexts = append(groupTexts, g.text)
		}

		updateStatus(fmt.Sprintf("Generating %d consolidated audio chunks (FPT.AI)", len(groupTexts)), 20)
		groupPaths, ttsErr := h.audioService.GenerateAudioChunks(
			groupTexts,
			req.Voice,
			req.SpeakingSpeed,
			jobID,
			h.cfg.MaxConcurrentTTSRequests,
		)
		if ttsErr != nil {
			h.markJobFailed(jobID, fmt.Errorf("FPT TTS failed: %w", ttsErr))
			return
		}

		// Split group audio back into segment-level audio files
		audioPaths = make([]string, len(segments))
		for gIdx, groupPath := range groupPaths {
			group := ttsGroups[gIdx]

			// If only one segment in this group, use the file directly
			if len(group.segIdxs) == 1 {
				audioPaths[group.segIdxs[0]] = groupPath
				continue
			}

			// Proportional split based on character count
			// Since TTS is highly linear, this is very reliable for FPT.AI
			totalCharsInGroup := 0
			for _, l := range group.charLens {
				totalCharsInGroup += l
			}

			groupDur, _ := utils.GetAudioDuration(groupPath)
			var currentTime float64 = 0.0

			groupDir := filepath.Dir(groupPath)
			for sIdx, segIdx := range group.segIdxs {
				segDur := (float64(group.charLens[sIdx]) / float64(totalCharsInGroup)) * groupDur
				segPath := filepath.Join(groupDir, fmt.Sprintf("chunk_split_%03d_%03d.mp3", gIdx, sIdx))

				if err := utils.ExtractAudioSegment(groupPath, currentTime, segDur, segPath); err != nil {
					log.Printf("[Job %s] Failed to split group audio at %d: %v", jobID, segIdx, err)
					audioPaths[segIdx] = groupPath // Fallback to full group audio (not ideal but safe)
				} else {
					audioPaths[segIdx] = segPath
				}
				currentTime += segDur
			}
		}
	}

	if err != nil {
		h.markJobFailed(jobID, fmt.Errorf("audio generation failed: %w", err))
		return
	}

	// Step 2b: Generate Subtitles
	updateStatus("Generating subtitles", 32)
	if _, err := h.GenerateSRT(jobID, audioPaths, audioTexts, filepath.Join(tempDir, "output"), req.Platform); err != nil {
		log.Printf("[Job %s] Failed to generate subtitles: %v", jobID, err)
		// Don't fail the whole job, just log error
	}

	// Step 3: Merge audio
	updateStatus("Merging audio", 42)
	mergedAudioPath := filepath.Join(tempDir, "output", "merged_audio.mp3")
	if err := h.audioService.MergeAudioFiles(audioPaths, mergedAudioPath); err != nil {
		h.markJobFailed(jobID, fmt.Errorf("audio merge failed: %w", err))
		return
	}

	// Step 4: Stock video per segment (always stock mode now)
	var mergedVideoPath string

	updateStatus("Preparing per-segment stock videos", 50)

	// Collect real duration of every audio chunk
	realDurations := make([]float64, len(audioPaths))
	for i, ap := range audioPaths {
		d, err := utils.GetAudioDuration(ap)
		if err != nil {
			log.Printf("[Job %s] Could not get duration of chunk %d: %v (using estimate 5s)", jobID, i, err)
			d = 5.0
		}
		realDurations[i] = d
	}

	// Extract keywords specific per-segment mapped from JSON
	segKeywords := make([]string, len(segments))
	for i, seg := range segments {
		segKeywords[i] = seg.VisualPrompt
		// Fallback just in case Gemini returned empty query
		if strings.TrimSpace(segKeywords[i]) == "" {
			segKeywords[i] = h.textProcessor.ExtractKeywordsFromText(seg.Text, req.StockKeywords)
		}
		log.Printf("[Job %s] Segment %d stock video keywords: %q", jobID, i, segKeywords[i])
	}

	// Fetch + trim a stock video per segment in parallel (using configured max concurrency)
	segVideoPaths := make([]string, len(segments))
	segErrors := make([]error, len(segments))
	sem := make(chan struct{}, h.cfg.MaxConcurrentVideoRequests)
	var wg sync.WaitGroup

	for i := range segments {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			updateStatus(fmt.Sprintf("Fetching stock video for segment %d/%d", idx+1, len(segments)), 50+idx*30/len(segments))

			// Create a per-segment context with timeout (3 mins per segment should be plenty)
			segCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			vp, err := h.stockVideoService.PrepareSegmentVideo(
				segCtx,
				segKeywords[idx],
				segments[idx].VisualDescription,
				req.T2VModel,
				req.T2VProvider,
				realDurations[idx],
				jobID,
				idx,
				orientation,
			)
			if err != nil {
				segErrors[idx] = err
				log.Printf("[Job %s] Segment %d video error: %v", jobID, idx, err)
			} else {
				segVideoPaths[idx] = vp
			}
		}(i)
	}
	wg.Wait()

	// Check for segment errors (strict: all must succeed or use fallback)
	var finalSegPaths []string
	for i, err := range segErrors {
		if err != nil {
			h.markJobFailed(jobID, fmt.Errorf("segment %d video generation failed critically: %w", i, err))
			return
		}
		if segVideoPaths[i] == "" {
			h.markJobFailed(jobID, fmt.Errorf("segment %d returned empty video path", i))
			return
		}
		finalSegPaths = append(finalSegPaths, segVideoPaths[i])
	}

	if len(finalSegPaths) != len(segments) {
		h.markJobFailed(jobID, fmt.Errorf("segment count mismatch: got %d, want %d", len(finalSegPaths), len(segments)))
		return
	}

	// Concatenate all segment videos into one video track
	updateStatus("Concatenating segment videos", 82)
	concatVideoPath := filepath.Join(tempDir, "output", "segments_concat.mp4")
	if err := utils.ConcatVideosNoAudio(finalSegPaths, concatVideoPath); err != nil {
		h.markJobFailed(jobID, fmt.Errorf("segment video concat failed: %w", err))
		return
	}
	mergedVideoPath = concatVideoPath

	// Step 8: Compose final video
	updateStatus("Composing final video with audio", 90)
	composedPath := filepath.Join(tempDir, "output", "final_video_composed.mp4")
	if err := h.composerService.ComposeVideoWithAudio(mergedVideoPath, mergedAudioPath, composedPath); err != nil {
		h.markJobFailed(jobID, fmt.Errorf("composition failed: %w", err))
		return
	}
	finalVideoPath := composedPath // start with composed path

	// Step 8.5: Burn subtitles for premium look
	updateStatus("Burning premium subtitles", 92)
	srtPath := filepath.Join(tempDir, "output", "subtitles.srt")
	videoWithSubs := filepath.Join(tempDir, "output", "final_with_subtitles.mp4")
	if _, err := os.Stat(srtPath); err == nil {
		if err := utils.BurnSubtitles(composedPath, srtPath, videoWithSubs, orientation); err != nil {
			log.Printf("[Job %s] Subtitle burning failed: %v", jobID, err)
		} else {
			finalVideoPath = videoWithSubs
		}
	}

	// Step 9: Add Intro/Outro ONLY for youtube
	updateStatus("Adding intro/outro", 95)

	introPath := "static/intro_video.mp4"
	outroPath := "static/outro_video.mp4"

	concatList := h.BuildFinalConcatList(req.Platform, introPath, outroPath, finalVideoPath)

	if len(concatList) > 1 {
		finalWithIntroOutro := filepath.Join(tempDir, "output", "final_complete.mp4")
		if err := utils.ConcatVideos(concatList, finalWithIntroOutro); err != nil {
			h.markJobFailed(jobID, fmt.Errorf("failed to add intro/outro: %w", err))
			return
		}
		finalVideoPath = finalWithIntroOutro
	}

	// Step 10: Save to ai-videos/{platform}/{content-name}/
	updateStatus("Saving video to output folder", 98)
	savedPath, err := h.saveToOutputFolder(finalVideoPath, req.Platform, req.ContentName)
	if err != nil {
		// Non-fatal: log but don't fail the job
		log.Printf("[Job %s] Warning: could not save to output folder: %v", jobID, err)
		savedPath = ""
	} else {
		log.Printf("[Job %s] Video saved to: %s", jobID, savedPath)
	}

	// Complete
	updateStatus("Complete", 100)
	h.jobsMux.Lock()
	if job, exists := h.jobs[jobID]; exists {
		job.Status = "completed"
		job.VideoPath = finalVideoPath
		job.SavedPath = savedPath
		job.UpdatedAt = time.Now()
	}
	h.jobsMux.Unlock()

	log.Printf("[Job %s] Video generation completed successfully", jobID)
}

// saveToOutputFolder copies the final video to ai-videos/{platform}/{contentName}/
func (h *VideoHandler) saveToOutputFolder(srcPath, platform, contentName string) (string, error) {
	destDir := filepath.Join(h.cfg.OutputDir, platform, contentName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	destPath := filepath.Join(destDir, "final_video.mp4")

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create dest file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Return relative path for display
	return filepath.Join("ai-videos", platform, contentName, "final_video.mp4"), nil
}

// BuildFinalConcatList returns the list of video paths to concatenate based on platform.
// Only "youtube" platform includes the intro and outro if they exist on disk.
func (h *VideoHandler) BuildFinalConcatList(platform, introPath, outroPath, mainVideoPath string) []string {
	var concatList []string

	if platform == "youtube" {
		if _, err := os.Stat(introPath); err == nil {
			concatList = append(concatList, introPath)
		}
	}

	concatList = append(concatList, mainVideoPath)

	if platform == "youtube" {
		if _, err := os.Stat(outroPath); err == nil {
			concatList = append(concatList, outroPath)
		}
	}

	return concatList
}

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	// Lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric chars except hyphens
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	s = re.ReplaceAllString(s, "")
	// Collapse multiple hyphens
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	// Trim hyphens from ends
	s = strings.Trim(s, "-")
	// Limit length
	if len(s) > 60 {
		s = s[:60]
	}
	if s == "" {
		s = "content"
	}
	return s
}

// markJobFailed marks a job as failed
func (h *VideoHandler) markJobFailed(jobID string, err error) {
	log.Printf("[Job %s] FAILED: %v", jobID, err)
	h.jobsMux.Lock()
	if job, exists := h.jobs[jobID]; exists {
		job.Status = "failed"
		job.Error = err
		job.UpdatedAt = time.Now()
	}
	h.jobsMux.Unlock()
}

// GenerateSRT generates SRT subtitle file from audio chunks
func (h *VideoHandler) GenerateSRT(jobID string, audioPaths []string, texts []string, outputDir string, platform string) (string, error) {
	srtPath := filepath.Join(outputDir, "subtitles.srt")
	file, err := os.Create(srtPath)
	if err != nil {
		return "", fmt.Errorf("failed to create SRT file: %w", err)
	}
	defer file.Close()

	// Calculate initial offset (Always 0.0 because subtitles are burned into main content before intro/outro)
	currentOffset := 0.0

	for i, audioPath := range audioPaths {
		if i >= len(texts) {
			break
		}

		duration, err := utils.GetAudioDuration(audioPath)
		if err != nil {
			return "", fmt.Errorf("failed to get audio duration for %s: %w", audioPath, err)
		}

		start := currentOffset
		end := currentOffset + duration
		currentOffset += duration

		// Format timestamp: HH:MM:SS,mmm
		startStr := utils.FormatSRTTimestamp(start)
		endStr := utils.FormatSRTTimestamp(end)

		// Write to file
		fmt.Fprintf(file, "%d\n%s --> %s\n%s\n\n", i+1, startStr, endStr, texts[i])
	}

	return srtPath, nil
}
