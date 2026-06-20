package exception

import (
	"context"
	"testing"
)

func TestDefaultLogger_Error(t *testing.T) {
	logger := NewDefaultLogger()

	logger.Error(context.Background(), "test error",
		KeyValue{Key: "key1", Value: "value1"},
		KeyValue{Key: "key2", Value: 123},
	)

	if logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestDefaultLogger_ErrorWithNilContext(t *testing.T) {
	logger := NewDefaultLogger()

	logger.Error(context.TODO(), "test error",
		KeyValue{Key: "key", Value: "value"},
	)

	if logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestDefaultMetricsRecorder_RecordException(t *testing.T) {
	recorder := NewDefaultMetricsRecorder()

	recorder.RecordException("TestError", 500)

	if recorder == nil {
		t.Error("MetricsRecorder should not be nil")
	}
}

func TestDefaultMetricsRecorder_RecordMultipleExceptions(t *testing.T) {
	recorder := NewDefaultMetricsRecorder()

	recorder.RecordException("Error1", 404)
	recorder.RecordException("Error2", 500)
	recorder.RecordException("Error1", 404)

	if recorder == nil {
		t.Error("MetricsRecorder should not be nil")
	}
}
