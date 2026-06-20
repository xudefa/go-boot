// Package docs 提供 API 文档功能。
package docs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DocsHandler 文档处理器
type DocsHandler struct {
	title       string
	description string
	version     string
	basePath    string
}

// NewDocsHandler 创建文档处理器
func NewDocsHandler(title, description, version string) *DocsHandler {
	return &DocsHandler{
		title:       title,
		description: description,
		version:     version,
	}
}

// SwaggerUIHandler Swagger UI handler
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

// RedocHandler ReDoc handler
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

// SwaggerJSONHandler Swagger JSON documentation handler
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

// RegisterRoutes registers documentation routes to the HTTP router
func (h *DocsHandler) RegisterRoutes(mux interface {
	Handle(pattern string, handler http.Handler)
}) {
	mux.Handle("/docs/", http.HandlerFunc(h.SwaggerUIHandler))
	mux.Handle("/docs/swagger", http.HandlerFunc(h.SwaggerUIHandler))
	mux.Handle("/docs/redoc", http.HandlerFunc(h.RedocHandler))
	mux.Handle("/docs/swagger.json", http.HandlerFunc(h.SwaggerJSONHandler))
}
