# 视频爬虫 (Video Crawler)

一个基于 Go 语言开发的视频爬虫服务，提供 HTTP API 接口来爬取网页内容，并包含完整的 Vue 3 前端管理系统。

## 功能特性

- 🚀 基于 `Gin` 框架的高性能 HTTP 服务器
- 🕷️ 可配置的网页爬虫
- 🔧 从 YAML 文件加载配置（默认 `configs/config.yaml`）
- 📊 健康检查接口
- 🎯 完整的 RESTful API
- 🌐 内置 Vue 3 前端应用（SPA）
- 📦 静态文件自动嵌入到二进制文件中
- 👥 用户管理系统
- 🎬 视频资源管理
- 📺 观影和观看历史功能
- 🔐 JWT 身份验证

## 项目结构

```
video-crawler/
├── cmd/video-crawler/     # 主程序入口
├── internal/              # 内部包
│   ├── app/              # 应用主逻辑（Gin 启动与路由）
│   ├── config/           # 配置管理（YAML 加载）
│   ├── handler/          # HTTP处理器（Gin）
│   ├── controllers/      # 控制器层
│   ├── services/         # 业务逻辑层
│   ├── entities/         # 数据实体
│   ├── middleware/       # 中间件
│   └── static/           # 静态文件处理（嵌入前端）
├── frontend/             # Vue 3 前端项目
│   ├── src/              # 前端源码
│   ├── dist/             # 构建输出（自动嵌入）
│   └── package.json      # 前端依赖
├── configs/              # 配置文件
│   ├── config.yaml       # 主配置文件
│   ├── users.json        # 用户数据
│   ├── video-source.json # 视频源配置
│   └── *.example.*       # 示例配置文件
├── pkg/                  # 可导出的包
├── docs/                 # 文档
├── scripts/              # 脚本文件
├── test/                 # 测试文件
├── go.mod               # Go模块文件
├── Makefile             # 构建脚本
└── README.md            # 项目说明
```

## 快速开始

### 环境要求

- Go 1.19 或更高版本
- Node.js 16 或更高版本（用于前端开发）

### 安装和运行

1. 克隆项目
```bash
git clone <your-repo-url>
cd video-crawler
```

2. 配置环境
```bash
# 复制示例配置文件
cp configs/config.example.yaml configs/config.yaml
cp configs/users.example.json configs/users.json
cp configs/video-source.example.json configs/video-source.json

# 编辑配置文件，设置你的 JWT 密钥和用户信息
```

3. 构建前端（可选，已预构建）
```bash
cd frontend && npm install && npm run build
```

4. 运行项目
```bash
# 使用 Makefile
make dev

# 或直接运行
go run cmd/video-crawler/main.go
```

5. 访问服务
- 前端应用: http://localhost:8080
- 健康检查: http://localhost:8080/health
- API 接口: http://localhost:8080/api

### 使用 Makefile

```bash
# 开发模式
make dev

# 构建所有平台
make build-all

# 构建前端
make build-frontend

# 清理构建文件
make clean

# 查看帮助
make help
```

## 主要功能

### 用户管理
- 用户注册和登录
- 管理员权限控制
- 登录历史记录

### 视频资源管理
- 视频源站点配置
- 爬虫规则设置
- 站点状态检查

### 观影功能
- 视频搜索
- 多站点类型支持
- 搜索历史记录

### 观看历史
- 观看进度记录
- 历史查询
- 用户隔离

## API 接口

### 用户相关
- `POST /api/user/register` - 用户注册
- `POST /api/user/login` - 用户登录
- `GET /api/user/list` - 获取用户列表（管理员）
- `GET /api/user/detail/:id` - 获取用户详情
- `PUT /api/user/save` - 保存用户信息

