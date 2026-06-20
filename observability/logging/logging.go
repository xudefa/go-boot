package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	SetLevel(level Level)
	SetOutput(output io.Writer)
	WithFields(fields ...Field) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value any
}

// String 字符串字段
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error 错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// defaultLogger 默认日志实现
type defaultLogger struct {
	level  Level
	output io.Writer
	fields []Field
	mu     sync.Mutex
}

// NewLogger 创建日志器
func NewLogger() Logger {
	return &defaultLogger{
		level:  INFO,
		output: os.Stdout,
	}
}

// SetLevel 设置日志级别
func (l *defaultLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置输出
func (l *defaultLogger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// WithFields 添加字段
func (l *defaultLogger) WithFields(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &defaultLogger{
		level:  l.level,
		output: l.output,
		fields: newFields,
	}
}

// Debug 调试日志
func (l *defaultLogger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

// Info 信息日志
func (l *defaultLogger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

// Warn 警告日志
func (l *defaultLogger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

// Error 错误日志
func (l *defaultLogger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

// Fatal 致命日志
func (l *defaultLogger) Fatal(msg string, fields ...Field) {
	l.log(FATAL, msg, fields...)
	os.Exit(1)
}

// log 内部日志方法
func (l *defaultLogger) log(level Level, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	allFields := make([]Field, len(l.fields)+len(fields))
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, _ = fmt.Fprintf(l.output, "[%s] [%s] %s", timestamp, level.String(), msg)

	for _, f := range allFields {
		_, _ = fmt.Fprintf(l.output, " %s=%v", f.Key, f.Value)
	}
	_, _ = fmt.Fprintln(l.output)
}
