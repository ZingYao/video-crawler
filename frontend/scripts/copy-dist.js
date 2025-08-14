#!/usr/bin/env node

import { copyFileSync, mkdirSync, readdirSync, statSync, existsSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// 获取项目根目录
const projectRoot = join(__dirname, '..', '..');
const sourceDir = join(__dirname, '..', 'dist');
const targetDir = join(projectRoot, 'internal', 'static', 'dist');

console.log('📋 开始拷贝前端构建文件...');
console.log(`源目录: ${sourceDir}`);
console.log(`目标目录: ${targetDir}`);

// 递归拷贝目录
function copyDir(src, dest) {
  // 创建目标目录
  if (!existsSync(dest)) {
    mkdirSync(dest, { recursive: true });
  }

  // 读取源目录
  const items = readdirSync(src);

  for (const item of items) {
    const srcPath = join(src, item);
    const destPath = join(dest, item);
    const stat = statSync(srcPath);

    if (stat.isDirectory()) {
      // 递归拷贝子目录
      copyDir(srcPath, destPath);
    } else {
      // 拷贝文件
      copyFileSync(srcPath, destPath);
      console.log(`✅ 已拷贝: ${item}`);
    }
  }
}

try {
  // 检查源目录是否存在
  if (!existsSync(sourceDir)) {
    console.error('❌ 源目录不存在:', sourceDir);
    process.exit(1);
  }

  // 删除目标目录（如果存在）
  if (existsSync(targetDir)) {
    const { rmSync } = await import('fs');
    rmSync(targetDir, { recursive: true, force: true });
    console.log('🗑️  已清理旧的目标目录');
  }

  // 拷贝文件
  copyDir(sourceDir, targetDir);
  
  console.log('✅ 前端构建文件拷贝完成！');
  console.log(`📁 目标位置: ${targetDir}`);
} catch (error) {
  console.error('❌ 拷贝失败:', error.message);
  process.exit(1);
}
