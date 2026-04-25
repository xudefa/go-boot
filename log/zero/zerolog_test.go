package zero

import (
	"context"
	"github.com/xudefa/go-boot/log"
	"testing"
)

func TestZerologConfig(t *testing.T) {
	cfg := &ZerologConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
	}

	if cfg.Level != "debug" {
		t.Errorf("ZerologConfig.Level = %v, want debug", cfg.Level)
	}
}

func TestNewZerologAdapter(t *testing.T) {
	cfg := &ZerologConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
	}

	adapter := NewZerologAdapter(cfg)
	if adapter == nil {
		t.Error("NewZerologAdapter() returned nil")
	}
}

func TestNewZerologAdapterConsole(t *testing.T) {
	cfg := &ZerologConfig{
		Level:  "debug",
		Format: "console",
	}

	adapter := NewZerologAdapter(cfg)
	if adapter == nil {
		t.Error("NewZerologAdapter() returned nil")
	}
}

func TestNewZerologAdapterWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  ZerologOption
	}{
		{"WithZerologLevel", WithZerologLevel(log.DebugLevel)},
		{"WithZerologFormat", WithZerologFormat("json")},
		{"WithZerologTimeFormat", WithZerologTimeFormat("2006-01-02")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewZerologAdapter(nil, tt.opt)
			if adapter == nil {
				t.Error("option failed")
			}
		})
	}
}

func TestZerologAdapterLogLevels(t *testing.T) {
	adapter := NewZerologAdapter(nil)
	ctx := context.Background()

	adapter.Debug(ctx, "debug message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Info(ctx, "info message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Warn(ctx, "warn message", log.KeyValue{Key: "k", Value: "v"})
	adapter.Error(ctx, "error message", log.KeyValue{Key: "k", Value: "v"})
}

func TestZerologAdapterWith(t *testing.T) {
	adapter := NewZerologAdapter(nil)
	ctx := context.Background()

	newAdapter := adapter.With(ctx, log.KeyValue{Key: "k", Value: "v"})
	if newAdapter == nil {
		t.Error("With() returned nil")
	}
}

func TestZerologAdapterSync(t *testing.T) {
	adapter := NewZerologAdapter(&ZerologConfig{Format: "json"})
	adapter.Sync()
}

func TestZerologAdapterImplementsInterface(t *testing.T) {
	var _ log.Logger = (*ZerologAdapter)(nil)
}

func TestZerologLoggerBuilder(t *testing.T) {
	b := NewZerologLoggerBuilder()
	cfg := &ZerologConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "2006-01-02",
	}
	b.cfg = cfg

	if b == nil {
		t.Error("NewZerologLoggerBuilder() returned nil")
	}
}

func TestZerologLoggerBuilderBuild(t *testing.T) {
	b := NewZerologLoggerBuilder()
	b.cfg = &ZerologConfig{Level: "info", Format: "text"}

	logger := b.Build()
	if logger == nil {
		t.Error("Build() returned nil")
	}
}

func TestToZerologLevel(t *testing.T) {
	adapter := &ZerologAdapter{}

	tests := []struct {
		input    log.Level
		expected string
	}{
		{log.DebugLevel, "debug"},
		{log.InfoLevel, "info"},
		{log.WarnLevel, "warn"},
		{log.ErrorLevel, "error"},
		{log.DPanicLevel, "debug"},
		{log.PanicLevel, "panic"},
		{log.FatalLevel, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			lvl := adapter.toZerologLevel(tt.input)
			if lvl.String() != tt.expected {
				t.Errorf("toZerologLevel() = %v, want %v", lvl, tt.expected)
			}
		})
	}
}

func TestToZerologFields(t *testing.T) {
	adapter := &ZerologAdapter{}
	keys := []log.KeyValue{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: 123},
		{Key: "k3", Value: true},
	}

	fields := adapter.toZerologFields(keys)
	if len(fields) != 3 {
		t.Errorf("toZerologFields() returned %d fields, want 3", len(fields))
	}
	if fields["k1"] != "v1" {
		t.Errorf("toZerologFields() k1 = %v", fields["k1"])
	}
}

func TestMustZerolog(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustZerolog() panicked unexpectedly: %v", r)
			}
		}()
		MustZerolog(nil)
	}()

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustZerolog() did not panic with error")
			}
		}()
		MustZerolog(context.DeadlineExceeded)
	}()
}
