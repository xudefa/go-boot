package exception

import (
	"encoding/json"
	"testing"
	"time"
)

func TestErrorResponse_MarshalJSON(t *testing.T) {
	resp := &ErrorResponse{
		Code:      404,
		Message:   "Not found",
		RequestID: "req-123",
		TraceID:   "trace-456",
		Details:   map[string]string{"id": "123"},
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["code"].(float64) != 404 {
		t.Errorf("Expected code 404, got %v", result["code"])
	}
	if result["message"].(string) != "Not found" {
		t.Errorf("Expected message 'Not found', got %v", result["message"])
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse(500, "Internal error", "req-123", "trace-456", nil)

	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
	if resp.Message != "Internal error" {
		t.Errorf("Expected message 'Internal error', got %s", resp.Message)
	}
	if resp.Timestamp == 0 {
		t.Error("Expected timestamp to be set")
	}
}
