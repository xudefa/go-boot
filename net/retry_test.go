package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExponentialBackoff_ShouldRetry_Error(t *testing.T) {
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)

	// 网络错误应该重试
	if !strategy.ShouldRetry(nil, context.DeadlineExceeded, 0) {
		t.Error("expected to retry on network error")
	}
}

func TestExponentialBackoff_ShouldRetry_StatusCodes(t *testing.T) {
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second, 500, 502, 503)

	tests := []struct {
		statusCode  int
		shouldRetry bool
	}{
		{200, false},
		{400, false},
		{404, false},
		{500, true},
		{502, true},
		{503, true},
	}

	for _, tt := range tests {
		resp := &HttpResponse{StatusCode: tt.statusCode}
		result := strategy.ShouldRetry(resp, nil, 0)
		if result != tt.shouldRetry {
			t.Errorf("status %d: expected %v, got %v", tt.statusCode, tt.shouldRetry, result)
		}
	}
}

func TestExponentialBackoff_Delay(t *testing.T) {
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)

	// 延迟应该在合理范围内（由于抖动，不能严格比较大小）
	delay0 := strategy.Delay(0)
	delay1 := strategy.Delay(1)
	delay2 := strategy.Delay(2)

	// 延迟不应该超过最大值
	if delay0 > 10*time.Second || delay1 > 10*time.Second || delay2 > 10*time.Second {
		t.Errorf("expected all delays <= 10s")
	}

	// 平均延迟应该递增（多次采样）
	var avg0, avg1, avg2 time.Duration
	for i := 0; i < 100; i++ {
		avg0 += strategy.Delay(0)
		avg1 += strategy.Delay(1)
		avg2 += strategy.Delay(2)
	}
	avg0 /= 100
	avg1 /= 100
	avg2 /= 100

	if avg1 <= avg0 {
		t.Errorf("expected avg delay1 > avg0, got %v <= %v", avg1, avg0)
	}
	if avg2 <= avg1 {
		t.Errorf("expected avg delay2 > avg1, got %v <= %v", avg2, avg1)
	}
}

func TestFixedDelay_ShouldRetry(t *testing.T) {
	strategy := NewFixedDelay(1*time.Second, 500, 503)

	// 网络错误应该重试
	if !strategy.ShouldRetry(nil, context.DeadlineExceeded, 0) {
		t.Error("expected to retry on network error")
	}

	// 500 应该重试
	resp500 := &HttpResponse{StatusCode: 500}
	if !strategy.ShouldRetry(resp500, nil, 0) {
		t.Error("expected to retry on 500")
	}

	// 200 不应该重试
	resp200 := &HttpResponse{StatusCode: 200}
	if strategy.ShouldRetry(resp200, nil, 0) {
		t.Error("expected not to retry on 200")
	}
}

func TestFixedDelay_Delay(t *testing.T) {
	strategy := NewFixedDelay(2 * time.Second)

	delay0 := strategy.Delay(0)
	delay1 := strategy.Delay(1)
	delay2 := strategy.Delay(2)

	// 固定延迟应该相同
	if delay0 != 2*time.Second || delay1 != 2*time.Second || delay2 != 2*time.Second {
		t.Errorf("expected all delays to be 2s, got %v, %v, %v", delay0, delay1, delay2)
	}
}

func TestRetryableClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_RetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success after retries"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(5),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableClient_MaxAttemptsExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(2),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected 500 after max attempts, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_OnRetryCallback(t *testing.T) {
	retryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	callbackCalled := false

	retryClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
		WithOnRetry(func(attempt int, resp *HttpResponse, err error) {
			callbackCalled = true
		}),
	)

	_, err := retryClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !callbackCalled {
		t.Error("expected onRetry callback to be called")
	}
}

func TestRetryableClient_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(10),
		WithRetryStrategy(NewExponentialBackoff(100*time.Millisecond, 1*time.Second)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	resp, err := retryClient.Get(ctx, "/")
	if err != context.DeadlineExceeded {
		t.Errorf("expected context deadline exceeded, got %v", err)
	}
	if resp == nil {
		t.Error("expected response to be non-nil")
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// 初始状态应该是关闭
	if cb.GetState() != CircuitClosed {
		t.Errorf("expected initial state to be closed, got %v", cb.GetState())
	}

	// 记录失败
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != CircuitClosed {
		t.Errorf("expected closed after 2 failures, got %v", cb.GetState())
	}

	// 第 3 次失败应该打开断路器
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Errorf("expected open after 3 failures, got %v", cb.GetState())
	}

	// 打开状态不应该允许请求
	if cb.AllowRequest() {
		t.Error("expected not to allow requests when open")
	}
}

