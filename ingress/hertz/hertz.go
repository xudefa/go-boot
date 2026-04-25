// Package hertz 基于 CloudWeGo Hertz 框架提供 HTTP 服务器实现。
//
// 该包将 Hertz 框架与 go-boot 容器系统集成，
// 支持依赖注入和统一日志记录。
package hertz

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/network/netpoll"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/ingress"
	"github.com/xudefa/go-boot/log"
)

// ServerBeanId 是在容器中注册 Hertz 服务器的 Bean 标识符。
const ServerBeanId = "hertz"

// Hertz 是基于 CloudWeave Hertz 框架的 HTTP 服务器。
//
// 它将 Hertz 服务器与 go-boot 容器集成，
// 提供依赖注入和中间件管理功能。
type Hertz struct {
	container   core.Container
	logger      log.Logger
	hertzConfig *HertzConfig
	server      *server.Hertz
	middlewares []app.HandlerFunc
	initialized bool
}

// HertzConfig 定义 Hertz 服务器的配置选项。
//
// 支持通过配置文件使用 mapstructure 标签进行配置。
type HertzConfig struct {
	// HostPorts 指定服务器监听的地址，格式为 "host:port"。
	HostPorts string `mapstructure:"hostPorts" default:"0.0.0.0:8080" json:"hostPorts"`
	// Transport 选择网络传输实现。
	Transport string `mapstructure:"transport" json:"transport"`

	// ReadTimeout 设置读取整个请求的最大时长。
	ReadTimeout time.Duration `mapstructure:"readTimeout" default:"10s" json:"readTimeout"`
	// WriteTimeout 设置写入响应的最大时长。
	WriteTimeout time.Duration `mapstructure:"writeTimeout" default:"10s" json:"writeTimeout"`
	// IdleTimeout 设置空闲连接的最大时长。
	IdleTimeout time.Duration `mapstructure:"idleTimeout" default:"60s" json:"idleTimeout"`
	// KeepAlive 启用 TCP keep-alive 保持持久连接。
	KeepAlive bool `mapstructure:"keepAlive" default:"true" json:"keepAlive"`
	// BathPath 设置路由注册的基路径。
	BathPath string `mapstructure:"bathPath" default:"/" json:"bathPath"`
}

// New 创建一个新的 Hertz 服务器实例。
//
// 参数:
//   - container: go-boot 依赖注入容器
//   - hertzConfig: 服务器配置选项
//   - l: 请求日志记录器
//
// 返回一个未初始化的 Hertz 服务器。
func New(container core.Container, hertzConfig *HertzConfig, l log.Logger) *Hertz {
	return &Hertz{
		container:   container,
		hertzConfig: hertzConfig,
		middlewares: make([]app.HandlerFunc, 0),
		logger:      l,
	}
}

// Run 启动 HTTP 服务器并阻塞，直到收到关闭信号。
//
// 它使用配置选项初始化 Hertz 服务器，将其注册到容器中，并开始处理请求。
// 服务器将阻塞直到收到 SIGINT 或 SIGTERM 信号。
//
// 如果服务器已初始化或注册失败，则返回错误。
func (h *Hertz) Run() error {
	if h.initialized {
		return fmt.Errorf("server already initialized")
	}
	h.server = server.New(
		server.WithHostPorts(h.hertzConfig.HostPorts),

		server.WithTransport(func(options *config.Options) network.Transporter {
			return netpoll.NewTransporter(options)
		}),

		server.WithReadTimeout(h.hertzConfig.ReadTimeout),
		server.WithWriteTimeout(h.hertzConfig.WriteTimeout),
		server.WithIdleTimeout(h.hertzConfig.IdleTimeout),
		server.WithKeepAlive(true),

		server.WithMaxRequestBodySize(10<<20),
		server.WithMaxKeepBodySize(8<<20),

		server.WithBasePath("/"),
		server.WithRedirectTrailingSlash(true),
		server.WithRedirectFixedPath(true),
		server.WithHandleMethodNotAllowed(true),
		server.WithGetOnly(false),
	)
	h.registerMiddlewares(h.logger)
	h.initialized = true

	err := h.container.Register(ServerBeanId, core.Bean(h.server))
	if err != nil {
		h.logger.Error(context.Background(), "hertz start run failed")
		return err
	}
	go h.server.Spin()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// Server 返回底层的 Hertz 服务器实例。
//
// 返回已配置的 Hertz 服务器以供直接访问。
func (h *Hertz) Server() *server.Hertz {
	return h.server
}

var _ ingress.Server = (*Hertz)(nil)

func (h *Hertz) Use(m any) ingress.Server {
	if handler, ok := m.(app.HandlerFunc); ok {
		h.middlewares = append(h.middlewares, handler)
	}
	return h
}

func (h *Hertz) UseGlobal(m any) ingress.Server {
	if handler, ok := m.(app.HandlerFunc); ok {
		h.middlewares = append([]app.HandlerFunc{handler}, h.middlewares...)
	}
	return h
}

func (h *Hertz) Register(fn func(core.Container) error) ingress.Server {
	if err := fn(h.container); err != nil {
		panic(fmt.Errorf("register handler failed: %w", err))
	}
	return h
}

// Container 返回 go-boot 容器实例。
//
// 返回用于依赖注入的容器。
func (h *Hertz) Container() core.Container {
	return h.container
}

// registerMiddlewares 注册内置中间件和用户定义的中间件。
//
// 内置中间件包括:
//   - LoggerMiddleware: 记录请求详情
//   - RecoveryMiddleware: 从 panic 中恢复
func (h *Hertz) registerMiddlewares(log log.Logger) {
	h.server.Use(LoggerMiddleware(log))
	h.server.Use(RecoveryMiddleware(log))

	for _, m := range h.middlewares {
		h.server.Use(m)
	}
}

// LoggerMiddleware 创建一个记录每个传入请求的中间件。
//
// 日志包括: 方法、路径、响应状态码和请求耗时。
//
// 参数:
//   - l: 日志输出实例
//
// 返回 Hertz 中间件处理器。
func LoggerMiddleware(l log.Logger) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		path := log.KeyValue{Key: "path", Value: string(ctx.Request.URI().Path())}
		method := log.KeyValue{Key: "method", Value: string(ctx.Request.Method())}
		status := log.KeyValue{Key: "status", Value: ctx.Response.StatusCode()}
		timeUse := log.KeyValue{Key: "time", Value: time.Since(start)}
		l.Info(c, "receive request ", method, path, status, timeUse)
		ctx.Next(c)
	}
}

// RecoveryMiddleware 创建一个从 panic 中恢复的中间件。
//
// 如果处理链中发生 panic，它会记录错误
// 并返回 500 Internal Server Error 响应。
//
// 参数:
//   - l: 错误日志输出实例
//
// 返回 Hertz 中间件处理器。
func RecoveryMiddleware(l log.Logger) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				err := log.KeyValue{Key: "error", Value: r}
				l.Info(c, "Recovered from panic ", err)
				ctx.String(500, "Internal Server Error")
				ctx.Abort()
			}
		}()
		ctx.Next(c)
	}
}
