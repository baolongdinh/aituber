# Backend Golang Skill

## TRIGGER
Read this file before writing ANY Go code: handlers, services, repositories, models, middleware, DTOs.

---
name: golang-backend
description: >
  Production-grade Golang backend architecture skill with Gin, GORM, and PostgreSQL.
  Use this skill for writing, reviewing, refactoring, or optimizing ANY Go backend code — including
  handlers, services, repositories, models, DTOs, middleware, error handling, config, or database logic.
  Trigger on any Go file creation or edit, API endpoint design, database schema questions, or when the
  user asks to "clean up", "restructure", "optimize", or "add a feature" to a Go backend.
  FOR HIGH-PERFORMANCE WORK: Load `performance/SKILL.md` for audits, profiling, and deep optimization.
---

# 🦫 Golang Backend Architecture Skill

## IDENTITY
You are a Senior Golang Engineer specializing in production-grade backend development.
You apply SOLID principles, write clear, testable, and extensible code.
You NEVER write "just to get it done" code — every file, every function has clear purpose.

---

## PROJECT STACK
- **Language**: Go 1.22+
- **Framework**: Gin (HTTP router)
- **ORM**: GORM v2
- **Database**: PostgreSQL
- **Auth**: JWT (golang-jwt/jwt)
- **Config**: godotenv / viper
- **Validation**: go-playground/validator

---

## 📁 MANDATORY FOLDER STRUCTURE

```
backend/
├── 
│   main.go                  # App startup, wire dependencies
├── internal/                        # Non-exportable code
│   ├── handler/                     # Layer 1: HTTP — receive request, return response
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   └── handler.go               # Base handler struct if needed
│   ├── service/                     # Layer 2: Business Logic — handle domain logic
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   └── interfaces.go            # ALL Service interface definitions
│   ├── repository/                  # Layer 3: Data Access — only talks to DB
│   │   ├── user_repository.go
│   │   └── interfaces.go            # ALL Repository interface definitions
│   ├── model/                       # GORM DB models
│   │   ├── user.go
│   │   └── base_model.go            # Common fields: ID, CreatedAt, UpdatedAt, DeletedAt
│   ├── dto/                         # Data Transfer Objects (request/response shapes)
│   │   ├── auth_dto.go
│   │   └── user_dto.go
│   ├── middleware/                  # Gin middlewares
│   │   ├── auth_middleware.go
│   │   ├── cors_middleware.go
│   │   └── logger_middleware.go
│   └── router/
│       └── router.go                # Declare all routes
├── pkg/                             # Shared packages, reusable
│   ├── database/
│   │   └── postgres.go              # Init DB connection
│   ├── jwt/
│   │   └── jwt.go                   # JWT helpers
│   ├── response/
│   │   └── response.go              # Standard HTTP response helpers
│   ├── apperror/
│   │   └── apperror.go              # Custom error types
│   └── validator/
│       └── validator.go             # Input validation helpers
├── migrations/
│   ├── 001_create_users_table.sql
│   └── 002_create_nfts_table.sql
- `config/` — Load & parse config from env
- `performance/` — Performance optimization patterns & benchmarking

**Folder rules:**
- `internal/` — CANNOT be imported from packages outside this module
- `pkg/` — pure utilities, don't import from `internal/`
- `cmd/` — only contains `main.go`, wire all dependencies here

---

## 🏗️ N-LAYER ARCHITECTURE & SOLID

### Immutable Data Flow
```
HTTP Request
    ↓
Middleware (auth, cors, logging)
    ↓
Handler (parse & validate input → call Service → format response)
    ↓
Service (business logic → call Repository)
    ↓
Repository (SQL queries via GORM → return Model)
    ↓
Database (PostgreSQL)
```

### SOLID in Go Practice

**S — Single Responsibility**
```go
// ✅ RIGHT: Each struct does 1 thing
type UserRepository struct { db *gorm.DB }   // only handles DB
type UserService struct { repo UserRepo }     // only handles business
type UserHandler struct { svc UserService }   // only handles HTTP

// ❌ WRONG: Handler calls DB directly
func (h *UserHandler) GetUser(c *gin.Context) {
    h.db.Find(&user) // NO — violates SRP
}
```

**O — Open/Closed (use Interface in `interfaces.go`)**
```go
// internal/repository/interfaces.go
// Extend by implementing interface, don't modify existing code
type UserRepository interface {
    FindByID(ctx context.Context, id uint) (*model.User, error)
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    Create(ctx context.Context, user *model.User) error
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id uint) error
}
```

**D — Dependency Inversion**
```go
// Service depends on Interface, not concrete type
type UserService struct {
    repo repository.UserRepository  // ← Interface, not *UserRepositoryImpl
    jwt  jwt.Manager
}

