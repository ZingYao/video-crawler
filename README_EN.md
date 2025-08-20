# Video Crawler

中文说明 | [English README](README_EN.md)

A Go + Vue3 based visual video crawler / Lua runtime platform. It features real‑time Lua debugging, chainable HTML parsing, Monaco editor with local assets, and chunked/SSE streaming outputs.

## Tech Stack (Updated)

- Backend (Go)
  - Gin (HTTP)
  - Native net/http crawler (returns *http.Response) with realistic browser headers; forwards frontend headers (skips Cookie/Host/Content-Length)
  - gopher-lua (Lua engine):
    - Streaming output channel with timestamped `print`/`log`
    - Captures top-level `return` into `map[string]interface{}` and streams formatted JSON as `[RESULT]` in order
    - Injected functions:
      - HTTP: `http_get`, `http_post`, `set_headers`, `set_cookies`, `set_user_agent`, `set_random_user_agent`, `get_user_agent`, `set_ua_2_current_request_ua`
      - HTML chain: `parse_html`, and on Document/Selection: `select`, `select_one`, `first`, `eq`, `parent`, `children`, `next`, `prev`, `attr`, `text`, `html`
      - Utils: `sleep(ms)`, `split(s, sep)`, `trim(s)`, `json_encode(value, indent?)` (boolean/number/string indent), `json_decode(json)`
    - Auto-decompress response body (gzip/deflate)
  - goquery (HTML parsing; chain API exposed to Lua via userdata)
  - github.com/lib4u/fake-useragent (random UA)
  - Logs: print to console when `env=dev`

- Frontend (Vue3 + TS + Vite)
  - Ant Design Vue, Pinia
  - @guolao/vue-monaco-editor with local Monaco assets (no CDN)
  - Unified green theme; editor/logs side-by-side with draggable splitter + persistence; colored logs; clear logs; F5 debug; block Cmd/Ctrl+S
  - Global interceptor: when `code === 6`, show "Login expired", logout, then redirect to login after delay
  - Lua docs drawer (right side, no mask, editable simultaneously)

## Structure

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

## Lua Debug APIs

- Chunked: `POST /api/lua/test`
- SSE: `POST /api/lua/test-sse`

Body:
```json
{ "script": "print('hello'); return { x = 1 }" }
```

Output order guaranteed: `[INFO]` → Lua `[PRINT]/[LOG]` → `[RESULT]` → final `[INFO]`.

## Frontend Editing Page

- Fields: site name, domain, sort, source type
- Lua editor: default template; required function checks (`search_video`, `get_video_detail`, `get_play_video_detail`)
- Drafts: auto-save; restore/delete prompt (double confirm for delete)
- UX: F5 debug, block Cmd/Ctrl+S, resizable panes with persistence, clear logs, auto-scroll

## Config

`configs/config.yaml` key:
```yaml
env: dev  # print logs to console in dev
```

## License

MIT License

---

Looking for Chinese docs? See [README.md](README.md).