### 视频源管理
- `GET /api/video-source/list` - 获取视频源列表
- `GET /api/video-source/detail/:id` - 获取视频源详情
- `POST /api/video-source/save` - 保存视频源
- `DELETE /api/video-source/delete/:id` - 删除视频源
- `GET /api/video-source/check-status/:id` - 检查站点状态

### Lua 脚本实时调试与链式 HTML API（新增）
- 新增后端 Lua 引擎与长连接流式输出，`print`/`log` 均带毫秒时间戳，统一格式：`[PRINT][YYYY-MM-DD HH:mm:ss.SSS] ...`、`[LOG][...] ...`
- 新增接口：
  - `POST /api/lua/test`：HTTP chunked 流式输出
  - `POST /api/lua/test-sse`：SSE 方式输出
- 注入函数：`sleep(ms)`、HTTP 请求函数（`http_get`/`http_post`/`set_headers`/`set_cookies`/`set_user_agent`/`set_random_user_agent`）
- HTML 解析改为链式 API：
  - `parse_html(html) -> Document`
  - `Document:select(css)`、`Document:select_one(css)`、`Document:html()`、`Document:text()`
  - `Selection:select(css)`、`Selection:select_one(css)`、`Selection:first()`、`Selection:eq(i)`、`Selection:parent()`、`Selection:children()`、`Selection:next()`、`Selection:prev()`、`Selection:attr(name)`、`Selection:html()`、`Selection:text()`
- `http_get/http_post` 返回体 `body` 去除转义显示：JSON 自动解码再紧凑编码；非 JSON 直接按 UTF-8 输出，避免 `\\uXXXX`。

前端新增页面 `VideoSourceEditView.vue` 的 Lua 调试面板：
- 集成 Monaco Editor（本地资源加载，无需 CDN），主题与页面统一绿色风格。
- “打开文档/填充完整 Demo/脚本调试”三个按钮置于标题栏右侧，统一绿色主题。
- 调试输出支持自动滚动与不同前缀着色（PRINT/LOG/INFO/ERROR）。
- “保存/创建”按钮样式统一为绿色主按钮。

### 历史记录
- `GET /api/history/video` - 获取观看历史
- `GET /api/history/search` - 获取搜索历史
- `GET /api/history/login` - 获取登录历史

## 配置说明

### 主配置文件 (configs/config.yaml)
```yaml
server:
  host: "localhost"
  port: 8080
  jwt_secret: "your-jwt-secret-here"
  jwt_expire: 72
env: dev # 新增：dev 时日志输出到控制台
```

### 用户配置文件 (configs/users.json)
包含用户信息，密码使用 MD5 加密存储。

### 视频源配置 (configs/video-source.json)
包含爬虫规则和站点配置信息，定义了如何从不同视频站点提取数据。

## 开发

### 前端开发
```bash
cd frontend
npm install
npm run dev  # 开发模式
npm run build  # 构建生产版本
```

### 后端开发
```bash
go test ./...
go build -o bin/video-crawler cmd/video-crawler/main.go
```

### 完整构建流程
```bash
# 使用 Makefile
make build

# 或手动构建
cd frontend && npm run build
cd .. && go build -o bin/video-crawler cmd/video-crawler/main.go
```

## 部署

### 使用 Makefile 构建
```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 构建特定平台
make build-linux
make build-darwin
make build-windows
```

### Docker 部署（可选）
```bash
# 构建镜像
docker build -t video-crawler .

# 运行容器
docker run -p 8080:8080 -v ./configs:/app/configs video-crawler
```

## 安全注意事项

1. **配置文件安全**：
   - 不要将 `configs/users.json` 和 `configs/config.yaml` 提交到版本控制
   - 使用示例配置文件作为模板
   - 在生产环境中使用强密码和安全的 JWT 密钥

2. **数据安全**：
   - 历史记录文件包含用户敏感信息
   - 确保服务器文件系统权限正确设置

3. **网络安全**：
   - 在生产环境中使用 HTTPS
   - 配置适当的防火墙规则

## 许可证

MIT License