// ✅ Constructor injection
func NewUserService(repo repository.UserRepository, jwt jwt.Manager) *UserService {
    return &UserService{repo: repo, jwt: jwt}
}
```

--- 

## ⚡ RAPID DEVELOPMENT (PRO-LEVEL) 

Use the enhanced automation script `skills/backend/scripts/go-scaffold.py` to quickly generate a new domain. This version is **Modular**—you can customize the base code by editing templates in `skills/backend/templates/`.

**Command:** 
```bash 
python3 skills/backend/scripts/go-scaffold.py --domain <name> 
``` 

**Database Migrations:**
Use the migration generator to create paired SQL files:
```bash
python3 skills/backend/scripts/migration-gen.py <migration_name>
```

**Common Patterns library:**
Check `skills/backend/patterns/` for production-ready logic you can drop in:
- `pagination.go`: Standard GORM pagination wrapper.
- `transaction.go`: Optional transaction wrapper.

**Benefits:** 
- **Modularity**: Templates are externalized for project-specific overrides.
- **Consistency**: Ensures 100% adherence to N-layer architecture. 
- **Efficiency**: Saves ~20 mins of boilerplate writing. 

---

## 📦 MANDATORY PATTERNS

### 1. Base Model with Enhanced Features
```go
// internal/model/base_model.go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    UUID      string         `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex" json:"uuid"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // soft delete
    CreatedBy uint           `json:"created_by,omitempty"`
    UpdatedBy uint           `json:"updated_by,omitempty"`
}

// Common methods for all models
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
    if b.UUID == "" {
        b.UUID = uuid.New().String()
    }
    return nil
}
```

### 2. Advanced GORM Model Patterns
```go
// internal/models/collection.go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Collection struct {
    Id            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"` 
    Name          string     `json:"name" gorm:"type:varchar(255);"` 
    Username      string     `json:"username" gorm:"type:varchar(255);index:idx_collection_username;priority:1"` 
    Description   string     `json:"description" gorm:"type:text;"` 
    Thumbnail     string     `json:"thumbnail" gorm:"type:varchar(300);"` 
    Votes         []Vote     `json:"votes" gorm:"foreignKey:CollectionId;references:Id;constraint:OnDelete:CASCADE"` 
    Datasets      []*Dataset `json:"datasets" gorm:"many2many:collection_datasets;"` 
    Models        []*Model   `json:"models" gorm:"many2many:collection_models;"` 
    UserId        uuid.UUID  `json:"user_id" gorm:"type:uuid;index:idx_collection_user_id;priority:2"` 
    AuthorInfo    Info       `gorm:"foreignKey:BelongsToId;references:UserId;constraint:OnDelete:CASCADE"` 
    IsVotedByUser bool       `json:"is_voted_by_user" gorm:"is_voted_by_user"` 
    VotesCount    int64      `json:"votes_count" gorm:"votes_count"` 
    ModelsCount   int64      `json:"models_count" gorm:"models_count"` 
    DatasetsCount int64      `json:"datasets_count" gorm:"datasets_count"` 
    Visibility    string     `json:"visibility" gorm:"type:varchar(50);not null;default:'public'"` 
    Violated      bool       `json:"violated" gorm:"type:boolean;"` 
    CreatedAt     time.Time  `json:"created_at" gorm:"type:timestamp;"` 
    UpdatedAt     time.Time  `json:"updated_at" gorm:"type:timestamp;"` 
}

// BeforeCreate hook to generate UUID
func (c *Collection) BeforeCreate(tx *gorm.DB) error {
    if c.Id == uuid.Nil {
        c.Id = uuid.New()
    }
    return nil
}

// TableName specifies the table name for Collection model
func (Collection) TableName() string {
    return "collections"
}

// Related Models
type Vote struct {
    Id           uuid.UUID `gorm:"type:uuid;primary_key"`
    CollectionId uuid.UUID `gorm:"type:uuid;index"`
    UserId       uuid.UUID `gorm:"type:uuid;index"`
    VoteType     string    `gorm:"type:varchar(10)"` // up, down
    CreatedAt    time.Time `gorm:"type:timestamp"`
}

type Dataset struct {
    Id          uuid.UUID `gorm:"type:uuid;primary_key"`
    Name        string    `gorm:"type:varchar(255)"`
    Description string    `gorm:"type:text"`
    CreatedAt   time.Time `gorm:"type:timestamp"`
}

type Model struct {
    Id          uuid.UUID `gorm:"type:uuid;primary_key"`
    Name        string    `gorm:"type:varchar(255)"`
    Description string    `gorm:"type:text"`
    CreatedAt   time.Time `gorm:"type:timestamp"`
}

type Info struct {
    BelongsToId uuid.UUID `gorm:"type:uuid;primary_key"`
    DisplayName string    `gorm:"type:varchar(255)"`
    Avatar      string    `gorm:"type:varchar(500)"`
    Bio         string    `gorm:"type:text"`
}
```

### 3. Database Connection with Enhanced GORM Setup
```go
// pkg/db/connection.go
package db

