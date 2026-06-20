// Package config 提供配置验证功能。
//
// 内置 Validator 接口和 DefaultValidator 实现，支持必填字段检查、
// 最小/最大值限制、正则表达式匹配、枚举值校验和自定义验证规则。
package config

import (
	"regexp"
	"slices"
)

// Validator 配置验证器接口.
//
// 定义配置验证的统一方式,支持多种验证规则.
type Validator interface {
	// Validate 验证配置数据.
	//
	// 参数:
	//   - target: 待验证的配置数据,通常为map[string]any
	//
	// 返回:
	//   - error: 验证错误,如果验证通过返回nil
	Validate(target any) error
}

// Rules 验证规则结构体.
//
// 存储配置验证的各种规则.
// 字段说明:
//   - Required: 必填字段列表
//   - Min: 最小值限制,如 map[string]int{"port": 1}
//   - Max: 最大值限制,如 map[string]int{"port": 65535}
//   - Regex: 正则表达式,如 map[string]string{"email": "^...$"}
//   - Enum: 枚举值,如 map[string][]any{"status": []any{"active", "inactive"}}
//   - Custom: 自定义验证函数
type Rules struct {
	Required []string
	Min      map[string]int
	Max      map[string]int
	Regex    map[string]string
	Enum     map[string][]any
	Custom   map[string]func(any) error
	regexMap map[string]*regexp.Regexp // 缓存编译后的正则表达式
}

// ValidationError 验证错误结构体.
//
// 携带验证失败的字段和错误信息.
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现error接口.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// DefaultValidator 默认验证器实现.
//
// 提供链式API来构建验证规则.
type DefaultValidator struct {
	rules Rules
}

// NewValidator 创建默认验证器实例.
//
// 返回:
//   - *DefaultValidator: 验证器实例
func NewValidator() *DefaultValidator {
	return &DefaultValidator{
		rules: Rules{
			Min:      make(map[string]int),
			Max:      make(map[string]int),
			Regex:    make(map[string]string),
			Enum:     make(map[string][]any),
			Custom:   make(map[string]func(any) error),
			regexMap: make(map[string]*regexp.Regexp),
		},
	}
}

// Validate 验证配置数据.
//
// 参数:
//   - target: 待验证的配置数据,必须为map[string]any类型
//
// 返回:
//   - error: 验证错误,如果验证通过返回nil
func (v *DefaultValidator) Validate(target any) error {
	if target == nil {
		return &ValidationError{Field: "", Message: "target is nil"}
	}
	data, ok := target.(map[string]any)
	if !ok {
		return &ValidationError{Field: "", Message: "target must be map[string]any"}
	}

	// 验证必填字段
	for _, field := range v.rules.Required {
		if _, exists := data[field]; !exists {
			return &ValidationError{Field: field, Message: "required field missing"}
		}
	}

	// 验证最小值（支持 int 和 float64）
	for field, minInt := range v.rules.Min {
		if val, ok := data[field].(int); ok && val < minInt {
			return &ValidationError{Field: field, Message: "value below minimum"}
		}
		if val, ok := data[field].(float64); ok && int(val) < minInt {
			return &ValidationError{Field: field, Message: "value below minimum"}
		}
	}

	// 验证最大值（支持 int 和 float64）
	for field, maxInt := range v.rules.Max {
		if val, ok := data[field].(int); ok && val > maxInt {
			return &ValidationError{Field: field, Message: "value above maximum"}
		}
		if val, ok := data[field].(float64); ok && int(val) > maxInt {
			return &ValidationError{Field: field, Message: "value above maximum"}
		}
	}

	// 验证正则表达式
	for field, pattern := range v.rules.Regex {
		if val, ok := data[field].(string); ok {
			re, ok := v.rules.regexMap[pattern]
			if !ok {
				return &ValidationError{Field: field, Message: "regex pattern not compiled"}
			}
			if !re.MatchString(val) {
				return &ValidationError{Field: field, Message: "regex not matched"}
			}
		}
	}

	// 验证枚举值
	for field, allowed := range v.rules.Enum {
		if val, exists := data[field]; exists {
			found := slices.Contains(allowed, val)
			if !found {
				return &ValidationError{Field: field, Message: "value not in enum"}
			}
		}
	}

	// 执行自定义验证
	for field, fn := range v.rules.Custom {
		if val, exists := data[field]; exists {
			if err := fn(val); err != nil {
				return &ValidationError{Field: field, Message: err.Error()}
			}
		}
	}

	return nil
}

// AddRequired 添加必填字段.
//
// 参数:
//   - fields: 字段名列表
//
// 返回:
//   - *DefaultValidator: 验证器实例(支持链式调用)
func (v *DefaultValidator) AddRequired(fields ...string) *DefaultValidator {
	v.rules.Required = append(v.rules.Required, fields...)
	return v
}

// AddMin 添加最小值限制.
//
// 参数:
//   - field: 字段名
//   - min: 最小值
//
// 返回:
//   - *DefaultValidator: 验证器实例
func (v *DefaultValidator) AddMin(field string, min int) *DefaultValidator {
	v.rules.Min[field] = min
	return v
}

// AddMax 添加最大值限制.
//
// 参数:
//   - field: 字段名
//   - max: 最大值
//
// 返回:
//   - *DefaultValidator: 验证器实例
func (v *DefaultValidator) AddMax(field string, max int) *DefaultValidator {
	v.rules.Max[field] = max
	return v
}

// AddRegex 添加正则表达式限制.
//
// 参数:
//   - field: 字段名
//   - pattern: 正则表达式
//
// 返回:
//   - *DefaultValidator: 验证器实例
func (v *DefaultValidator) AddRegex(field string, pattern string) *DefaultValidator {
	v.rules.Regex[field] = pattern
	v.rules.regexMap[pattern] = regexp.MustCompile(pattern)
	return v
}

// AddEnum 添加枚举值限制.
//
// 参数:
//   - field: 字段名
//   - values: 允许的值列表
//
// 返回:
//   - *DefaultValidator: 验证器实例
func (v *DefaultValidator) AddEnum(field string, values ...any) *DefaultValidator {
	v.rules.Enum[field] = values
	return v
}

// AddCustomRule 添加自定义验证规则.
//
// 参数:
//   - field: 字段名
//   - fn: 自定义验证函数
//
// 返回:
//   - *DefaultValidator: 验证器实例
func (v *DefaultValidator) AddCustomRule(field string, fn func(any) error) *DefaultValidator {
	v.rules.Custom[field] = fn
	return v
}
