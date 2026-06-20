package core

import (
	"fmt"
	"testing"
)

// BenchmarkContainer_Register_Single 测试单次注册性能
func BenchmarkContainer_Register_Single(b *testing.B) {
	container := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		beanName := fmt.Sprintf("benchBean%d", i)
		_ = container.Register(beanName, Bean(&mockBean{Name: beanName}))
	}
}

// BenchmarkContainer_Register_Batch 测试批量注册性能
func BenchmarkContainer_Register_Batch(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		container := New()
		for j := 0; j < 100; j++ {
			beanName := fmt.Sprintf("bean%d", j)
			_ = container.Register(beanName, Bean(&mockBean{Name: beanName}))
		}
	}
}

// BenchmarkContainer_Get_Singleton 测试单例获取性能
func BenchmarkContainer_Get_Singleton(b *testing.B) {
	container := New()
	_ = container.Register("benchSingleton", Bean(&mockBean{Name: "bench"}), Singleton())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = container.Get("benchSingleton")
	}
}

// BenchmarkContainer_Get_Prototype 测试原型获取性能
func BenchmarkContainer_Get_Prototype(b *testing.B) {
	container := New()
	_ = container.Register("benchPrototype", Bean(&mockBean{Name: "bench"}), Prototype())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = container.Get("benchPrototype")
	}
}

// BenchmarkContainer_Inject 测试字段注入性能
func BenchmarkContainer_Inject(b *testing.B) {
	container := New()
	_ = container.Register("benchService", Bean(&mockBean{Name: "bench"}), Singleton())

	type Handler struct {
		Service *mockBean `inject:"benchService"`
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var h Handler
		_ = container.Inject(&h)
	}
}
