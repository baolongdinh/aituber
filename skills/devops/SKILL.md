# DevOps & Binary Deployment Skill

## TRIGGER
Read this file when dealing with binary deployment, Docker containerization, CI/CD, environment management, or infrastructure setup using binary builds.

---

## IDENTITY
You are a Senior DevOps Engineer specializing in binary-based deployments.
You ensure applications are deployed securely, efficiently, and with proper monitoring using optimized binary builds.
You never deploy "just to get it working" — every deployment must be repeatable and monitored.

---

--- 

## ⚡ DEVOPS AUTOMATION (PRO-LEVEL) 

Use these specialized scripts to manage the application lifecycle: 

### 1. Environment Check 
Verify your development or CI environment is properly set up. 
**Command:** 
```bash 
python3 scripts/env-check.py 
``` 

### 2. Deployment Helper 
Automate the build, compilation, and packaging of the full stack. 
**Command:** 
```bash 
python3 scripts/deploy-helper.py 
``` 

### 3. Docker Generator 
Generate optimized Dockerfiles and docker-compose configurations. 
**Command:** 
```bash 
python3 scripts/docker-gen.py 
``` 

## 🏗️ BINARY-FIRST DEPLOYMENT STRATEGY

### Backend Binary Build & Dockerfile
```dockerfile
# backend/Dockerfile
FROM ubuntu:22.04

USER root
RUN mkdir /app
WORKDIR /app

# Install certificates for external API calls (Gemini/AIOZ RPC)
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy binary from bin folder (context is root of repo, so path is ./bin/...)
# Note: Makefile will build to ../bin from backend dir, which equates to ./bin from root context
COPY ./bin/ai-playground-backend .

EXPOSE 8080

CMD ["/app/ai-playground-backend"]
```

### Backend Makefile
```makefile
# backend/Makefile
BIN_NAME=ai-playground-backend
CONTAINER_NAME=aioz-ai-playground-backend

build:
	@echo "Build $(BIN_NAME) execute...."
	@mkdir -p ../bin
	@GOARCH=amd64 GOOS=linux go build -o ../bin/$(BIN_NAME) .
	@echo "Build $(BIN_NAME) execute: done"

test:
	@echo "Run $(BIN_NAME) test...."
	@go test -v ./...
	@echo "Run $(BIN_NAME) test: done "

lint:
	@echo "Run $(BIN_NAME) lint..."
	@go vet ./...
	@echo "Runt $(BIN_NAME) lint: done"

run: build
	@clear
	@../bin/$(BIN_NAME)

restart: build
	@docker restart $(CONTAINER_NAME)

stop-container:
	@docker stop $(CONTAINER_NAME)

start-container:
	@docker restart $(CONTAINER_NAME)
```

### Root Docker Compose
```yaml
# docker-compose.yml (root level)
version: '3.8'

services:
  ai-playground-db:
    image: postgres:15-alpine
    container_name: aioz-ai-playground-db
    restart: always
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-aioz}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-aioz123}
      POSTGRES_DB: ${POSTGRES_DB:-ai_playground}
    # ports:
    #   - "5432:5432"
    expose:
      - '5432'
    volumes:
      - ./backend/postgres_data:/var/lib/postgresql/data

  ai-playground-backend:
    build:
      context: .
      dockerfile: ./backend/Dockerfile
    container_name: aioz-ai-playground-backend
    restart: always
    depends_on:
      - ai-playground-db
    ports:
      - "8892:8892"
    env_file:
      - ./backend/.env
    environment:
      POSTGRES_HOST: ai-playground-db
      POSTGRES_PORT: 5432
      POSTGRES_USER: ${POSTGRES_USER:-aioz}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-aioz123}
      POSTGRES_DB: ${POSTGRES_DB:-ai_playground}
      LOCAL_UPLOAD_PATH: /app/uploads
    volumes:
      - ./backend/uploads:/app/uploads

  # ai-playground-frontend:
  #   image: nginx:alpine
  #   container_name: aioz-ai-playground-frontend
  #   restart: always
  #   ports:
  #     - "80:80"
  #   volumes:
  #     - ./frontend/dist:/usr/share/nginx/html
  #     - ./frontend/nginx.conf:/etc/nginx/conf.d/default.conf
```

