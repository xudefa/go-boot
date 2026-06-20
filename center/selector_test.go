package center

import (
	"testing"
)

// TestRandomSelect_Empty_ReturnsErrNoInstances 验证 RandomSelect 在传入 nil 或空列表时返回 ErrNoInstances。
func TestRandomSelect_Empty_ReturnsErrNoInstances(t *testing.T) {
	_, err := RandomSelect.Select(nil)
	if err != ErrNoInstances {
		t.Fatalf("expected ErrNoInstances, got %v", err)
	}
	_, err = RandomSelect.Select([]InstanceInfo{})
	if err != ErrNoInstances {
		t.Fatalf("expected ErrNoInstances, got %v", err)
	}
}

// TestRandomSelect_ReturnsOneInstance 验证 RandomSelect 从多个实例中随机返回一个。
// 循环 100 次确保每次都能正确返回实例之一，验证随机选择的基本正确性。
func TestRandomSelect_ReturnsOneInstance(t *testing.T) {
	instances := []InstanceInfo{
		{ServiceName: "srv", ID: "1", Host: "a", Port: 1},
		{ServiceName: "srv", ID: "2", Host: "b", Port: 2},
	}
	for range 100 {
		got, err := RandomSelect.Select(instances)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != "1" && got.ID != "2" {
			t.Fatalf("unexpected instance ID: %s", got.ID)
		}
	}
}

// TestRoundRobinSelect_Empty_ReturnsErrNoInstances 验证 RoundRobinSelect 在传入 nil 时返回 ErrNoInstances。
func TestRoundRobinSelect_Empty_ReturnsErrNoInstances(t *testing.T) {
	_, err := RoundRobinSelect.Select(nil)
	if err != ErrNoInstances {
		t.Fatalf("expected ErrNoInstances, got %v", err)
	}
}

// TestRoundRobinSelect_CyclesThroughInstances 验证 RoundRobinSelect 按轮询顺序依次返回实例。
// 对于 3 个实例，期望输出顺序为 "1","2","3","1","2","3"，验证轮询调度算法的正确性。
func TestRoundRobinSelect_CyclesThroughInstances(t *testing.T) {
	instances := []InstanceInfo{
		{ServiceName: "srv", ID: "1", Host: "a", Port: 1},
		{ServiceName: "srv", ID: "2", Host: "b", Port: 2},
		{ServiceName: "srv", ID: "3", Host: "c", Port: 3},
	}
	expected := []string{"1", "2", "3", "1", "2", "3"}
	for i, want := range expected {
		got, err := RoundRobinSelect.Select(instances)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != want {
			t.Fatalf("round %d: expected ID %s, got %s", i, want, got.ID)
		}
	}
}

// TestWeightedRandomSelect_Empty_ReturnsErrNoInstances 验证 WeightedRandomSelect 在传入 nil 时返回 ErrNoInstances。
func TestWeightedRandomSelect_Empty_ReturnsErrNoInstances(t *testing.T) {
	s := NewWeightedRandomSelect()
	_, err := s.Select(nil)
	if err != ErrNoInstances {
		t.Fatalf("expected ErrNoInstances, got %v", err)
	}
}

// TestWeightedRandomSelect_ReturnsOneInstance 验证 WeightedRandomSelect 按权重概率返回实例。
// 两个实例权重分别为 1 和 9，循环 1000 次后确认两个实例都被选中过，验证权重随机算法的基本正确性。
func TestWeightedRandomSelect_ReturnsOneInstance(t *testing.T) {
	s := NewWeightedRandomSelect()
	instances := []InstanceInfo{
		{ServiceName: "srv", ID: "1", Host: "a", Port: 1, Weight: 1},
		{ServiceName: "srv", ID: "2", Host: "b", Port: 2, Weight: 9},
	}
	counts := map[string]int{"1": 0, "2": 0}
	for range 1000 {
		got, err := s.Select(instances)
		if err != nil {
			t.Fatal(err)
		}
		counts[got.ID]++
	}
	if counts["1"] == 0 || counts["2"] == 0 {
		t.Fatalf("expected both instances selected, got %v", counts)
	}
}

// TestSelectors_ConcurrentSafe 验证所有选择器（RandomSelect、RoundRobinSelect、WeightedRandomSelect）的并发安全性。
// 对每个选择器启动 10 个 goroutine，每个执行 100 次 Select 操作，验证在并发场景下不会出现数据竞争或崩溃。
func TestSelectors_ConcurrentSafe(t *testing.T) {
	instances := []InstanceInfo{
		{ServiceName: "srv", ID: "1", Host: "a", Port: 1},
		{ServiceName: "srv", ID: "2", Host: "b", Port: 2},
	}
	selectors := []Selector{RandomSelect, RoundRobinSelect, NewWeightedRandomSelect()}
	for _, sel := range selectors {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			done := make(chan struct{})
			for range 10 {
				go func() {
					for range 100 {
						_, _ = sel.Select(instances)
					}
					done <- struct{}{}
				}()
			}
			for range 10 {
				<-done
			}
		})
	}
}