func TestCircuitBreaker_ResetTimeout(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	// 打开断路器
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Fatalf("expected open, got %v", cb.GetState())
	}

	// 等待重置超时
	time.Sleep(150 * time.Millisecond)

	// 应该允许请求（半开状态）
	if !cb.AllowRequest() {
		t.Error("expected to allow request after reset timeout")
	}
	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("expected half-open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Second)

	// 打开断路器
	cb.RecordFailure()
	cb.RecordFailure()

	// 等待重置超时
	time.Sleep(100 * time.Millisecond)
	cb.AllowRequest() // 转换到半开

	// 记录成功应该关闭断路器
	cb.RecordSuccess()
	if cb.GetState() != CircuitClosed {
		t.Errorf("expected closed after success, got %v", cb.GetState())
	}
}

func TestCircuitBreakerClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(3),
		WithCircuitResetTimeout(1*time.Second),
	)

	resp, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCircuitBreakerClient_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	fallbackCalled := false

	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(2),
		WithCircuitResetTimeout(1*time.Second),
		WithFallback(func(ctx context.Context) (*HttpResponse, error) {
			fallbackCalled = true
			return &HttpResponse{StatusCode: 200, Body: []byte("fallback")}, nil
		}),
	)

	// 触发断路器打开
	for i := 0; i < 3; i++ {
		_, _ = cbClient.Get(context.Background(), "/")
	}

	// 断路器打开后应该调用降级函数
	resp, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fallbackCalled {
		t.Error("expected fallback to be called")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected fallback response 200, got %d", resp.StatusCode)
	}
}

func TestCircuitBreakerClient_NoFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(2),
		WithCircuitResetTimeout(1*time.Second),
	)

	// 触发断路器打开
	for i := 0; i < 3; i++ {
		_, _ = cbClient.Get(context.Background(), "/")
	}

	// 断路器打开后没有降级函数应该返回错误
	_, err := cbClient.Get(context.Background(), "/")
	if err == nil {
		t.Error("expected error when circuit breaker is open without fallback")
	}
}

func TestRetryableClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(1),
	)

	resp, err := retryClient.Post(context.Background(), "/", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client, WithMaxAttempts(1))

	resp, err := retryClient.Put(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client, WithMaxAttempts(1))

	resp, err := retryClient.Delete(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCircuitBreakerClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	cbClient := NewCircuitBreakerClient(client)

	resp, err := cbClient.Post(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryableClient_Close(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	retryClient := NewRetryableClient(client)

	err := retryClient.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestCircuitBreakerClient_Close(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	cbClient := NewCircuitBreakerClient(client)

	err := cbClient.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestExponentialBackoff_DefaultValues(t *testing.T) {
	strategy := NewExponentialBackoff(0, 0)

	if strategy.baseDelay <= 0 {
		t.Error("expected default base delay")
	}
	if strategy.maxDelay <= 0 {
		t.Error("expected default max delay")
	}
	if len(strategy.retryableStatus) == 0 {
		t.Error("expected default retryable status")
	}
}

func TestFixedDelay_DefaultValues(t *testing.T) {
	strategy := NewFixedDelay(0)

	if strategy.delay <= 0 {
		t.Error("expected default delay")
	}
	if len(strategy.retryableStatus) == 0 {
		t.Error("expected default retryable status")
	}
}

func TestCircuitBreaker_DefaultValues(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)

	if cb.maxFailures <= 0 {
		t.Error("expected default max failures")
	}
	if cb.resetTimeout <= 0 {
		t.Error("expected default reset timeout")
	}
}

func TestRetryableClient_NetworkError(t *testing.T) {
	// 使用无效地址触发网络错误
	client := NewClient("http://localhost:99999")
	retryClient := NewRetryableClient(client,
		WithMaxAttempts(2),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	_, err := retryClient.Get(context.Background(), "/")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestCircuitBreakerClient_GetCircuitState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	cbClient := NewCircuitBreakerClient(client)

	state := cbClient.GetCircuitState()
	if state != CircuitClosed {
		t.Errorf("expected closed, got %v", state)
	}
}
