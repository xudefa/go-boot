package log

import (
	"context"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{DPanicLevel, "dpanic"},
		{PanicLevel, "panic"},
		{FatalLevel, "fatal"},
		{Level(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type mockLogger struct {
	lastMsg  string
	lastKeys []KeyValue
}

func (m *mockLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	m.lastMsg = msg
	m.lastKeys = keys
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	m.lastKeys = keys
	return m
}

func TestLoggerOption(t *testing.T) {
	mock := &mockLogger{}
	opt := WithLogger(mock)
	cfg := &loggerConfig{}
	opt(cfg)

	if cfg.logger != mock {
		t.Error("WithLogger option did not set logger correctly")
	}
}

func TestBuildOptions(t *testing.T) {
	mock := &mockLogger{}
	logger := Build(WithLogger(mock))

	if logger != mock {
		t.Error("Build() did not apply options correctly")
	}
}
