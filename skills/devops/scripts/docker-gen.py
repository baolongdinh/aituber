#!/usr/bin/env python3
import os

DOCKERFILE_TEMPLATE = """# --- Stage 1: Build Backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# --- Stage 2: Build Frontend ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# --- Stage 3: Final Image ---
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

# Copy backend
COPY --from=backend-builder /app/server .
COPY backend/migrations ./migrations

# Copy frontend
COPY --from=frontend-builder /app/dist ./public

EXPOSE 8080
CMD ["./server"]
"""

DOCKER_COMPOSE_TEMPLATE = """version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_URL=postgres://user:pass@db:5432/dapp?sslmode=disable
    depends_on:
      - db

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
      - POSTGRES_DB=dapp
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
"""

def main():
    print("🐳 GENERATING DOCKER CONFIGURATION\n")
    
    with open("Dockerfile", "w") as f:
        f.write(DOCKERFILE_TEMPLATE)
    print("✅ Created Dockerfile (Multi-stage optimized)")
    
    with open("docker-compose.yml", "w") as f:
        f.write(DOCKER_COMPOSE_TEMPLATE)
    print("✅ Created docker-compose.yml")
    
    print("\n🚀 Ready to run: docker-compose up --build")

if __name__ == "__main__":
    main()
