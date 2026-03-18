package handler

import (
	"aituber/config"
	"aituber/internal/service"
	"aituber/pkg/response"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	cfg       *config.Config
	videoSvc  service.VideoService
	jobSvc    service.JobService
	workflow  service.IVideoWorkflow // We'll need to define this in internal/service
	scriptSvc service.IScriptGenerator
}

func NewVideoHandler(cfg *config.Config, videoSvc service.VideoService, jobSvc service.JobService, workflow service.IVideoWorkflow, scriptSvc service.IScriptGenerator) *VideoHandler {
	return &VideoHandler{
		cfg:       cfg,
		videoSvc:  videoSvc,
		jobSvc:    jobSvc,
		workflow:  workflow,
		scriptSvc: scriptSvc,
	}
}

// GetMyVideos godoc
// @Summary Get current user's video gallery
// @Tags Gallery
func (h *VideoHandler) GetMyVideos(c *gin.Context) {
	userID := c.GetString("user_id")
	platform := c.Query("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	videos, total, err := h.videoSvc.GetGallery(c.Request.Context(), userID, platform, page, limit)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch gallery")
		return
	}

	response.Paginated(c, videos, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	})
}

// GetMyTasks godoc
// @Summary Get current user's job history/status
// @Tags Gallery
func (h *VideoHandler) GetMyTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	platform := c.Query("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	jobs, total, err := h.jobSvc.ListUserJobs(c.Request.Context(), userID, platform, page, limit)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch tasks")
		return
	}

	response.Paginated(c, jobs, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	})
}

// TogglePublish godoc
// @Summary Toggle public status of a video
// @Tags Explore
func (h *VideoHandler) TogglePublish(c *gin.Context) {
	videoID := c.Param("id")
	userID := c.GetString("user_id")

	isPublic, err := h.videoSvc.TogglePublic(c.Request.Context(), videoID, userID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	response.OK(c, gin.H{"is_public": isPublic})
}

// GetExplore godoc
// @Summary Get public explore feed
// @Tags Explore
func (h *VideoHandler) GetExplore(c *gin.Context) {
	platform := c.Query("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	videos, total, err := h.videoSvc.GetExplore(c.Request.Context(), platform, page, limit)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch explore feed")
		return
	}

	response.Paginated(c, videos, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	})
}

// GetActiveTask godoc
// @Summary Get user's current pending task (job or series)
// @Tags Gallery
func (h *VideoHandler) GetActiveTask(c *gin.Context) {
	userID := c.GetString("user_id")
	platform := c.Query("platform")

	job, series, err := h.jobSvc.GetActiveTask(c.Request.Context(), userID, platform)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch active task")
		return
	}

	if series != nil {
		response.OK(c, gin.H{
			"type":      "series",
			"series_id": series.ID,
			"status":    series.Status,
		})
		return
	}

	if job != nil {
		response.OK(c, gin.H{
			"type":   "job",
			"job_id": job.ID,
			"status": job.Status,
		})
		return
	}

	response.OK(c, gin.H{"type": "none"})
}

// Generate handles POST /api/v1/generate
func (h *VideoHandler) Generate(c *gin.Context) {
	userID := c.GetString("user_id")
	var req service.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request: "+err.Error())
		return
	}

	// Basic validation
	if req.Platform != "youtube" && req.Platform != "tiktok" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "platform must be 'youtube' or 'tiktok'")
		return
	}
	if req.Topic == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "topic is required")
		return
	}

	// Speaking speed force
	req.SpeakingSpeed = 0

	// Content name slug
	if req.ContentName == "" {
		req.ContentName = slugify(req.Topic)
	} else {
		req.ContentName = slugify(req.ContentName)
	}
	req.ContentName = fmt.Sprintf("%s-%s", req.ContentName, time.Now().Format("0102-1504"))

	// Create job in DB via service
	job, err := h.jobSvc.CreateJob(c.Request.Context(), userID, req.Platform, req.ContentName, req.Topic, req.Voice, req.TTSProvider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create job: "+err.Error())
		return
	}

	// Start background generation
	go h.workflow.StartGeneration(job.ID, req)

	response.OK(c, gin.H{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// GetStatus handles GET /api/v1/status/:job_id
func (h *VideoHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	job, err := h.jobSvc.GetJob(c.Request.Context(), jobID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch job status")
		return
	}
	if job == nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "job not found")
		return
	}

	resp := gin.H{
		"status":       job.Status,
		"progress":     job.Progress,
		"current_step": job.CurrentStep,
	}
	if job.Status == "completed" {
		resp["video_url"] = job.SavedPath
		resp["video_path"] = job.VideoPath
		resp["saved_path"] = job.SavedPath
		resp["thumbnail_url"] = job.ThumbnailURL
	}
	if job.Status == "failed" && job.ErrorMsg != nil {
		resp["error"] = *job.ErrorMsg
	}

	response.OK(c, resp)
}

// Download handles GET /api/v1/download/:job_id
func (h *VideoHandler) Download(c *gin.Context) {
	jobID := c.Param("job_id")
	job, err := h.jobSvc.GetJob(c.Request.Context(), jobID)
	if err != nil || job == nil || job.Status != "completed" || job.VideoPath == "" {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "video not found or job not completed")
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=video_%s.mp4", jobID))
	c.File(job.VideoPath)
}

// slugify helper
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
