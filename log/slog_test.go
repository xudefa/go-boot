package log

import (
	"bytes"
	"context"
	"testing"
)

func TestNewSlogLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(
		WithLevel(DebugLevel),
		WithFormat("json"),
		WithTimeFormat("2006-01-02"),
		WithAddSource(false),
		WithOutput(buf),
	)
	if logger == nil {
		t.Error("NewSlogLogger() returned nil")
	}
}

func TestNewSlogLoggerTextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(
		WithLevel(DebugLevel),
		WithFormat("text"),
		WithOutput(buf),
	)
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
		{"WithLevel", WithLevel(DebugLevel)},
		{"WithFormat", WithFormat("json")},
		{"WithTimeFormat", WithTimeFormat("2006-01-02")},
		{"WithAddSource", WithAddSource(true)},
		{"WithOutput", WithOutput(buf)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewSlogLogger(tt.opt)
			if logger == nil {
				t.Error("option failed")
			}
		})
	}
}

func TestSlogLoggerLogLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(WithOutput(buf))
	ctx := context.Background()

	logger.Debug(ctx, "debug message", KeyValue{Key: "k", Value: "v"})
	logger.Info(ctx, "info message", KeyValue{Key: "k", Value: "v"})
	logger.Warn(ctx, "warn message", KeyValue{Key: "k", Value: "v"})
	logger.Error(ctx, "error message", KeyValue{Key: "k", Value: "v"})
}

func TestSlogLoggerLogLevelsWithDPanicPanicFatal(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(WithOutput(buf))
	ctx := context.Background()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("recovered from DPanic: %v", r)
			}
		}()
		logger.DPanic(ctx, "dpanic message", KeyValue{Key: "k", Value: "v"})
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("recovered from Panic: %v", r)
			}
		}()
		logger.Panic(ctx, "panic message", KeyValue{Key: "k", Value: "v"})
	}()

	logger.Fatal(ctx, "fatal message", KeyValue{Key: "k", Value: "v"})
}

func TestSlogLoggerWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(WithOutput(buf))
	ctx := context.Background()

	newLogger := logger.With(ctx, KeyValue{Key: "k", Value: "v"})
	if newLogger == nil {
		t.Error("With() returned nil")
	}
}

func TestSlogLoggerWithMultipleKeys(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(WithOutput(buf))
	ctx := context.Background()

	newLogger := logger.With(
		ctx,
		KeyValue{Key: "k1", Value: "v1"},
		KeyValue{Key: "k2", Value: 123},
		KeyValue{Key: "k3", Value: true},
	)
	if newLogger == nil {
		t.Error("With() returned nil")
	}
}

func TestSlogLoggerSync(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSlogLogger(WithOutput(buf))

	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestSlogLoggerImplementsInterface(t *testing.T) {
	var _ Logger = (*SlogLogger)(nil)
}

func TestToLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"error", ErrorLevel},
		{"invalid", InfoLevel},
		{"", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ToLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSlogLoggerToSlogLevel(t *testing.T) {
	logger := &SlogLogger{}

	tests := []struct {
		input    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{DPanicLevel, "ERROR+2"},
		{PanicLevel, "ERROR+4"},
		{FatalLevel, "ERROR+6"},
		{Level(100), "INFO"},
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

func TestSlogLoggerChainedOptions(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := NewSlogLogger(
		WithLevel(DebugLevel),
		WithFormat("json"),
		WithTimeFormat("2006-01-02"),
		WithAddSource(true),
		WithOutput(buf),
	)

	if logger == nil {
		t.Error("chained options failed")
	}
}