import (
    "fmt"
    "log"
    "os"
    "time"

    "your-project/initializers"
    "your-project/internal/models"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func ConnectDB(config *initializers.Config) *gorm.DB {
    var err error
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", 
        config.DBHost, config.DBUserName, config.DBUserPassword, config.DBName, config.DBPort)

    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
        logger.Config{
            SlowThreshold:             200 * time.Millisecond,
            LogLevel:                  logger.Warn,
            IgnoreRecordNotFoundError: true,  // Ignore ErrRecordNotFound error for logger
            ParameterizedQueries:      false, // Don't include params in the SQL log
            Colorful:                  true,  // Enable color
        },
    )
    
    DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })
    
    if err != nil {
        log.Fatal("Failed to connect to the Database")
    }

    // Auto-migrate all models
    if err := DB.AutoMigrate(
        &models.Discussion{},
        &models.Comment{},
        &models.Like{},
        &models.Download{},
        &models.Collection{},
        &models.Vote{},
        &models.Reaction{},
        &models.Notification{},
        &models.Metadata{},
        &models.ContentConfig{},
        &models.ReportedComment{},
        &models.ReportedDiscussion{},
        &models.ReportedCollection{},
    ); err != nil {
        panic(err)
    }

    log.Println("🚀 Connected successfully to the Database 🚀")
    return DB
}

// GetDBInstance returns singleton DB instance
var DBInstance *gorm.DB

func InitDB(config *initializers.Config) {
    DBInstance = ConnectDB(config)
}

func GetDB() *gorm.DB {
    return DBInstance
}
```

### 2. Enhanced Standard Response
```go
// pkg/response/response.go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
    RequestID string    `json:"request_id,omitempty"`
}

type ErrorInfo struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type Meta struct {
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
    HasNext    bool  `json:"has_next"`
    HasPrev    bool  `json:"has_prev"`
}

// Enhanced response helpers
func OK(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Success:   true,
        Data:      data,
        RequestID: c.GetString("request_id"),
    })
}

func Fail(c *gin.Context, status int, err *apperror.AppError) {
    c.JSON(status, Response{
        Success:   false,
        Error:     &ErrorInfo{Code: err.Code, Message: err.Message, Details: err.Details},
        RequestID: c.GetString("request_id"),
    })
}

func Paginated(c *gin.Context, data interface{}, meta Meta) {
    c.JSON(http.StatusOK, Response{
        Success:   true,
        Data:      data,
        Meta:      &meta,
        RequestID: c.GetString("request_id"),
    })
}
```

### 3. Enhanced AppError with Error Codes
```go
// pkg/apperror/apperror.go
type AppError struct {
    Code    string `json:"code"`           // Error code for frontend handling
    Message string `json:"message"`        // User-facing message
    Details string `json:"details,omitempty"` // Additional context
    Err     error  `json:"-"`               // Internal error for logging
    HTTPStatus int `json:"-"`              // HTTP status code
}

func (e *AppError) Error() string { return e.Message }

// Error codes constants
const (
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeUnauthorized     = "UNAUTHORIZED"
    ErrCodeForbidden        = "FORBIDDEN"
    ErrCodeBadRequest       = "BAD_REQUEST"
    ErrCodeConflict         = "CONFLICT"
    ErrCodeInternal         = "INTERNAL_ERROR"
    ErrCodeValidationFailed = "VALIDATION_FAILED"
    ErrCodeRateLimited      = "RATE_LIMITED"
    ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Enhanced constructors
func NotFound(msg string) *AppError {
    return &AppError{Code: ErrCodeNotFound, Message: msg, HTTPStatus: http.StatusNotFound}
}

func Unauthorized(msg string) *AppError {
    return &AppError{Code: ErrCodeUnauthorized, Message: msg, HTTPStatus: http.StatusUnauthorized}
}

func BadRequest(msg string) *AppError {
    return &AppError{Code: ErrCodeBadRequest, Message: msg, HTTPStatus: http.StatusBadRequest}
}

func Conflict(msg string) *AppError {
    return &AppError{Code: ErrCodeConflict, Message: msg, HTTPStatus: http.StatusConflict}
}

func ValidationFailed(msg string, details string) *AppError {
    return &AppError{Code: ErrCodeValidationFailed, Message: msg, Details: details, HTTPStatus: http.StatusBadRequest}
}

func Internal(err error, msg string) *AppError {
    return &AppError{Code: ErrCodeInternal, Message: msg, Err: err, HTTPStatus: http.StatusInternalServerError}
}

func RateLimited(msg string) *AppError {
    return &AppError{Code: ErrCodeRateLimited, Message: msg, HTTPStatus: http.StatusTooManyRequests}
}
```

### 4. Enhanced Handler Pattern with Validation & Tracing
```go
// internal/handler/user_handler.go
type UserHandler struct {
    userSvc service.UserService
    validator *validator.Validate
    logger    *logrus.Logger
}

func NewUserHandler(userSvc service.UserService, validator *validator.Validate, logger *logrus.Logger) *UserHandler {
    return &UserHandler{
        userSvc:   userSvc,
        validator: validator,
        logger:    logger,
    }
}

func (h *UserHandler) GetUser(c *gin.Context) {
    // Extract and validate input
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid user id"))
        return
    }

    // Call service
    user, appErr := h.userSvc.GetByID(c.Request.Context(), uint(id))
    if appErr != nil {
        h.logger.WithFields(logrus.Fields{
            "user_id": id,
            "error":  appErr.Err,
        }).Error("Failed to get user")
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    response.OK(c, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req dto.CreateUserRequest
    
    // Bind and validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid request body"))
        return
    }
    
    if err := h.validator.Struct(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.ValidationFailed("validation failed", err.Error()))
        return
    }

    // Call service
    user, appErr := h.userSvc.Create(c.Request.Context(), req)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    // Return 201 for created resource
    c.JSON(http.StatusCreated, response.Success{
        Success:   true,
        Data:      user,
        RequestID: c.GetString("request_id"),
    })
}
```

### 5. Enhanced Service Pattern with Context & Caching
```go
// internal/service/user_service.go
type UserService interface {
    GetByID(ctx context.Context, id uint) (*dto.UserResponse, *apperror.AppError)
    Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, *apperror.AppError)
    Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*dto.UserResponse, *apperror.AppError)
    Delete(ctx context.Context, id uint) *apperror.AppError
    List(ctx context.Context, params dto.ListUsersRequest) (*dto.ListUsersResponse, *apperror.AppError)
}

