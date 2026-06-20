package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// RuleBuilder 验证规则构建器
type RuleBuilder struct {
	rules    []string
	messages map[string]string
}

// NewRuleBuilder 创建规则构建器
func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{
		rules:    make([]string, 0),
		messages: make(map[string]string),
	}
}

// Required 添加必填验证
func (b *RuleBuilder) Required() *RuleBuilder {
	b.rules = append(b.rules, "required")
	return b
}

// Min 添加最小值/长度验证
func (b *RuleBuilder) Min(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("min=%d", n))
	return b
}

// Max 添加最大值/长度验证
func (b *RuleBuilder) Max(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("max=%d", n))
	return b
}

// Len 添加固定长度验证
func (b *RuleBuilder) Len(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("len=%d", n))
	return b
}

// Email 添加邮箱格式验证
func (b *RuleBuilder) Email() *RuleBuilder {
	b.rules = append(b.rules, "email")
	return b
}

// URL 添加URL格式验证
func (b *RuleBuilder) URL() *RuleBuilder {
	b.rules = append(b.rules, "url")
	return b
}

// IP 添加IP地址验证
func (b *RuleBuilder) IP() *RuleBuilder {
	b.rules = append(b.rules, "ip")
	return b
}

// Gt 添加大于验证
func (b *RuleBuilder) Gt(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("gt=%d", n))
	return b
}

// Gte 添加大于等于验证
func (b *RuleBuilder) Gte(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("gte=%d", n))
	return b
}

// Lt 添加小于验证
func (b *RuleBuilder) Lt(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("lt=%d", n))
	return b
}

// Lte 添加小于等于验证
func (b *RuleBuilder) Lte(n int) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("lte=%d", n))
	return b
}

// OneOf 添加枚举值验证
func (b *RuleBuilder) OneOf(options ...string) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("oneof=%s", strings.Join(options, " ")))
	return b
}

// Regexp 添加正则表达式验证
func (b *RuleBuilder) Regexp(pattern string) *RuleBuilder {
	b.rules = append(b.rules, fmt.Sprintf("regexp=%s", pattern))
	return b
}

// CustomMessage 设置自定义错误消息
func (b *RuleBuilder) CustomMessage(message string) *RuleBuilder {
	for _, rule := range b.rules {
		b.messages[rule] = message
	}
	return b
}

// Build 构建验证规则字符串
func (b *RuleBuilder) Build() string {
	return strings.Join(b.rules, ",")
}

// BuildWithMessages 构建验证规则并返回消息映射
func (b *RuleBuilder) BuildWithMessages() (string, map[string]string) {
	return b.Build(), b.messages
}

// ValidatorChain 验证器链，支持多个对象连续验证
type ValidatorChain struct {
	validators       []Validator
	stopOnFirstError bool
}

// NewValidatorChain 创建验证器链
func NewValidatorChain() *ValidatorChain {
	return &ValidatorChain{
		validators: make([]Validator, 0),
	}
}

// StopOnFirstError 设置遇到第一个错误时停止
func (c *ValidatorChain) StopOnFirstError() *ValidatorChain {
	c.stopOnFirstError = true
	return c
}

// Add 添加验证器
func (c *ValidatorChain) Add(v Validator) *ValidatorChain {
	c.validators = append(c.validators, v)
	return c
}

// AddStruct 添加结构体验证器
func (c *ValidatorChain) AddStruct(obj interface{}) *ValidatorChain {
	c.validators = append(c.validators, &structValidator{obj: obj})
	return c
}

// AddValue 添加值验证器
func (c *ValidatorChain) AddValue(value interface{}, rules string) *ValidatorChain {
	c.validators = append(c.validators, &valueValidator{value: value, rules: rules})
	return c
}

// Validate 执行验证链
func (c *ValidatorChain) Validate() error {
	var allErrors ValidationErrors

	for _, v := range c.validators {
		err := v.Validate(nil)
		if err != nil {
			if validationErrs, ok := err.(ValidationErrors); ok {
				allErrors = append(allErrors, validationErrs...)
				if c.stopOnFirstError {
					return allErrors
				}
			} else {
				allErrors = append(allErrors, ValidationError{
					Field:   "unknown",
					Message: err.Error(),
				})
				if c.stopOnFirstError {
					return allErrors
				}
			}
		}
	}

	if len(allErrors) > 0 {
		return allErrors
	}
	return nil
}

// structValidator 结构体验证器适配器
type structValidator struct {
	obj interface{}
}

func (v *structValidator) Validate(obj interface{}) error {
	return ValidateStruct(v.obj)
}

// valueValidator 值验证器适配器
type valueValidator struct {
	value interface{}
	rules string
}

func (v *valueValidator) Validate(obj interface{}) error {
	return Validate(v.value, v.rules)
}

// RegexCache 正则表达式缓存
type RegexCache struct {
	cache map[string]*regexp.Regexp
}

// NewRegexCache 创建正则表达式缓存
func NewRegexCache() *RegexCache {
	return &RegexCache{
		cache: make(map[string]*regexp.Regexp),
	}
}

// Get 获取或编译正则表达式
func (c *RegexCache) Get(pattern string) (*regexp.Regexp, error) {
	if re, exists := c.cache[pattern]; exists {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	c.cache[pattern] = re
	return re, nil
}

// MustGet 获取或编译正则表达式，失败则panic
func (c *RegexCache) MustGet(pattern string) *regexp.Regexp {
	re, err := c.Get(pattern)
	if err != nil {
		panic(fmt.Sprintf("invalid regex pattern: %s, error: %v", pattern, err))
	}
	return re
}

// Clear 清空缓存
func (c *RegexCache) Clear() {
	c.cache = make(map[string]*regexp.Regexp)
}

// Size 获取缓存大小
func (c *RegexCache) Size() int {
	return len(c.cache)
}

// 全局正则表达式缓存
var defaultRegexCache = NewRegexCache()

// GetRegexCache 获取全局正则表达式缓存
func GetRegexCache() *RegexCache {
	return defaultRegexCache
}
