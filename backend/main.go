package main

import (
	"aituber/config"
	"aituber/internal/model"
	"aituber/internal/repository"
	"aituber/internal/router"
	"aituber/internal/service"
	"aituber/pkg/database"
	"aituber/pkg/jwtutil" // Dummy change to align text if needed, but really just removing duplicate
	"aituber/utils"
	"fmt"
	"log"
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

	// Initialize Database
	db, err := database.Connect(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Migrate all models
	if err := database.AutoMigrate(db, &model.User{}, &model.Job{}, &model.Series{}, &model.Video{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT Manager
	jwtManager := jwtutil.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenExpiry)

	// Create Gin router
	r := gin.Default()

	// Setup CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files from OutputDir
	r.Static("/ai-videos", cfg.OutputDir)
	log.Printf("Serving static files from: %s", cfg.OutputDir)

	// --- SETUP DEPENDENCY INJECTION ---
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)
	seriesRepo := repository.NewSeriesRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	// 2. Services
	jobSvc := service.NewJobService(jobRepo, seriesRepo, videoRepo)

	// 3. Legacy Core Services (Working with new JobService)
	ttsPool := utils.NewAPIKeyPool(cfg.TTSAPIKeys)
	var videoPool *utils.APIKeyPool
	if len(cfg.VideoAPIKeys) > 0 {
		videoPool = utils.NewAPIKeyPool(cfg.VideoAPIKeys)
	} else {
		videoPool = utils.NewAPIKeyPool([]string{"placeholder"})
	}

	textProcessor := service.NewTextProcessor(cfg.AudioChunkSize, cfg.VideoSegmentDuration)
	audioService := service.NewAudioService(
		ttsPool,
		cfg.ElevenLabsAPIKey,
		cfg.TempDir,
		cfg.AudioBitrate,
		cfg.AudioSampleRate,
		cfg.AudioCrossfadeDuration,
	)
	videoProcessor := service.NewVideoProcessor(
		videoPool,
		cfg.TempDir,
		cfg.VideoBitrate,
		cfg.VideoResolution,
		cfg.VideoFPS,
		cfg.VideoTransitionDuration,
	)
	geminiService := service.NewGeminiService(cfg.GeminiAPIKeys)
	hfService := service.NewHuggingFaceService(cfg.HuggingFaceTokens)
	stockVideoService := service.NewStockVideoService(cfg.PexelsAPIKey, cfg.TempDir, cfg.CacheDir, geminiService, hfService, cfg.LocalHubURL, cfg.RemoteHubURL, cfg.RemoteHubToken)
	composerService := service.NewComposerService(cfg.VideoBitrate)

	// 4. Orchestrator Workflow (Updated to use new JobService)
	workflowSvc := service.NewVideoWorkflowService(
		cfg,
		jobSvc,
		textProcessor,
		audioService,
		videoProcessor,
		stockVideoService,
		composerService,
		geminiService,
	)

	// 5. Router Setup (Setup v1 API with Auth)
	router.Setup(r, router.RouterConfig{
		Cfg:       cfg,
		DB:        db,
		JWT:       jwtManager,
		JobSvc:    jobSvc,
		Workflow:  workflowSvc,
		ScriptSvc: geminiService,
	})

	// Keep variables used
	_ = userRepo

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
