# Makefile Guide (English) | [中文](MAKEFILE_README.md) | [Back to README](README_EN.md)

The Makefile unifies frontend and backend build, run and packaging workflows. It supports two build lines: HTTP server and Wails desktop, and provides multi‑platform binaries.

## Conventions
- Frontend build: All backend build targets depend on `frontend` bundle (`npm run build`).
- Artifacts directory: `bin/`.
- Version info: Injects `Version/BuildTime/GitCommit` via ldflags.

## Targets Overview

- Basics
  - `deps`: Install Go & frontend deps
  - `fmt` / `vet` / `lint`: Format & static checks
  - `clean`: Clean artifacts

- Run
  - `run-http`: Run standalone HTTP server (dev)
  - `run-wails`: Wails dev mode (hot reload)
  - `dev`: Frontend dev + backend main.go (dev only)

- Build (current platform)
  - `build-http`: HTTP server binary (current platform)
  - `build-wails`: Wails desktop app (current platform)

- Matrix builds
  - `build-http-all`: HTTP server for common Linux/macOS/Windows archs
    - linux/amd64, linux/arm64, linux/386, linux/arm
    - darwin/amd64, darwin/arm64
    - windows/amd64, windows/386, windows/arm64
  - `build-wails-all`: Wails desktop for common desktop platforms (Linux/macOS/Windows)
  - `build-wails-ios` / `build-wails-android`: Mobile placeholders printing required SDK notes (see Wails README)

- Release
  - `release`: Packages multi‑platform HTTP server artifacts (zip/tar.gz)

## Typical commands
```bash
# Install deps
make deps

# HTTP server: current platform
make build-http

# HTTP server: multi-platform
make build-http-all

# Wails desktop: current platform
make build-wails

# Wails desktop: desktop platforms
make build-wails-all

# Run Wails dev
make run-wails

# Run HTTP dev
make run-http

# Create release packages (HTTP)
make release
```

## Notes
- iOS/Android Wails builds require local SDK/NDK & signing toolchains. The Makefile provides entry points only; see `README_WAILS_EN.md` for details.
- In CI/CD, call these targets to produce artifacts for multiple platforms.
