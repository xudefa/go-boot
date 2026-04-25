package slog

import (
	"bytes"
	"context"
	"github.com/xudefa/go-boot/log"
	"testing"
)

func TestConfigLevelValue(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"error", "ERROR"},
		{"invalid", "INFO"},
		{"", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cfg := &Config{Level: tt.level}
			lvl := cfg.LevelValue().String()
			if lvl != tt.expected {
				t.Errorf("Config.LevelValue() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

func TestNewSlogLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
		AddSource:  false,
	}

	logger := NewSlogLogger(cfg, WithOutput(buf))
	if logger == nil {
		t.Error("NewSlogLogger() returned nil")
	}
}

func TestNewSlogLoggerTextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:  "debug",
		Format: "text",
	}

	logger := NewSlogLogger(cfg, WithOutput(buf))
	if logger == nil {
		t.Error("NewSlogLogger() returned nil")
	}
}

func TestSlogLoggerOptions(t *testing.T) {
	buf := &bytes.Buffer{}

	tests := []struct {
		name string
		opt  Option
	}{
		{"WithLevel", WithLevel(log.DebugLevel)},
		{"WithFormat", WithFormat("json")},
		{"WithTimeFormat", WithTimeFormat("2006-01-02")},
		{"WithAddSource", WithAddSource(true)},
		{"WithOutput", WithOutput(buf)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSlogLogger(nil, tt.opt)
			if logger == nil {
				t.Error("option failed")
			}
		})
	}
}

func TestSlogLoggerLogLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))
	ctx := context.Background()

	logger.Debug(ctx, "debug message", log.KeyValue{Key: "k", Value: "v"})
	logger.Info(ctx, "info message", log.KeyValue{Key: "k", Value: "v"})
	logger.Warn(ctx, "warn message", log.KeyValue{Key: "k", Value: "v"})
	logger.Error(ctx, "error message", log.KeyValue{Key: "k", Value: "v"})
}

func TestSlogLoggerLogLevelsWithDPanicPanicFatal(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))
	ctx := context.Background()

	func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		logger.DPanic(ctx, "dpanic message", log.KeyValue{Key: "k", Value: "v"})
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		logger.Panic(ctx, "panic message", log.KeyValue{Key: "k", Value: "v"})
	}()

	logger.Fatal(ctx, "fatal message", log.KeyValue{Key: "k", Value: "v"})
}

func TestSlogLoggerWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))
	ctx := context.Background()

	newLogger := logger.With(ctx, log.KeyValue{Key: "k", Value: "v"})
	if newLogger == nil {
		t.Error("With() returned nil")
	}
}

func TestSlogLoggerWithMultipleKeys(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))
	ctx := context.Background()

	newLogger := logger.With(
		ctx,
		log.KeyValue{Key: "k1", Value: "v1"},
		log.KeyValue{Key: "k2", Value: 123},
		log.KeyValue{Key: "k3", Value: true},
	)
	if newLogger == nil {
		t.Error("With() returned nil")
	}
}

func TestSlogLoggerSync(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))

	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestSlogLoggerImplementsInterface(t *testing.T) {
	var _ log.Logger = (*SlogLogger)(nil)
}

func TestToLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected log.Level
	}{
		{"debug", log.DebugLevel},
		{"info", log.InfoLevel},
		{"warn", log.WarnLevel},
		{"error", log.ErrorLevel},
		{"invalid", log.InfoLevel},
		{"", log.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toLevel(tt.input)
			if got != tt.expected {
				t.Errorf("toLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSlogLoggerToSlogLevel(t *testing.T) {
	logger := &SlogLogger{}

	tests := []struct {
		input    log.Level
		expected string
	}{
		{log.DebugLevel, "DEBUG"},
		{log.InfoLevel, "INFO"},
		{log.WarnLevel, "WARN"},
		{log.ErrorLevel, "ERROR"},
		{log.DPanicLevel, "DEBUG"},
		{log.PanicLevel, "DEBUG"},
		{log.FatalLevel, "DEBUG"},
		{log.Level(100), "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			lvl := logger.toSlogLevel(tt.input)
			if lvl.String() != tt.expected {
				t.Errorf("toSlogLevel() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

func TestSlogLoggerToSlogAttr(t *testing.T) {
	logger := &SlogLogger{}

	attr := logger.toSlogAttr("key", "value")
	if attr.Key != "key" {
		t.Errorf("toSlogAttr() key = %v, want key", attr.Key)
	}
}

func TestSlogLoggerChainedOptions(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := NewSlogLogger(nil,
		WithLevel(log.DebugLevel),
		WithFormat("json"),
		WithTimeFormat("2006-01-02"),
		WithAddSource(true),
		WithOutput(buf),
	)

	if logger == nil {
		t.Error("chained options failed")
	}
}

func TestSlogLoggerNilConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(nil, WithOutput(buf))

	if logger == nil {
		t.Error("NewSlogLogger(nil) returned nil")
	}
}
