package core

import (
	"reflect"
)

const (
	TagName          = "inject"
	ComponentTag     = "component"
	ConfigurationTag = "configuration"
	ServiceTag       = "service"
	RepositoryTag    = "repository"
)

// Component 结构体组件标签
//
// 标记在结构体上,用于自动扫描注册
type Component struct {
	Name string
}

// IsComponent 检查类型是否为组件
//
// 参数:
//   - t: 反射类型
//
// 返回值:
//   - bool: 是否为组件
func IsComponent(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Component" {
			return true
		}
	}
	return false
}

// GetComponentName 获取组件名称
//
// 参数:
//   - t: 反射类型
//   - fieldName: 字段名
//
// 返回值:
//   - string: 组件名称
func GetComponentName(t reflect.Type, fieldName string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return fieldName
	}
	tag := f.Tag.Get(ComponentTag)
	if tag == "" {
		return fieldName
	}
	if tag == "-" {
		return ""
	}
	return tag
}

// GetInjectTag 获取注入标签
//
// 参数:
//   - t: 反射类型
//   - fieldName: 字段名
//
// 返回值:
//   - string: 注入标签值
func GetInjectTag(t reflect.Type, fieldName string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return ""
	}
	return f.Tag.Get(TagName)
}

// Configuration 结构体配置组件标签
//
// 标记在结构体上,用于自动扫描注册
type Configuration struct {
	Name string
}

// Service 结构体服务组件标签
//
// 标记在结构体上,用于自动扫描注册
type Service struct {
	Name string
}

// Repository 结构体仓储组件标签
//
// 标记在结构体上,用于自动扫描注册
type Repository struct {
	Name string
}

// IsConfiguration 检查类型是否为配置组件
func IsConfiguration(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Configuration" {
			return true
		}
	}
	return false
}

// IsService 检查类型是否为服务组件
func IsService(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Service" {
			return true
		}
	}
	return false
}

// IsRepository 检查类型是否为仓储组件
func IsRepository(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Repository" {
			return true
		}
	}
	return false
}

// GetConfigurationName 获取配置组件名称
func GetConfigurationName(t reflect.Type, fieldName string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return fieldName
	}
	tag := f.Tag.Get(ConfigurationTag)
	if tag == "" {
		return fieldName
	}
	if tag == "-" {
		return ""
	}
	return tag
}

// GetServiceName 获取服务组件名称
func GetServiceName(t reflect.Type, fieldName string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return fieldName
	}
	tag := f.Tag.Get(ServiceTag)
	if tag == "" {
		return fieldName
	}
	if tag == "-" {
		return ""
	}
	return tag
}

// GetRepositoryName 获取仓储组件名称
func GetRepositoryName(t reflect.Type, fieldName string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return fieldName
	}
	tag := f.Tag.Get(RepositoryTag)
	if tag == "" {
		return fieldName
	}
	if tag == "-" {
		return ""
	}
	return tag
}
