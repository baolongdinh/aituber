BIN_NAME=backend-service
CONTAINER_NAME=backend-service

build:
	@echo "Build $(BIN_NAME) execute...."
	@GOARCH=amd64 GOOS=linux go build -tags timetzdata -o ../bin/$(BIN_NAME) .
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