package zap

import (
	"context"
	"github.com/xudefa/go-boot/log"
	"os"
	"testing"
)

func TestZapConfig(t *testing.T) {
	cfg := &ZapConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
		AddCaller:  true,
		CallerSkip: 1,
	}

	if cfg.Level != "debug" {
		t.Errorf("ZapConfig.Level = %v, want debug", cfg.Level)
	}
}

func TestNewZapAdapter(t *testing.T) {
	cfg := &ZapConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
	}

	adapter := NewZapAdapter(cfg)
	if adapter == nil {
		t.Error("NewZapAdapter() returned nil")
	}
}

func TestNewZapAdapterWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  ZapOption
	}{
		{"WithZapLevel", WithZapLevel(log.DebugLevel)},

		{"WithZapAddCaller", WithZapAddCaller(true)},
		{"WithZapCallerSkip", WithZapCallerSkip(2)},
		{"WithZapTimeFormat", WithZapTimeFormat("2006-01-02")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewZapAdapter(nil, tt.opt)
			if adapter == nil {
				t.Error("option failed")
			}
		})
	}
}

func TestZapAdapterLogLevels(t *testing.T) {
	adapter := NewZapAdapter(nil)
	ctx := context.Background()

	adapter.Debug(ctx, "debug message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Info(ctx, "info message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Warn(ctx, "warn message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Error(ctx, "error message", log.KeyValue{Key: "k", Value: "v"})
}

func TestZapAdapterWith(t *testing.T) {
	adapter := NewZapAdapter(nil)
	ctx := context.Background()

	newAdapter := adapter.With(ctx, log.KeyValue{Key: "k", Value: "v"})
	if newAdapter == nil {
		t.Error("With() returned nil")
	}
}

func TestZapAdapterSync(t *testing.T) {
	adapter := NewZapAdapter(&ZapConfig{Format: "json"})
	adapter.Sync()
}

func TestZapAdapterImplementsInterface(t *testing.T) {
	var _ log.Logger = (*ZapAdapter)(nil)
}

func TestZapLoggerBuilder(t *testing.T) {
	b := NewZapLoggerBuilder()
	cfg := &ZapConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
		AddCaller:  true,
		CallerSkip: 1,
	}
	b.cfg = cfg

	if b == nil {
		t.Error("NewZapLoggerBuilder() returned nil")
	}
}

func TestZapLoggerBuilderBuild(t *testing.T) {
	b := NewZapLoggerBuilder()
	b.cfg = &ZapConfig{Level: "info", Format: "text"}

	logger := b.Build()
	if logger == nil {
		t.Error("Build() returned nil")
	}
}

func TestToZapLevel(t *testing.T) {
	adapter := &ZapAdapter{}

	tests := []struct {
		input    log.Level
		expected string
	}{
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.DPanicLevel, "debug"},
		{log.PanicLevel, "fatal"},
		{log.FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			lvl := adapter.toZapLevel(tt.input)
			if lvl.String() != tt.expected {
				t.Errorf("toZapLevel() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

func TestToZapFields(t *testing.T) {
	adapter := &ZapAdapter{}
	keys := []log.KeyValue{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: 123},
		{Key: "k3", Value: true},
	}

	fields := adapter.toZapFields(keys)
	if len(fields) != 3 {
		t.Errorf("toZapFields() returned %d fields, want 3", len(fields))
	}
}

func TestMustZap(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustZap() panicked unexpectedly: %v", r)
			}
		}()
		MustZap(nil)
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustZap() did not panic with error")
			}
		}()
		MustZap(os.ErrPermission)
	}()
}
