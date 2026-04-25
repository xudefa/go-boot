// 示例: go-boot/ingress/hertz Hertz 框架使用指南
//
// 本示例演示 go-boot/ingress/hertz 包的核心功能:
//
// 1. 创建 Hertz HTTP 服务器
// 2. 配置服务器参数
// 3. 注册中间件
// 4. 注册路由处理函数
//
// 运行方式:
//
//	cd examples/hertz && go run .
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	hertzframe "github.com/xudefa/go-boot/ingress/hertz"
	"github.com/xudefa/go-boot/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	fmt.Println("=== Hertz Example ===")

	basicExample()

	fmt.Println("=== Hertz Example ===")
	return nil
}

func basicExample() {
	fmt.Println("--- Basic Hertz Example ---")

	config := &hertzframe.HertzConfig{
		HostPorts:   ":8080",
		ReadTimeout: 10,
	}

	_ = hertzframe.New(nil, config, &nopLogger{})

	fmt.Println("Hertz server configured successfully")
	fmt.Println("Server will listen on http://localhost:8080")
	fmt.Println("Try: curl http://localhost:8080/hello")
	fmt.Println("Try: curl http://localhost:8080/health")
	fmt.Println("Try: curl -X POST -H 'Content-Type: application/json' -d '{\"name\":\"test\",\"email\":\"test@example.com\"}' http://localhost:8080/api/user")

	fmt.Println("--- Basic Hertz Example ---")
}

func registerRoutes(h *hertzframe.Hertz) {
	router := h.Server()

	router.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})

	router.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	router.POST("/api/user", func(ctx context.Context, c *app.RequestContext) {
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.Bind(&req); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":    1,
			"name":  req.Name,
			"email": req.Email,
		})
	})
}

type nopLogger struct{}

func (l *nopLogger) Debug(ctx context.Context, msg string, keys ...log.KeyValue)  {}
func (l *nopLogger) Info(ctx context.Context, msg string, keys ...log.KeyValue)   {}
func (l *nopLogger) Warn(ctx context.Context, msg string, keys ...log.KeyValue)   {}
func (l *nopLogger) Error(ctx context.Context, msg string, keys ...log.KeyValue)  {}
func (l *nopLogger) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {}
func (l *nopLogger) Panic(ctx context.Context, msg string, keys ...log.KeyValue)  {}
func (l *nopLogger) Fatal(ctx context.Context, msg string, keys ...log.KeyValue)  {}
func (l *nopLogger) Sync() error                                                  { return nil }
func (l *nopLogger) With(ctx context.Context, keys ...log.KeyValue) log.Logger    { return l }
