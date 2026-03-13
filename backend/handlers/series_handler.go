package handlers

import (
	"aituber/config"
	"aituber/models"
	"aituber/services"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SeriesHandler handles multi-part series video generation
type SeriesHandler struct {
	cfg           *config.Config
	jobManager    services.IJobManager
	workflow      services.IVideoWorkflow
	geminiService services.IScriptGenerator

	seriesMu sync.RWMutex
	series   map[string]*models.SeriesJobStatus
}

// NewSeriesHandler creates a SeriesHandler sharing services
func NewSeriesHandler(
	cfg *config.Config,
	jobManager services.IJobManager,
	workflow services.IVideoWorkflow,
	gemini services.IScriptGenerator,
) *SeriesHandler {
	return &SeriesHandler{
		cfg:           cfg,
		jobManager:    jobManager,
		workflow:      workflow,
		geminiService: gemini,
		series:        make(map[string]*models.SeriesJobStatus),
	}
}

// GenerateSeries handles POST /api/generate-series
func (sh *SeriesHandler) GenerateSeries(c *gin.Context) {
	var req models.SeriesGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate platform
	if req.Platform != "youtube" && req.Platform != "tiktok" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform must be 'youtube' or 'tiktok'"})
		return
	}

	// Validate num_parts
	if req.NumParts < 2 || req.NumParts > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "num_parts must be between 2 and 20"})
		return
	}

	// Gemini required for series
	if !sh.geminiService.HasKeys() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GEMINI_API_KEYS required for series generation"})
		return
	}

	// Default speaking speed
	// Force speaking speed to 0.8 for FPT TTS (hard-coded)
	req.SpeakingSpeed = 0.8

	// Slug content name
	baseName := req.ContentName
	if baseName == "" {
		baseName = slugify(req.Topic)
	} else {
		baseName = slugify(baseName)
	}

	seriesID := uuid.New().String()

	// Init blank SeriesJobStatus
	parts := make([]*models.SeriesPartStatus, req.NumParts)
	for i := range parts {
		parts[i] = &models.SeriesPartStatus{
			PartIndex: i,
			Status:    "queued",
		}
	}

	job := &models.SeriesJobStatus{
		SeriesID:      seriesID,
		Topic:         req.Topic,
		NumParts:      req.NumParts,
		Platform:      req.Platform,
		ContentName:   baseName,
		Voice:         req.Voice,
		SpeakingSpeed: req.SpeakingSpeed,
		TTSProvider:   req.TTSProvider,
		T2VModel:      req.T2VModel,
		T2VProvider:   req.T2VProvider,
		Status:        "processing",
		Parts:         parts,
		Scripts:       make([][]models.VideoSegment, req.NumParts),
		ChildJobIDs:   make([]string, req.NumParts),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	sh.seriesMu.Lock()
	sh.series[seriesID] = job
	sh.seriesMu.Unlock()

	// Start processing in background
	go sh.processSeriesGeneration(seriesID, req)

	c.JSON(http.StatusAccepted, models.SeriesGenerateResponse{
		SeriesID: seriesID,
		Status:   "processing",
		NumParts: req.NumParts,
	})
}

// GetSeriesStatus handles GET /api/series-status/:series_id
func (sh *SeriesHandler) GetSeriesStatus(c *gin.Context) {
	seriesID := c.Param("series_id")

	sh.seriesMu.RLock()
	job, exists := sh.series[seriesID]
	sh.seriesMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	// Calculate overall progress
	var totalProgress int
	for _, p := range job.Parts {
		totalProgress += p.Progress
	}
	overallProgress := 0
	if len(job.Parts) > 0 {
		overallProgress = totalProgress / len(job.Parts)
	}

	c.JSON(http.StatusOK, gin.H{
		"series_id":        job.SeriesID,
		"topic":            job.Topic,
		"platform":         job.Platform,
		"status":           job.Status,
		"overall_progress": overallProgress,
		"num_parts":        job.NumParts,
		"parts":            job.Parts,
	})
}

