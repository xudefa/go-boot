package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewComponentScanner 测试创建组件扫描器
// 验证：NewComponentScanner 函数返回非 nil 的扫描器实例
func TestNewComponentScanner(t *testing.T) {
	t.Parallel()
	scanner := NewComponentScanner(".")
	if scanner == nil {
		t.Error("NewComponentScanner() returned nil")
	}
}

// TestComponentScanner_Scan 测试扫描器扫描功能
// 验证：扫描器能够正确识别各种组件注释并注册到容器
func TestComponentScanner_Scan(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// @Component
type UserService struct {
	Name string
}

// @Configuration
type AppConfig struct {
	Port int
}

// @Service
type UserRepo struct {
	Table string
}

// @Component("customName")
type TestComponent struct {
	Field string
}

// RegularStruct 不是组件
type RegularStruct struct {
	Field string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if !container.Has("userService") {
		t.Error("userService should be registered")
	}

	if !container.Has("appConfig") {
		t.Error("appConfig should be registered")
	}

	if !container.Has("userRepo") {
		t.Error("userRepo should be registered")
	}

	if !container.Has("customName") {
		t.Error("customName should be registered")
	}

	if container.Has("regularStruct") {
		t.Error("regularStruct should not be registered")
	}
}

// TestComponentScanner_Scan_NilDoc 测试扫描无注释的结构体
// 验证：无组件注释的结构体不会被注册到容器
func TestComponentScanner_Scan_NilDoc(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type NoComment struct {
	Name string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if container.Has("noComment") {
		t.Error("noComment should not be registered")
	}
}

// TestComponentScanner_WithCustomName 测试自定义名称的组件
// 验证：注释中指定名称的组件使用指定名称注册
func TestComponentScanner_WithCustomName(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// @Component("myCustomService")
type MyService struct {
	Name string
}

// @Component
type AnotherService struct {
	Address string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if !container.Has("myCustomService") {
		t.Error("myCustomService should be registered with custom name")
	}
	if !container.Has("anotherService") {
		t.Error("anotherService should be registered")
	}
}

// TestComponentScanner_LazyLookup_GetAll 测试懒加载 bean 的 GetAll 可见性
// 验证：当 struct 包含 AST 无法重建的字段类型时（如 *ast.Ident 非基础类型），
// 懒加载路径的 bean 在首次 Get 后应能通过 GetAll 按接口类型找到
//
// BUG 2：Factory(nil) 导致 ConcreteType 为空，GetAll(type) 遍历 def 时跳过此类 bean
func TestComponentScanner_LazyLookup_GetAll(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "lazy.go")
	content := `package test

type DepService struct {
	Name string
}

// @Component
type MyService struct {
	Dep *DepService
	Val string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	if err := scanner.Scan(container); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if !container.Has("myService") {
		t.Fatal("myService should be registered")
	}

	// 首次 Get 触发工厂，补齐 ConcreteType
	bean, err := container.Get("myService")
	if err != nil {
		t.Fatalf("Get(myService) error = %v", err)
	}
	if bean == nil {
		t.Fatal("Get(myService) returned nil")
	}

	// GetAll by interface or concrete type should find myService
	// After the lazy ConcreteType fix, typeToIDs should have the entry
	for _, info := range container.ListBeans() {
		if info.ID == "myService" {
			if info.Type == "" || info.Type == "<nil>" {
				t.Error("BUG: myService has no type info in ListBeans - GetAll will miss it")
			}
			return
		}
	}
	t.Error("myService not found in ListBeans")
}

// TestComponentScanner_LazyLookup_TypeFillsAfterGet 验证懒加载 bean 在 Get 后类型信息被填充
func TestComponentScanner_LazyLookup_TypeFillsAfterGet(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "lazy2.go")
	content := `package test

type Config struct {
	Host string
}

// @Service
type UserService struct {
	Cfg *Config
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	if err := scanner.Scan(container); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Get 前 type 应为空
	info := getBeanInfo(container, "userService")
	if info == nil {
		t.Fatal("userService not registered")
	}

	// 首次 Get
	if _, err := container.Get("userService"); err != nil {
		t.Fatalf("Get(userService) error = %v", err)
	}

	// Get 后 type 应被补齐
	newInfo := getBeanInfo(container, "userService")
	if newInfo == nil {
		t.Fatal("userService not found after Get")
	}
	if newInfo.Type == "" || newInfo.Type == "<nil>" {
		t.Error("BUG: userService type was not filled after Get")
	}
}

func getBeanInfo(c Container, id string) *BeanInfo {
	for _, info := range c.ListBeans() {
		if info.ID == id {
			return &info
		}
	}
	return nil
}

// TestToFirstCharLower 测试首字母小写转换函数
// 验证：能够正确将字符串首字母转为小写
func TestToFirstCharLower(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"UserService", "userService"},
		{"Config", "config"},
		{"A", "a"},
		{"", ""},
		{"ABC", "aBC"},
	}

	for _, tt := range tests {
		result := toFirstCharLower(tt.input)
		if result != tt.expected {
			t.Errorf("toFirstCharLower(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
