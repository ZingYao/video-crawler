#!/bin/bash

# 视频聚合项目构建脚本

set -e

echo "🚀 开始构建视频聚合项目..."

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 构建前端（会自动拷贝到后端）
echo "📦 构建前端..."
cd frontend
if [ ! -d "node_modules" ]; then
    echo "📥 安装前端依赖..."
    npm install
fi
npm run build
cd ..

# 验证前端文件是否已拷贝
if [ ! -d "internal/static/dist" ]; then
    echo "❌ 前端构建文件未找到，手动拷贝..."
    rm -rf internal/static/dist
    cp -r frontend/dist internal/static/
fi

# 构建后端
echo "🔨 构建后端..."
go build -o bin/video-crawler cmd/video-crawler/main.go

echo "✅ 构建完成！"
echo "📁 二进制文件位置: bin/video-crawler"
echo "🌐 运行命令: ./bin/video-crawler"
