package core

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestContainer_StressHighConcurrencyGet 测试高并发获取 Bean
func TestContainer_StressHighConcurrencyGet(t *testing.T) {
	t.Parallel()
	container := New()

	_ = container.Register("stressTestBean", Bean(&mockBean{Name: "stress"}), Singleton())

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := container.Get("stressTestBean")
			if err != nil {
				t.Errorf("Failed to get bean: %v", err)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	t.Logf("10000 concurrent Get() calls took %v", elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("Performance degradation: 10000 Get() calls took %v", elapsed)
	}
}

// TestContainer_StressHighConcurrencyRegister 测试高并发注册 Bean
func TestContainer_StressHighConcurrencyRegister(t *testing.T) {
	t.Parallel()
	container := New()

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			beanName := fmt.Sprintf("bean%d", id)
			err := container.Register(beanName, Bean(&mockBean{Name: beanName}))
			if err != nil && err != ErrDuplicateBean {
				t.Errorf("Failed to register bean: %v", err)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	t.Logf("1000 concurrent Register() calls took %v", elapsed)
}

// TestContainer_StressPrototypeCreation 测试原型作用域的高频创建
func TestContainer_StressPrototypeCreation(t *testing.T) {
	t.Parallel()
	container := New()

	_ = container.Register("prototypeBean", Bean(&mockBean{Name: "proto"}), Prototype())

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bean, err := container.Get("prototypeBean")
			if err != nil {
				t.Errorf("Failed to get bean: %v", err)
			}
			if bean == nil {
				t.Error("Expected non-nil bean")
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	t.Logf("10000 prototype bean creations took %v", elapsed)

	if elapsed > 10*time.Second {
		t.Errorf("Performance degradation: 10000 prototype creations took %v", elapsed)
	}
}
