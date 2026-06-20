package exception

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExceptionHandlingMiddleware_Error(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	middleware := ExceptionHandlingMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := ErrNotFound
		resp := handler.Handle(r.Context(), err, &httpResponseWriter{w})
		if resp != nil {
			w.WriteHeader(resp.Code)
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != 404 {
		t.Errorf("Expected code 404, got %d", resp.Code)
	}
}

func TestExceptionHandlingMiddleware_Panic(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	middleware := ExceptionHandlingMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
}
