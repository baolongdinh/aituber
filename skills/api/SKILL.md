# API Design & Development Skill

## TRIGGER
Read this file when designing APIs, implementing endpoints, creating API documentation, or optimizing API performance.

---

## IDENTITY
You are an API Architect focused on building scalable, maintainable, and developer-friendly APIs.
You design APIs that are intuitive, well-documented, and follow industry best practices.
You never create APIs without proper versioning, documentation, and error handling.

---

## 🏗️ API DESIGN PRINCIPLES

### RESTful API Design
```go
// internal/handler/nft_handler.go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "your-project/internal/dto"
    "your-project/internal/service"
    "your-project/pkg/response"
    "your-project/pkg/validator"
)

type NFTHandler struct {
    nftService service.NFTService
    validator  *validator.CustomValidator
}

func NewNFTHandler(nftService service.NFTService, validator *validator.CustomValidator) *NFTHandler {
    return &NFTHandler{
        nftService: nftService,
        validator:  validator,
    }
}

// GET /api/v1/nfts - List NFTs with pagination and filtering
func (h *NFTHandler) ListNFTs(c *gin.Context) {
    // Parse query parameters with defaults
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    sortBy := c.DefaultQuery("sort", "created_at")
    sortOrder := c.DefaultQuery("order", "desc")
    
    // Validate pagination
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }
    
    // Parse filters
    filters := dto.ListNFTsFilter{
        OwnerID:     c.Query("owner_id"),
        Category:    c.Query("category"),
        MinPrice:    c.Query("min_price"),
        MaxPrice:    c.Query("max_price"),
        Status:      c.DefaultQuery("status", "active"),
        Search:      c.Query("search"),
    }
    
    // Validate filters
    if err := h.validator.ValidateStruct(filters); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid filter parameters")
        return
    }
    
    // Create pagination info
    pagination := dto.Pagination{
        Page:     page,
        Limit:    limit,
        SortBy:   sortBy,
        SortDesc: sortOrder == "desc",
    }
    
    // Get NFTs
    nfts, total, err := h.nftService.ListNFTs(c.Request.Context(), filters, pagination)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to fetch NFTs")
        return
    }
    
    // Build response
    response.Success(c, http.StatusOK, dto.ListNFTsResponse{
        NFTs: nfts,
        Meta: response.Meta{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: (total + int64(limit) - 1) / int64(limit),
        },
    })
}

// GET /api/v1/nfts/:id - Get NFT by ID
func (h *NFTHandler) GetNFT(c *gin.Context) {
    id := c.Param("id")
    
    // Validate ID format
    if !isValidUUID(id) {
        response.Error(c, http.StatusBadRequest, "Invalid NFT ID format")
        return
    }
    
    nft, err := h.nftService.GetNFT(c.Request.Context(), id)
    if err != nil {
        if err == service.ErrNFTNotFound {
            response.Error(c, http.StatusNotFound, "NFT not found")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to fetch NFT")
        return
    }
    
    response.Success(c, http.StatusOK, nft)
}

// POST /api/v1/nfts - Create new NFT
func (h *NFTHandler) CreateNFT(c *gin.Context) {
    var req dto.CreateNFTRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Validate request
    if err := h.validator.ValidateStruct(req); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed")
        return
    }
    
    // Get user ID from context (set by auth middleware)
    userID := c.GetUint("user_id")
    
    nft, err := h.nftService.CreateNFT(c.Request.Context(), userID, req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to create NFT")
        return
    }
    
    response.Success(c, http.StatusCreated, nft)
}

// PUT /api/v1/nfts/:id - Update NFT
func (h *NFTHandler) UpdateNFT(c *gin.Context) {
    id := c.Param("id")
    
    if !isValidUUID(id) {
        response.Error(c, http.StatusBadRequest, "Invalid NFT ID format")
        return
    }
    
    var req dto.UpdateNFTRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    if err := h.validator.ValidateStruct(req); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed")
        return
    }
    
    userID := c.GetUint("user_id")
    
    nft, err := h.nftService.UpdateNFT(c.Request.Context(), id, userID, req)
    if err != nil {
        if err == service.ErrNFTNotFound {
            response.Error(c, http.StatusNotFound, "NFT not found")
            return
        }
        if err == service.ErrUnauthorized {
            response.Error(c, http.StatusForbidden, "Not authorized to update this NFT")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to update NFT")
        return
    }
    
    response.Success(c, http.StatusOK, nft)
}

// DELETE /api/v1/nfts/:id - Delete NFT
func (h *NFTHandler) DeleteNFT(c *gin.Context) {
    id := c.Param("id")
    
    if !isValidUUID(id) {
        response.Error(c, http.StatusBadRequest, "Invalid NFT ID format")
        return
    }
    
    userID := c.GetUint("user_id")
    
    err := h.nftService.DeleteNFT(c.Request.Context(), id, userID)
    if err != nil {
        if err == service.ErrNFTNotFound {
            response.Error(c, http.StatusNotFound, "NFT not found")
            return
        }
        if err == service.ErrUnauthorized {
            response.Error(c, http.StatusForbidden, "Not authorized to delete this NFT")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to delete NFT")
        return
    }
    
    response.Success(c, http.StatusNoContent, nil)
}

func isValidUUID(id string) bool {
    // Simple UUID validation (in production, use proper UUID library)
    return len(id) == 36 && id[8] == '-' && id[13] == '-' && id[18] == '-' && id[23] == '-'
}
```

