package condition

import "strings"

// CompositeOption 复合条件选项函数
type CompositeOption func(*compositeConfig)

// compositeConfig 复合条件配置
type compositeConfig struct {
	description string
	lazy        bool
}

// WithDescription 设置条件描述
func WithDescription(desc string) CompositeOption {
	return func(c *compositeConfig) {
		c.description = desc
	}
}

// WithLazyEvaluation 启用惰性求值（短路优化）
// All 和 Any 默认已经支持短路优化，此选项用于显式声明
func WithLazyEvaluation() CompositeOption {
	return func(c *compositeConfig) {
		c.lazy = true
	}
}

// All 返回一个复合条件，当所有子条件都匹配时返回 true（逻辑与运算）。
// 执行逻辑：遍历所有子条件，依次调用 Matches 方法：
//  1. 只要有一个子条件不匹配，立即短路返回 false（不再继续判断后续条件）
//  2. 所有子条件都匹配时返回 true
//
// 该行为与编程语言中的 && 运算符一致，支持短路优化。
func All(conditions ...Condition) Condition {
	return &allCondition{conditions: conditions}
}

// AllWithOptions 返回一个带选项的复合条件
func AllWithOptions(opts ...CompositeOption) func(conditions ...Condition) Condition {
	config := &compositeConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return func(conditions ...Condition) Condition {
		return &allCondition{
			conditions:  conditions,
			description: config.description,
		}
	}
}

// allCondition 逻辑与复合条件
type allCondition struct {
	conditions  []Condition // 子条件列表
	description string      // 条件描述
}

func (a *allCondition) Matches(ctx ConditionContext) bool {
	for _, c := range a.conditions {
		if !c.Matches(ctx) {
			return false
		}
	}
	return true
}

func (a *allCondition) String() string {
	if a.description != "" {
		return "All(" + a.description + ")"
	}
	return "All(" + joinConditions(a.conditions, ", ") + ")"
}

// Any 返回一个复合条件，当任一子条件匹配时返回 true（逻辑或运算）。
// 执行逻辑：遍历所有子条件，依次调用 Matches 方法：
//  1. 只要有一个子条件匹配，立即短路返回 true（不再继续判断后续条件）
//  2. 所有子条件都不匹配时返回 false
//
// 该行为与编程语言中的 || 运算符一致，支持短路优化。
func Any(conditions ...Condition) Condition {
	return &anyCondition{conditions: conditions}
}

// anyCondition 逻辑或复合条件
type anyCondition struct {
	conditions  []Condition // 子条件列表
	description string      // 条件描述
}

func (a *anyCondition) Matches(ctx ConditionContext) bool {
	for _, c := range a.conditions {
		if c.Matches(ctx) {
			return true
		}
	}
	return false
}

func (a *anyCondition) String() string {
	if a.description != "" {
		return "Any(" + a.description + ")"
	}
	return "Any(" + joinConditions(a.conditions, ", ") + ")"
}

// Not 返回一个复合条件，对子条件的匹配结果取反（逻辑非运算）。
// 执行逻辑：调用子条件的 Matches 方法，然后对其结果取反：
//  1. 子条件匹配时返回 false
//  2. 子条件不匹配时返回 true
//
// 该行为与编程语言中的 ! 运算符一致。
func Not(condition Condition) Condition {
	return &notCondition{condition: condition}
}

// notCondition 逻辑非复合条件
type notCondition struct {
	condition Condition // 被取反的子条件
}

func (n *notCondition) Matches(ctx ConditionContext) bool {
	return !n.condition.Matches(ctx)
}

func (n *notCondition) String() string {
	return "Not(" + n.condition.String() + ")"
}

// joinConditions 将条件列表格式化为可读字符串，用指定的分隔符连接每个条件的 String() 输出。
func joinConditions(conditions []Condition, sep string) string {
	var result strings.Builder
	for i, c := range conditions {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(c.String())
	}
	return result.String()
}