type userServiceImpl struct {
    repo      repository.UserRepository
    cache     cache.Cache  // Redis cache interface
    logger    *logrus.Logger
    metrics   *metrics.Metrics
}

func NewUserService(repo repository.UserRepository, cache cache.Cache, logger *logrus.Logger, metrics *metrics.Metrics) UserService {
    return &userServiceImpl{
        repo:    repo,
        cache:   cache,
        logger:  logger,
        metrics: metrics,
    }
}

func (s *userServiceImpl) GetByID(ctx context.Context, id uint) (*dto.UserResponse, *apperror.AppError) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%d", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
        var user dto.UserResponse
        if json.Unmarshal([]byte(cached), &user) == nil {
            s.metrics.Increment("user.cache_hit")
            return &user, nil
        }
    }

    // Get from repository
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NotFound("user not found")
        }
        s.logger.WithError(err).WithField("user_id", id).Error("Failed to get user from repository")
        return nil, apperror.Internal(err, "failed to get user")
    }

    // Convert to DTO
    userResp := dto.ToUserResponse(user)
    
    // Cache the result
    if data, err := json.Marshal(userResp); err == nil {
        s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
    }
    
    s.metrics.Increment("user.cache_miss")
    return userResp, nil
}

func (s *userServiceImpl) Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, *apperror.AppError) {
    // Check if user already exists
    existing, _ := s.repo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return nil, apperror.Conflict("user with this email already exists")
    }

    // Create user
    user := &model.User{
        Email:    req.Email,
        Username: req.Username,
        Password: s.hashPassword(req.Password), // Hash password
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        s.logger.WithError(err).Error("Failed to create user")
        return nil, apperror.Internal(err, "failed to create user")
    }

    // Invalidate cache
    s.cache.DeletePattern(ctx, "user:list:*")
    
    return dto.ToUserResponse(user), nil
}

func (s *userServiceImpl) hashPassword(password string) string {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        s.logger.WithError(err).Error("Failed to hash password")
        return ""
    }
    return string(hash)
}

func (s *userServiceImpl) Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*dto.UserResponse, *apperror.AppError) {
    // Get user from repository
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NotFound("user not found")
    }
    
    // Count total records
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Apply sorting
    if params.SortBy != "" {
        direction := "asc"
        if params.SortDir == "desc" {
            direction = "desc"
        }
        query = query.Order(fmt.Sprintf("%s %s", params.SortBy, direction))
    }
    
    // Apply pagination
    offset := (params.Page - 1) * params.Limit
    err := query.Offset(offset).Limit(params.Limit).Find(&users).Error
    
    return users, total, err
}

func (r *userRepositoryImpl) SoftDelete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

### 7. Enhanced DTO Pattern with Validation & Transformation
```go
// internal/dto/user_dto.go

// Request DTOs — validate input
type CreateUserRequest struct {
    Email    string `json:"email"    validate:"required,email,max=255"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    FirstName string `json:"first_name" validate:"required,max=100"`
    LastName  string `json:"last_name"  validate:"required,max=100"`
}

// Update DTOs — partial updates allowed
type UpdateUserRequest struct {
    Email     *string `json:"email,omitempty"     validate:"omitempty,email,max=255"`
    Username  *string `json:"username,omitempty"  validate:"omitempty,min=3,max=50,alphanum"`
    FirstName *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
    LastName  *string `json:"last_name,omitempty"  validate:"omitempty,max=100"`
}

