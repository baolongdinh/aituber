package router

import (
	"aituber/config"
	"aituber/internal/handler"
	"aituber/internal/middleware"
	"aituber/internal/repository"
	"aituber/internal/service"
	"aituber/pkg/jwtutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouterConfig struct {
	Cfg       *config.Config
	DB        *gorm.DB
	JWT       *jwtutil.Manager
	JobSvc    service.JobService
	Workflow  service.IVideoWorkflow   // From existing services
	ScriptSvc service.IScriptGenerator // From existing services
}

func Setup(r *gin.Engine, c RouterConfig) {
	// Repositories
	userRepo := repository.NewUserRepository(c.DB)
	videoRepo := repository.NewVideoRepository(c.DB)

	// Services
	authSvc := service.NewAuthService(userRepo, c.JWT)
	videoSvc := service.NewVideoService(videoRepo)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(c.JWT)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userRepo, authSvc)
	videoHandler := handler.NewVideoHandler(c.Cfg, videoSvc, c.JobSvc, c.Workflow, c.ScriptSvc)
	seriesHandler := handler.NewSeriesHandler(c.Cfg, c.JobSvc, videoSvc, c.Workflow, c.ScriptSvc)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiV1 := r.Group("/api/v1")
	{
		// Public Auth
		auth := apiV1.Group("/auth")
		{
			auth.GET("/nonce", authHandler.GetNonce)
			auth.POST("/login", authHandler.LoginWithWallet)
		}

		// Public Explore
		apiV1.GET("/explore", videoHandler.GetExplore)
		apiV1.GET("/videos/:id", videoHandler.GetMyVideos) // Details (can be public/private check in svc)

		// Protected Routes
		protected := apiV1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// User Profile
			protected.GET("/me", userHandler.GetMe)
			protected.PUT("/me/profile", userHandler.UpdateProfile)

			// Generation
			protected.POST("/generate", videoHandler.Generate)
			protected.POST("/series/generate", seriesHandler.GenerateSeries)
			protected.GET("/series/status/:id", seriesHandler.GetSeriesStatus)

			// Tasks & Gallery
			protected.GET("/me/tasks", videoHandler.GetMyTasks)
			protected.GET("/me/active-task", videoHandler.GetActiveTask)
			protected.GET("/me/videos", videoHandler.GetMyVideos)
			protected.GET("/status/:job_id", videoHandler.GetStatus)
			protected.DELETE("/status/:job_id", videoHandler.Cancel)
			protected.POST("/status/:job_id/resume", videoHandler.Resume)
			protected.POST("/videos/:id/publish", videoHandler.TogglePublish)
		}

		// Downloads (can be protected or signed URL, for now simple)
		apiV1.GET("/download/:job_id", videoHandler.Download)
	}
}
