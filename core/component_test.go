package core

import (
	"reflect"
	"testing"
)

type testComponent struct {
	Component
	Name string
}

type testConfiguration struct {
	Configuration
	Value string
}

type testService struct {
	Service
	Host string
}

type testRepository struct {
	Repository
	Table string
}

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
			typeVal:  reflect.TypeOf(structWithComponent{}),
			expected: true,
		},
		{
			name:     "struct without Component field",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeOf("string"),
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
			typeVal:  reflect.TypeOf(structWithConfiguration{}),
			expected: true,
		},
		{
			name:     "struct without Configuration field",
			typeVal:  reflect.TypeOf(struct{ Value string }{}),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeOf(123),
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
			typeVal:  reflect.TypeOf(structWithService{}),
			expected: true,
		},
		{
			name:     "struct without Service field",
			typeVal:  reflect.TypeOf(struct{ Host string }{}),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeOf(true),
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
			typeVal:  reflect.TypeOf(structWithRepository{}),
			expected: true,
		},
		{
			name:     "struct without Repository field",
			typeVal:  reflect.TypeOf(struct{ Table string }{}),
			expected: false,
		},
		{
			name:     "non-struct type",
			typeVal:  reflect.TypeOf([]int{}),
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
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Component",
			expected: "customComponent",
		},
		{
			name:     "without component tag",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
			field:    "Name",
			expected: "Name",
		},
		{
			name:     "non-existent field",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
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
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Service",
			expected: "myService",
		},
		{
			name:     "with empty inject tag",
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Logger",
			expected: "",
		},
		{
			name:     "without inject tag",
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Name",
			expected: "",
		},
		{
			name:     "non-existent field",
			typeVal:  reflect.TypeOf(testStruct{}),
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
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Configuration",
			expected: "customConfig",
		},
		{
			name:     "without configuration tag",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
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
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Service",
			expected: "customService",
		},
		{
			name:     "without service tag",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
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
			typeVal:  reflect.TypeOf(testStruct{}),
			field:    "Repository",
			expected: "customRepo",
		},
		{
			name:     "without repository tag",
			typeVal:  reflect.TypeOf(struct{ Name string }{}),
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