---

## 🚀 ROOT DEPLOYMENT MAKEFILE

### Master Deployment Makefile
```makefile
# Makefile (root level)
# Binary and Container Names
BIN_NAME=ai-playground-backend
CONTAINER_NAME=aioz-ai-playground-backend

# Remote Deployment Config
REMOTE_USER=aioz-ai-hub
REMOTE_HOST=10.0.0.154
REMOTE_PATH=~/Desktop/dapps/ai-playground
REMOTE_PASSWORD=Hub@aioz1

# Deployment Commands
BASE_SSH_COMMAND = sshpass -p '$(REMOTE_PASSWORD)' ssh -p 22 -o StrictHostKeyChecking=no $(REMOTE_USER)@$(REMOTE_HOST)
BASE_SCP_COMMAND = sshpass -p '$(REMOTE_PASSWORD)' scp -P 22 -o StrictHostKeyChecking=no

# Colors
COLOR_RESET = \033[0m
COLOR_BLUE = \033[0;34m
COLOR_GREEN = \033[0;32m
COLOR_YELLOW = \033[0;33m

.PHONY: help clean build-backend build-frontend build-all init stop \
        deploy-backend deploy-frontend deploy-all logs logs-backend logs-frontend

## Help command
help:
	@echo "$(COLOR_BLUE)Available commands:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)make init$(COLOR_RESET)            - First time setup (create directories and upload initial files)"
	@echo "  $(COLOR_GREEN)make build-backend$(COLOR_RESET)   - Build backend locally"
	@echo "  $(COLOR_GREEN)make build-frontend$(COLOR_RESET)  - Build frontend locally"
	@echo "  $(COLOR_GREEN)make build-all$(COLOR_RESET)       - Build both"
	@echo "  $(COLOR_GREEN)make deploy-backend$(COLOR_RESET)  - Deploy backend only"
	@echo "  $(COLOR_GREEN)make deploy-frontend$(COLOR_RESET) - Deploy frontend only"
	@echo "  $(COLOR_GREEN)make deploy-all$(COLOR_RESET)      - Deploy all services"
	@echo "  $(COLOR_GREEN)make stop$(COLOR_RESET)            - Stop all services on remote"
	@echo "  $(COLOR_GREEN)make logs$(COLOR_RESET)            - View all logs"

## Clean build artifacts
clean:
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf ./bin/
	@mkdir -p ./bin/
	@echo "$(COLOR_GREEN)✓ Clean completed$(COLOR_RESET)"

## Build Commands
build-backend: clean
	@echo "$(COLOR_YELLOW)Building backend...$(COLOR_RESET)"
	@mkdir -p ./bin
	@cd backend && GOARCH=amd64 GOOS=linux go build -o ../bin/$(BIN_NAME) .
	@echo "$(COLOR_GREEN)✓ Backend build completed$(COLOR_RESET)"

build-frontend:
	@echo "$(COLOR_YELLOW)Building frontend...$(COLOR_RESET)"
	@cd frontend && npm install && npm run build
	@echo "$(COLOR_GREEN)✓ Frontend build completed$(COLOR_RESET)"

build-all: build-backend build-frontend
	@echo "$(COLOR_GREEN)✓ All builds completed$(COLOR_RESET)"

## Initial Setup (Checking & Creating Remote Dirs)
init:
	@echo "$(COLOR_YELLOW)Initializing remote server...$(COLOR_RESET)"
	@$(BASE_SSH_COMMAND) "mkdir -p $(REMOTE_PATH)/backend/postgres_data $(REMOTE_PATH)/bin $(REMOTE_PATH)/frontend/dist"
	
	@echo "Uploading initial configuration..."
	@$(BASE_SCP_COMMAND) ./docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/docker-compose.yml
	@$(BASE_SCP_COMMAND) ./backend/.env $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/backend/.env
	@$(BASE_SCP_COMMAND) ./backend/Dockerfile $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/backend/Dockerfile
	@$(BASE_SCP_COMMAND) ./frontend/nginx.conf $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/frontend/nginx.conf
	
	@echo "$(COLOR_GREEN)✓ Initialization completed$(COLOR_RESET)"

## Upload Commands (Internal)
upload-backend:
	@echo "$(COLOR_YELLOW)Uploading backend binaries & config...$(COLOR_RESET)"
	@$(BASE_SCP_COMMAND) -r ./bin/* $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/bin/
	@$(BASE_SCP_COMMAND) ./backend/.env $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/backend/.env
	@$(BASE_SCP_COMMAND) ./backend/Dockerfile $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/backend/Dockerfile
	@echo "$(COLOR_GREEN)✓ Backend upload completed$(COLOR_RESET)"

upload-frontend:
	@echo "$(COLOR_YELLOW)Uploading frontend assets...$(COLOR_RESET)"
	@$(BASE_SSH_COMMAND) "mkdir -p $(REMOTE_PATH)/frontend/dist-new"
	@$(BASE_SCP_COMMAND) -r ./frontend/dist/* $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/frontend/dist-new/
	@$(BASE_SCP_COMMAND) ./frontend/nginx.conf $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/frontend/nginx.conf
	@echo "$(COLOR_GREEN)✓ Frontend upload completed$(COLOR_RESET)"

## Deploy Commands
deploy-backend: build-backend upload-backend
	@echo "$(COLOR_YELLOW)Deploying backend...$(COLOR_RESET)"
	@$(BASE_SCP_COMMAND) ./docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/docker-compose.yml
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose up -d --build --force-recreate ai-playground-backend ai-playground-db"
	@echo "$(COLOR_GREEN)✓ Backend deployment completed$(COLOR_RESET)"

deploy-frontend: build-frontend upload-frontend
	@echo "$(COLOR_YELLOW)Deploying frontend...$(COLOR_RESET)"
	# Zero-downtime-ish swap
	@$(BASE_SSH_COMMAND) "rm -rf $(REMOTE_PATH)/frontend/dist.old && mv $(REMOTE_PATH)/frontend/dist $(REMOTE_PATH)/frontend/dist.old 2>/dev/null || true"
	@$(BASE_SSH_COMMAND) "mv $(REMOTE_PATH)/frontend/dist-new $(REMOTE_PATH)/frontend/dist"
	@$(BASE_SCP_COMMAND) ./docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/docker-compose.yml
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose up -d --force-recreate ai-playground-frontend"
	@echo "$(COLOR_GREEN)✓ Frontend deployment completed$(COLOR_RESET)"

deploy-frontend-vercel:
	@echo "$(COLOR_YELLOW)Deploying frontend to Vercel...$(COLOR_RESET)"
	@cd frontend && npx vercel --prod && cd ..
	@echo "$(COLOR_GREEN)✓ Frontend deployment to Vercel completed$(COLOR_RESET)"

deploy-all: build-all
	@echo "$(COLOR_YELLOW)Deploying full stack...$(COLOR_RESET)"
	@make upload-backend
	@make upload-frontend
	@$(BASE_SCP_COMMAND) ./docker-compose.yml $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)/docker-compose.yml
	
	@echo "Swapping frontend assets..."
	@$(BASE_SSH_COMMAND) "rm -rf $(REMOTE_PATH)/frontend/dist.old && mv $(REMOTE_PATH)/frontend/dist $(REMOTE_PATH)/frontend/dist.old 2>/dev/null || true"
	@$(BASE_SSH_COMMAND) "mv $(REMOTE_PATH)/frontend/dist-new $(REMOTE_PATH)/frontend/dist"
	
	@echo "Restarting containers..."
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose up -d --build --remove-orphans"
	@echo "$(COLOR_GREEN)✓ Full deployment completed$(COLOR_RESET)"

## Service Control
stop:
	@echo "$(COLOR_YELLOW)Stopping services...$(COLOR_RESET)"
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose stop"
	@echo "$(COLOR_GREEN)✓ Services stopped$(COLOR_RESET)"

## Logs
logs:
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose logs -f"

logs-backend:
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose logs -f ai-playground-backend ai-playground-db"

logs-frontend:
	@$(BASE_SSH_COMMAND) "cd $(REMOTE_PATH) && docker compose logs -f ai-playground-frontend"
```

