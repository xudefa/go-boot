package core

import (
	"reflect"
	"testing"
)

// TestIsComponent 测试 IsComponent 函数
// 验证：能够正确识别嵌入 Component 类型的结构体
func TestIsComponent(t *testing.T) {
	type structWithComponent struct {
		Component
		Name string
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		expected bool
	}{
		{
			name:     "struct with Component field",
			typeVal:  reflect.TypeFor[structWithComponent](),
			expected: true,
		},
		{
			name:     "struct without Component field",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeFor[string](),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsComponent(tt.typeVal)
			if result != tt.expected {
				t.Errorf("IsComponent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsConfiguration 测试 IsConfiguration 函数
// 验证：能够正确识别嵌入 Configuration 类型的结构体
func TestIsConfiguration(t *testing.T) {
	type structWithConfiguration struct {
		Configuration
		Value string
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		expected bool
	}{
		{
			name:     "struct with Configuration field",
			typeVal:  reflect.TypeFor[structWithConfiguration](),
			expected: true,
		},
		{
			name:     "struct without Configuration field",
			typeVal:  reflect.TypeFor[struct{ Value string }](),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeFor[int](),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfiguration(tt.typeVal)
			if result != tt.expected {
				t.Errorf("IsConfiguration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsService 测试 IsService 函数
// 验证：能够正确识别嵌入 Service 类型的结构体
func TestIsService(t *testing.T) {
	type structWithService struct {
		Service
		Host string
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		expected bool
	}{
		{
			name:     "struct with Service field",
			typeVal:  reflect.TypeFor[structWithService](),
			expected: true,
		},
		{
			name:     "struct without Service field",
			typeVal:  reflect.TypeFor[struct{ Host string }](),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeFor[bool](),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsService(tt.typeVal)
			if result != tt.expected {
				t.Errorf("IsService() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsRepository 测试 IsRepository 函数
// 验证：能够正确识别嵌入 Repository 类型的结构体
func TestIsRepository(t *testing.T) {
	type structWithRepository struct {
		Repository
		Table string
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		expected bool
	}{
		{
			name:     "struct with Repository field",
			typeVal:  reflect.TypeFor[structWithRepository](),
			expected: true,
		},
		{
			name:     "struct without Repository field",
			typeVal:  reflect.TypeFor[struct{ Table string }](),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeFor[[]int](),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRepository(tt.typeVal)
			if result != tt.expected {
				t.Errorf("IsRepository() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetComponentName 测试 GetComponentName 函数
// 验证：能够正确获取组件标签中指定的名称
func TestGetComponentName(t *testing.T) {
	type testStruct struct {
		Component string `component:"customComponent"`
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		field    string
		expected string
	}{
		{
			name:     "with component tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Component",
			expected: "customComponent",
		},
		{
			name:     "without component tag",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			field:    "Name",
			expected: "Name",
		},
		{
			name:     "non-existent field",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			field:    "NonExistent",
			expected: "NonExistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetComponentName(tt.typeVal, tt.field)
			if result != tt.expected {
				t.Errorf("GetComponentName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetInjectTag 测试 GetInjectTag 函数
// 验证：能够正确获取字段的 inject 标签值
func TestGetInjectTag(t *testing.T) {
	type testStruct struct {
		Service string `inject:"myService"`
		Logger  string `inject:""`
		Name    string
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		field    string
		expected string
	}{
		{
			name:     "with inject tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Service",
			expected: "myService",
		},
		{
			name:     "with empty inject tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Logger",
			expected: "",
		},
		{
			name:     "without inject tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Name",
			expected: "",
		},
		{
			name:     "non-existent field",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "NonExistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInjectTag(tt.typeVal, tt.field)
			if result != tt.expected {
				t.Errorf("GetInjectTag() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetConfigurationName 测试 GetConfigurationName 函数
// 验证：能够正确获取配置组件标签中指定的名称
func TestGetConfigurationName(t *testing.T) {
	type testStruct struct {
		Configuration string `configuration:"customConfig"`
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		field    string
		expected string
	}{
		{
			name:     "with configuration tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Configuration",
			expected: "customConfig",
		},
		{
			name:     "without configuration tag",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			field:    "Name",
			expected: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigurationName(tt.typeVal, tt.field)
			if result != tt.expected {
				t.Errorf("GetConfigurationName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetServiceName 测试 GetServiceName 函数
// 验证：能够正确获取服务组件标签中指定的名称
func TestGetServiceName(t *testing.T) {
	type testStruct struct {
		Service string `service:"customService"`
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		field    string
		expected string
	}{
		{
			name:     "with service tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Service",
			expected: "customService",
		},
		{
			name:     "without service tag",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			field:    "Name",
			expected: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetServiceName(tt.typeVal, tt.field)
			if result != tt.expected {
				t.Errorf("GetServiceName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetRepositoryName 测试 GetRepositoryName 函数
// 验证：能够正确获取仓储组件标签中指定的名称
func TestGetRepositoryName(t *testing.T) {
	type testStruct struct {
		Repository string `repository:"customRepo"`
	}

	tests := []struct {
		name     string
		typeVal  reflect.Type
		field    string
		expected string
	}{
		{
			name:     "with repository tag",
			typeVal:  reflect.TypeFor[testStruct](),
			field:    "Repository",
			expected: "customRepo",
		},
		{
			name:     "without repository tag",
			typeVal:  reflect.TypeFor[struct{ Name string }](),
			field:    "Name",
			expected: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRepositoryName(tt.typeVal, tt.field)
			if result != tt.expected {
				t.Errorf("GetRepositoryName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
