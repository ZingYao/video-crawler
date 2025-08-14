# 视频爬虫前端

基于 Vue 3 + TypeScript + Vite 构建的现代化前端应用。

## 功能特性

- 🎨 现代化的用户界面设计
- 🔐 用户认证系统（登录/登出）
- 📱 响应式设计，支持移动端
- 🚀 基于 Vite 的快速开发体验
- 📦 TypeScript 支持
- 🎯 Pinia 状态管理
- 🛣️ Vue Router 路由管理

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **样式**: CSS3 + Flexbox/Grid
- **代码规范**: ESLint + Prettier

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:5173

### 构建生产版本

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

## 项目结构

```
frontend/
├── src/
│   ├── assets/          # 静态资源
│   ├── components/      # 公共组件
│   ├── router/          # 路由配置
│   ├── stores/          # Pinia 状态管理
│   ├── views/           # 页面组件
│   ├── App.vue          # 根组件
│   └── main.ts          # 入口文件
├── public/              # 公共文件
├── package.json         # 依赖配置
└── vite.config.ts       # Vite 配置
```

## 页面路由

- `/` - 首页（需要登录）
- `/login` - 登录页面
- `/register` - 注册页面

## 认证系统

### 默认登录信息
- 用户名: `admin`
- 密码: `admin`

### 注册功能
- 支持新用户注册
- 密码长度至少6位
- 用户名长度至少3位
- 注册后需要管理员批准才能登录

### 功能特性
- 自动登录状态保持
- 路由守卫保护
- 本地存储支持
- 真实API接口集成
- 前端密码MD5加密保护

## API 集成

前端集成了完整的后端 API：
- 用户登录: `POST /api/user/login`
- 用户注册: `POST /api/user/register`
- 健康检查: `GET /health`
- API 信息: `GET /api`

## 开发说明

1. 确保后端服务运行在 `http://localhost:8080`
2. 前端开发服务器运行在 `http://localhost:5173`
3. 支持热重载和快速开发
