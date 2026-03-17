# Backend Go Patterns

## TRIGGER
Đọc file này khi viết Go code phức tạp hơn: middleware, pagination, transactions, Go modules setup.

---

## GO MODULE SETUP

```bash
go mod init github.com/<org>/<project>-backend

# Core dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get github.com/joho/godotenv
go get golang.org/x/crypto
```

---

## CONFIG PATTERN

```go
// config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Port string
    Env  string // "development" | "production"
}

type DatabaseConfig struct {
    URL          string
    MaxOpenConns int
    MaxIdleConns int
}

type JWTConfig struct {
    Secret     string
    ExpireHours int
}

func Load() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        // OK in production (uses system env)
    }
    return &Config{
        Server: ServerConfig{
            Port: getEnv("PORT", "8080"),
            Env:  getEnv("APP_ENV", "development"),
        },
        Database: DatabaseConfig{
            URL:          mustGetEnv("DATABASE_URL"),
            MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
        },
        JWT: JWTConfig{
            Secret:      mustGetEnv("JWT_SECRET"),
            ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
        },
    }, nil
}
```

---

## PAGINATION PATTERN

```go
// dto/pagination_dto.go
type PaginationQuery struct {
    Page  int `form:"page"  binding:"min=1"`
    Limit int `form:"limit" binding:"min=1,max=100"`
}

func (p *PaginationQuery) SetDefaults() {
    if p.Page == 0  { p.Page = 1 }
    if p.Limit == 0 { p.Limit = 20 }
}

func (p *PaginationQuery) Offset() int {
    return (p.Page - 1) * p.Limit
}

// Repository: paginated query
func (r *nftRepository) ListByOwner(ctx context.Context, ownerID uint, q *dto.PaginationQuery) ([]*model.NFT, int64, error) {
    var nfts []*model.NFT
    var total int64

    query := r.db.WithContext(ctx).Model(&model.NFT{}).Where("owner_id = ?", ownerID)
    query.Count(&total)
    err := query.Offset(q.Offset()).Limit(q.Limit).Order("created_at DESC").Find(&nfts).Error
    return nfts, total, err
}
```

---

## AUTH MIDDLEWARE PATTERN

```go
// internal/middleware/auth_middleware.go
func AuthRequired(jwtService jwt.JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            response.Unauthorized(c, "missing or invalid authorization header")
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwtService.Verify(token)
        if err != nil {
            response.Unauthorized(c, "invalid or expired token")
            c.Abort()
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("walletAddress", claims.WalletAddress)
        c.Next()
    }
}

// Helpers để lấy claims trong handler
func GetUserID(c *gin.Context) uint {
    return c.MustGet("userID").(uint)
}
```

---

## ROUTER PATTERN

```go
// internal/router/router.go
func Setup(
    r *gin.Engine,
    authMiddleware gin.HandlerFunc,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    nftHandler *handler.NFTHandler,
) {
    api := r.Group("/api/v1")
    {
        // Public routes
        auth := api.Group("/auth")
        {
            auth.POST("/login",    authHandler.Login)
            auth.POST("/register", authHandler.Register)
        }

        // Protected routes
        protected := api.Group("/")
        protected.Use(authMiddleware)
        {
            users := protected.Group("/users")
            {
                users.GET("/me",      userHandler.GetMe)
                users.PUT("/me",      userHandler.UpdateMe)
            }

            nfts := protected.Group("/nfts")
            {
                nfts.GET("",         nftHandler.List)
                nfts.GET("/:id",     nftHandler.GetByID)
                nfts.POST("",        nftHandler.Create)
                nfts.POST("/:id/buy",nftHandler.Buy)
            }
        }
    }
}
```

---

## .ENV.EXAMPLE TEMPLATE

```env
# Server
PORT=8080
APP_ENV=development

# Database
DATABASE_URL=postgres://user:password@localhost:5432/dapp_db?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# JWT
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRE_HOURS=24

# Redis (optional)
REDIS_URL=redis://localhost:6379

# Blockchain
CONTRACT_ADDRESS=0x...
RPC_URL=https://mainnet.infura.io/v3/...
```
