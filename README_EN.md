# Video Crawler

[Chinese README](README.md) | English

A Go + Vue3 based visual video crawler / script runtime platform. It now supports both Lua (gopher-lua) and JavaScript (goja) engines, with real‑time debugging, chainable HTML parsing, Monaco editor with local assets, and chunked/SSE streaming outputs. The watch page uses Plyr + hls.js for HLS playback.

## Tech Stack (Updated)

- Backend (Go)
  - Gin (HTTP)
  - Native net/http crawler (returns *http.Response) with realistic browser headers; forwards frontend headers (skips Cookie/Host/Content-Length)
  - Lua engine (gopher-lua):
    - Streaming output with timestamps
    - Captures top-level `return` into `map[string]interface{}` and streams as `[RESULT]`
    - Injections:
      - HTTP: `http_get/http_post/set_headers/set_cookies/set_user_agent/set_random_user_agent/get_user_agent/set_ua_2_current_request_ua`
      - HTML chain: `parse_html` and selector helpers on Document/Selection
      - Utils: `sleep/trim/split/json_encode/json_decode`
    - Security: dangerous `io/os/package` functions disabled; only safe ones like `os.time/os.exit/os.clock` allowed with friendly messages
  - JavaScript engine (goja):
    - Synchronous `fetch(url, { method, headers, body, timeout, redirect })`
      - Response: `ok/status/statusText/url/headers/text()/json()/arrayBuffer()/clone()`
      - Headers: `get/has/keys/values/entries/forEach`
    - HTTP & UA helpers (camelCase): `httpGet/httpPost/setHeaders/setCookies/setUserAgent/setRandomUserAgent/getUserAgent/setUaToCurrentRequestUa`
    - DOM parsing via goquery: `parseHtml(html)` → Document/Element with `querySelector/querySelectorAll/getElementById/getElementsByTagName/getElementsByClassName/text/html/attr/innerText/innerHTML/getAttribute`
    - Full `console` API (`log/info/warn/error/debug/trace/time/timeEnd/assert/group/groupCollapsed/groupEnd/count/countReset/table/dir/dirxml/clear`) with streaming back to frontend
    - Security sandbox: no `os/fs/child_process` or local file access
  - goquery (HTML parsing)
  - github.com/lib4u/fake-useragent (random UA)
  - Logs: print to console when `env=dev`

- Frontend (Vue3 + TS + Vite)
  - Ant Design Vue, Pinia
  - @guolao/vue-monaco-editor with local Monaco assets (no CDN), lazy-loaded; absolute worker paths fixed
  - Player: Plyr + hls.js, with HLS, playback rate select (mobile-friendly), next/prev episode, source/episode tabs, progress persistence/resume, auto screen orientation in fullscreen on mobile
  - Unified green theme; editor/logs side-by-side with draggable splitter + persistence; colored logs; clear logs; F5 debug; block Cmd/Ctrl+S
  - Global interceptor: when `code === 6`, show "Login expired", logout, then redirect to login after delay
  - Docs drawer: LuaDocs & JSDocs, auto-switch by selected engine

## Structure (Simplified)

```
video-crawler/
├── cmd/video-crawler/
├── internal/
│   ├── app/              # Gin bootstrap & routes
│   ├── config/           # config (with env)
│   ├── handler/          # route registration
│   ├── controllers/      # include /api/lua/test & /test-sse
│   ├── services/         # Lua execution & strict output ordering
│   ├── crawler/          # native net/http + realistic headers + random UA
│   ├── lua/              # Lua engine (injections, chain HTML, return capture)
│   ├── entities/ middleware/ utils/ ...
│   └── static/dist/      # bundled frontend
├── frontend/
│   ├── src/
│   ├── public/monaco/
│   └── scripts/copy-monaco.js
├── configs/
├── scripts/
├── Makefile
└── README.md / README_EN.md
```

## Quick Start

1) Init configs
```bash
cp configs/config.example.yaml configs/config.yaml
cp configs/users.example.json configs/users.json
cp configs/video-source.example.json configs/video-source.json
```

2) Build frontend (local Monaco assets)
```bash
cd frontend && npm install && npm run build
```

3) Start backend
```bash
make dev
# or
go run cmd/video-crawler/main.go
```

4) Visit
- Web: http://localhost:8080
- API: http://localhost:8080/api

5) Playback / Watch

- Search page: input keyword; search only sites with "normal" status
- Result card: whole card is clickable to start watching; keep "Original Site" button
- Watch page: auto select first source & first episode if no cache; support switching source/episode, auto play and resume progress

## Debug APIs

- Lua (Chunked): `POST /api/lua/test`
- Lua (SSE): `POST /api/lua/test-sse`
- JavaScript (Chunked): `POST /api/js/test`
- JavaScript (SSE): `POST /api/js/test-sse`

Body:
```json
{ "script": "print('hello'); return { x = 1 }" }
```

Output order guaranteed: `[INFO]` → Lua `[PRINT]/[LOG]` → `[RESULT]` → final `[INFO]`.

## Frontend Editing Page

- Fields: site name, domain, sort, source type, crawler engine (Lua/JavaScript), status
- Editor switches language & docs by engine; provides default template and Demo for both Lua and JS; switching language can auto fill Demo when no existing code for that language
- Required functions: `search_video`, `get_video_detail`, `get_play_video_detail`
- Drafts: auto-save; restore/delete prompt (double confirm for delete)
- UX: F5 debug, block Cmd/Ctrl+S, resizable panes with persistence, clear logs, auto-scroll

### JavaScript Script Guidelines

- Global methods (camelCase): `httpGet`, `httpPost`, `setHeaders`, `setCookies`, `setUserAgent`, `setRandomUserAgent`, `getUserAgent`, `setUaToCurrentRequestUa`, `fetch`
- DOM: `parseHtml(html)` → `Document`/`Element` with `querySelector/querySelectorAll/.../text/html/attr` helpers
- Console: full `console` API; output streams back to the debug panel
- Demo: the "Fill Demo" button contains examples calling all provided APIs

## Config

`configs/config.yaml` key:
```yaml
env: dev  # print logs to console in dev
```

## License

MIT License

---

## Player Interaction Optimizations (Designed for Chinese user habits)

To improve mobile and touch scenarios, we tuned the video playback interactions (both Plyr and native HTML5 video):

- Double-click to Play/Pause: toggle playback on double click (replaces double-click seek).
- Long-press 2x: press and hold for 500ms to enter 2x speed, release to restore; mutually exclusive with progress dragging to avoid mis-touches.
- Horizontal drag to seek: supports repeated back-and-forth; ignore when vertical movement exceeds 1/4 of container height or when starting area is within top/bottom 1/6 (prevents pulling notification bar or system gestures).
- Persistent progress bar during drag: once dragging starts, keep the bar visible until finger lifts; temporary pauses while holding will not hide the bar.
- Mobile control simplification: hide volume slider on mobile; keep mute and speed settings for a cleaner UI.

All optimizations apply to both Plyr and native video to deliver a consistent mobile experience across engines.