// List DTOs — pagination and filtering
type ListUsersRequest struct {
    Page     int    `json:"page"     form:"page"     validate:"min=1"`
    Limit    int    `json:"limit"    form:"limit"    validate:"min=1,max=100"`
    Search   string `json:"search"   form:"search"   validate:"max=100"`
    SortBy   string `json:"sort_by"  form:"sort_by"  validate:"oneof=created_at updated_at username email"`
    SortDir  string `json:"sort_dir" form:"sort_dir"  validate:"oneof=asc desc"`
}

// Response DTOs — NEVER return Model directly
type UserResponse struct {
    ID        uint      `json:"id"`
    UUID      string    `json:"uuid"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    Avatar    string    `json:"avatar,omitempty"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    // NO Password field
}

type ListUsersResponse struct {
    Users []UserResponse `json:"users"`
    Meta  response.Meta  `json:"meta"`
}

// Enhanced mapper functions
func ToUserResponse(u *model.User) *UserResponse {
    return &UserResponse{
        ID:        u.ID,
        UUID:      u.UUID,
        Email:     u.Email,
        Username:  u.Username,
        FirstName: u.FirstName,
        LastName:  u.LastName,
        Avatar:    u.Avatar,
        Status:    u.Status,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}

func ToListUsersResponse(users []*model.User, total int64, params ListUsersRequest) *ListUsersResponse {
    userResponses := make([]UserResponse, len(users))
    for i, user := range users {
        userResponses[i] = *ToUserResponse(user)
    }
    
    totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))
    
    return &ListUsersResponse{
        Users: userResponses,
        Meta: response.Meta{
            Page:       params.Page,
            Limit:      params.Limit,
            Total:      total,
            TotalPages: totalPages,
            HasNext:    params.Page < totalPages,
            HasPrev:    params.Page > 1,
        },
    }
}

### 8. Enhanced Router Setup with Middleware & Rate Limiting
```go
// internal/router/router.go
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    nftHandler *handler.NFTHandler,
    authMiddleware *middleware.AuthMiddleware,
    rateLimitMiddleware *middleware.RateLimitMiddleware,
    corsMiddleware *middleware.CorsMiddleware,
    loggerMiddleware *middleware.LoggerMiddleware,
) *gin.Engine {
    r := gin.New()
    
    // Global middleware
    r.Use(gin.Recovery())
    r.Use(corsMiddleware.Handle())
    r.Use(loggerMiddleware.Handle())
    r.Use(middleware.RequestID()) // Add request ID for tracing
    
    // Health check endpoint (no auth required)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
    })
    
    // API versioning
    v1 := r.Group("/api/v1")
    {
        // Apply rate limiting to all API endpoints
        v1.Use(rateLimitMiddleware.Handle())
        
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
            auth.POST("/forgot-password", authHandler.ForgotPassword)
            auth.POST("/reset-password", authHandler.ResetPassword)
        }
        
        // Public NFT endpoints
        nfts := v1.Group("/nfts")
        {
            nfts.GET("", nftHandler.ListNFTs)        // Public list
            nfts.GET("/:id", nftHandler.GetNFT)       // Public details
            nfts.GET("/:id/owners", nftHandler.GetNFTOwners) // Public ownership
        }
        
        // Protected routes (require authentication)
        protected := v1.Group("/")
        protected.Use(authMiddleware.Authenticate())
        {
            // User management
            users := protected.Group("/users")
            {
                users.GET("/me", userHandler.GetMe)
                users.PUT("/me", userHandler.UpdateMe)
                users.POST("/me/avatar", userHandler.UploadAvatar)
                users.DELETE("/me", userHandler.DeleteMe)
            }
            
            // NFT management (authenticated)
            myNFTs := protected.Group("/my-nfts")
            {
                myNFTs.GET("", nftHandler.GetMyNFTs)
                myNFTs.POST("/:id/transfer", nftHandler.TransferNFT)
                myNFTs.POST("/:id/list", nftHandler.ListNFTForSale)
            }
        }
        
        // Admin routes (require admin role)
        admin := v1.Group("/admin")
        admin.Use(authMiddleware.Authenticate())
        admin.Use(authMiddleware.RequireRole("admin"))
        {
            admin.GET("/users", userHandler.ListUsers)
            admin.PUT("/users/:id/status", userHandler.UpdateUserStatus)
            admin.GET("/analytics", userHandler.GetAnalytics)
        }
    }
    
    return r
}
```

### 9. Enhanced Dependency Injection with Configuration
### 10. Enhanced Swagger/OpenAPI Documentation

```go
// pkg/docs/docs.go — Swagger configuration
package docs

import (
	"github.com/swaggo/swag"
)

const DocTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
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
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {},
    "definitions": {},
    "securityDefinitions": {
        "Bearer": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "DApp API",
	Description:      "Production-grade DApp API with comprehensive documentation",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  DocTemplate,
}

type s struct{}

func (s *s) ReadDoc() string {
	t, err := swag.ReadDoc(SwaggerInfo)
	if err != nil {
		return ""
	}
	return t
}

func init() {
	swag.Register(swag.Name, &s{})
}
```

