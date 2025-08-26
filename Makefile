# Makefile for video-crawler

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# 变量定义
BINARY_NAME=video-crawler
BUILD_DIR=bin
MAIN_PATH=cmd/http-server/main.go
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
PLATFORMS=linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/386 windows/arm64 android/amd64 android/arm64 android/386 android/arm

# 默认目标
.PHONY: all
all: clean build-frontend build-all

# 构建前端
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build
	@echo "Frontend build completed"

# 构建当前平台（不启用 CGO）
.PHONY: build
build: build-frontend
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建所有平台（仅 Android 开启 CGO）
.PHONY: build-all
build-all: build-frontend
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		if [ "$$GOOS" = "android" ]; then \
			continue; \
		fi; \
		BINARY_NAME_FULL="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; fi; \
		CGO=0; \
		echo "Building for $$GOOS/$$GOARCH (CGO_ENABLED=$$CGO)..."; \
		CGO_ENABLED=$$CGO GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$BINARY_NAME_FULL $(MAIN_PATH) || exit 1; \
	done; \
	BUILD_DIR=$(BUILD_DIR) BINARY_NAME=$(BINARY_NAME) MAIN_PATH=$(MAIN_PATH) VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) scripts/build_android.sh; \
	echo "All platform builds completed in $(BUILD_DIR)/"

# 构建特定平台（非 Android 关闭 CGO）
.PHONY: build-linux
build-linux: build-frontend
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64 386 arm; do \
		echo "linux/$$arch (CGO_ENABLED=0)"; \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-$$arch $(MAIN_PATH) || exit 1; \
	done; \
	echo "Linux builds completed"

.PHONY: build-darwin
build-darwin: build-frontend
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "darwin/$$arch (CGO_ENABLED=0)"; \
		CGO_ENABLED=0 GOOS=darwin GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-$$arch $(MAIN_PATH) || exit 1; \
	done; \
	echo "macOS builds completed"

.PHONY: build-windows
build-windows: build-frontend
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 386 arm64; do \
		echo "windows/$$arch (CGO_ENABLED=0)"; \
		CGO_ENABLED=0 GOOS=windows GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-$$arch.exe $(MAIN_PATH) || exit 1; \
	done; \
	echo "Windows builds completed"

.PHONY: build-android
build-android: build-frontend
	@echo "Building for Android..."
	@BUILD_DIR=$(BUILD_DIR) BINARY_NAME=$(BINARY_NAME) MAIN_PATH=$(MAIN_PATH) VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) scripts/build_android.sh

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
	$(GOCMD) clean
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

# 安装项目（本机 CGO=1）
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Install completed"

# 创建发布包
.PHONY: release
release: build-all
	@echo "Creating release packages..."
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; fi; \
		RELEASE_NAME="$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then \
			zip -j release/$$RELEASE_NAME.zip $(BUILD_DIR)/$$BINARY_NAME_FULL configs/ README.md; \
		else \
			tar -czf release/$$RELEASE_NAME.tar.gz -C $(BUILD_DIR) $$BINARY_NAME_FULL -C ../../ configs/ README.md; \
		fi; \
		echo "Created release: release/$$RELEASE_NAME.*"; \
	done
	@echo "Release packages created in release/"

