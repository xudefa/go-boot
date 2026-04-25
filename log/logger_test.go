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

func TestSetDefault(t *testing.T) {
	original := defaultLogger
	mock := &mockLogger{}

	SetDefault(mock)

	if defaultLogger != mock {
		t.Error("SetDefault() failed to set defaultLogger")
	}

	SetDefault(original)
}

func TestDefaultLogger(t *testing.T) {
	mock := &mockLogger{}
	original := defaultLogger
	defaultLogger = mock

	got := DefaultLogger()
	if got != mock {
		t.Error("DefaultLogger() did not return the expected logger")
	}

	defaultLogger = original
}

func TestNopLogger(t *testing.T) {
	n := &nopLogger{}
	ctx := context.Background()

	n.Debug(ctx, "debug", KeyValue{Key: "k", Value: "v"})
	n.Info(ctx, "info", KeyValue{Key: "k", Value: "v"})
	n.Warn(ctx, "warn", KeyValue{Key: "k", Value: "v"})
	n.Error(ctx, "error", KeyValue{Key: "k", Value: "v"})
	n.DPanic(ctx, "dpanic", KeyValue{Key: "k", Value: "v"})
	n.Panic(ctx, "panic", KeyValue{Key: "k", Value: "v"})
	n.Fatal(ctx, "fatal", KeyValue{Key: "k", Value: "v"})

	if err := n.Sync(); err != nil {
		t.Errorf("nopLogger.Sync() error = %v", err)
	}

	w := n.With(ctx, KeyValue{Key: "k", Value: "v"})
	if w == nil {
		t.Error("nopLogger.With() returned nil")
	}
}

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Error("NewBuilder() returned nil")
	}

	logger := b.Build()
	if logger == nil {
		t.Error("builder.Build() returned nil")
	}
}

func TestBuildWithLogger(t *testing.T) {
	mock := &mockLogger{}
	b := NewBuilder()
	b.Logger = mock

	logger := b.Build()
	if logger != mock {
		t.Error("Build() did not return the set logger")
	}
}

func TestBuildWithNilLogger(t *testing.T) {
	b := &Builder{nil}
	logger := b.Build()
	if _, ok := logger.(*nopLogger); !ok {
		t.Error("Build() did not return nopLogger when Logger is nil")
	}
}

func TestLoggerOption(t *testing.T) {
	mock := &mockLogger{}
	opt := WithLogger(mock)
	b := &Builder{}
	opt(b)

	if b.Logger != mock {
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

func TestMust(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Must() panicked unexpectedly: %v", r)
			}
		}()
		Must(nil)
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Must() did not panic with error")
			}
		}()
		Must(context.DeadlineExceeded)
	}()
}

func TestKeyValue(t *testing.T) {
	kv := KeyValue{Key: "test", Value: "value"}
	if kv.Key != "test" || kv.Value != "value" {
		t.Error("KeyValue set incorrectly")
	}
}
