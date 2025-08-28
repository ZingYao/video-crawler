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
    // 检查是否在Wails环境中
    var isWails = process.env.WAILS_ENV === 'true' || process.env.NODE_ENV === 'production';
    return {
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
                    rewrite: function (path) { return path.replace(/^\/api/, '/api'); },
                    configure: function (proxy, options) {
                        proxy.on('proxyReq', function (proxyReq, req, res) {
                            // 检测当前请求的域名
                            var host = req.headers.host || '';
                            var referer = req.headers.referer || '';
                            console.log('代理请求信息:', {
                                host: host,
                                referer: referer,
                                url: req.url,
                                method: req.method
                            });
                            // 如果是Wails环境（wails://协议），尝试获取后端端口
                            if (host.includes('wails.localhost') || referer.includes('wails://')) {
                                console.log('检测到Wails环境，尝试获取后端端口');
                                // 这里我们可以尝试从Wails API获取端口
                                // 但由于这是服务器端的代理配置，我们无法直接调用Wails API
                                // 所以这里只是记录信息，实际的端口获取在客户端进行
                            }
                        });
                        proxy.on('error', function (err, req, res) {
                            console.log('代理错误:', {
                                error: err.message,
                                host: req.headers.host,
                                url: req.url
                            });
                        });
                    }
                }
            }
        },
        // 临时开启 console 方法用于调试
        esbuild: {
            drop: [], // 临时禁用 console 过滤
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
                },
            },
            // 生产环境使用 terser，通常体积更小
            minify: 'terser',
            terserOptions: {
                compress: {
                    drop_console: false, // 临时禁用 console 过滤
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
    };
});