---

## 🔧 BINARY OPTIMIZATION

### Production Build Optimization
```bash
# backend/Makefile (enhanced)
BIN_NAME=ai-playground-backend
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -w -s"

build:
	@echo "Building $(BIN_NAME) version $(VERSION)..."
	@mkdir -p ../bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ../bin/$(BIN_NAME) .
	@echo "Build $(BIN_NAME) completed: $(VERSION)"

build-debug:
	@echo "Building $(BIN_NAME) with debug info..."
	@mkdir -p ../bin
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -gcflags="all=-N -l" -o ../bin/$(BIN_NAME)-debug .
	@echo "Debug build completed"

build-release:
	@echo "Building $(BIN_NAME) for release..."
	@mkdir -p ../bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo $(LDFLAGS) -o ../bin/$(BIN_NAME) .
	@echo "Release build completed"

size:
	@echo "Binary size analysis:"
	@ls -lh ../bin/$(BIN_NAME)
	@file ../bin/$(BIN_NAME)
```

### Optimized Dockerfile for Binary
```dockerfile
# backend/Dockerfile (optimized)
FROM ubuntu:22.04

# Use non-root user for security
RUN groupadd -r appuser && useradd -r -g appuser appuser

USER root
RUN mkdir /app && chown appuser:appuser /app
WORKDIR /app

# Install minimal runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Copy and set permissions
COPY --chown=appuser:appuser ./bin/ai-playground-backend .
RUN chmod +x /ai-playground-backend

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /ai-playground-backend health || exit 1

EXPOSE 8080

CMD ["/ai-playground-backend"]
```