```go
// pkg/swagger/swagger.go — Use existing DTO models directly
package swagger

import (
    "time"
    "your-project/internal/dto"
    "your-project/pkg/response"
)

// Use existing DTO models for Swagger documentation
type SwaggerError struct {
    Code    string `json:"code" example:"NOT_FOUND"`
    Message string `json:"message" example:"Resource not found"`
    Details string `json:"details,omitempty" example:"User with ID 123 not found"`
}

type SwaggerResponse struct {
    Success   bool        `json:"success" example:"true"`
    Data      interface{} `json:"data,omitempty"`
    Error     *SwaggerError `json:"error,omitempty"`
    Meta      *response.Meta `json:"meta,omitempty"`
    RequestID string      `json:"request_id" example:"req_1640995200000_abc123def"`
}

// Reuse existing DTO models - no duplication needed
// Use dto.CreateUserRequest directly in Swagger annotations
// Use dto.UpdateUserRequest directly in Swagger annotations
// Use dto.ListUsersRequest directly in Swagger annotations
// Use dto.UserResponse directly in Swagger annotations

// Example usage in annotations:
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Success 200 {object} response.Response{data=[]dto.UserResponse,meta=response.Meta}
```

```go
// internal/handler/user_handler.go — Enhanced with Swagger annotations using existing DTOs
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/swaggo/swag"
    _ "your-project/docs" // Import docs
    "your-project/internal/dto"
    "your-project/pkg/response"
)

// @title DApp API
// @version 1.0
// @description Production-grade DApp API with comprehensive documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with email, username, and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 409 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /users [post]
// @Security Bearer
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req dto.CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid request body"))
        return
    }
    
    if err := h.validator.Struct(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.ValidationFailed("validation failed", err.Error()))
        return
    }

    user, appErr := h.userSvc.Create(c.Request.Context(), req)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    c.JSON(http.StatusCreated, response.Response{
        Success:   true,
        Data:      user,
        RequestID: c.GetString("request_id"),
    })
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve user information by their unique ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /users/{id} [get]
// @Security Bearer
func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid user id"))
        return
    }

    user, appErr := h.userSvc.GetByID(c.Request.Context(), uint(id))
    if appErr != nil {
        h.logger.WithFields(logrus.Fields{
            "user_id": id,
            "error":  appErr.Err,
        }).Error("Failed to get user")
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    response.OK(c, user)
}

// ListUsers godoc
// @Summary List users with pagination
// @Description Retrieve a paginated list of users with optional filtering and sorting
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param search query string false "Search term for username or email"
// @Param sort_by query string false "Sort field" Enums(created_at,updated_at,username,email) default(created_at)
// @Param sort_dir query string false "Sort direction" Enums(asc,desc) default(desc)
// @Success 200 {object} response.Response{data=[]dto.UserResponse,meta=response.Meta}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /users [get]
// @Security Bearer
func (h *UserHandler) ListUsers(c *gin.Context) {
    var req dto.ListUsersRequest
    
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid query parameters"))
        return
    }
    
    if err := h.validator.Struct(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.ValidationFailed("validation failed", err.Error()))
        return
    }

    // Set defaults
    if req.Page == 0 {
        req.Page = 1
    }
    if req.Limit == 0 {
        req.Limit = 20
    }

    users, appErr := h.userSvc.List(c.Request.Context(), req)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    response.Paginated(c, users.Users, users.Meta)
}

// UpdateUser godoc
// @Summary Update user information
// @Description Update user profile information (partial update supported)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest false "User update data"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /users/{id} [put]
// @Security Bearer
func (h *UserHandler) UpdateUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid user id"))
        return
    }

    var req dto.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid request body"))
        return
    }

    user, appErr := h.userSvc.Update(c.Request.Context(), uint(id), req)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    response.OK(c, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete user account (sets deleted_at timestamp)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /users/{id} [delete]
// @Security Bearer
func (h *UserHandler) DeleteUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid user id"))
        return
    }

    appErr := h.userSvc.Delete(c.Request.Context(), uint(id))
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }

    c.Status(http.StatusNoContent)
}
```
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize logger
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    if cfg.Environment == "development" {
        logger.SetFormatter(&logrus.TextFormatter{})
    } else {
        logger.SetFormatter(&logrus.JSONFormatter{})
    }
    
    // Initialize database
    db, err := database.NewPostgres(cfg.Database.URL, logger)
    if err != nil {
        logger.WithError(err).Fatal("Failed to connect to database")
    }
    
    // Initialize Redis cache
    cache, err := redis.NewCache(cfg.Redis.URL, logger)
    if err != nil {
        logger.WithError(err).Fatal("Failed to connect to Redis")
    }
    
    // Initialize metrics
    metrics := metrics.New(cfg.Metrics.Enabled)
    
    // Initialize validator
    validator := validator.New()
    
    // Initialize repositories
    userRepo := repository.NewUserRepository(db, logger)
    nftRepo := repository.NewNFTRepository(db, logger)
    
    // Initialize JWT manager
    jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration)
    
    // Initialize services
    userSvc := service.NewUserService(userRepo, cache, logger, metrics)
    authSvc := service.NewAuthService(userRepo, jwtManager, cache, logger)
    nftSvc := service.NewNFTService(nftRepo, cache, logger, metrics)
    
    // Initialize handlers
    userHandler := handler.NewUserHandler(userSvc, validator, logger)
    authHandler := handler.NewAuthHandler(authSvc, validator, logger)
    nftHandler := handler.NewNFTHandler(nftSvc, validator, logger)
    
    // Initialize middleware
    authMw := middleware.NewAuthMiddleware(jwtManager, userRepo)
    rateLimitMw := middleware.NewRateLimitMiddleware(cache, cfg.RateLimit)
    corsMw := middleware.NewCorsMiddleware(cfg.CORS)
    loggerMw := middleware.NewLoggerMiddleware(logger)
    
    // Setup router
    router := router.SetupRouter(
        authHandler, userHandler, nftHandler,
        authMw, rateLimitMw, corsMw, loggerMw,
    )
    
    // Start server with graceful shutdown
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:   60 * time.Second,
    }
    
    // Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.WithError(err).Fatal("Failed to start server")
        }
    }()
    
    logger.WithField("port", cfg.Server.Port).Info("Server started")
    
    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        logger.WithError(err).Error("Server forced to shutdown")
    }
    
    logger.Info("Server exited")
}

