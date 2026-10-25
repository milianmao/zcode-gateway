package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// buildCommit 由构建时 -X main.buildCommit=<sha> 注入，用于静态资源版本号
var buildCommit = "dev"

// assetVersion 返回静态资源查询串：构建变了 URL 就变，绕开浏览器/CDN 旧缓存
func assetVersion() string {
	if buildCommit == "" || buildCommit == "dev" {
		return ""
	}
	return "?v=" + buildCommit
}

//go:embed web
var webAssets embed.FS

func (s *Server) webHandler() http.Handler {
	sub, err := fs.Sub(webAssets, "web")
	if err != nil {
		return http.NotFoundHandler()
	}
	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/v1/") {
			http.NotFound(w, r)
			return
		}
		if !s.adminAuthed(r) {
			// Only the login page and its minimal assets are reachable
			// before authentication.
			if path == "/login.js" || path == "/style.css" {
				files.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			r.URL.Path = "/login.html"
		}
		if path == "/" || path == "/index.html" {
			raw, err := webAssets.ReadFile("web/index.html")
			if err == nil {
				version := assetVersion()
				html := strings.ReplaceAll(string(raw), `href="/style.css"`, `href="/style.css`+version+`"`)
				html = strings.ReplaceAll(html, `src="/app.js"`, `src="/app.js`+version+`"`)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-store")
				_, _ = w.Write([]byte(html))
				return
			}
		}
		if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".html") {
			// 面板资源是内嵌的，构建一变必须让浏览器/CDN 重新取
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
		}
		files.ServeHTTP(w, r)
	})
}
