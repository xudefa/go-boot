// Package validation 提供 HTTP 请求验证能力
//
// 支持：
//   - Query/Header/Body 参数验证
//   - 多种验证类型：required, string, number, email, regex, enum, min, max, length
//   - 快速失败模式
//   - JSON Body 验证
//   - 自定义错误消息
//
// 使用示例：
//
//	validator, _ := validation.NewRequestValidator(validation.ValidationConfig{
//	    Source: "query",
//	    Rules: []validation.ValidationRule{
//	        {Field: "page", Type: "required"},
//	        {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
//	    },
//	})
//
//	result := validator.Validate(req)
//	if !result.Valid {
//	    // 处理验证错误
//	}
package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule 验证规则
type ValidationRule struct {
	// Field 字段名称
	Field string `json:"field"`

	// Type 验证类型：required, string, number, email, regex, enum, min, max, length
	Type string `json:"type"`

	// Value 验证值（用于 enum, regex 等）
	Value string `json:"value,omitempty"`

	// Min 最小值
	Min *float64 `json:"min,omitempty"`

	// Max 最大值
	Max *float64 `json:"max,omitempty"`

	// MinLength 最小长度
	MinLength *int `json:"min_length,omitempty"`

	// MaxLength 最大长度
	MaxLength *int `json:"max_length,omitempty"`

	// Pattern 正则表达式
	Pattern string `json:"pattern,omitempty"`

	// Message 自定义错误消息
	Message string `json:"message,omitempty"`

	// In 枚举值
	In []string `json:"in,omitempty"`
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	// Rules 验证规则
	Rules []ValidationRule `json:"rules"`

	// Source 验证来源：query, header, body
	Source string `json:"source"`

	// FailFast 快速失败（遇到第一个错误就停止）
	FailFast bool `json:"fail_fast"`
}

// RuleValidationError 验证错误
type RuleValidationError struct {
	// Field 字段名称
	Field string `json:"field"`

	// Message 错误消息
	Message string `json:"message"`

	// Type 错误类型
	Type string `json:"type"`
}

// RuleValidationResult 验证结果
type RuleValidationResult struct {
	// Valid 是否通过验证
	Valid bool `json:"valid"`

	// Errors 错误列表
	Errors []RuleValidationError `json:"errors,omitempty"`
}

// RequestValidator HTTP 请求验证器
type RequestValidator struct {
	config ValidationConfig
	regexs map[string]*regexp.Regexp
}

// NewRequestValidator 创建请求验证器
func NewRequestValidator(config ValidationConfig) (*RequestValidator, error) {
	v := &RequestValidator{
		config: config,
		regexs: make(map[string]*regexp.Regexp),
	}

	// 预编译正则表达式
	for i, rule := range config.Rules {
		if rule.Pattern != "" {
			regex, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern for field %s: %w", rule.Field, err)
			}
			v.regexs[rule.Field] = regex
			config.Rules[i].Pattern = "" // 清空，使用预编译的
		}
	}

	return v, nil
}

// GetConfig 获取验证配置
func (v *RequestValidator) GetConfig() ValidationConfig {
	return v.config
}

// Validate 验证请求
func (v *RequestValidator) Validate(req *http.Request) *RuleValidationResult {
	result := &RuleValidationResult{Valid: true}

	for _, rule := range v.config.Rules {
		var value string

		// 根据来源提取值
		switch v.config.Source {
		case "query":
			value = req.URL.Query().Get(rule.Field)
		case "header":
			value = req.Header.Get(rule.Field)
		case "body":
			// body 验证需要特殊处理，这里简化
			continue
		}

		// 执行验证
		if err := v.validateRule(rule, value); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, RuleValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    rule.Type,
			})

			if v.config.FailFast {
				return result
			}
		}
	}

	return result
}