---

## 📦 ENVIRONMENT MANAGEMENT

### Environment Configuration
```bash
# .env.example (root level)
# Database Configuration
POSTGRES_USER=aioz
POSTGRES_PASSWORD=aioz123
POSTGRES_DB=ai_playground

# Backend Configuration
BACKEND_PORT=8892
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRES_IN=24h

# External APIs
GEMINI_API_KEY=your-gemini-api-key
AIOZ_RPC_URL=https://eth-dataseed.aioz.network

# File Upload
MAX_FILE_SIZE=10MB
UPLOAD_PATH=/app/uploads

# Deployment
REMOTE_USER=aioz-ai-hub
REMOTE_HOST=10.0.0.154
REMOTE_PATH=~/Desktop/dapps/ai-playground
```

### Environment-specific Scripts
```bash
#!/bin/bash
# scripts/deploy.sh

set -e

ENVIRONMENT=${1:-production}
echo "Deploying to $ENVIRONMENT"

case $ENVIRONMENT in
  "development")
    make build-backend
    docker compose up -d ai-playground-backend ai-playground-db
    ;;
  "staging")
    make deploy-backend
    ;;
  "production")
    make deploy-all
    ;;
  *)
    echo "Unknown environment: $ENVIRONMENT"
    echo "Usage: $0 [development|staging|production]"
    exit 1
    ;;
esac

echo "Deployment to $ENVIRONMENT completed"
```

---

## 🔍 MONITORING & LOGGING

### Health Check Implementation
```go
// internal/handler/health_handler.go
package handler

import (
    "net/http"
    "runtime"
    "time"
    
    "github.com/gin-gonic/gin"
)

type HealthResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version"`
    Uptime    string    `json:"uptime"`
    Memory    Memory    `json:"memory"`
    Build     BuildInfo `json:"build"`
}

type Memory struct {
    Alloc      uint64 `json:"alloc"`
    TotalAlloc uint64 `json:"totalAlloc"`
    Sys        uint64 `json:"sys"`
    NumGC      uint32 `json:"numGC"`
}

type BuildInfo struct {
    Version   string `json:"version"`
    BuildTime string `json:"build_time"`
    GoVersion string `json:"go_version"`
}

var startTime = time.Now()

