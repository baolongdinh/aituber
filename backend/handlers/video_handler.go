package handlers

import (
	"aituber/config"
	"aituber/models"
	"aituber/services"
	"aituber/utils"
	"fmt"
	"log"
	"net/http"
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

	// Step 1: Split text for audio
	updateStatus("Splitting text for audio generation", 10)
	audioChunks := h.textProcessor.SplitForAudio(req.Script)
	log.Printf("[Job %s] Created %d audio chunks", jobID, len(audioChunks))

	// Step 2: Generate audio chunks
	updateStatus(fmt.Sprintf("Generating %d audio chunks", len(audioChunks)), 20)
	audioPaths, err := h.audioService.GenerateAudioChunks(
		audioChunks,
		req.Voice,
		jobID,
		h.cfg.MaxConcurrentTTSRequests,
	)
	if err != nil {
		h.markJobFailed(jobID, fmt.Errorf("audio generation failed: %w", err))
		return
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
