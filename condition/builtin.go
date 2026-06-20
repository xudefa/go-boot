package condition

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xudefa/go-boot/environment"
)

// OnProperty 创建基于配置属性的条件
//
// 如果只传 key，当该属性存在且不为空时匹配。
// 如果传 key 和 value，当该属性等于 value 时匹配。
func OnProperty(key string, expectedValue ...string) Condition {
	return &propertyCondition{
		key:           key,
		expectedValue: expectedValue,
	}
}

// propertyCondition 基于配置属性的条件实现
type propertyCondition struct {
	key           string   // 配置键名
	expectedValue []string // 期望值（为空时仅检查键是否存在）
}

func (p *propertyCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	val, ok := env.GetProperty(p.key)
	if !ok {
		return false
	}
	if len(p.expectedValue) == 0 {
		s, ok := val.(string)
		return ok && s != ""
	}
	return valAsString(val) == p.expectedValue[0]
}

func (p *propertyCondition) String() string {
	if len(p.expectedValue) > 0 {
		return fmt.Sprintf("OnProperty(%s=%s)", p.key, p.expectedValue[0])
	}
	return fmt.Sprintf("OnProperty(%s)", p.key)
}

// OnMissingProperty 创建基于属性不存在的条件
//
// 当指定的配置键不存在时匹配。
func OnMissingProperty(key string) Condition {
	return &missingPropertyCondition{key: key}
}

// missingPropertyCondition 基于属性不存在的条件实现
type missingPropertyCondition struct {
	key string // 配置键名
}

func (m *missingPropertyCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	_, ok := env.GetProperty(m.key)
	return !ok
}

func (m *missingPropertyCondition) String() string {
	return fmt.Sprintf("OnMissingProperty(%s)", m.key)
}

// OnBean 创建基于 Bean 存在的条件
//
// 当容器中存在指定 ID 的 Bean 时匹配。
func OnBean(beanID string) Condition {
	return &beanCondition{beanID: beanID, missing: false}
}

// OnMissingBean 创建基于 Bean 不存在的条件
//
// 当容器中不存在指定 ID 的 Bean 时匹配。
func OnMissingBean(beanID string) Condition {
	return &beanCondition{beanID: beanID, missing: true}
}

// beanCondition 基于 Bean 存在性的条件实现
type beanCondition struct {
	beanID  string // Bean ID
	missing bool   // true 表示检查不存在
}

func (b *beanCondition) Matches(ctx ConditionContext) bool {
	container := ctx.Container()
	has := container.Has(b.beanID)
	if b.missing {
		return !has
	}
	return has
}

func (b *beanCondition) String() string {
	if b.missing {
		return fmt.Sprintf("OnMissingBean(%s)", b.beanID)
	}
	return fmt.Sprintf("OnBean(%s)", b.beanID)
}

// OnProfile 创建基于 Profile 的条件
//
// 当指定 Profile 被激活时匹配。支持否定前缀 "!"，如 "!dev" 表示非 dev 环境时匹配。
func OnProfile(profile string) Condition {
	return &profileCondition{profile: profile}
}

// profileCondition 基于 Profile 的条件实现
type profileCondition struct {
	profile string // Profile 名称，支持 "!" 否定前缀
}

func (p *profileCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	type profileAcceptor interface {
		AcceptsProfile(profile string) bool
	}
	if pa, ok := env.(profileAcceptor); ok {
		return pa.AcceptsProfile(p.profile)
	}
	negate := strings.HasPrefix(p.profile, "!")
	return negate
}

func (p *profileCondition) String() string {
	return fmt.Sprintf("OnProfile(%s)", p.profile)
}

// OnClass 创建基于类路径的条件
//
// 当指定类在编译依赖中存在时匹配。
func OnClass(className string) Condition {
	return &classCondition{className: className, missing: false}
}

// OnMissingClass 创建基于类路径不存在的条件
//
// 当指定类在编译依赖中不存在时匹配。
func OnMissingClass(className string) Condition {
	return &classCondition{className: className, missing: true}
}

// classCondition 基于类路径的条件实现
type classCondition struct {
	className string // 类名
	missing   bool   // true 表示检查不存在
}

func (c *classCondition) Matches(ctx ConditionContext) bool {
	cl := ctx.ClassLoader()
	has := cl.HasClass(c.className)
	if c.missing {
		return !has
	}
	return has
}

func (c *classCondition) String() string {
	if c.missing {
		return fmt.Sprintf("OnMissingClass(%s)", c.className)
	}
	return fmt.Sprintf("OnClass(%s)", c.className)
}

// OnPropertyPrefix 创建基于配置前缀存在的条件
//
// 当存在以指定前缀开头的配置键时匹配。
func OnPropertyPrefix(prefix string) Condition {
	return &propertyPrefixCondition{prefix: prefix}
}

// propertyPrefixCondition 基于配置前缀的条件实现
type propertyPrefixCondition struct {
	prefix string // 配置键前缀
}

func (p *propertyPrefixCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	type propertySourceLister interface {
		GetPropertySources() []environment.PropertySource
	}
	if lister, ok := env.(propertySourceLister); ok {
		for _, src := range lister.GetPropertySources() {
			for _, key := range listKeys(src) {
				if strings.HasPrefix(key, p.prefix) {
					return true
				}
			}
		}
	}
	return false
}

func (p *propertyPrefixCondition) String() string {
	return fmt.Sprintf("OnPropertyPrefix(%s)", p.prefix)
}

// Custom 创建自定义条件
//
// 通过传入的评估函数决定是否匹配，name 用于日志和调试输出。
func Custom(name string, evaluator func(ctx ConditionContext) bool) Condition {
	return &customCondition{name: name, evaluator: evaluator}
}

// customCondition 自定义条件实现
type customCondition struct {
	name      string                          // 条件名称
	evaluator func(ctx ConditionContext) bool // 评估函数
}

func (c *customCondition) Matches(ctx ConditionContext) bool {
	return c.evaluator(ctx)
}

func (c *customCondition) String() string {
	return fmt.Sprintf("Custom(%s)", c.name)
}

// valAsString 将任意值转换为字符串
//
// 支持 string、bool、int、float64 类型，其他类型返回空字符串。
func valAsString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// keyLister 可列举键的接口，用于 OnPropertyPrefix 条件
type keyLister interface {
	Keys() []string
}

// listKeys 尝试列举 PropertySource 的键
func listKeys(src any) []string {
	if kl, ok := src.(keyLister); ok {
		return kl.Keys()
	}
	return nil
}
