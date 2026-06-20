package center

import (
	"context"
	"testing"
)

func TestRegistryBuilder_Defaults(t *testing.T) {
	builder := NewRegistryBuilder()

	if builder.heartbeat != 30 {
		t.Errorf("expected default heartbeat 30, got %d", builder.heartbeat)
	}

	if builder.metadata == nil {
		t.Error("expected non-nil metadata")
	}
}

func TestRegistryBuilder_ChainConfig(t *testing.T) {
	reg, err := NewRegistryBuilder().
		Type("consul").
		Address("http://localhost:8500").
		Token("test-token").
		Heartbeat(15).
		Metadata("env", "prod").
		Metadata("version", "1.0").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRegistryBuilder_MissingAddress(t *testing.T) {
	_, err := NewRegistryBuilder().Build()
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestRegistryBuilder_WithCustomRegistry(t *testing.T) {
	customReg := NewMemoryRegistry()

	reg, err := NewRegistryBuilder().
		Registry(customReg).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRegistryBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing address")
		}
	}()

	NewRegistryBuilder().MustBuild()
}

func TestInstanceBuilder_BasicInstance(t *testing.T) {
	info := NewInstanceBuilder("user-service").
		ID("instance-1").
		Host("192.168.1.1").
		Port(8080).
		Weight(5).
		Metadata("env", "prod").
		Build()

	if info.ServiceName != "user-service" {
		t.Errorf("expected ServiceName 'user-service', got %s", info.ServiceName)
	}

	if info.ID != "instance-1" {
		t.Errorf("expected ID 'instance-1', got %s", info.ID)
	}

	if info.Host != "192.168.1.1" {
		t.Errorf("expected Host '192.168.1.1', got %s", info.Host)
	}

	if info.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", info.Port)
	}

	if info.Weight != 5 {
		t.Errorf("expected Weight 5, got %d", info.Weight)
	}

	if info.Metadata["env"] != "prod" {
		t.Errorf("expected Metadata['env'] 'prod', got %s", info.Metadata["env"])
	}

	if !info.Healthy {
		t.Error("expected instance to be healthy by default")
	}
}

func TestInstanceBuilder_DefaultValues(t *testing.T) {
	info := NewInstanceBuilder("test-service").Build()

	if info.Weight != 1 {
		t.Errorf("expected default Weight 1, got %d", info.Weight)
	}

	if !info.Healthy {
		t.Error("expected default Healthy true")
	}

	if info.Metadata == nil {
		t.Error("expected non-nil Metadata")
	}
}

func TestInstanceBuilder_InvalidWeight(t *testing.T) {
	info := NewInstanceBuilder("test-service").Weight(-1).Build()

	if info.Weight != 1 {
		t.Errorf("expected Weight to remain 1 for negative value, got %d", info.Weight)
	}
}

func TestSelectorBuilder_RandomStrategy(t *testing.T) {
	sel, err := NewSelectorBuilder().
		Strategy("random").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sel == nil {
		t.Fatal("expected non-nil selector")
	}
}

func TestSelectorBuilder_RoundRobinStrategy(t *testing.T) {
	sel, err := NewSelectorBuilder().
		Strategy("round_robin").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sel == nil {
		t.Fatal("expected non-nil selector")
	}
}

func TestSelectorBuilder_WeightedStrategy(t *testing.T) {
	sel, err := NewSelectorBuilder().
		Strategy("weighted").
		Weight("instance-1", 10).
		Weight("instance-2", 20).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sel == nil {
		t.Fatal("expected non-nil selector")
	}
}

func TestSelectorBuilder_InvalidStrategy(t *testing.T) {
	_, err := NewSelectorBuilder().Strategy("invalid").Build()
	if err == nil {
		t.Error("expected error for invalid strategy")
	}
}

func TestSelectorBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid strategy")
		}
	}()

	NewSelectorBuilder().Strategy("invalid").MustBuild()
}

func TestMemoryRegistry_RegisterAndDiscover(t *testing.T) {
	reg := NewMemoryRegistry()
	ctx := context.Background()

	info := NewInstanceBuilder("user-service").
		ID("instance-1").
		Host("192.168.1.1").
		Port(8080).
		Build()

	err := reg.Register(ctx, info)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	instances, err := reg.Discover(ctx, "user-service")
	if err != nil {
		t.Fatalf("failed to discover: %v", err)
	}

	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}

	if instances[0].ID != "instance-1" {
		t.Errorf("expected ID 'instance-1', got %s", instances[0].ID)
	}
}

