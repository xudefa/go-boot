package loadbalancer

import (
	"testing"
)

func TestRoundRobin(t *testing.T) {
	lb := NewRoundRobin()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081", Weight: 1},
		{ID: "2", URL: "http://localhost:8082", Weight: 1},
		{ID: "3", URL: "http://localhost:8083", Weight: 1},
	}

	// 应该按顺序轮询
	b1, _ := lb.Next(backends)
	if b1.URL != "http://localhost:8081" {
		t.Errorf("expected first backend to be 8081, got %s", b1.URL)
	}

	b2, _ := lb.Next(backends)
	if b2.URL != "http://localhost:8082" {
		t.Errorf("expected second backend to be 8082, got %s", b2.URL)
	}

	b3, _ := lb.Next(backends)
	if b3.URL != "http://localhost:8083" {
		t.Errorf("expected third backend to be 8083, got %s", b3.URL)
	}

	// 循环回到第一个
	b4, _ := lb.Next(backends)
	if b4.URL != "http://localhost:8081" {
		t.Errorf("expected fourth backend to be 8081, got %s", b4.URL)
	}
}

func TestRoundRobinEmpty(t *testing.T) {
	lb := NewRoundRobin()
	_, err := lb.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestRandom(t *testing.T) {
	lb := NewRandom()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081", Weight: 1},
		{ID: "2", URL: "http://localhost:8082", Weight: 1},
	}

	// 随机选择应该总是返回某个后端
	b, err := lb.Next(backends)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if b == nil {
		t.Error("expected a backend, got nil")
	}
}

func TestWeightedRoundRobin(t *testing.T) {
	lb := NewWeightedRoundRobin()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081", Weight: 3},
		{ID: "2", URL: "http://localhost:8082", Weight: 1},
	}

	// 前 3 次应该选择权重高的后端
	count1 := 0
	count2 := 0
	for i := 0; i < 4; i++ {
		b, _ := lb.Next(backends)
		if b.URL == "http://localhost:8081" {
			count1++
		} else {
			count2++
		}
	}

	if count1 != 3 {
		t.Errorf("expected backend1 to be selected 3 times, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("expected backend2 to be selected 1 time, got %d", count2)
	}
}

func TestLeastConnections(t *testing.T) {
	lb := NewLeastConnections()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081", Active: 5},
		{ID: "2", URL: "http://localhost:8082", Active: 2},
		{ID: "3", URL: "http://localhost:8083", Active: 8},
	}

	b, _ := lb.Next(backends)
	if b.URL != "http://localhost:8082" {
		t.Errorf("expected backend with least connections (8082), got %s", b.URL)
	}
}

func TestIPHash(t *testing.T) {
	lb := NewIPHash()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081"},
		{ID: "2", URL: "http://localhost:8082"},
		{ID: "3", URL: "http://localhost:8083"},
	}

	// 相同 IP 应该总是选择相同的后端
	b1, _ := lb.NextByIP(backends, "192.168.1.1")
	b2, _ := lb.NextByIP(backends, "192.168.1.1")
	if b1.URL != b2.URL {
		t.Errorf("expected same backend for same IP, got %s and %s", b1.URL, b2.URL)
	}

	// 不同 IP 可能选择不同后端
	b3, _ := lb.NextByIP(backends, "192.168.1.2")
	_ = b3 // 可能相同也可能不同
}

func TestStickySession(t *testing.T) {
	lb := NewStickySession("SESSIONID")
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081"},
		{ID: "2", URL: "http://localhost:8082"},
	}

	// 相同会话 ID 应该选择相同的后端
	b1, _ := lb.NextWithSession(backends, "session-123")
	b2, _ := lb.NextWithSession(backends, "session-123")
	if b1.URL != b2.URL {
		t.Errorf("expected same backend for same session, got %s and %s", b1.URL, b2.URL)
	}

	// 不同会话 ID 可能选择不同后端
	b3, _ := lb.NextWithSession(backends, "session-456")
	_ = b3
}

func TestHealthAware(t *testing.T) {
	inner := NewRoundRobin()
	lb := NewHealthAware(inner)

	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081", Health: HealthDown},
		{ID: "2", URL: "http://localhost:8082", Health: HealthUp},
		{ID: "3", URL: "http://localhost:8083", Health: HealthUp},
	}

	// 应该只选择健康的后端
	b, _ := lb.Next(backends)
	if b.Health != HealthUp {
		t.Errorf("expected healthy backend, got %s with health %v", b.URL, b.Health)
	}
}

func TestAdaptiveWeight(t *testing.T) {
	lb := NewAdaptiveWeight()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081"},
		{ID: "2", URL: "http://localhost:8082"},
	}

	// 记录后端 1 的快速响应
	lb.RecordRequest("http://localhost:8081", 10, false)
	lb.RecordRequest("http://localhost:8081", 15, false)

	// 记录后端 2 的慢响应
	lb.RecordRequest("http://localhost:8082", 500, false)
	lb.RecordRequest("http://localhost:8082", 600, false)

	// 多次选择，后端 1 应该被选中的概率更高
	count1 := 0
	for i := 0; i < 100; i++ {
		b, _ := lb.Next(backends)
		if b.URL == "http://localhost:8081" {
			count1++
		}
	}

	// 后端 1 应该被选中超过 50% 的次数
	if count1 <= 50 {
		t.Errorf("expected faster backend to be selected more often, got %d/100", count1)
	}
}

func TestResponseTimeWeighted(t *testing.T) {
	lb := NewResponseTimeWeighted()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081"},
		{ID: "2", URL: "http://localhost:8082"},
	}

	// 记录响应时间
	lb.RecordResponseTime("http://localhost:8081", 10)
	lb.RecordResponseTime("http://localhost:8082", 100)

	// 多次选择，快速后端应该被选中的概率更高
	count1 := 0
	for i := 0; i < 100; i++ {
		b, _ := lb.Next(backends)
		if b.URL == "http://localhost:8081" {
			count1++
		}
	}

	if count1 <= 50 {
		t.Errorf("expected faster backend to be selected more often, got %d/100", count1)
	}
}

func TestConsistentHash(t *testing.T) {
	lb := NewConsistentHash()
	backends := []*ServiceInstance{
		{ID: "1", URL: "http://localhost:8081"},
		{ID: "2", URL: "http://localhost:8082"},
	}

	// 相同键应该选择相同的后端
	b1, _ := lb.NextByKey(backends, "user-123")
	b2, _ := lb.NextByKey(backends, "user-123")
	if b1.URL != b2.URL {
		t.Errorf("expected same backend for same key, got %s and %s", b1.URL, b2.URL)
	}
}
