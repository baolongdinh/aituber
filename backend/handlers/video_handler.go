package handlers

import (
	"aituber/config"
	"aituber/models"
	"aituber/services"
	"aituber/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VideoHandler handles video generation requests
type VideoHandler struct {
	cfg               *config.Config
	jobManager        services.IJobManager
	workflow          services.IVideoWorkflow
	geminiSVC         services.IScriptGenerator
	textProcessor     *services.TextProcessor
	audioService      *services.AudioService
	videoService      *services.VideoService
	geminiService     *services.GeminiService
	hfService         *services.HuggingFaceService
	stockVideoService *services.StockVideoService
	composerService   *services.ComposerService
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
	stockVideoService := services.NewStockVideoService(cfg.PexelsAPIKey, cfg.TempDir, cfg.CacheDir, geminiService, hfService, cfg.LocalHubURL)
	composerService := services.NewComposerService(cfg.VideoBitrate)

	// Create job manager and workflow
	jobManager := services.NewJobManager()
	workflow := services.NewVideoWorkflowService(cfg, jobManager, textProcessor, audioService, videoService, stockVideoService, composerService, geminiService)

	return &VideoHandler{
		cfg:               cfg,
		jobManager:        jobManager,
		workflow:          workflow,
		geminiSVC:         geminiService,
		textProcessor:     textProcessor,
		audioService:      audioService,
		videoService:      videoService,
		geminiService:     geminiService,
		hfService:         hfService,
		stockVideoService: stockVideoService,
		composerService:   composerService,
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
	if req.Script == "" && !h.geminiSVC.HasKeys() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No GEMINI_API_KEYS configured — cannot auto-generate script. Please provide a pre-written script or add GEMINI_API_KEYS to .env"})
		return
	}

	// Set default speaking speed if not provided
	if req.SpeakingSpeed == 0 {
		if req.Platform == "tiktok" {
			req.SpeakingSpeed = 1.2
		} else {
			req.SpeakingSpeed = 1.0
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
	req.ContentName = fmt.Sprintf("%s-%s", req.ContentName, time.Now().Format("0102-1504"))

	// Generate job ID and register job
	jobID := uuid.New().String()
	h.jobManager.CreateJob(jobID, req.Platform, req.ContentName)

	// Start background processing via Orchestrator
	go h.workflow.StartGeneration(jobID, req)

	// Return job ID immediately
	c.JSON(http.StatusOK, models.GenerateResponse{
		JobID:  jobID,
		Status: "processing",
	})
}

// GetStatus handles GET /api/status/:job_id
func (h *VideoHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	job, exists := h.jobManager.GetJob(jobID)
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

	job, exists := h.jobManager.GetJob(jobID)
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

	job, exists := h.jobManager.GetJob(jobID)
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

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	s = re.ReplaceAllString(s, "")
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
	}
	if s == "" {
		s = "content"
	}
	return s
}
