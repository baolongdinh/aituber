package handler

import (
	"aituber/config"
	"aituber/internal/model"
	"aituber/internal/service"
	"aituber/pkg/response"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// SeriesHandler handles multi-part series video generation
type SeriesHandler struct {
	cfg       *config.Config
	jobSvc    service.JobService
	videoSvc  service.VideoService
	workflow  service.IVideoWorkflow
	scriptSvc service.IScriptGenerator
}

func NewSeriesHandler(cfg *config.Config, jobSvc service.JobService, videoSvc service.VideoService, workflow service.IVideoWorkflow, scriptSvc service.IScriptGenerator) *SeriesHandler {
	return &SeriesHandler{
		cfg:       cfg,
		jobSvc:    jobSvc,
		videoSvc:  videoSvc,
		workflow:  workflow,
		scriptSvc: scriptSvc,
	}
}

// GenerateSeries handles POST /api/v1/series/generate
func (h *SeriesHandler) GenerateSeries(c *gin.Context) {
	userID := c.GetString("user_id")
	var req service.SeriesGenerateRequest // Need to define this in internal/service
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request: "+err.Error())
		return
	}

	// Validation
	if req.Platform != "youtube" && req.Platform != "tiktok" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "platform must be 'youtube' or 'tiktok'")
		return
	}
	if req.NumParts < 2 || req.NumParts > 20 {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "num_parts must be between 2 and 20")
		return
	}
	if !h.scriptSvc.HasKeys() {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "script generator keys missing")
		return
	}

	req.SpeakingSpeed = 0

	// Create Series in DB
	series, err := h.jobSvc.CreateSeries(c.Request.Context(), userID, req.Topic, req.Platform, req.ContentName, req.NumParts)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create series: "+err.Error())
		return
	}

	// Start background processing
	go h.processSeriesGeneration(series.ID, userID, req)

	response.OK(c, gin.H{
		"series_id": series.ID,
		"status":    series.Status,
		"num_parts": series.NumParts,
	})
}

// GetSeriesStatus handles GET /api/v1/series/:id
func (h *SeriesHandler) GetSeriesStatus(c *gin.Context) {
	seriesID := c.Param("id")
	series, err := h.jobSvc.GetSeries(c.Request.Context(), seriesID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch series")
		return
	}
	if series == nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "series not found")
		return
	}

	response.OK(c, series)
}

// Internal worker logic (ported from series_handler.go)
func (h *SeriesHandler) processSeriesGeneration(seriesID, userID string, req service.SeriesGenerateRequest) {
	ctx := context.Background()
	log.Printf("[Series %s] Starting generation...", seriesID)

	// 1. Generate outline
	outlines, err := h.scriptSvc.GenerateSeriesOutline(req.Topic, req.Platform, req.NumParts) // Need to add to IScriptGenerator
	if err != nil {
		log.Printf("[Series %s] Outline failed: %v", seriesID, err)
		_ = h.jobSvc.UpdateSeriesStatus(ctx, seriesID, "failed")
		return
	}

	// 2. Pre-create all job records in DB synchronous so the UI can show them immediately
	var jobs []*model.Job
	for i := 0; i < req.NumParts; i++ {
		partName := fmt.Sprintf("%s - Part %d", req.ContentName, i+1)
		// Try to use the outline title if available
		if i < len(outlines) && outlines[i].Title != "" {
			partName = outlines[i].Title
		}

		job, err := h.jobSvc.CreateSeriesPartJob(ctx, userID, seriesID, i+1, req.Platform, partName, req.Topic, req.Voice, req.TTSProvider)
		if err != nil {
			log.Printf("[Series %s] Failed to pre-create part %d job: %v", seriesID, i, err)
			continue
		}
		jobs = append(jobs, job)
	}

	// 3. Process parts in parallel
	var wg sync.WaitGroup
	for i := 0; i < req.NumParts; i++ {
		if i >= len(jobs) {
			continue // skip if we failed to create this job
		}

		wg.Add(1)
		go func(idx int, currentJob *model.Job) {
			defer wg.Done()

			// Update status to show script generation has started
			_ = h.jobSvc.UpdateProgress(ctx, currentJob.ID, "Generating script...", 5)

			// Generate part script
			genScript, err := h.scriptSvc.GenerateSeriesPartScript(req.Topic, req.Platform, outlines, idx)
			if err != nil {
				log.Printf("[Series %s] Part %d script failed: %v", seriesID, idx, err)
				_ = h.jobSvc.MarkFailed(ctx, currentJob.ID, err)
				return
			}

			partName := currentJob.ContentName
			// If Gemini generated a better title, update the job
			if genScript.Title != "" && genScript.Title != partName {
				partName = genScript.Title
				_ = h.jobSvc.UpdateJobTitle(ctx, currentJob.ID, partName)
			}

			_ = h.jobSvc.UpdateProgress(ctx, currentJob.ID, "Preparing video generation...", 10)

			// Prepare generation request for this part
			genReq := service.GenerateRequest{
				Platform:      req.Platform,
				Topic:         req.Topic,
				ContentName:   partName,
				Voice:         req.Voice,
				SpeakingSpeed: req.SpeakingSpeed,
				TTSProvider:   req.TTSProvider,
				T2VModel:      req.T2VModel,
				T2VProvider:   req.T2VProvider,
				Segments:      genScript.Segments,
			}

			// Start individual video generation workflow
			h.workflow.StartGeneration(currentJob.ID, genReq)
		}(i, jobs[i])
	}
	wg.Wait()
	_ = h.jobSvc.UpdateSeriesStatus(ctx, seriesID, "completed")
}
