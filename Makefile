# Makefile for video-crawler

# 变量定义
BINARY_NAME=video-crawler
WAILS_BINARY_NAME=video-crawler-desktop
HTTP_BINARY_NAME=video-crawler-server
BUILD_DIR=bin
MAIN_PATH=main.go
HTTP_MAIN_PATH=cmd/http-server/main.go
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

# 支持的平台和架构（HTTP 服务）
HTTP_PLATFORMS=linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/386 windows/arm64
# Wails 桌面应用平台（通过 wails build 交叉编译）
WAILS_PLATFORMS=darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/amd64 linux/arm64

# 默认目标
.PHONY: all
all: clean build-frontend build-http-all build-wails-all

# 构建前端
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build
	@echo "Frontend build completed"

# 构建当前平台（后端：Wails入口）
.PHONY: build
build: build-frontend
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建Wails桌面应用（当前平台）
.PHONY: build-wails
build-wails: build-frontend
	@echo "Building Wails desktop application (current platform)..."
	@mkdir -p $(BUILD_DIR)
	wails build -o $(BUILD_DIR)/$(WAILS_BINARY_NAME)
	@echo "Wails build completed: $(BUILD_DIR)/$(WAILS_BINARY_NAME)"

# 构建HTTP服务（当前平台）
.PHONY: build-http
build-http: build-frontend
	@echo "Building HTTP server (current platform)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(HTTP_BINARY_NAME) $(HTTP_MAIN_PATH)
	@echo "HTTP server build completed: $(BUILD_DIR)/$(HTTP_BINARY_NAME)"

# HTTP 服务：构建所有平台
.PHONY: build-http-all
build-http-all: build-frontend
	@echo "Building HTTP server for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(HTTP_PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(HTTP_BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$BINARY_NAME_FULL $(HTTP_MAIN_PATH); \
	 done
	@echo "HTTP server builds completed in $(BUILD_DIR)/"

# Wails 桌面应用：构建常用桌面平台
.PHONY: build-wails-all
build-wails-all: build-frontend
	@echo "Building Wails desktop for common platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(WAILS_PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		echo "Wails building for $$GOOS/$$GOARCH..."; \
		wails build -platform $$GOOS/$$GOARCH -o $(BUILD_DIR)/$(WAILS_BINARY_NAME)-$$GOOS-$$GOARCH || exit 1; \
	 done
	@echo "Wails builds completed in $(BUILD_DIR)/"

# iOS/Android 构建占位（需要对应SDK，提供提示）
.PHONY: build-wails-ios build-wails-android
build-wails-ios:
	@echo "iOS build requires Xcode, iOS SDK and Wails mobile support. Refer to README_WAILS.md."

build-wails-android:
	@echo "Android build requires Android SDK/NDK and Wails mobile support. Refer to README_WAILS.md."

# 运行项目
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# 运行Wails开发模式
.PHONY: run-wails
run-wails:
	@echo "Running Wails in development mode..."
	wails dev

# 运行HTTP服务
.PHONY: run-http
run-http:
	@echo "Running HTTP server..."
	$(GOCMD) run $(HTTP_MAIN_PATH)

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

# 创建发布包（HTTP 服务）
.PHONY: release
release: build-http-all
	@echo "Creating release packages..."
	@mkdir -p release
	@for platform in $(HTTP_PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(HTTP_BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; \
		fi; \
		RELEASE_NAME="$(HTTP_BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH"; \
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
	@echo "  build             - Build current platform (Wails entry)"
	@echo "  build-http        - Build HTTP server (current platform)"
	@echo "  build-wails       - Build Wails desktop (current platform)"
	@echo "  build-http-all    - Build HTTP server for Linux/macOS/Windows common archs"
	@echo "  build-wails-all   - Build Wails desktop for Linux/macOS/Windows common archs"
	@echo "  build-wails-ios   - Print requirements for iOS build"
	@echo "  build-wails-android- Print requirements for Android build"
	@echo "  build-frontend    - Build frontend only"
	@echo ""
	@echo "Development targets:"
	@echo "  run               - Run the application"
	@echo "  run-http          - Run HTTP server"
	@echo "  run-wails         - Run Wails dev"
	@echo "  dev               - Run frontend dev + backend"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean             - Clean build files"
	@echo "  deps              - Download dependencies (Go + Node.js)"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"
	@echo "  lint              - Run fmt and vet"
	@echo "  install           - Install to GOPATH/bin"
	@echo "  release           - Create release packages (HTTP server)"
	@echo "  version           - Show version information"
	@echo "  help              - Show this help message"
