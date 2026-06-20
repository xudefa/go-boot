package validation

import (
	"reflect"
	"sync"
	"testing"
)

// TestNewValidatorRegistry 测试创建验证器注册表
func TestNewValidatorRegistry(t *testing.T) {
	registry := NewValidatorRegistry()
	if registry == nil {
		t.Fatal("NewValidatorRegistry 返回 nil")
	}
}

// TestRegisterFunc 测试注册函数式验证器
func TestRegisterFunc(t *testing.T) {
	registry := NewValidatorRegistry()

	validatorFunc := func(field reflect.Value, param string) (bool, string) {
		return true, ""
	}

	registry.RegisterFunc("test", validatorFunc)

	validator, ok := registry.GetFunc("test")
	if !ok {
		t.Error("无法获取已注册的验证器")
	}

	if validator == nil {
		t.Error("验证器为 nil")
	}
}

// TestUnregister 测试注销验证器
func TestUnregister(t *testing.T) {
	registry := NewValidatorRegistry()

	registry.RegisterFunc("test", func(reflect.Value, string) (bool, string) {
		return true, ""
	})

	registry.Unregister("test")

	_, ok := registry.GetFunc("test")
	if ok {
		t.Error("验证器应该已被注销")
	}
}

// TestConcurrency 测试并发安全性
func TestConcurrency(t *testing.T) {
	registry := NewValidatorRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			registry.RegisterFunc(string(rune('a'+n%26)), func(reflect.Value, string) (bool, string) {
				return true, ""
			})
		}(i)
	}

	wg.Wait()

	validator, ok := registry.GetFunc("a")
	if !ok {
		t.Error("无法获取验证器")
	}

	if validator == nil {
		t.Error("验证器为 nil")
	}
}

// TestMultipleValidators 测试多个验证器
func TestMultipleValidators(t *testing.T) {
	registry := NewValidatorRegistry()

	registry.RegisterFunc("required", func(field reflect.Value, param string) (bool, string) {
		return field.Interface() != "", "字段不能为空"
	})

	registry.RegisterFunc("min", func(field reflect.Value, param string) (bool, string) {
		str := field.String()
		return len(str) >= 2, "字段长度不能少于2"
	})

	required, ok := registry.GetFunc("required")
	if !ok {
		t.Error("无法获取 required 验证器")
	}

	min, ok := registry.GetFunc("min")
	if !ok {
		t.Error("无法获取 min 验证器")
	}

	_, msg1 := required(reflect.ValueOf(""), "")
	if msg1 != "字段不能为空" {
		t.Errorf("预期消息 '字段不能为空'，得到 '%s'", msg1)
	}

	_, msg2 := min(reflect.ValueOf("a"), "")
	if msg2 != "字段长度不能少于2" {
		t.Errorf("预期消息 '字段长度不能少于2'，得到 '%s'", msg2)
	}
}