// internal/router/router.go — Enhanced with Swagger
func SetupRouter(
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    nftHandler *handler.NFTHandler,
    authMiddleware *middleware.AuthMiddleware,
    rateLimitMiddleware *middleware.RateLimitMiddleware,
    corsMiddleware *middleware.CorsMiddleware,
    loggerMiddleware *middleware.LoggerMiddleware,
) *gin.Engine {
    r := gin.New()
    
    // Global middleware
    r.Use(gin.Recovery())
    r.Use(corsMiddleware.Handle())
    r.Use(loggerMiddleware.Handle())
    r.Use(middleware.RequestID())
    
    // Swagger documentation
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.GET("/swagger.json", func(c *gin.Context) {
        c.Data(http.StatusOK, "application/json", []byte(docs.SwaggerInfo.ReadDoc()))
    })
    
    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok", 
            "timestamp": time.Now(),
            "version": docs.SwaggerInfo.Version,
        })
    })
    
    // API versioning
    v1 := r.Group("/api/v1")
    {
        v1.Use(rateLimitMiddleware.Handle())
        
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
            auth.POST("/forgot-password", authHandler.ForgotPassword)
            auth.POST("/reset-password", authHandler.ResetPassword)
        }
        
        // Public NFT endpoints
        nfts := v1.Group("/nfts")
        {
            nfts.GET("", nftHandler.ListNFTs)
            nfts.GET("/:id", nftHandler.GetNFT)
            nfts.GET("/:id/owners", nftHandler.GetNFTOwners)
        }
        
        // Protected routes
        protected := v1.Group("/")
        protected.Use(authMiddleware.Authenticate())
        {
            // User management
            users := protected.Group("/users")
            {
                users.GET("/me", userHandler.GetMe)
                users.PUT("/me", userHandler.UpdateMe)
                users.POST("/me/avatar", userHandler.UploadAvatar)
                users.DELETE("/me", userHandler.DeleteMe)
            }
            
            // NFT management
            myNFTs := protected.Group("/my-nfts")
            {
                myNFTs.GET("", nftHandler.GetMyNFTs)
                myNFTs.POST("/:id/transfer", nftHandler.TransferNFT)
                myNFTs.POST("/:id/list", nftHandler.ListNFTForSale)
            }
        }
        
        // Admin routes
        admin := v1.Group("/admin")
        admin.Use(authMiddleware.Authenticate())
        admin.Use(authMiddleware.RequireRole("admin"))
        {
            admin.GET("/users", userHandler.ListUsers)
            admin.GET("/users/:id", userHandler.GetUser)
            admin.PUT("/users/:id", userHandler.UpdateUser)
            admin.DELETE("/users/:id", userHandler.DeleteUser)
            admin.PUT("/users/:id/status", userHandler.UpdateUserStatus)
            admin.GET("/analytics", userHandler.GetAnalytics)
        }
    }
    
    return r
}

// cmd/server/main.go — Enhanced with Swagger generation
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "github.com/swaggo/swag"
    _ "your-project/docs" // Import docs
)