### API Versioning Strategy
```go
// internal/router/versioned_router.go
package router

import (
    "github.com/gin-gonic/gin"
    
    "your-project/internal/handler"
    "your-project/pkg/middleware"
)

type VersionedRouter struct {
    v1Handlers *handler.V1Handlers
    v2Handlers *handler.V2Handlers
}

func NewVersionedRouter(v1 *handler.V1Handlers, v2 *handler.V2Handlers) *VersionedRouter {
    return &VersionedRouter{
        v1Handlers: v1,
        v2Handlers: v2,
    }
}

func (r *VersionedRouter) SetupRoutes() *gin.Engine {
    gin.SetMode(gin.ReleaseMode)
    engine := gin.New()
    
    // Global middleware
    engine.Use(gin.Logger())
    engine.Use(gin.Recovery())
    engine.Use(middleware.CORSMiddleware())
    engine.Use(middleware.SecurityHeadersMiddleware())
    engine.Use(middleware.PerformanceMiddleware())
    
    // Health check (version-independent)
    engine.GET("/health", r.v1Handlers.Health.Check)
    engine.GET("/api/health", r.v1Handlers.Health.Check)
    
    // API v1 routes
    v1 := engine.Group("/api/v1")
    {
        r.setupV1Routes(v1)
    }
    
    // API v2 routes (with breaking changes)
    v2 := engine.Group("/api/v2")
    {
        r.setupV2Routes(v2)
    }
    
    // Latest version (redirects to v2)
    latest := engine.Group("/api/latest")
    {
        r.setupV2Routes(latest)
    }
    
    return engine
}

func (r *VersionedRouter) setupV1Routes(rg *gin.RouterGroup) {
    // Public routes
    rg.GET("/nfts", r.v1Handlers.NFT.ListNFTs)
    rg.GET("/nfts/:id", r.v1Handlers.NFT.GetNFT)
    
    // Protected routes
    protected := rg.Group("/")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/nfts", r.v1Handlers.NFT.CreateNFT)
        protected.PUT("/nfts/:id", r.v1Handlers.NFT.UpdateNFT)
        protected.DELETE("/nfts/:id", r.v1Handlers.NFT.DeleteNFT)
        protected.GET("/users/profile", r.v1Handlers.User.GetProfile)
        protected.PUT("/users/profile", r.v1Handlers.User.UpdateProfile)
    }
    
    // Auth routes
    rg.POST("/auth/register", r.v1Handlers.Auth.Register)
    rg.POST("/auth/login", r.v1Handlers.Auth.Login)
    rg.POST("/auth/logout", r.v1Handlers.Auth.Logout)
    rg.POST("/auth/refresh", r.v1Handlers.Auth.RefreshToken)
}

func (r *VersionedRouter) setupV2Routes(rg *gin.RouterGroup) {
    // Enhanced v2 routes with additional features
    rg.GET("/nfts", r.v2Handlers.NFT.ListNFTsV2)        // Enhanced filtering
    rg.GET("/nfts/:id", r.v2Handlers.NFT.GetNFTV2)        // Enhanced response
    rg.GET("/nfts/:id/history", r.v2Handlers.NFT.GetNFTHistory) // New endpoint
    
    // Protected routes
    protected := rg.Group("/")
    protected.Use(middleware.AuthMiddleware())
    protected.Use(middleware.RateLimitMiddleware(100, 60)) // Stricter rate limiting
    {
        protected.POST("/nfts", r.v2Handlers.NFT.CreateNFTV2) // Enhanced validation
        protected.PUT("/nfts/:id", r.v2Handlers.NFT.UpdateNFTV2)
        protected.DELETE("/nfts/:id", r.v2Handlers.NFT.DeleteNFTV2)
        protected.POST("/nfts/:id/transfer", r.v2Handlers.NFT.TransferNFT) // New endpoint
        
        // Batch operations
        protected.POST("/nfts/batch", r.v2Handlers.NFT.BatchCreateNFTs)
        protected.PUT("/nfts/batch", r.v2Handlers.NFT.BatchUpdateNFTs)
    }
}
```

