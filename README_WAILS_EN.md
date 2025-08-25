# Wails Desktop Mode | [中文](README_WAILS.md) | [Back to English README](README_EN.md)

This project supports two modes:

- HTTP service (standalone backend + browser frontend)
- Wails desktop app (bundled frontend; backend runs as a local HTTP server on a random port)

## Highlights

- Reuses the same services (`internal/services/`) in both modes
- In Wails mode, a Gin HTTP server is started on a random port; the frontend calls the same REST APIs
- CORS enabled for WebView → local HTTP access
- Optional auth: when `auth.require_login=false`, all APIs skip JWT; UI hides user-related menus
- Unified data dir for `video-source.json`, `users.json`, and histories (search/video/login)

## Run

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Dev
$GOPATH/bin/wails dev

# Build
$GOPATH/bin/wails build
```

Make targets (if available):

```bash
make run-wails
make build-wails
make run-http
make build-http
```

## Backend Port

- Printed in logs, e.g. `Starting gin server on 0.0.0.0:57124`
- Also shown at the top of the in-app “API Docs” page

## API Docs (Wails menu)

- Open the left menu “API Docs” or visit route `/api-docs`
- Shows the current port; categorized API specs with body/response examples and curl/JS/Python snippets

## Config Files

- Prefer `VIDEO_CRAWLER_CONFIG_DIR/config.yaml` when set
- Fallback to `configs/config.yaml`
- Default `auth.require_login=false` when missing

## Integration Summary

- Dual entrypoints: Wails `main.go`, HTTP `cmd/http-server/main.go`
- Frontend env detection and unified API helper (`frontend/src/utils/api.ts`)
- Different route/menu behavior in no-login mode

### Unified API usage (TS)
```ts
import { configAPI, videoSourceAPI, scriptAPI } from '@/utils/api'
const cfg = await configAPI.getConfig()
const list = await videoSourceAPI.getList()
const save = await videoSourceAPI.saveVideoSource(data)
const lua  = await scriptAPI.testLua(script, method, params)
```

### Login control & routes
- No-login: hide user menus; login/register/user management → 404; others allowed
- Require-login: full auth and role checks
