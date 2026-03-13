package main

import (
	"aituber/config"
	"aituber/handlers"
	"aituber/services"
	"aituber/utils"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded: %s", cfg)

	// Create Gin router
	router := gin.Default()

	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// --- SETUP DEPENDENCY INJECTION ---
	// 1. API pools
	ttsPool := utils.NewAPIKeyPool(cfg.TTSAPIKeys)
	var videoPool *utils.APIKeyPool
	if len(cfg.VideoAPIKeys) > 0 {
		videoPool = utils.NewAPIKeyPool(cfg.VideoAPIKeys)
	} else {
		videoPool = utils.NewAPIKeyPool([]string{"placeholder"})
	}

	// 2. Job Manager
	jobManager := services.NewJobManager()

	// 3. Core Services
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
	stockVideoService := services.NewStockVideoService(cfg.PexelsAPIKey, cfg.TempDir, cfg.CacheDir, geminiService, hfService, cfg.LocalHubURL, cfg.RemoteHubURL, cfg.RemoteHubToken)
	composerService := services.NewComposerService(cfg.VideoBitrate)

	// 4. Orchestrator Workflow
	workflowSvc := services.NewVideoWorkflowService(
		cfg,
		jobManager,
		textProcessor,
		audioService,
		videoService,
		stockVideoService,
		composerService,
		geminiService,
	)

	// 5. Initialize handlers
	videoHandler := handlers.NewVideoHandler(cfg)
	seriesHandler := handlers.NewSeriesHandler(cfg, jobManager, workflowSvc, geminiService)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/generate", videoHandler.Generate)
		api.GET("/status/:job_id", videoHandler.GetStatus)
		api.GET("/download/:job_id", videoHandler.Download)
		api.GET("/download-subtitle/:job_id", videoHandler.DownloadSubtitle)

		// Series routes
		api.POST("/generate-series", seriesHandler.GenerateSeries)
		api.GET("/series-status/:series_id", seriesHandler.GetSeriesStatus)
		api.POST("/retry-series-part/:series_id/:part_index", seriesHandler.RetrySeriesPart)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
