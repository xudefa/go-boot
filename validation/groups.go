package validation

import (
	"errors"
	"reflect"
	"strings"
)

// GroupRule 验证组规则
type GroupRule struct {
	GroupName string
	Rules     []string
	Inherited bool
}

// GroupedTagValidator 支持验证组的标签验证器
type GroupedTagValidator struct {
	registry      *ValidatorRegistry
	defaultGroups []string
}

// NewGroupedValidator 创建新的分组验证器
func NewGroupedValidator(registry *ValidatorRegistry) *GroupedTagValidator {
	return &GroupedTagValidator{
		registry: registry,
	}
}

// SetDefaultGroups 设置默认验证组
func (v *GroupedTagValidator) SetDefaultGroups(groups ...string) {
	v.defaultGroups = groups
}

// ValidateWithGroups 使用指定组验证对象
func (v *GroupedTagValidator) ValidateWithGroups(obj interface{}, groups ...string) error {
	if obj == nil {
		return nil
	}

	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("validation: only structs are supported")
	}

	var errs ValidationErrors

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		groupRules := v.parseGroupRules(tag)
		resolvedRules := v.resolveInheritedRules(groupRules, groups)

		if len(resolvedRules) == 0 && len(v.defaultGroups) > 0 {
			resolvedRules = v.resolveInheritedRules(groupRules, v.defaultGroups)
		}

		if len(resolvedRules) > 0 {
			fieldErrors := v.validateFieldWithRules(field, resolvedRules, fieldName, obj)
			errs = append(errs, fieldErrors...)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Validate 使用默认组验证对象
func (v *GroupedTagValidator) Validate(obj interface{}) error {
	if len(v.defaultGroups) > 0 {
		return v.ValidateWithGroups(obj, v.defaultGroups...)
	}
	return v.ValidateWithGroups(obj, "default")
}

// parseGroupRules 解析验证组标签
func (v *GroupedTagValidator) parseGroupRules(tag string) []GroupRule {
	var rules []GroupRule
	groups := strings.Split(tag, ";")

	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		parts := strings.SplitN(group, ":", 2)
		if len(parts) != 2 {
			continue
		}

		groupName := strings.TrimSpace(parts[0])
		ruleStr := strings.TrimSpace(parts[1])

		ruleList := strings.Split(ruleStr, ",")
		var cleanedRules []string
		for _, r := range ruleList {
			r = strings.TrimSpace(r)
			if r != "" {
				cleanedRules = append(cleanedRules, r)
			}
		}

		inherited := false
		for _, r := range cleanedRules {
			if _, ok := v.GetGroup(r); ok {
				inherited = true
				break
			}
		}

		rules = append(rules, GroupRule{
			GroupName: groupName,
			Rules:     cleanedRules,
			Inherited: inherited,
		})
	}

	return rules
}

// resolveInheritedRules 解析继承的规则
func (v *GroupedTagValidator) resolveInheritedRules(rules []GroupRule, groups []string) []string {
	var resolvedRules []string

	for _, group := range groups {
		for _, rule := range rules {
			if rule.GroupName == group {
				if rule.Inherited {
					for _, r := range rule.Rules {
						if _, ok := v.GetGroup(r); ok {
							inheritedRules := v.resolveInheritedRules(rules, []string{r})
							resolvedRules = append(resolvedRules, inheritedRules...)
						} else {
							resolvedRules = append(resolvedRules, r)
						}
					}
				} else {
					resolvedRules = append(resolvedRules, rule.Rules...)
				}
			}
		}
	}

	return resolvedRules
}

// GetGroup 获取组是否存在
func (v *GroupedTagValidator) GetGroup(name string) (bool, bool) {
	return false, false
}

// validateFieldWithRules 使用指定规则验证字段
func (v *GroupedTagValidator) validateFieldWithRules(field reflect.Value, rules []string, fieldName string, obj interface{}) []ValidationError {
	baseValidator := NewTagValidatorWithRegistry(v.registry)
	tag := strings.Join(rules, ",")
	return baseValidator.validateField(field, tag, fieldName, obj)
}
