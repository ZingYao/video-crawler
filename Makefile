# Makefile for video-crawler

# 变量定义
BINARY_NAME=video-crawler
BUILD_DIR=bin
MAIN_PATH=cmd/video-crawler/main.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"

# 支持的平台和架构
PLATFORMS=linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/386 windows/arm64 android/amd64 android/arm64

# 默认目标
.PHONY: all
all: clean build-frontend build-all

# 构建前端
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build
	@echo "Frontend build completed"

# 构建当前平台
.PHONY: build
build: build-frontend
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建所有平台
.PHONY: build-all
build-all: build-frontend
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$BINARY_NAME_FULL $(MAIN_PATH); \
	done
	@echo "All platform builds completed in $(BUILD_DIR)/"

# 构建特定平台
.PHONY: build-linux
build-linux: build-frontend
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=386 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 $(MAIN_PATH)
	GOOS=linux GOARCH=arm $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm $(MAIN_PATH)
	@echo "Linux builds completed"

.PHONY: build-darwin
build-darwin: build-frontend
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "macOS builds completed"

.PHONY: build-windows
build-windows: build-frontend
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	GOOS=windows GOARCH=386 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe $(MAIN_PATH)
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PATH)
	@echo "Windows builds completed"

.PHONY: build-android
build-android: build-frontend
	@echo "Building for Android..."
	@mkdir -p $(BUILD_DIR)
	GOOS=android GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-android-amd64 $(MAIN_PATH)
	GOOS=android GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-android-arm64 $(MAIN_PATH)
	@echo "Android builds completed"

# 运行项目
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# 开发模式运行（包含前端热重载）
.PHONY: dev
dev:
	@echo "Starting development mode..."
	@echo "Frontend will be available at http://localhost:5173"
	@echo "Backend will be available at http://localhost:8080"
	@cd frontend && npm run dev &
	@sleep 3
	$(GOCMD) run $(MAIN_PATH)

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf frontend/dist
	@rm -rf internal/static/dist
	@echo "Clean completed"

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Downloading frontend dependencies..."
	@cd frontend && npm install
	@echo "All dependencies downloaded"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GOCMD) fmt ./...
	@echo "Formatting frontend code..."
	@cd frontend && npm run format 2>/dev/null || echo "No frontend formatter configured"

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# 代码检查
.PHONY: lint
lint: fmt vet

# 安装项目
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Install completed"

# 创建发布包
.PHONY: release
release: build-all
	@echo "Creating release packages..."
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; \
		fi; \
		RELEASE_NAME="$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			zip -j release/$$RELEASE_NAME.zip $(BUILD_DIR)/$$BINARY_NAME_FULL configs/ README.md; \
		else \
			tar -czf release/$$RELEASE_NAME.tar.gz -C $(BUILD_DIR) $$BINARY_NAME_FULL -C ../../ configs/ README.md; \
		fi; \
		echo "Created release: release/$$RELEASE_NAME.*"; \
	done
	@echo "Release packages created in release/"

# 显示版本信息
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# 帮助信息
.PHONY: help
help:
	@echo "Video Crawler Build System"
	@echo "=========================="
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-all      - Build for all platforms (Linux, macOS, Windows, Android)"
	@echo "  build-linux    - Build for Linux (amd64, arm64, 386, arm)"
	@echo "  build-darwin   - Build for macOS (amd64, arm64)"
	@echo "  build-windows  - Build for Windows (amd64, 386, arm64)"
	@echo "  build-android  - Build for Android (amd64, arm64)"
	@echo "  build-frontend - Build frontend only"
	@echo ""
	@echo "Development targets:"
	@echo "  run            - Run the application"
	@echo "  dev            - Run in development mode (frontend + backend)"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean          - Clean build files"
	@echo "  deps           - Download dependencies (Go + Node.js)"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run fmt and vet"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  release        - Create release packages"
	@echo "  version        - Show version information"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Supported platforms:"
	@echo "  Linux:   amd64, arm64, 386, arm"
	@echo "  macOS:   amd64, arm64"
	@echo "  Windows: amd64, 386, arm64"
	@echo "  Android: amd64, arm64"
