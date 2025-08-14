package static

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var staticFiles embed.FS

// GetStaticFS 获取静态文件系统
func GetStaticFS() http.FileSystem {
	// 直接将 dist 目录作为静态文件根目录
	// 返回 dist 目录下的文件系统
	fs, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		panic("无法获取 dist 目录: " + err.Error())
	}
	return http.FS(fs)
}

// ServeStatic 处理静态文件请求（保留用于兼容性）
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	// 获取请求路径
	requestPath := r.URL.Path

	// 如果路径是根路径，返回 index.html
	if requestPath == "/" {
		requestPath = "/index.html"
	}

	// 构建文件路径
	filePath := "dist" + requestPath

	// 尝试从静态文件系统获取文件
	file, err := staticFiles.Open(filePath)
	if err != nil {
		// 如果文件不存在，返回 index.html（SPA 路由）
		file, err = staticFiles.Open("dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 设置正确的 Content-Type
	contentType := getContentType(requestPath)
	if contentType != "" {
		w.Header().Add("Content-Type", contentType)
	}

	// 返回文件内容
	w.Write(content)
}

// getContentType 根据文件扩展名获取 Content-Type
func getContentType(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".ico":
		return "image/x-icon"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return ""
	}
}
