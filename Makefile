# MailMole Makefile
# Provides convenient commands for building, running, and managing MailMole

# Variables
BINARY_NAME=mailmole
DOCKER_IMAGE=mailmole
GO_VERSION=$(shell go version | awk '{print $$3}')
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=dev -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Colors
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build run clean install test help docker-build docker-run

# Default target
all: help

## help: Show this help message
help:
	@echo "$(BLUE)MailMole Makefile Commands:$(NC)"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
	@echo ""
	@echo "$(BLUE)Examples:$(NC)"
	@echo "  make build          # Build the binary"
	@echo "  make run            # Run with terminal UI"
	@echo "  make web            # Run with web dashboard"
	@echo "  make install        # Install dependencies and build"

## build: Build the MailMole binary
build:
	@echo "$(BLUE)🔨 Building MailMole...$(NC)"
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)✅ Build complete: ./$(BINARY_NAME)$(NC)"

## run: Run MailMole with terminal UI
run: build
	@echo "$(BLUE)🚀 Starting MailMole (Terminal UI)...$(NC)"
	@./$(BINARY_NAME)

## web: Run MailMole with web dashboard
web: build
	@echo "$(BLUE)🌐 Starting MailMole Web Dashboard...$(NC)"
	@echo "$(YELLOW)   Access at: http://localhost:8080$(NC)"
	@./$(BINARY_NAME) -web :8080

## web-only: Run only web dashboard (no TUI)
web-only: build
	@echo "$(BLUE)🌐 Starting MailMole Web Dashboard (only)...$(NC)"
	@echo "$(YELLOW)   Access at: http://localhost:8080$(NC)"
	@./$(BINARY_NAME) -web :8080 -web-only

## install: Install dependencies and build
install:
	@echo "$(BLUE)📦 Installing dependencies...$(NC)"
	@bash install.sh

## clean: Remove build artifacts
clean:
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f mailmole.log
	@rm -f migration_state.json
	@echo "$(GREEN)✅ Clean complete$(NC)"

## test: Run tests
test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@go test -v ./...

## fmt: Format Go code
fmt:
	@echo "$(BLUE)📝 Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Formatting complete$(NC)"

## lint: Run linter
lint:
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## deps: Download and verify dependencies
deps:
	@echo "$(BLUE)📥 Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✅ Dependencies ready$(NC)"

## update: Update Go dependencies
update:
	@echo "$(BLUE)🔄 Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## release: Build release binaries for multiple platforms
release: clean
	@echo "$(BLUE)📦 Building release binaries...$(NC)"
	
	@echo "$(YELLOW)  Building for Linux (amd64)...$(NC)"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_amd64 .
	
	@echo "$(YELLOW)  Building for Linux (arm64)...$(NC)"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_arm64 .
	
	@echo "$(YELLOW)  Building for macOS (amd64)...$(NC)"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_amd64 .
	
	@echo "$(YELLOW)  Building for macOS (arm64)...$(NC)"
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_arm64 .
	
	@echo "$(YELLOW)  Building for Windows (amd64)...$(NC)"
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_windows_amd64.exe .
	
	@echo "$(GREEN)✅ Release binaries built in dist/$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)✅ Docker image built: $(DOCKER_IMAGE):latest$(NC)"

## docker-run: Run MailMole in Docker
docker-run: docker-build
	@echo "$(BLUE)🐳 Running MailMole in Docker...$(NC)"
	@docker run -it --rm -p 8080:8080 -v $(PWD)/data:/data $(DOCKER_IMAGE):latest

## docker-web: Run MailMole web dashboard in Docker
docker-web: docker-build
	@echo "$(BLUE)🐳 Running MailMole Web Dashboard in Docker...$(NC)"
	@echo "$(YELLOW)   Access at: http://localhost:8080$(NC)"
	@docker run -it --rm -p 8080:8080 -v $(PWD)/data:/data $(DOCKER_IMAGE):latest -web :8080 -web-only

## setup: Initial project setup
setup:
	@echo "$(BLUE)⚙️  Setting up MailMole project...$(NC)"
	@mkdir -p logs exports backups dist
	@echo "$(GREEN)✅ Project directories created$(NC)"
	@echo "$(BLUE)📦 Run 'make install' or './install.sh' to complete installation$(NC)"

## dev: Run in development mode with hot reload (requires air)
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "$(BLUE)🔄 Starting development server with hot reload...$(NC)"; \
		air; \
	else \
		echo "$(YELLOW)⚠️  air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(BLUE)   Falling back to regular build...$(NC)"; \
		make run; \
	fi

## info: Show project information
info:
	@echo "$(BLUE)📊 Project Information:$(NC)"
	@echo "   Binary:    $(BINARY_NAME)"
	@echo "   Go Version: $(GO_VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Git Commit: $(GIT_COMMIT)"
	@echo ""
	@if [ -f $(BINARY_NAME) ]; then \
		echo "$(GREEN)   Status: Built ✓$(NC)"; \
	else \
		echo "$(RED)   Status: Not built yet$(NC)"; \
	fi

# Default target if no arguments given
.DEFAULT_GOAL := help
