// 示例: go-boot/ingress/gin Gin 框架使用指南
//
// 本示例演示 go-boot/ingress/gin 包的核心功能:
//
// 1. 创建 Gin HTTP 服务器
// 2. 配置服务器参数
// 3. 注册中间件
// 4. 注册路由处理函数
//
// 运行方式:
//
//	cd examples/gin && go run .
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginframe "github.com/xudefa/go-boot/ingress/gin"
	"github.com/xudefa/go-boot/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	fmt.Println("=== Gin Example ===")

	basicExample()

	fmt.Println("=== Gin Example ===")
	return nil
}

func basicExample() {
	fmt.Println("--- Basic Gin Example ---")

	config := &ginframe.GinConfig{
		HostPorts:    ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Mode:         "debug",
	}

	container := ginframe.New(nil, config, &nopLogger{})

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.POST("/api/user", func(c *gin.Context) {
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":    1,
			"name":  req.Name,
			"email": req.Email,
		})
	})

	_ = container
	_ = router

	fmt.Println("Gin server configured at http://localhost:8080")
	fmt.Println("Try: curl http://localhost:8080/hello")
	fmt.Println("Try: curl http://localhost:8080/health")
	fmt.Println("Try: curl -X POST -H 'Content-Type: application/json' -d '{\"name\":\"test\",\"email\":\"test@example.com\"}' http://localhost:8080/api/user")

	fmt.Println("--- Basic Gin Example ---")
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