---

## 📝 API DOCUMENTATION

### OpenAPI/Swagger Specification
```go
// pkg/docs/docs.go
package docs

import (
    "github.com/swaggo/swag"
)

const docTemplate = `{
    "schemes": ["http", "https"],
    "swagger": "2.0",
    "info": {
        "description": "AIOZ DApp RESTful API",
        "title": "AIOZ DApp API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "2.0"
    },
    "host": "localhost:8080",
    "basePath": "/api/v2",
    "paths": {},
    "definitions": {}
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = struct {
    Version     string
    Host        string
    BasePath    string
    Schemes     []string
    Title       string
    Description string
}{
    Version:     "2.0",
    Host:        "localhost:8080",
    BasePath:    "/api/v2",
    Schemes:     []string{},
    Title:       "AIOZ DApp API",
    Description: "AIOZ DApp RESTful API",
}

func init() {
    swag.Register(swag.Name, &SwaggerInfo)
}

func ReadDoc() string {
    return docTemplate
}
```

### Enhanced Handler with Swagger Annotations
```go
// internal/handler/nft_handler_v2.go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "your-project/internal/dto"
    "your-project/internal/service"
    "your-project/pkg/response"
    "your-project/pkg/validator"
)

// @title AIOZ DApp API
// @version 2.0
// @description AIOZ DApp RESTful API with comprehensive NFT marketplace functionality
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v2

type NFTHandlerV2 struct {
    nftService service.NFTService
    validator  *validator.CustomValidator
}

// ListNFTsV2 godoc
// @Summary List NFTs with enhanced filtering
// @Description Get a paginated list of NFTs with advanced filtering options including price range, categories, and search functionality.
// @Tags NFT
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Param sort query string false "Sort field" Enums(created_at, price, name) default(created_at)
// @Param order query string false "Sort order" Enums(asc, desc) default(desc)
// @Param owner_id query string false "Filter by owner ID"
// @Param category query string false "Filter by category"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Param status query string false "Filter by status" Enums(active, sold, pending) default(active)
// @Param search query string false "Search in name and description"
// @Success 200 {object} response.Response{data=dto.ListNFTsV2Response} "Successfully retrieved NFTs"
// @Failure 400 {object} response.Response{error=response.SwaggerError} "Bad request"
// @Failure 500 {object} response.Response{error=response.SwaggerError} "Internal server error"
// @Router /nfts [get]
func (h *NFTHandlerV2) ListNFTsV2(c *gin.Context) {
    // Parse and validate query parameters
    params, err := h.parseListParams(c)
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // Get NFTs with enhanced filtering
    nfts, total, err := h.nftService.ListNFTsV2(c.Request.Context(), params)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to fetch NFTs")
        return
    }
    
    // Build enhanced response
    response.Success(c, http.StatusOK, dto.ListNFTsV2Response{
        NFTs: nfts,
        Meta: response.Meta{
            Page:       params.Page,
            Limit:      params.Limit,
            Total:      total,
            TotalPages: (total + int64(params.Limit) - 1) / int64(params.Limit),
        },
        Filters: dto.FiltersApplied{
            Category: params.Category,
            PriceRange: dto.PriceRange{
                Min: params.MinPrice,
                Max: params.MaxPrice,
            },
            Status: params.Status,
            Search: params.Search,
        },
    })
}