// processSeriesGeneration is the background worker for series generation.
func (sh *SeriesHandler) processSeriesGeneration(seriesID string, req models.SeriesGenerateRequest) {
	log.Printf("[Series %s] Starting: topic=%q parts=%d platform=%s", seriesID, req.Topic, req.NumParts, req.Platform)

	updateSeries := func(status string) {
		sh.seriesMu.Lock()
		if s, ok := sh.series[seriesID]; ok {
			s.Status = status
			s.UpdatedAt = time.Now()
		}
		sh.seriesMu.Unlock()
	}

	updatePart := func(idx int, fn func(*models.SeriesPartStatus)) {
		sh.seriesMu.Lock()
		if s, ok := sh.series[seriesID]; ok && idx < len(s.Parts) {
			fn(s.Parts[idx])
			s.UpdatedAt = time.Now()
		}
		sh.seriesMu.Unlock()
	}

	// ── Step 1: Generate series outline ──────────────────────────
	log.Printf("[Series %s] Generating outline...", seriesID)
	outlines, err := sh.geminiService.GenerateSeriesOutline(req.Topic, req.Platform, req.NumParts)
	if err != nil {
		log.Printf("[Series %s] Outline generation failed: %v", seriesID, err)
		updateSeries("failed")
		return
	}

	// Populate part titles from outline
	for i, o := range outlines {
		if i >= req.NumParts {
			break
		}
		idx := i
		title := o.Title
		updatePart(idx, func(p *models.SeriesPartStatus) { p.Title = title })
	}

	// ── Step 2 & 3: Gen script then render each part IN PARALLEL ────────
	log.Printf("[Series %s] Processing %d parts in parallel...", seriesID, req.NumParts)
	scripts := make([][]models.VideoSegment, req.NumParts)

	var wg sync.WaitGroup

	for i := 0; i < req.NumParts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// --- Script generation ---
			updatePart(idx, func(p *models.SeriesPartStatus) {
				p.Status = "processing"
				p.CurrentStep = "Generating script"
				p.Progress = 5
			})

			segs, err := sh.geminiService.GenerateSeriesPartScript(req.Topic, req.Platform, outlines, idx)
			if err != nil {
				log.Printf("[Series %s] Part %d script failed: %v", seriesID, idx+1, err)
				errStr := err.Error()
				updatePart(idx, func(p *models.SeriesPartStatus) {
					p.Status = "failed"
					p.Error = &errStr
				})
				// Part fails, but others can continue
				return
			}

			scripts[idx] = segs
			sh.seriesMu.Lock()
			if s, ok := sh.series[seriesID]; ok {
				s.Scripts[idx] = segs
			}
			sh.seriesMu.Unlock()

			updatePart(idx, func(p *models.SeriesPartStatus) {
				p.CurrentStep = "Script ready"
				p.Progress = 15
			})

			// --- Video render immediately after script ---
			log.Printf("[Series %s] Rendering part %d/%d...", seriesID, idx+1, req.NumParts)
			sh.runPartGeneration(seriesID, idx)
		}(i)
	}

	// Wait for all parts to finish their generation
	wg.Wait()

	// ── Step 4: Mark series final status ─────────────────────────
	sh.updateOverallStatus(seriesID)

	sh.seriesMu.RLock()
	finalStatus := "completed"
	if s, ok := sh.series[seriesID]; ok {
		finalStatus = s.Status
	}
	sh.seriesMu.RUnlock()
	log.Printf("[Series %s] Generation finished with status: %s", seriesID, finalStatus)
}

// updateOverallStatus recalculates the series status based on part statuses
func (sh *SeriesHandler) updateOverallStatus(seriesID string) {
	sh.seriesMu.Lock()
	defer sh.seriesMu.Unlock()

	job, ok := sh.series[seriesID]
	if !ok {
		return
	}

	completed := 0
	failed := 0
	processing := 0
	for _, p := range job.Parts {
		switch p.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		default:
			processing++
		}
	}

	if processing > 0 {
		job.Status = "processing"
	} else if failed == 0 {
		job.Status = "completed"
	} else if completed == 0 {
		job.Status = "failed"
	} else {
		job.Status = "partial_failed"
	}
	job.UpdatedAt = time.Now()
}

