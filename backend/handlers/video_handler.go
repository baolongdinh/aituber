package handlers

import (
	"aituber/config"
	"aituber/models"
	"aituber/services"
	"aituber/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// In-memory job tracking
	jobs    map[string]*models.JobStatus
	jobsMux sync.RWMutex
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(cfg *config.Config) *VideoHandler {
	// Create API key pools
	ttsPool := utils.NewAPIKeyPool(cfg.TTSAPIKeys)
	videoPool := utils.NewAPIKeyPool(cfg.VideoAPIKeys)

	// Initialize services
	textProcessor := services.NewTextProcessor(cfg.AudioChunkSize, cfg.VideoSegmentDuration)

	audioService := services.NewAudioService(
		ttsPool,
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

	stockVideoService := services.NewStockVideoService(cfg.PexelsAPIKey, cfg.TempDir)

	composerService := services.NewComposerService(cfg.VideoBitrate)

	return &VideoHandler{
		cfg:               cfg,
		textProcessor:     textProcessor,
		audioService:      audioService,
		videoService:      videoService,
		stockVideoService: stockVideoService,
		composerService:   composerService,
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

	// Validate input
	if req.Script == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Script is required"})
		return
	}
	if len(req.Script) > h.cfg.MaxTextLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Script too long (max %d chars)", h.cfg.MaxTextLength)})
		return
	}

	// Set default speaking speed if not provided
	if req.SpeakingSpeed == 0 {
		req.SpeakingSpeed = 1.0
	}
	// Validate speaking speed range
	if req.SpeakingSpeed < 0.5 || req.SpeakingSpeed > 2.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Speaking speed must be between 0.5 and 2.0"})
		return
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create job status
	job := &models.JobStatus{
		JobID:       jobID,
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

	// Construct path to subtitles.srt
	// Assuming it's in the same directory as the final video but we need to find the temp dir
	// Since we don't store temp dir in job status (bad design but let's work around it),
	// we reconstruct it: tempDir/jobID/output/subtitles.srt
	// Wait, we need h.cfg.TempDir
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

// processVideoGeneration processes video generation in background
func (h *VideoHandler) processVideoGeneration(jobID string, req models.GenerateRequest) {
	// Helper function to update status
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

	updateStatus("Creating temporary directories", 5)

	// Create temp directories
	tempDir, err := utils.CreateTempDir(h.cfg.TempDir, jobID)
	if err != nil {
		h.markJobFailed(jobID, fmt.Errorf("failed to create temp dir: %w", err))
		return
	}

	// Step 1: Split text for audio (and subtitles)
	updateStatus("Splitting text for audio generation", 10)
	audioChunks := h.textProcessor.SplitForSubtitles(req.Script)
	log.Printf("[Job %s] Created %d audio chunks (subtitle segments)", jobID, len(audioChunks))

	// Step 2: Generate audio chunks
	updateStatus(fmt.Sprintf("Generating %d audio chunks", len(audioChunks)), 20)
	audioPaths, err := h.audioService.GenerateAudioChunks(
		audioChunks,
		req.Voice,
		req.SpeakingSpeed,
		jobID,
		h.cfg.MaxConcurrentTTSRequests,
	)
	if err != nil {
		h.markJobFailed(jobID, fmt.Errorf("audio generation failed: %w", err))
		return
	}

	// Step 2b: Generate Subtitles
	updateStatus("Generating subtitles", 30)
	if _, err := h.GenerateSRT(jobID, audioPaths, audioChunks, filepath.Join(tempDir, "output")); err != nil {
		log.Printf("[Job %s] Failed to generate subtitles: %v", jobID, err)
		// Don't fail the whole job, just log error
	}

	// Step 3: Merge audio
	updateStatus("Merging audio with crossfade", 40)
	mergedAudioPath := filepath.Join(tempDir, "output", "merged_audio.mp3")
	if err := h.audioService.MergeAudioFiles(audioPaths, mergedAudioPath); err != nil {
		h.markJobFailed(jobID, fmt.Errorf("audio merge failed: %w", err))
		return
	}

	// Step 4: Video Generation (AI or Stock)
	var mergedVideoPath string

	if req.VideoSource == "stock" {
		updateStatus("Preparing stock video", 50)

		// Get audio duration
		audioDuration, err := utils.GetVideoDuration(mergedAudioPath) // Works for audio too
		if err != nil {
			h.markJobFailed(jobID, fmt.Errorf("failed to get audio duration: %w", err))
			return
		}

		// Prepare stock video (search -> download -> loop -> trim)
		stockKeywords := req.StockKeywords
		if stockKeywords == "" {
			stockKeywords = "nature technology abstract" // Default fallback
		}

		mergedVideoPath, err = h.stockVideoService.PrepareStockVideo(stockKeywords, audioDuration, jobID)
		if err != nil {
			h.markJobFailed(jobID, fmt.Errorf("stock video preparation failed: %w", err))
			return
		}

	} else {
		// AI Video Generation Workflow
		updateStatus("Splitting text for video segments", 45)
		videoSegments := h.textProcessor.SplitForVideo(req.Script)
		log.Printf("[Job %s] Created %d video segments", jobID, len(videoSegments))

		// Step 5: Generate video prompts
		updateStatus("Generating video prompts", 50)
		prompts, err := h.videoService.GenerateVideoPrompts(videoSegments, req.VideoStyle)
		if err != nil {
			h.markJobFailed(jobID, fmt.Errorf("prompt generation failed: %w", err))
			return
		}

		// Step 6: Generate videos
		updateStatus(fmt.Sprintf("Generating %d video segments", len(videoSegments)), 55)
		durations := make([]float64, len(videoSegments))
		for i, seg := range videoSegments {
			durations[i] = seg.EstimatedDuration
		}

		videoPaths, err := h.videoService.GenerateVideos(
			prompts,
			durations,
			jobID,
			h.cfg.MaxConcurrentVideoRequests,
		)
		if err != nil {
			log.Printf("[Job %s] Video generation error: %v", jobID, err)
			h.markJobFailed(jobID, fmt.Errorf("video generation failed: %w", err))
			return
		}

		// Step 7: Merge videos
		updateStatus("Merging video segments with transitions", 80)
		mergedVideoPath = filepath.Join(tempDir, "output", "merged_video.mp4")
		if err := h.videoService.MergeVideos(videoPaths, mergedVideoPath); err != nil {
			h.markJobFailed(jobID, fmt.Errorf("video merge failed: %w", err))
			return
		}
	}

	// Step 8: Compose final video
	updateStatus("Composing final video with audio", 90)
	finalVideoPath := filepath.Join(tempDir, "output", "final_video.mp4")
	if err := h.composerService.ComposeVideoWithAudio(mergedVideoPath, mergedAudioPath, finalVideoPath); err != nil {
		h.markJobFailed(jobID, fmt.Errorf("composition failed: %w", err))
		return
	}

	// Step 9: Add Intro/Outro if they exist
	updateStatus("Adding intro/outro", 95)

	// Define paths relative to backend execution directory
	introPath := "static/intro_video.mp4"
	outroPath := "static/outro_video.mp4"

	concatList := []string{}

	// Check Intro
	if _, err := os.Stat(introPath); err == nil {
		concatList = append(concatList, introPath)
	}

	// Add Main Video
	concatList = append(concatList, finalVideoPath)

	// Check Outro
	if _, err := os.Stat(outroPath); err == nil {
		concatList = append(concatList, outroPath)
	}

	// If we have more than just the main video, concat them
	if len(concatList) > 1 {
		finalWithIntroOutro := filepath.Join(tempDir, "output", "final_complete.mp4")
		if err := utils.ConcatVideos(concatList, finalWithIntroOutro); err != nil {
			h.markJobFailed(jobID, fmt.Errorf("failed to add intro/outro: %w", err))
			return
		}
		// Update final video path
		finalVideoPath = finalWithIntroOutro
	}

	// Complete
	updateStatus("Complete", 100)
	h.jobsMux.Lock()
	if job, exists := h.jobs[jobID]; exists {
		job.Status = "completed"
		job.VideoPath = finalVideoPath
		job.UpdatedAt = time.Now()
	}
	h.jobsMux.Unlock()

	log.Printf("[Job %s] Video generation completed successfully", jobID)
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
func (h *VideoHandler) GenerateSRT(jobID string, audioPaths []string, texts []string, outputDir string) (string, error) {
	srtPath := filepath.Join(outputDir, "subtitles.srt")
	file, err := os.Create(srtPath)
	if err != nil {
		return "", fmt.Errorf("failed to create SRT file: %w", err)
	}
	defer file.Close()

	// Calculate initial offset (Intro duration)
	currentOffset := 0.0
	introPath := "static/intro_video.mp4"
	if _, err := os.Stat(introPath); err == nil {
		duration, err := utils.GetVideoDuration(introPath)
		if err == nil {
			currentOffset = duration
		} else {
			log.Printf("Failed to get intro duration: %v", err)
		}
	}

	for i, audioPath := range audioPaths {
		if i >= len(texts) {
			break
		}

		duration, err := utils.GetAudioDuration(audioPath)
		if err != nil {
			return "", fmt.Errorf("failed to get audio duration for %s: %w", audioPath, err)
		}

		// Account for crossfade overlap for all chunks except the first one
		if i > 0 {
			currentOffset -= h.cfg.AudioCrossfadeDuration
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