# HTTP all-arch build including Android
.PHONY: build-http-all
build-http-all: build-frontend
	@echo "Building HTTP service for Linux/macOS/Windows/Android (all archs)..."
	@mkdir -p $(BUILD_DIR)
	@API=21; \
	for platform in linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/386 windows/arm64 android/amd64 android/arm64 android/386 android/arm; do \
		IFS='/' read -r GOOS GOARCH <<< "$$platform"; \
		BINARY_NAME_FULL="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then BINARY_NAME_FULL="$$BINARY_NAME_FULL.exe"; fi; \
		CGO=0; GOARM_VAL=""; CC_BIN=""; \
		if [ "$$GOOS" = "android" ]; then \
			CGO=1; \
			case "$$GOARCH" in \
				amd64) TRIPLE=x86_64-linux-android ;; \
				386)   TRIPLE=i686-linux-android ;; \
				arm64) TRIPLE=aarch64-linux-android ;; \
				arm)   TRIPLE=armv7a-linux-androideabi ; GOARM_VAL=7 ;; \
				*)     TRIPLE="" ;; \
			esac; \
			if [ -n "$$TRIPLE" ] && [ -n "$$ANDROID_NDK_HOME" ]; then \
				PREBUILT_DIR=$$(ls -d "$$ANDROID_NDK_HOME"/toolchains/llvm/prebuilt/* 2>/dev/null | head -n 1); \
				if [ -n "$$PREBUILT_DIR" ]; then CC_BIN="$$PREBUILT_DIR/bin/$$TRIPLE$$API-clang"; fi; \
			fi; \
		fi; \
		LDVAL="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"; \
		echo "Building for $$GOOS/$$GOARCH (CGO_ENABLED=$$CGO)"; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=$$CGO GOARM=$$GOARM_VAL CC="$$CC_BIN" $(GOBUILD) -ldflags "$$LDVAL" -o $(BUILD_DIR)/$$BINARY_NAME_FULL $(MAIN_PATH) || exit 1; \
	done; \
	echo "HTTP multi-arch builds completed in $(BUILD_DIR)/"

# Wails builds: Windows, Linux, macOS; macOS packaged to DMG
.PHONY: build-wails
build-wails:
	@echo "Building Wails app for Windows, Linux, macOS..."
	@which wails >/dev/null 2>&1 || { echo "wails 未安装，请先安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest"; exit 1; }
	@wails build -platform windows/amd64,windows/arm64,linux/amd64,linux/arm64,darwin/universal,darwin/arm64,darwin/amd64
	@$(MAKE) prefix-wails-artifacts
	@$(MAKE) package-macos-dmg-all

.PHONY: build-wails-windows
build-wails-windows:
	@which wails >/dev/null 2>&1 || { echo "wails 未安装"; exit 1; }
	@wails build -platform windows/amd64,windows/arm64
	@$(MAKE) prefix-wails-artifacts

.PHONY: build-wails-linux
build-wails-linux:
	@which wails >/dev/null 2>&1 || { echo "wails 未安装"; exit 1; }
	@wails build -platform linux/amd64,linux/arm64
	@$(MAKE) prefix-wails-artifacts

.PHONY: build-wails-macos
build-wails-macos:
	@which wails >/dev/null 2>&1 || { echo "wails 未安装"; exit 1; }
	@$(MAKE) wails-build-macos-one P="darwin/universal" SUFFIX="darwin-universal"
	@$(MAKE) wails-build-macos-one P="darwin/arm64" SUFFIX="darwin-arm64"
	@$(MAKE) wails-build-macos-one P="darwin/amd64" SUFFIX="darwin-amd64"

