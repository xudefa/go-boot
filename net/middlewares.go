package net

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/xudefa/go-boot/log"
)

const (
	requestIDHeader = "X-Request-Id" // 请求 ID 响应头名称
)

type requestIDKey struct{} // 请求 ID 上下文键类型

// RequestIDFromContext 从上下文中获取请求ID
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// SetRequestIDToContext 将请求ID设置到上下文中
func SetRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GenerateRequestID 生成唯一的请求ID
func GenerateRequestID() string {
	return fmt.Sprintf("%d-%d-%d",
		time.Now().UnixNano(),
		rand.Int63(),
		rand.Intn(10000))
}

// RequestIDMiddleware 请求ID中间件
// 优先读取 X-Request-Id 请求头，如果不存在则自动生成
// 请求ID会被设置到响应头和上下文中
func RequestIDMiddleware(logger log.Logger) MiddlewareFunc {
	return func(ctx HandlerContext) {
		requestID := ctx.Header(requestIDHeader)
		if requestID == "" {
			requestID = GenerateRequestID()
		}

		ctx.SetHeader(requestIDHeader, requestID)
		newCtx := SetRequestIDToContext(ctx.Context(), requestID)
		ctx.SetContext(newCtx)

		if logger != nil {
			logger = logger.With(ctx.Context(), log.KeyValue{Key: "request_id", Value: requestID})
		}

		ctx.Next()
	}
}

// TraceIDExtractor 用于从上下文中提取 TraceID 的函数
type TraceIDExtractor func(ctx context.Context) string

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Logger           log.Logger
	SamplingRate     float64 // 采样率，0.0-1.0，默认1.0（全采样）
	IgnorePaths      []string
	TraceIDExtractor TraceIDExtractor // 可选的 TraceID 提取器
}

// AccessLogMiddleware 访问日志中间件
// 记录 method、path、status、latency、request_id、trace_id
func AccessLogMiddleware(config AccessLogConfig) MiddlewareFunc {
	if config.Logger == nil {
		config.Logger = log.Build()
	}
	if config.SamplingRate <= 0 || config.SamplingRate > 1 {
		config.SamplingRate = 1.0
	}

	return func(ctx HandlerContext) {
		startTime := time.Now()

		ctx.Next()

		if config.SamplingRate < 1.0 && rand.Float64() > config.SamplingRate {
			return
		}

		path := ctx.RequestURI()
		for _, ignorePath := range config.IgnorePaths {
			if path == ignorePath {
				return
			}
		}

		latency := time.Since(startTime)
		requestID := RequestIDFromContext(ctx.Context())

		var traceID string
		if config.TraceIDExtractor != nil {
			traceID = config.TraceIDExtractor(ctx.Context())
		}

		config.Logger.Info(ctx.Context(), "access log",
			log.KeyValue{Key: "method", Value: ctx.RequestMethod()},
			log.KeyValue{Key: "path", Value: path},
			log.KeyValue{Key: "latency", Value: latency.String()},
			log.KeyValue{Key: "request_id", Value: requestID},
			log.KeyValue{Key: "trace_id", Value: traceID},
		)
	}
}

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code      int         `json:"code"`              // 错误码
	Message   string      `json:"message"`           // 错误消息
	RequestID string      `json:"requestId"`         // 请求 ID
	TraceID   string      `json:"traceId"`           // 链路追踪 ID
	Details   interface{} `json:"details,omitempty"` // 错误详情（可选）
}

// ErrorMiddlewareConfig 错误中间件配置
type ErrorMiddlewareConfig struct {
	Logger           log.Logger       // 日志记录器
	TraceIDExtractor TraceIDExtractor // 可选的 TraceID 提取器
}

// ErrorMiddleware 统一错误响应中间件
// 捕获处理过程中的panic，并返回标准化的错误响应
func ErrorMiddleware(config ErrorMiddlewareConfig) MiddlewareFunc {
	if config.Logger == nil {
		config.Logger = log.Build()
	}

	return func(ctx HandlerContext) {
		defer func() {
			if r := recover(); r != nil {
				requestID := RequestIDFromContext(ctx.Context())

				var traceID string
				if config.TraceIDExtractor != nil {
					traceID = config.TraceIDExtractor(ctx.Context())
				}

				errMsg := fmt.Sprintf("%v", r)
				config.Logger.Error(ctx.Context(), "panic recovered",
					log.KeyValue{Key: "request_id", Value: requestID},
					log.KeyValue{Key: "error", Value: errMsg},
				)

				response := ErrorResponse{
					Code:      http.StatusInternalServerError,
					Message:   "Internal Server Error",
					RequestID: requestID,
					TraceID:   traceID,
					Details:   errMsg,
				}

				ctx.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()

		ctx.Next()
	}
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string      // 允许的源列表，"*" 表示全部
	AllowMethods     []string      // 允许的 HTTP 方法列表
	AllowHeaders     []string      // 允许的请求头列表
	AllowCredentials bool          // 是否允许携带凭据
	MaxAge           time.Duration // 预检请求缓存时间
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORSMiddleware CORS 中间件
//
// 处理跨域请求，根据配置设置 Access-Control-* 响应头。
// 对于 OPTIONS 预检请求，直接返回 204 No Content。
func CORSMiddleware(config CORSConfig) MiddlewareFunc {
	return func(ctx HandlerContext) {
		origin := ctx.Header("Origin")
		if origin == "" {
			ctx.Next()
			return
		}

		allowed := false
		for _, allowedOrigin := range config.AllowOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			ctx.Next()
			return
		}

		ctx.SetHeader("Access-Control-Allow-Origin", origin)

		if config.AllowCredentials {
			ctx.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowMethods) > 0 {
			methods := ""
			for i, method := range config.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			ctx.SetHeader("Access-Control-Allow-Methods", methods)
		}

		if len(config.AllowHeaders) > 0 {
			headers := ""
			for i, header := range config.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			ctx.SetHeader("Access-Control-Allow-Headers", headers)
		}

		if config.MaxAge > 0 {
			ctx.SetHeader("Access-Control-Max-Age", fmt.Sprintf("%d", int(config.MaxAge.Seconds())))
		}

		if ctx.RequestMethod() == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

// ErrorResponseConfig 错误响应配置
type ErrorResponseConfig struct {
	TraceIDExtractor TraceIDExtractor // 可选的 TraceID 提取器
}

// WrapErrorResponse 包装错误响应
func WrapErrorResponse(ctx context.Context, code int, message string, details interface{}, config ErrorResponseConfig) ErrorResponse {
	requestID := RequestIDFromContext(ctx)

	var traceID string
	if config.TraceIDExtractor != nil {
		traceID = config.TraceIDExtractor(ctx)
	}

	return ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		TraceID:   traceID,
		Details:   details,
	}
}

// WriteErrorResponse 写入错误响应
func WriteErrorResponse(ctx HandlerContext, code int, message string, details interface{}, config ErrorResponseConfig) {
	response := WrapErrorResponse(ctx.Context(), code, message, details, config)
	ctx.AbortWithStatusJSON(code, response)
}

// WriteErrorResponseSimple 使用默认配置写入错误响应
func WriteErrorResponseSimple(ctx HandlerContext, code int, message string, details interface{}) {
	WriteErrorResponse(ctx, code, message, details, ErrorResponseConfig{})
}