func TestMemoryRegistry_Deregister(t *testing.T) {
	reg := NewMemoryRegistry()
	ctx := context.Background()

	info := NewInstanceBuilder("user-service").
		ID("instance-1").
		Host("192.168.1.1").
		Port(8080).
		Build()

	_ = reg.Register(ctx, info)
	_ = reg.Deregister(ctx, info)

	instances, err := reg.Discover(ctx, "user-service")
	if err != nil {
		t.Fatalf("failed to discover: %v", err)
	}

	if len(instances) != 0 {
		t.Errorf("expected 0 instances after deregister, got %d", len(instances))
	}
}

func TestMemoryRegistry_Watch(t *testing.T) {
	reg := NewMemoryRegistry()
	ctx := context.Background()

	ch, err := reg.Watch(ctx, "user-service")
	if err != nil {
		t.Fatalf("failed to watch: %v", err)
	}

	if ch == nil {
		t.Fatal("expected non-nil channel")
	}

	info := NewInstanceBuilder("user-service").
		ID("instance-1").
		Host("192.168.1.1").
		Port(8080).
		Build()

	_ = reg.Register(ctx, info)

	select {
	case instances := <-ch:
		if len(instances) != 1 {
			t.Errorf("expected 1 instance in watch, got %d", len(instances))
		}
	default:
		t.Error("expected watch channel to receive update")
	}
}

func TestMemoryRegistry_DiscoverOnlyHealthy(t *testing.T) {
	reg := NewMemoryRegistry()
	ctx := context.Background()

	healthy := NewInstanceBuilder("user-service").
		ID("instance-1").
		Host("192.168.1.1").
		Port(8080).
		Build()

	unhealthy := NewInstanceBuilder("user-service").
		ID("instance-2").
		Host("192.168.1.2").
		Port(8081).
		Build()
	unhealthy.Healthy = false

	_ = reg.Register(ctx, healthy)
	_ = reg.Register(ctx, unhealthy)

	instances, err := reg.Discover(ctx, "user-service")
	if err != nil {
		t.Fatalf("failed to discover: %v", err)
	}

	if len(instances) != 1 {
		t.Errorf("expected 1 healthy instance, got %d", len(instances))
	}

	if instances[0].ID != "instance-1" {
		t.Errorf("expected ID 'instance-1', got %s", instances[0].ID)
	}
}

func TestRandomSelector_Select(t *testing.T) {
	sel := NewRandomSelector()
	instances := []InstanceInfo{
		{ID: "instance-1"},
		{ID: "instance-2"},
	}

	instance, err := sel.Select(instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instance.ID != "instance-1" {
		t.Errorf("expected ID 'instance-1', got %s", instance.ID)
	}
}

func TestRandomSelector_EmptyInstances(t *testing.T) {
	sel := NewRandomSelector()

	_, err := sel.Select([]InstanceInfo{})
	if err == nil {
		t.Error("expected error for empty instances")
	}
}

func TestRoundRobinSelector_Select(t *testing.T) {
	sel := NewRoundRobinSelector()
	instances := []InstanceInfo{
		{ID: "instance-1"},
		{ID: "instance-2"},
	}

	// 第一次选择
	instance1, err := sel.Select(instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instance1.ID != "instance-1" {
		t.Errorf("expected ID 'instance-1', got %s", instance1.ID)
	}

	// 第二次选择
	instance2, err := sel.Select(instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instance2.ID != "instance-2" {
		t.Errorf("expected ID 'instance-2', got %s", instance2.ID)
	}

	// 第三次选择（循环）
	instance3, err := sel.Select(instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instance3.ID != "instance-1" {
		t.Errorf("expected ID 'instance-1' (cycle), got %s", instance3.ID)
	}
}

func TestRoundRobinSelector_EmptyInstances(t *testing.T) {
	sel := NewRoundRobinSelector()

	_, err := sel.Select([]InstanceInfo{})
	if err == nil {
		t.Error("expected error for empty instances")
	}
}

func TestWeightedSelector_Select(t *testing.T) {
	sel := NewWeightedSelector(map[string]int{
		"instance-1": 10,
		"instance-2": 20,
	})

	instances := []InstanceInfo{
		{ID: "instance-1"},
		{ID: "instance-2"},
	}

	instance, err := sel.Select(instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if instance.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestWeightedSelector_EmptyInstances(t *testing.T) {
	sel := NewWeightedSelector(nil)

	_, err := sel.Select([]InstanceInfo{})
	if err == nil {
		t.Error("expected error for empty instances")
	}
}
