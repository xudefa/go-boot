package logging

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)

	logger.Info("test message")
	if !strings.Contains(buf.String(), "INFO") {
		t.Error("expected INFO level in output")
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Error("expected message in output")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(WARN)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	if strings.Contains(buf.String(), "debug") {
		t.Error("expected debug message to be filtered")
	}
	if strings.Contains(buf.String(), "info") {
		t.Error("expected info message to be filtered")
	}
	if !strings.Contains(buf.String(), "warn") {
		t.Error("expected warn message in output")
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)

	logger.WithFields(String("key", "value")).Info("test")

	if !strings.Contains(buf.String(), "key=value") {
		t.Error("expected field in output")
	}
}

func TestLogger_ErrorField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)

	logger.Error("failed", Error(fmt.Errorf("test error")))

	if !strings.Contains(buf.String(), "error=test error") {
		t.Error("expected error field in output")
	}
}

func TestLogger_IntField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)

	logger.Info("count", Int("count", 42))

	if !strings.Contains(buf.String(), "count=42") {
		t.Error("expected int field in output")
	}
}
