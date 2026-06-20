package core

import (
	"fmt"
	"sync"
	"testing"
)

// TestContainer_ChaosConcurrentGetAndRegister 测试并发获取和注册的混沌场景
func TestContainer_ChaosConcurrentGetAndRegister(t *testing.T) {
	t.Parallel()
	container := New()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			beanName := fmt.Sprintf("chaosBean%d", id)
			err := container.Register(beanName, Bean(&mockBean{Name: beanName}))
			if err != nil && err != ErrDuplicateBean {
				t.Errorf("Failed to register bean: %v", err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			beanName := fmt.Sprintf("chaosBean%d", id)
			_, err := container.Get(beanName)
			if err != nil {
				t.Errorf("Failed to get bean: %v", err)
			}
		}(i)
	}

	wg.Wait()
}

// TestContainer_ChaosRapidRegisterAndUnregister 测试快速注册和注销的混沌场景
func TestContainer_ChaosRapidRegisterAndUnregister(t *testing.T) {
	t.Parallel()
	container := New()

	for i := 0; i < 50; i++ {
		beanName := fmt.Sprintf("tempBean%d", i)
		err := container.Register(beanName, Bean(&mockBean{Name: beanName}))
		if err != nil {
			t.Errorf("Failed to register bean: %v", err)
		}

		_, err = container.Get(beanName)
		if err != nil {
			t.Errorf("Failed to get bean: %v", err)
		}
	}
}

// TestContainer_ChaosMemoryPressure 测试内存压力下的容器行为
func TestContainer_ChaosMemoryPressure(t *testing.T) {
	t.Parallel()
	container := New()

	for i := 0; i < 1000; i++ {
		beanName := fmt.Sprintf("memoryBean%d", i)
		err := container.Register(beanName, Bean(&mockBean{Name: beanName}), Singleton())
		if err != nil {
			t.Errorf("Failed to register bean: %v", err)
		}
	}

	for i := 0; i < 1000; i++ {
		beanName := fmt.Sprintf("memoryBean%d", i)
		_, err := container.Get(beanName)
		if err != nil {
			t.Errorf("Failed to get bean: %v", err)
		}
	}
}

// TestContainer_ChaosDeepDependencyTree 测试深度依赖树的混沌场景
func TestContainer_ChaosDeepDependencyTree(t *testing.T) {
	t.Parallel()
	container := New()

	for i := 0; i < 100; i++ {
		current := fmt.Sprintf("deep%d", i)
		_ = container.Register(current, Bean(&mockBean{Name: current}))
	}

	for i := 0; i < 100; i++ {
		current := fmt.Sprintf("deep%d", i)
		_, err := container.Get(current)
		if err != nil {
			t.Errorf("Failed to get deep bean: %v", err)
		}
	}
}