// runPartGeneration handles the rendering of a single part.
// It assumes the script is already generated and stored in job.Scripts[idx].
func (sh *SeriesHandler) runPartGeneration(seriesID string, idx int) {
	sh.seriesMu.RLock()
	job, ok := sh.series[seriesID]
	sh.seriesMu.RUnlock()
	if !ok {
		return
	}

	if len(job.Scripts) <= idx || len(job.Scripts[idx]) == 0 {
		log.Printf("[Series %s] Part %d: No script found to render", seriesID, idx+1)
		return
	}

	script := job.Scripts[idx]

	// Build a GenerateRequest for this part
	genReq := models.GenerateRequest{
		Platform:      job.Platform,
		Topic:         job.Topic,
		Voice:         job.Voice,
		SpeakingSpeed: job.SpeakingSpeed,
		T2VModel:      job.T2VModel,
		T2VProvider:   job.T2VProvider,
		Segments:      script,
		ContentName:   fmt.Sprintf("%s-part%02d-%s", job.ContentName, idx+1, time.Now().Format("0102-1504")),
	}

	// Mint a real jobID and register it in JobManager
	jobID := uuid.New().String()
	sh.seriesMu.Lock()
	job.ChildJobIDs[idx] = jobID
	sh.seriesMu.Unlock()

	// Register the job in JobManager
	sh.jobManager.CreateJob(jobID, genReq.Platform, genReq.ContentName)

	// Progress bridge: forward VideoHandler job progress to our SeriesPartStatus
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(1 * time.Second):
				vj, exists := sh.jobManager.GetJob(jobID)
				if !exists {
					return
				}

				sh.seriesMu.Lock()
				if s, ok := sh.series[seriesID]; ok && idx < len(s.Parts) {
					p := s.Parts[idx]
					p.Progress = vj.Progress
					p.CurrentStep = vj.CurrentStep
					s.UpdatedAt = time.Now()
				}
				sh.seriesMu.Unlock()

				if vj.Status == "completed" || vj.Status == "failed" {
					return
				}
			}
		}
	}()

	// Start generation via workflow (this is usually blocking in the way it was originally used in parallel wg,
	// but workflow.StartGeneration is meant to be run async. Here we want to wait for it.)
	// Wait, the workflow.StartGeneration is NOT blocking. I should probably make a blocking version or just wait for status.
	sh.workflow.StartGeneration(jobID, genReq)

	// Wait for completion in this goroutine so wg.Done() works correctly
	for {
		vj, _ := sh.jobManager.GetJob(jobID)
		if vj.Status == "completed" || vj.Status == "failed" {
			break
		}
		time.Sleep(2 * time.Second)
	}
	close(done)

	// Sync final state
	vj, _ := sh.jobManager.GetJob(jobID)

	sh.seriesMu.Lock()
	if s, ok := sh.series[seriesID]; ok && idx < len(s.Parts) {
		p := s.Parts[idx]
		if vj.Status == "completed" {
			videoURL := fmt.Sprintf("/api/download/%s", jobID)
			savedPath := vj.SavedPath
			p.Status = "completed"
			p.Progress = 100
			p.CurrentStep = "Done"
			p.VideoURL = &videoURL
			p.SavedPath = &savedPath
			log.Printf("[Series %s] Part %d completed: %s", seriesID, idx+1, vj.VideoPath)
		} else {
			errStr := "render failed"
			if vj.Error != nil {
				errStr = vj.Error.Error()
			}
			p.Status = "failed"
			p.Error = &errStr
			log.Printf("[Series %s] Part %d FAILED: %s", seriesID, idx+1, errStr)
		}
		s.UpdatedAt = time.Now()
	}
	sh.seriesMu.Unlock()

	// Update overall series status
	sh.updateOverallStatus(seriesID)
}

// RetrySeriesPart handles POST /api/retry-series-part/:series_id/:part_index
func (sh *SeriesHandler) RetrySeriesPart(c *gin.Context) {
	seriesID := c.Param("series_id")
	partIdxStr := c.Param("part_index")

	var partIdx int
	if _, err := fmt.Sscanf(partIdxStr, "%d", &partIdx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid part_index"})
		return
	}

	sh.seriesMu.RLock()
	job, exists := sh.series[seriesID]
	sh.seriesMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	if partIdx < 0 || partIdx >= len(job.Parts) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "part_index out of bounds"})
		return
	}

	sh.seriesMu.Lock()
	part := job.Parts[partIdx]
	if part.Status == "completed" || part.Status == "processing" {
		sh.seriesMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Part is already completed or processing"})
		return
	}

	if len(job.Scripts) <= partIdx || len(job.Scripts[partIdx]) == 0 {
		sh.seriesMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Script not found for this part. Cannot retry."})
		return
	}

	// Reset part status
	part.Status = "queued"
	part.Progress = 0
	part.CurrentStep = "Retrying..."
	part.Error = nil
	job.Status = "processing"
	job.UpdatedAt = time.Now()
	sh.seriesMu.Unlock()

	// Run in background
	go sh.runPartGeneration(seriesID, partIdx)

	c.JSON(http.StatusOK, gin.H{"status": "queued", "part_index": partIdx})
}
