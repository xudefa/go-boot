// Package gin 基于 Gin 框架提供 HTTP 服务器实现。
//
// 该包将 Gin 框架与 go-boot 容器系统集成，
// 支持依赖注入和统一日志记录。
package gin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/ingress"
	"github.com/xudefa/go-boot/log"
)

const ServerBeanId = "gin"

type Gin struct {
	container   core.Container
	logger      log.Logger
	ginConfig   *GinConfig
	router      *gin.Engine
	middlewares []gin.HandlerFunc
	initialized bool
}

type GinConfig struct {
	HostPorts    string        `mapstructure:"hostPorts" default:":8080" json:"hostPorts"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout" default:"10s" json:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout" default:"10s" json:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout" default:"60s" json:"idleTimeout"`
	Mode         string        `mapstructure:"mode" default:"debug" json:"mode"`
}

func New(container core.Container, ginConfig *GinConfig, l log.Logger) *Gin {
	return &Gin{
		container:   container,
		ginConfig:   ginConfig,
		middlewares: make([]gin.HandlerFunc, 0),
		logger:      l,
	}
}

var _ ingress.Server = (*Gin)(nil)

func (g *Gin) Run() error {
	if g.initialized {
		return fmt.Errorf("server already initialized")
	}

	gin.SetMode(g.ginConfig.Mode)
	g.router = gin.New()

	g.router.Use(LoggerMiddleware(g.logger))
	g.router.Use(RecoveryMiddleware(g.logger))

	for _, m := range g.middlewares {
		g.router.Use(m)
	}

	g.initialized = true

	err := g.container.Register(ServerBeanId, core.Bean(g.router))
	if err != nil {
		g.logger.Error(context.Background(), "gin register failed")
		return err
	}

	go func() {
		srv := &http.Server{
			Addr:         g.ginConfig.HostPorts,
			Handler:      g.router,
			ReadTimeout:  g.ginConfig.ReadTimeout,
			WriteTimeout: g.ginConfig.WriteTimeout,
			IdleTimeout:  g.ginConfig.IdleTimeout,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Error(context.Background(), "gin server run failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return nil
}

func (g *Gin) Use(m any) ingress.Server {
	if handler, ok := m.(gin.HandlerFunc); ok {
		g.middlewares = append(g.middlewares, handler)
	}
	return g
}

func (g *Gin) UseGlobal(m any) ingress.Server {
	if handler, ok := m.(gin.HandlerFunc); ok {
		g.middlewares = append([]gin.HandlerFunc{handler}, g.middlewares...)
	}
	return g
}

func (g *Gin) Register(fn func(core.Container) error) ingress.Server {
	if err := fn(g.container); err != nil {
		panic(fmt.Errorf("register handler failed: %w", err))
	}
	return g
}

func (g *Gin) Container() core.Container {
	return g.container
}

func LoggerMiddleware(l log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		l.Info(c.Request.Context(), "receive request",
			log.KeyValue{Key: "method", Value: method},
			log.KeyValue{Key: "path", Value: path},
			log.KeyValue{Key: "status", Value: c.Writer.Status()},
			log.KeyValue{Key: "time", Value: time.Since(start)},
		)
	}
}

func RecoveryMiddleware(l log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				l.Error(c.Request.Context(), "recovered from panic",
					log.KeyValue{Key: "error", Value: r},
				)
				c.String(500, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
