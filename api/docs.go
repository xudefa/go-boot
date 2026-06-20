// Package api 提供 API 文档功能，支持 Swagger UI 和 ReDoc 两种文档展示方式。
//
// # 功能特性
//
//   - Swagger UI：交互式 API 文档界面
//   - ReDoc：响应式 API 文档展示
//   - OpenAPI 3.0 规范 JSON 输出
//   - 可配置的应用信息（标题、描述、版本）
//
// # 使用示例
//
//	handler := api.NewDocsHandler("My API", "API Documentation", "1.0.0")
//	handler.SetBasePath("http://localhost:8080")
//
//	mux := http.NewServeMux()
//	handler.RegisterRoutes(mux)
//
//	http.ListenAndServe(":8080", mux)
//
// 访问以下路径查看文档：
//   - /docs/ 或 /docs/swagger：Swagger UI 界面
//   - /docs/redoc：ReDoc 界面
//   - /docs/swagger.json：OpenAPI 规范 JSON
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DocsHandler API 文档处理器
//
// 提供 Swagger UI、ReDoc 和 OpenAPI JSON 文档的 HTTP 处理。
// 支持自定义应用标题、描述、版本和基础路径。
type DocsHandler struct {
	title       string // 文档标题
	description string // 文档描述
	version     string // API 版本
	basePath    string // 基础路径
}

// NewDocsHandler 创建文档处理器
//
// 参数:
//   - title: API 文档标题
//   - description: API 文档描述
//   - version: API 版本号
//
// 返回值:
//   - *DocsHandler: 文档处理器实例
//
// 示例:
//
//	handler := api.NewDocsHandler("My API", "API Documentation", "1.0.0")
func NewDocsHandler(title, description, version string) *DocsHandler {
	return &DocsHandler{
		title:       title,
		description: description,
		version:     version,
	}
}

// SetBasePath 设置 API 基础路径
//
// 参数:
//   - basePath: API 基础路径，如 "http://localhost:8080"
func (h *DocsHandler) SetBasePath(basePath string) {
	h.basePath = basePath
}

// SwaggerUIHandler Swagger UI 文档处理函数
//
// 渲染 Swagger UI 交互式文档界面，从 /docs/swagger.json 加载 OpenAPI 规范。
//
// 示例:
//
//	mux.HandleFunc("/docs", handler.SwaggerUIHandler)
func (h *DocsHandler) SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>` + h.title + `</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@latest/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@latest/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/docs/swagger.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(swaggerHTML)); err != nil {
		fmt.Printf("[go-boot] failed to write swagger HTML: %v\n", err)
	}
}

// RedocHandler ReDoc 文档处理函数
//
// 渲染 ReDoc 响应式文档界面，从 /docs/swagger.json 加载 OpenAPI 规范。
//
// 示例:
//
//	mux.HandleFunc("/docs/redoc", handler.RedocHandler)
func (h *DocsHandler) RedocHandler(w http.ResponseWriter, r *http.Request) {
	redocHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>` + h.title + `</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://unpkg.com/redoc@latest/bundles/redoc.standalone.js"></script>
</head>
<body>
    <div id="redoc-container"></div>
    <script>
        Redoc.init('/docs/swagger.json', {}, document.getElementById('redoc-container'));
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(redocHTML)); err != nil {
		fmt.Printf("[go-boot] failed to write redoc HTML: %v\n", err)
	}
}

// SwaggerJSONHandler OpenAPI 规范 JSON 处理函数
//
// 返回 OpenAPI 3.0 规范的 JSON 文档，包含应用信息和空的 paths/components 结构。
// 可被 Swagger UI、ReDoc 等工具消费。
//
// 示例:
//
//	mux.HandleFunc("/docs/swagger.json", handler.SwaggerJSONHandler)
func (h *DocsHandler) SwaggerJSONHandler(w http.ResponseWriter, r *http.Request) {
	swaggerSpec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       h.title,
			"description": h.description,
			"version":     h.version,
		},
		"servers": []map[string]string{
			{
				"url": h.basePath,
			},
		},
		"paths": make(map[string]interface{}),
		"components": map[string]interface{}{
			"schemas": make(map[string]interface{}),
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(swaggerSpec); err != nil {
		fmt.Printf("[go-boot] failed to encode swagger JSON: %v\n", err)
	}
}

// RegisterRoutes 注册文档路由到 HTTP 路由器
//
// 自动注册以下路由：
//   - /docs/ 或 /docs/swagger：Swagger UI 界面
//   - /docs/redoc：ReDoc 界面
//   - /docs/swagger.json：OpenAPI 规范 JSON
//
// 参数:
//   - mux: HTTP 路由器，需实现 Handle(pattern string, handler http.Handler) 方法
//
// 示例:
//
//	mux := http.NewServeMux()
//	handler.RegisterRoutes(mux)
func (h *DocsHandler) RegisterRoutes(mux interface {
	Handle(pattern string, handler http.Handler)
}) {
	mux.Handle("/docs/", http.HandlerFunc(h.SwaggerUIHandler))
	mux.Handle("/docs/swagger", http.HandlerFunc(h.SwaggerUIHandler))
	mux.Handle("/docs/redoc", http.HandlerFunc(h.RedocHandler))
	mux.Handle("/docs/swagger.json", http.HandlerFunc(h.SwaggerJSONHandler))
}
