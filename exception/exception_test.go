package exception

import (
	"context"
	"testing"
)

type mockLogger struct{}

func (m *mockLogger) Error(ctx context.Context, msg string, keyValues ...KeyValue) {}

type mockMetrics struct{}

func (m *mockMetrics) RecordException(exceptionType string, statusCode int) {}

type mockWriter struct{}

func (m *mockWriter) SetStatusCode(code int)      {}
func (m *mockWriter) SetHeader(key, value string) {}
func (m *mockWriter) Write(data []byte) error     { return nil }

func TestLogger_Interface(t *testing.T) {
	var _ Logger = (*mockLogger)(nil)
}

func TestMetricsRecorder_Interface(t *testing.T) {
	var _ MetricsRecorder = (*mockMetrics)(nil)
}

func TestResponseWriter_Interface(t *testing.T) {
	var _ ResponseWriter = (*mockWriter)(nil)
}