// GetNFTV2 godoc
// @Summary Get NFT details with enhanced information
// @Description Retrieve detailed information about a specific NFT including ownership history, transaction records, and market analytics.
// @Tags NFT
// @Accept json
// @Produce json
// @Param id path string true "NFT ID"
// @Param include_history query bool false "Include transaction history" default(false)
// @Param include_analytics query bool false "Include market analytics" default(false)
// @Success 200 {object} response.Response{data=dto.NFTDetailV2Response} "Successfully retrieved NFT details"
// @Failure 400 {object} response.Response{error=response.SwaggerError} "Invalid NFT ID"
// @Failure 404 {object} response.Response{error=response.SwaggerError} "NFT not found"
// @Failure 500 {object} response.Response{error=response.SwaggerError} "Internal server error"
// @Router /nfts/{id} [get]
func (h *NFTHandlerV2) GetNFTV2(c *gin.Context) {
    id := c.Param("id")
    
    if !isValidUUID(id) {
        response.Error(c, http.StatusBadRequest, "Invalid NFT ID format")
        return
    }
    
    // Parse optional parameters
    includeHistory := c.DefaultQuery("include_history", "false") == "true"
    includeAnalytics := c.DefaultQuery("include_analytics", "false") == "true"
    
    // Get NFT with enhanced details
    nft, err := h.nftService.GetNFTV2(c.Request.Context(), id, dto.GetNFTOptions{
        IncludeHistory:   includeHistory,
        IncludeAnalytics: includeAnalytics,
    })
    
    if err != nil {
        if err == service.ErrNFTNotFound {
            response.Error(c, http.StatusNotFound, "NFT not found")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to fetch NFT")
        return
    }
    
    response.Success(c, http.StatusOK, nft)
}

// CreateNFTV2 godoc
// @Summary Create new NFT with enhanced validation
// @Description Create a new NFT with comprehensive validation, metadata standards, and automatic market analysis.
// @Tags NFT
// @Accept json
// @Produce json
// @Param nft body dto.CreateNFTV2Request true "NFT creation data"
// @Success 201 {object} response.Response{data=dto.NFTResponseV2} "Successfully created NFT"
// @Failure 400 {object} response.Response{error=response.SwaggerError} "Validation failed"
// @Failure 401 {object} response.Response{error=response.SwaggerError} "Unauthorized"
// @Failure 500 {object} response.Response{error=response.SwaggerError} "Internal server error"
// @Router /nfts [post]
// @Security ApiKeyAuth
func (h *NFTHandlerV2) CreateNFTV2(c *gin.Context) {
    var req dto.CreateNFTV2Request
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Enhanced validation
    if err := h.validator.ValidateStruct(req); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
        return
    }
    
    // Get user ID from context
    userID := c.GetUint("user_id")
    
    // Create NFT with enhanced features
    nft, err := h.nftService.CreateNFTV2(c.Request.Context(), userID, req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to create NFT")
        return
    }
    
    response.Success(c, http.StatusCreated, nft)
}

func (h *NFTHandlerV2) parseListParams(c *gin.Context) (*dto.ListNFTsV2Params, error) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }
    
    params := &dto.ListNFTsV2Params{
        Pagination: dto.Pagination{
            Page:     page,
            Limit:    limit,
            SortBy:   c.DefaultQuery("sort", "created_at"),
            SortDesc: c.DefaultQuery("order", "desc") == "desc",
        },
        Filters: dto.NFTFilters{
            OwnerID:     c.Query("owner_id"),
            Category:    c.Query("category"),
            MinPrice:    c.Query("min_price"),
            MaxPrice:    c.Query("max_price"),
            Status:      c.DefaultQuery("status", "active"),
            Search:      c.Query("search"),
        },
    }
    
    if err := h.validator.ValidateStruct(params); err != nil {
        return nil, err
    }
    
    return params, nil
}
```

---

## 🔒 API SECURITY

### Rate Limiting Implementation
```go
// internal/middleware/rate_limiter.go
package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     rate.Limit(rps),
        burst:    burst,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[key] = limiter
    }
    
    return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Use IP + endpoint as key for more granular rate limiting
        key := fmt.Sprintf("%s:%s", c.ClientIP(), c.FullPath())
        limiter := rl.getLimiter(key)
        
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "code":  "RATE_LIMIT_EXCEEDED",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// API Key authentication for external services
func APIKeyAuthMiddleware(validKeys map[string]string) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        
        if apiKey == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "API key required",
                "code":  "API_KEY_REQUIRED",
            })
            c.Abort()
            return
        }
        
        service, exists := validKeys[apiKey]
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid API key",
                "code":  "INVALID_API_KEY",
            })
            c.Abort()
            return
        }
        
        // Add service info to context
        c.Set("service", service)
        c.Next()
    }
}
```

---

## 📊 API MONITORING

### API Metrics Collection
```go
// pkg/metrics/api_metrics.go
package metrics

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // Request metrics
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_request_duration_seconds",
            Help:    "API request duration in seconds",
            Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "status_code"},
    )
    
    requestCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
    
    // Business metrics
    activeUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "api_active_users",
            Help: "Number of currently active users",
        },
    )
    
    nftOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_nft_operations_total",
            Help: "Total number of NFT operations",
        },
        []string{"operation", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(requestCount)
    prometheus.MustRegister(activeUsers)
    prometheus.MustRegister(nftOperations)
}

func APIMetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        statusCode := strconv.Itoa(c.Writer.Status())
        method := c.Request.Method
        endpoint := c.FullPath()
        
        // Record metrics
        requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
        requestCount.WithLabelValues(method, endpoint, statusCode).Inc()
        
        // Log slow requests
        if duration > 2.0 {
            gin.Logger.Warn("Slow API request",
                "method", method,
                "endpoint", endpoint,
                "duration", duration,
                "status", statusCode,
            )
        }
    }
}

func MetricsHandler() gin.HandlerFunc {
    return gin.WrapH(promhttp.Handler())
}

// Business metric helpers
func RecordNFTOperation(operation, status string) {
    nftOperations.WithLabelValues(operation, status).Inc()
}

func UpdateActiveUsers(count int) {
    activeUsers.Set(float64(count))
}
```

---

## 📋 API DEVELOPMENT CHECKLIST

### Design Phase Checklist
```
API Design:
[ ] RESTful principles are followed
[ ] Proper HTTP methods are used (GET, POST, PUT, DELETE)
[ ] Resource naming is consistent and intuitive
[ ] API versioning strategy is defined
[ ] Error handling strategy is designed
[ ] Authentication and authorization are planned
[ ] Rate limiting is considered
[ ] Data validation rules are defined
[ ] Pagination strategy is designed
[ ] Filtering and sorting options are planned

Documentation:
[ ] OpenAPI/Swagger specification is created
[ ] Endpoint documentation is comprehensive
[ ] Request/response examples are provided
[ ] Error codes are documented
[ ] Authentication methods are documented
[ ] Rate limits are documented
[ ] SDK/client library examples are provided
```

### Implementation Phase Checklist
```
Code Quality:
[ ] Input validation is implemented
[ ] Error handling is comprehensive
[ ] Logging is appropriate (not too verbose, not missing)
[ ] Code follows project conventions
[ ] Unit tests are written
[ ] Integration tests are written
[ ] API documentation is updated
[ ] Performance considerations are addressed
[ ] Security measures are implemented
[ ] Monitoring and metrics are added

Testing:
[ ] Happy path scenarios are tested
[ ] Error scenarios are tested
[ ] Edge cases are tested
[ ] Authentication/authorization is tested
[ ] Rate limiting is tested
[ ] Performance is tested
[ ] Documentation examples are verified
```

---

## DO / DON'T

✅ **DO**
- Follow RESTful principles consistently
- Use proper HTTP status codes
- Implement comprehensive input validation
- Provide clear, consistent error messages
- Document all endpoints thoroughly
- Use API versioning for breaking changes
- Implement proper authentication and authorization
- Monitor API performance and usage
- Use rate limiting to prevent abuse
- Provide pagination for large datasets

❌ **DON'T**
- NEVER use GET requests for state-changing operations
- NEVER return sensitive data in error messages
- NEVER skip input validation
- NEVER ignore security best practices
- NEVER create APIs without proper documentation
- NEVER use inconsistent naming conventions
- NEVER ignore performance implications
- NEVER expose internal implementation details
- NEVER hardcode configuration values
- NEVER skip proper error handling
