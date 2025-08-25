# Release Notes (master)

- Tag: v1.0
- Date: 2025-08-25
- Form: HTTP service only (no Wails desktop integration)

## Key Features
- Unified REST API via Gin (`/api/...`), with health/info endpoints
- Video source management: list/detail, save/delete, status check & manual set
- Script debugging (Lua/JS): streamed output (Chunked/SSE) with basic logs
- Frontend (Vue3 + Vite): source edit/manage, watch/history, unified styles
- Auth (optional): JWT, router guards & permissions

## Build & Distribution
- Makefile (HTTP service)
  - Single platform: `make build`
  - Multi-platform: `make build-all`
    - Linux: amd64/arm64/386/arm
    - macOS: amd64/arm64
    - Windows: amd64/386/arm64
    - Android: amd64/arm64 (optional)
  - Frontend build is an automatic dependency of backend builds (`build-frontend`)
- Artifact naming: `bin/video-crawler-{os}-{arch}[.exe]`

## Config & Run
- Config file: `configs/config.yaml` (contains `env`, `auth.require_login`, etc.)
- Run:
  - Combined dev: `make dev`
  - Backend only: `go run cmd/video-crawler/main.go`
  - Frontend only: `cd frontend && npm run dev`
- Proxy (optional):
  ```bash
  export https_proxy=http://127.0.0.1:7897 \
         http_proxy=http://127.0.0.1:7897 \
         all_proxy=socks5://127.0.0.1:7897
  ```

## Changelog
- feat: Video source management & script debug APIs
- feat: Frontend editing page & basic watch/history
- feat: Makefile multi-platform build & release scaffolding
- chore/docs: Structure and base README improvements
