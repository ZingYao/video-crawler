import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [
    vue(),
    ...(mode === 'development' ? [vueDevTools()] : []),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: '0.0.0.0', // 监听所有IP地址，支持局域网访问
    port: 5173, // 明确指定端口
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  },
  // 仅生产环境移除 console/debugger；开发环境保留日志便于调试
  esbuild: {
    drop: mode === 'production' ? ['console', 'debugger'] : [],
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('video.js') || id.includes('@videojs-player')) return 'videojs'
            if (id.includes('ant-design-vue')) return 'antd'
            // 仅当访问编辑页时才会通过 loader.init() 加载 monaco，此处避免入口预加载
            if (id.includes('@guolao/vue-monaco-editor')) return 'monaco'
            if (id.includes('vue')) return 'vue'
            return 'vendor'
          }
        },
      }
    },
    // 生产环境使用 terser，通常体积更小
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        passes: 2,
        pure_getters: true,
      },
      mangle: true,
    },
    cssCodeSplit: true,
    sourcemap: false, // 不生成sourcemap
    reportCompressedSize: false, // 不报告压缩大小
    chunkSizeWarningLimit: 1000,
  }
}))
