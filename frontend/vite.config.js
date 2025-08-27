var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
        if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
        }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
};
import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import vueDevTools from 'vite-plugin-vue-devtools';
// https://vite.dev/config/
export default defineConfig(function (_a) {
    var mode = _a.mode;
    return ({
        plugins: __spreadArray([
            vue()
        ], (mode === 'development' ? [vueDevTools()] : []), true),
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
                    rewrite: function (path) { return path.replace(/^\/api/, '/api'); }
                }
            }
        },
        // 临时保留console日志用于调试
        esbuild: {
            drop: [], // 临时保留所有console日志
        },
        build: {
            rollupOptions: {
                output: {
                    manualChunks: function (id) {
                        if (id.includes('node_modules')) {
                            if (id.includes('video.js') || id.includes('@videojs-player'))
                                return 'videojs';
                            if (id.includes('ant-design-vue'))
                                return 'antd';
                            // 仅当访问编辑页时才会通过 loader.init() 加载 monaco，此处避免入口预加载
                            if (id.includes('@guolao/vue-monaco-editor'))
                                return 'monaco';
                            if (id.includes('vue'))
                                return 'vue';
                            return 'vendor';
                        }
                    },
                }
            },
            // 生产环境使用 terser，通常体积更小
            minify: 'terser',
            terserOptions: {
                compress: {
                    drop_console: false, // 临时保留console日志
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
    });
});