// @title DApp API Documentation
// @version 1.0
// @description Production-grade DApp API with comprehensive Swagger documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.swagger.io/support
// @contact.email support@your-dapp.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Bearer token for authentication

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize logger
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    if cfg.Environment == "development" {
        logger.SetFormatter(&logrus.TextFormatter{})
    } else {
        logger.SetFormatter(&logrus.JSONFormatter{})
    }
    
    // Initialize database
    db, err := database.NewPostgres(cfg.Database.URL, logger)
    if err != nil {
        logger.WithError(err).Fatal("Failed to connect to database")
    }
    
    // Initialize Redis cache
    cache, err := redis.NewCache(cfg.Redis.URL, logger)
    if err != nil {
        logger.WithError(err).Fatal("Failed to connect to Redis")
    }
    
    // Initialize metrics
    metrics := metrics.New(cfg.Metrics.Enabled)
    
    // Initialize validator
    validator := validator.New()
    
    // Initialize repositories
    userRepo := repository.NewUserRepository(db, logger)
    nftRepo := repository.NewNFTRepository(db, logger)
    
    // Initialize JWT manager
    jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration)
    
    // Initialize services
    userSvc := service.NewUserService(userRepo, cache, logger, metrics)
    authSvc := service.NewAuthService(userRepo, jwtManager, cache, logger)
    nftSvc := service.NewNFTService(nftRepo, cache, logger, metrics)
    
    // Initialize handlers
    userHandler := handler.NewUserHandler(userSvc, validator, logger)
    authHandler := handler.NewAuthHandler(authSvc, validator, logger)
    nftHandler := handler.NewNFTHandler(nftSvc, validator, logger)
    
    // Initialize middleware
    authMw := middleware.NewAuthMiddleware(jwtManager, userRepo)
    rateLimitMw := middleware.NewRateLimitMiddleware(cache, cfg.RateLimit)
    corsMw := middleware.NewCorsMiddleware(cfg.CORS)
    loggerMw := middleware.NewLoggerMiddleware(logger)
    
    // Setup router
    router := router.SetupRouter(
        authHandler, userHandler, nftHandler,
        authMw, rateLimitMw, corsMw, loggerMw,
    )
    
    // Generate Swagger docs in development
    if cfg.Environment == "development" {
        go func() {
            for {
                // Watch for changes and regenerate docs
                time.Sleep(5 * time.Second)
                if _, err := swag.Init(swag.SetMarkdownFileDirectory("./docs")); err != nil {
                    logger.WithError(err).Error("Failed to generate Swagger docs")
                }
            }
        }()
    }
    
    // Start server with graceful shutdown
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:   60 * time.Second,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.WithError(err).Fatal("Failed to start server")
        }
    }()
    
    logger.WithFields(logrus.Fields{
        "port":    cfg.Server.Port,
        "docs_url": "http://localhost:" + cfg.Server.Port + "/swagger/index.html",
    }).Info("Server started with Swagger documentation")
    
    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        logger.WithError(err).Error("Server forced to shutdown")
    }
    
    logger.Info("Server exited")
}

// Makefile — Swagger generation commands
.PHONY: swag-init swag-gen swag-clean

# Initialize Swagger
swag-init:
	@echo "Initializing Swagger..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Generate Swagger docs
swag-gen:
	@echo "Generating Swagger documentation..."
	swag gen -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Clean Swagger docs
swag-clean:
	@echo "Cleaning Swagger docs..."
	rm -rf docs/

# Watch and regenerate (development)
swag-watch:
	@echo "Watching for changes and regenerating Swagger..."
	watch -n 2 'make swag-gen'

# Run with Swagger
development: swag-gen
	@echo "Starting development server with Swagger..."
	go run cmd/server/main.go

// .github/workflows/swagger.yml — CI/CD for Swagger
name: Generate Swagger Documentation

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  swagger:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.22
    
    - name: Install swag
      run: go install github.com/swaggo/swag/cmd/swag@latest
    
    - name: Generate Swagger docs
      run: make swag-gen
    
    - name: Commit and push docs
      run: |
        git config --local user.email "action@github.com"
        git config --local user.name "GitHub Action"
        git add docs/
        git diff --staged --quiet || git commit -m "docs: update swagger documentation"
        git push

// internal/docs/swagger/docs.go — Swagger documentation
package docs

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/swaggo/swag"
)

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Swagger{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "DApp API Documentation",
	Description:      "Production-grade DApp API with comprehensive Swagger documentation",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	fmt.Println(strings.Repeat("-", 80))
	log.Printf("server listening on port 8080")
	fmt.Println(strings.Repeat("-", 80))
}

// @title DApp API Documentation
// @version 1.0
// @description Production-grade DApp API with comprehensive Swagger documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.swagger.io/support
// @contact.email support@your-dapp.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Bearer token for authentication

func docTemplate(s *swag.Swagger) string {
	s.Info.Description = fmt.Sprintf("%s\n\n%v", s.Info.Description, s.Info.Description)
	return s.Info.Description
}

func main() {
	http.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
	))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// internal/docs/swagger/swagger.go — Swagger handler
package swagger

import (
	"github.com/swaggo/swag"
	"net/http"
)

func Handler(doc string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "doc.json") {
			http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
			return
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(swag.ReadDoc()); err != nil {
			http.Error(w, "Failed to read swagger doc", http.StatusInternalServerError)
			return
		}

		w.Write(buf.Bytes())
	}
}
