package exception

import (
	"testing"
)

func TestOption_WithLogger(t *testing.T) {
	logger := &mockLogger{}
	handler := NewDefaultExceptionHandler(WithLogger(logger))

	if handler.(*DefaultExceptionHandler).config.Logger == nil {
		t.Error("Logger should be set")
	}
}

func TestOption_WithMetricsRecorder(t *testing.T) {
	metrics := &mockMetrics{}
	handler := NewDefaultExceptionHandler(WithMetricsRecorder(metrics))

	if handler.(*DefaultExceptionHandler).config.MetricsRecorder == nil {
		t.Error("MetricsRecorder should be set")
	}
}

func TestOption_WithIncludeStackTrace(t *testing.T) {
	handler := NewDefaultExceptionHandler(WithIncludeStackTrace(true))

	if !handler.(*DefaultExceptionHandler).config.IncludeStackTrace {
		t.Error("IncludeStackTrace should be true")
	}
}
