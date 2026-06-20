// Package core 提供依赖注入 (DI) 容器、AOP 框架和组件扫描功能
//
// # 核心概念
//
//   - 容器 (Container): 管理 Bean 的注册、解析和生命周期
//   - 组件 (Component): 通过结构体嵌入 Component/Service/Repository/Configuration 标记组件
//   - 组件扫描: 扫描并自动注册带有组件标签的结构体
//
// # 组件标签
//
// 通过在结构体中嵌入特定类型实现组件标记:
//
//	type UserService struct {
//	    core.Service
//	}
package core

import (
	"reflect"
)

// 结构体标签常量定义
const (
	TagName          = "inject"        // 依赖注入标签名
	ComponentTag     = "component"     // 通用组件标签名
	ConfigurationTag = "configuration" // 配置组件标签名
	ServiceTag       = "service"       // 服务组件标签名
	RepositoryTag    = "repository"    // 仓储组件标签名
)

// hasEmbeddedField 检查结构体类型是否包含指定名称的嵌入字段
func hasEmbeddedField(t reflect.Type, fieldName string) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := range t.NumField() {
		if t.Field(i).Name == fieldName {
			return true
		}
	}
	return false
}

// getTagName 获取结构体字段指定标签的值
//
// 参数:
//   - t: 反射类型
//   - fieldName: 字段名
//   - tag: 标签名
//
// 返回值:
//   - string: 标签值，标签不存在或为空时返回 fieldName
func getTagName(t reflect.Type, fieldName string, tag string) string {
	f, ok := t.FieldByName(fieldName)
	if !ok {
		return fieldName
	}
	tagVal := f.Tag.Get(tag)
	if tagVal == "" {
		return fieldName
	}
	if tagVal == "-" {
		return ""
	}
	return tagVal
}

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
	return hasEmbeddedField(t, "Component")
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
	return getTagName(t, fieldName, ComponentTag)
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
	return hasEmbeddedField(t, "Configuration")
}

// IsService 检查类型是否为服务组件
func IsService(t reflect.Type) bool {
	return hasEmbeddedField(t, "Service")
}

// IsRepository 检查类型是否为仓储组件
func IsRepository(t reflect.Type) bool {
	return hasEmbeddedField(t, "Repository")
}

// GetConfigurationName 获取配置组件名称
func GetConfigurationName(t reflect.Type, fieldName string) string {
	return getTagName(t, fieldName, ConfigurationTag)
}

// GetServiceName 获取服务组件名称
func GetServiceName(t reflect.Type, fieldName string) string {
	return getTagName(t, fieldName, ServiceTag)
}

// GetRepositoryName 获取仓储组件名称
func GetRepositoryName(t reflect.Type, fieldName string) string {
	return getTagName(t, fieldName, RepositoryTag)
}