# Prefix all Wails build artifacts in build/bin with 'wails-'
.PHONY: prefix-wails-artifacts
prefix-wails-artifacts:
	@echo "Prefixing Wails artifacts with 'wails-'..."
	@set -e; \
	shopt -s nullglob; \
	for f in build/bin/*; do \
		base=$$(basename "$$f"); \
		case "$$base" in \
			wails-*) ;; \
			*) \
				dst="$$(dirname "$$f")/wails-$$base"; \
				echo "→ $$base -> $$(basename "$$dst")"; \
				mv "$$f" "$$dst"; \
				;; \
		esac; \
	done; \
	shopt -u nullglob

# Package all macOS .app bundles to DMG (universal/arm64/amd64)
.PHONY: package-macos-dmg-all
package-macos-dmg-all:
	@echo "Packaging all macOS .app bundles into DMGs..."
	@set -e; \
	shopt -s nullglob; \
	for APP_PATH in build/bin/*.app; do \
		APP_NAME=$$(basename "$$APP_PATH" .app); \
		DMG_PATH="build/bin/$$APP_NAME.dmg"; \
		echo "→ Creating DMG for $$APP_NAME"; \
		hdiutil create -volname "$$APP_NAME" -srcfolder "$$APP_PATH" -ov -format UDZO "$$DMG_PATH" >/dev/null && echo "   DMG created: $$DMG_PATH"; \
	done; \
	shopt -u nullglob

# Android app via android/build_android.sh
.PHONY: build-android-app
build-android-app:
	@echo "Building Android app via android/build_android.sh..."
	@bash android/build_android.sh

# 帮助信息
.PHONY: help
help:
	@echo "Video Crawler Build System"
	@echo "=========================="
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build              - Build for current platform (CGO on)"
	@echo "  build-all          - Build for all platforms (CGO on for host only; FORCE_CGO=1 to enable for all)"
	@echo "  build-http-all     - Build HTTP service for Linux/macOS/Windows/Android (all archs)"
	@echo "  build-linux        - Build for Linux variants"
	@echo "  build-darwin       - Build for macOS variants"
	@echo "  build-windows      - Build for Windows variants"
	@echo "  build-android      - Build for Android variants (CGO off by default)"
	@echo "  build-android-app  - Build Android APK via android/build_android.sh"
	@echo "  build-wails        - Build Wails app for Win/Linux/macOS (universal, arm64, amd64) and package DMGs"
	@echo "  build-wails-windows- Build Wails app for Windows"
	@echo "  build-wails-linux  - Build Wails app for Linux"
	@echo "  build-wails-macos  - Build Wails app for macOS (universal, arm64, amd64) and package DMGs"
	@echo ""
	@echo "Development targets:"
	@echo "  run                - Run the application"
	@echo "  dev                - Run in development mode (frontend + backend)"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean              - Clean build files"
	@echo "  deps               - Download dependencies (Go + Node.js)"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run fmt and vet"
	@echo "  install            - Install to GOPATH/bin"
	@echo "  release            - Create release packages"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "Supported platforms:"
	@echo "  Linux:   amd64, arm64, 386, arm"
	@echo "  macOS:   universal, amd64, arm64"
	@echo "  Windows: amd64, 386, arm64"
	@echo "  Android: amd64, arm64, 386, arm"

# Build one macOS arch, prefix artifacts, rename .app with suffix, and package DMG
.PHONY: wails-build-macos-one
wails-build-macos-one:
	@echo "Building Wails macOS: $(P)"
	@wails build -platform $(P)
	@$(MAKE) prefix-wails-artifacts
	@set -e; \
	shopt -s nullglob; \
	# pick the most recent .app just produced (before suffixing) \
	LATEST_APP=$$(ls -1t build/bin/*.app 2>/dev/null | head -n 1); \
	if [ -z "$$LATEST_APP" ]; then echo "未找到 .app 产物"; exit 1; fi; \
	DIR=$$(dirname "$$LATEST_APP"); \
	BASE=$$(basename "$$LATEST_APP" .app); \
	# if not already suffixed with current $(SUFFIX), rename \
	case "$$BASE" in \
		*$(SUFFIX)) TARGET_APP="$$DIR/$$BASE.app" ;; \
		*) TARGET_APP="$$DIR/$$BASE-$(SUFFIX).app"; mv "$$LATEST_APP" "$$TARGET_APP" ;; \
	esac; \
	DMG_PATH="$${TARGET_APP%.app}.dmg"; \
	echo "→ Packaging DMG: $$(basename "$$DMG_PATH")"; \
	hdiutil create -volname "$$BASE" -srcfolder "$$TARGET_APP" -ov -format UDZO "$$DMG_PATH" >/dev/null && echo "   DMG created: $$DMG_PATH"; \
	shopt -u nullglob
