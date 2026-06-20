package exception

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

// DefaultLogger 默认日志记录器实现
//
// DefaultLogger 是 Logger 接口的默认实现，使用标准库的 log 包。
// 它将日志输出到标准输出，并包含时间戳和文件位置信息。
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger 创建默认日志记录器
//
// 返回一个 DefaultLogger 实例，该实例会将日志输出到标准输出。
func NewDefaultLogger() Logger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[EXCEPTION] ", log.LstdFlags|log.Lshortfile),
	}
}

// Error 记录错误日志
//
// 将错误消息和键值对格式化后输出到日志。
// 键值对会以 "key=value" 的格式追加到消息后面。
func (l *DefaultLogger) Error(ctx context.Context, msg string, keyValues ...KeyValue) {
	logMsg := msg
	if len(keyValues) > 0 {
		logMsg += " | "
		for i, kv := range keyValues {
			if i > 0 {
				logMsg += ", "
			}
			logMsg += fmt.Sprintf("%s=%v", kv.Key, kv.Value)
		}
	}
	l.logger.Println(logMsg)
}

// DefaultMetricsRecorder 默认指标记录器实现
//
// DefaultMetricsRecorder 是 MetricsRecorder 接口的默认实现。
// 它使用内存中的 map 来记录异常计数，适合开发和测试环境。
type DefaultMetricsRecorder struct {
	counters map[string]int
	mu       sync.RWMutex
}

// NewDefaultMetricsRecorder 创建默认指标记录器
//
// 返回一个 DefaultMetricsRecorder 实例。
func NewDefaultMetricsRecorder() MetricsRecorder {
	return &DefaultMetricsRecorder{
		counters: make(map[string]int),
	}
}

// RecordException 记录异常指标
//
// 根据异常类型和状态码生成键，并增加对应的计数。
// 键的格式为 "exceptionType:statusCode"。
func (m *DefaultMetricsRecorder) RecordException(exceptionType string, statusCode int) {
	key := fmt.Sprintf("%s:%d", exceptionType, statusCode)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[key]++
}

// GetCount 获取计数（用于测试）
//
// 返回指定异常类型和状态码的计数。
// 注意：这个方法主要用于测试，生产环境应该使用更专业的指标系统。
func (m *DefaultMetricsRecorder) GetCount(exceptionType string, statusCode int) int {
	key := fmt.Sprintf("%s:%d", exceptionType, statusCode)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[key]
}