// validateRule 验证单个规则
func (v *RequestValidator) validateRule(rule ValidationRule, value string) error {
	switch rule.Type {
	case "required":
		if value == "" {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s is required", rule.Field))
		}

	case "string":
		if value == "" {
			return nil // 空值不验证
		}
		if rule.MinLength != nil && len(value) < *rule.MinLength {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %d characters", rule.Field, *rule.MinLength))
		}
		if rule.MaxLength != nil && len(value) > *rule.MaxLength {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %d characters", rule.Field, *rule.MaxLength))
		}

	case "number":
		if value == "" {
			return nil
		}
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
		}
		if rule.Min != nil && num < *rule.Min {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %f", rule.Field, *rule.Min))
		}
		if rule.Max != nil && num > *rule.Max {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %f", rule.Field, *rule.Max))
		}

	case "email":
		if value == "" {
			return nil
		}
		if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a valid email", rule.Field))
		}

	case "regex":
		if value == "" {
			return nil
		}
		if regex, ok := v.regexs[rule.Field]; ok {
			if !regex.MatchString(value) {
				return fmt.Errorf("%s", rule.MessageOrDefault("%s format is invalid", rule.Field))
			}
		}

	case "enum":
		if value == "" {
			return nil
		}
		for _, val := range rule.In {
			if value == val {
				return nil
			}
		}
		return fmt.Errorf("%s", rule.MessageOrDefault("%s must be one of %v", rule.Field, rule.In))

	case "min":
		if value == "" {
			return nil
		}
		if rule.Min != nil {
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
			}
			if num < *rule.Min {
				return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at least %f", rule.Field, *rule.Min))
			}
		}

	case "max":
		if value == "" {
			return nil
		}
		if rule.Max != nil {
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("%s", rule.MessageOrDefault("%s must be a number", rule.Field))
			}
			if num > *rule.Max {
				return fmt.Errorf("%s", rule.MessageOrDefault("%s must be at most %f", rule.Field, *rule.Max))
			}
		}

	case "length":
		if value == "" {
			return nil
		}
		if rule.MinLength != nil && len(value) < *rule.MinLength {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s length must be at least %d", rule.Field, *rule.MinLength))
		}
		if rule.MaxLength != nil && len(value) > *rule.MaxLength {
			return fmt.Errorf("%s", rule.MessageOrDefault("%s length must be at most %d", rule.Field, *rule.MaxLength))
		}

	default:
		return nil
	}

	return nil
}

// MessageOrDefault 获取自定义消息或默认消息
func (r ValidationRule) MessageOrDefault(format string, args ...interface{}) string {
	if r.Message != "" {
		return r.Message
	}
	return fmt.Sprintf(format, args...)
}

// ValidateJSONBody 验证 JSON body
func ValidateJSONBody(body []byte, rules []ValidationRule) *RuleValidationResult {
	result := &RuleValidationResult{Valid: true}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, RuleValidationError{
			Field:   "body",
			Message: "Invalid JSON body",
			Type:    "json",
		})
		return result
	}

	for _, rule := range rules {
		value, exists := data[rule.Field]
		valueStr := ""
		if exists {
			// 处理不同类型的值，保留原始类型信息
			switch v := value.(type) {
			case string:
				valueStr = v
			case float64: // JSON 中的数字都是 float64
				// 对于数值验证，保留原始值
				if rule.Type == "number" || rule.Type == "min" || rule.Type == "max" {
					valueStr = strconv.FormatFloat(v, 'f', -1, 64)
				} else {
					valueStr = strconv.FormatFloat(v, 'f', -1, 64)
				}
			case bool:
				valueStr = strconv.FormatBool(v)
			case nil:
				valueStr = ""
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
		}

		// 复用验证逻辑
		v := &RequestValidator{config: ValidationConfig{Rules: []ValidationRule{rule}}}
		if err := v.validateRule(rule, valueStr); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, RuleValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Type:    rule.Type,
			})
		}
	}

	return result
}

// ValidateHeaders 快速验证请求头
func ValidateHeaders(req *http.Request, rules []ValidationRule) *RuleValidationResult {
	v, err := NewRequestValidator(ValidationConfig{
		Source:   "header",
		Rules:    rules,
		FailFast: false,
	})
	if err != nil {
		return &RuleValidationResult{
			Valid: false,
			Errors: []RuleValidationError{{
				Field:   "config",
				Message: err.Error(),
				Type:    "config",
			}},
		}
	}
	return v.Validate(req)
}

// ValidateQuery 快速验证查询参数
func ValidateQuery(req *http.Request, rules []ValidationRule) *RuleValidationResult {
	v, err := NewRequestValidator(ValidationConfig{
		Source:   "query",
		Rules:    rules,
		FailFast: false,
	})
	if err != nil {
		return &RuleValidationResult{
			Valid: false,
			Errors: []RuleValidationError{{
				Field:   "config",
				Message: err.Error(),
				Type:    "config",
			}},
		}
	}
	return v.Validate(req)
}