func HealthCheck(c *gin.Context) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    response := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   getVersion(),
        Uptime:    time.Since(startTime).String(),
        Memory: Memory{
            Alloc:      m.Alloc,
            TotalAlloc: m.TotalAlloc,
            Sys:        m.Sys,
            NumGC:      m.NumGC,
        },
        Build: BuildInfo{
            Version:   getVersion(),
            BuildTime: getBuildTime(),
            GoVersion: runtime.Version(),
        },
    }

    c.JSON(http.StatusOK, response)
}

func getVersion() string {
    // This would be set at build time using ldflags
    if version := os.Getenv("APP_VERSION"); version != "" {
        return version
    }
    return "dev"
}

func getBuildTime() string {
    if buildTime := os.Getenv("BUILD_TIME"); buildTime != "" {
        return buildTime
    }
    return time.Now().Format("2006-01-02_15:04:05")
}
```

### Logging Configuration
```go
// pkg/logger/logger.go
package logger

import (
    "os"
    "github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func Init() {
    Logger = logrus.New()
    
    // Set output format
    Logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
    })
    
    // Set log level from environment
    level := os.Getenv("LOG_LEVEL")
    if level == "" {
        level = "info"
    }
    
    logLevel, err := logrus.ParseLevel(level)
    if err != nil {
        logLevel = logrus.InfoLevel
    }
    
    Logger.SetLevel(logLevel)
    
    // Set output
    Logger.SetOutput(os.Stdout)
    
    Logger.WithFields(logrus.Fields{
        "service": "ai-playground-backend",
        "version": getVersion(),
    }).Info("Logger initialized")
}
```

---

## 🚀 DEPLOYMENT BEST PRACTICES

### Pre-Deployment Checklist
```
[ ] Binary is built with production flags
[ ] Environment variables are configured
[ ] Docker images are built and tested
[ ] Database migrations are applied
[ ] Health checks are passing
[ ] SSL certificates are installed (if needed)
[ ] Monitoring is enabled
[ ] Log aggregation is configured
[ ] Backup strategies are in place
[ ] Security scanning is completed
[ ] Performance testing is passed
[ ] Rollback plan is ready
```

### Production Deployment Script
```bash
#!/bin/bash
# scripts/production-deploy.sh

set -e

echo "🚀 Starting production deployment..."

# Pre-deployment checks
echo "📋 Running pre-deployment checks..."
make test
make lint

# Build binaries
echo "🔨 Building binaries..."
make clean
make build-all

# Deploy to production
echo "📦 Deploying to production..."
make deploy-all

# Health check
echo "🏥 Checking service health..."
sleep 30
curl -f http://localhost:8892/health || {
    echo "❌ Health check failed!"
    exit 1
}

echo "✅ Production deployment completed successfully!"
```

---

## 📋 DEPLOYMENT CHECKLIST

### Development Environment
```
[ ] Local development setup complete
[ ] Docker installed and running
[ ] Environment variables configured
[ ] Database accessible
[ ] Binary builds successfully
[ ] Docker compose works locally
[ ] Health endpoint accessible
[ ] Logs are properly formatted
```

### Production Environment
```
[ ] Remote server access configured
[ ] SSH keys or password authentication
[ ] Docker installed on remote server
[ ] Required directories created
[ ] Environment variables set
[ ] SSL certificates configured
[ ] Firewall rules configured
[ ] Backup strategy implemented
[ ] Monitoring setup complete
[ ] Log rotation configured
[ ] Resource limits set appropriately
```

---

## DO / DON'T

✅ **DO**
- Use binary builds for consistent deployments
- Implement proper health checks
- Use non-root users in containers
- Set up proper monitoring and logging
- Implement rolling updates for zero downtime
- Use environment-specific configurations
- Set resource limits for containers
- Implement backup and recovery strategies
- Test deployments in staging first
- Use proper versioning and tagging

❌ **DON'T**
- NEVER deploy untested binaries
- NEVER hardcode secrets in Dockerfiles
- NEVER run containers as root user in production
- NEVER skip health checks in production
- NEVER ignore security scanning
- NEVER deploy without rollback plan
- NEVER use latest tags in production
- NEVER ignore resource limits
- NEVER skip backup verification
- NEVER deploy without proper monitoring
