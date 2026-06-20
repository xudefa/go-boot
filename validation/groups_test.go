package validation

import (
	"testing"
)

// TestNewGroupedValidator 测试创建分组验证器
func TestNewGroupedValidator(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	if validator == nil {
		t.Fatal("NewGroupedValidator 返回 nil")
	}
}

// TestSetDefaultGroups 测试设置默认组
func TestSetDefaultGroups(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	validator.SetDefaultGroups("base", "create")

	if len(validator.defaultGroups) != 2 {
		t.Errorf("预期 2 个默认组，得到 %d", len(validator.defaultGroups))
	}
}

// TestParseGroupRules 测试解析验证组标签
func TestParseGroupRules(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	tag := "base:required;create:min=2,max=50;update:min=1"
	rules := validator.parseGroupRules(tag)

	if len(rules) != 3 {
		t.Errorf("预期 3 个组规则，得到 %d", len(rules))
	}

	if rules[0].GroupName != "base" {
		t.Errorf("第一个组名应为 'base'，得到 '%s'", rules[0].GroupName)
	}

	if len(rules[0].Rules) != 1 || rules[0].Rules[0] != "required" {
		t.Errorf("base 组规则不正确: %v", rules[0].Rules)
	}
}

// TestValidateWithGroups 测试使用指定组验证
func TestValidateWithGroups(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"base:required;create:min=2,max=50;update:min=1"`
	}

	user := User{Name: "张三"}

	err := validator.ValidateWithGroups(user, "base")
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Name = "张"
	err = validator.ValidateWithGroups(user, "create")
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}
}

// TestValidateWithDefaultGroups 测试使用默认组验证
func TestValidateWithDefaultGroups(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)
	validator.SetDefaultGroups("base")

	type User struct {
		Name string `validate:"base:required;create:min=2"`
	}

	user := User{Name: "张三"}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Name = ""
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因名字为空而验证失败")
	}
}

// TestGroupInheritance 测试组继承
func TestGroupInheritance(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"base:required;create:base,min=2"`
	}

	user := User{Name: "张"}
	err := validator.ValidateWithGroups(user, "create")
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}
}

// TestMultipleGroups 测试多组验证
func TestMultipleGroups(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"base:required;create:min=2;update:min=1"`
	}

	user := User{Name: "张"}
	err := validator.ValidateWithGroups(user, "create", "update")
	if err == nil {
		t.Error("预期因名字太短而验证失败（create 组）")
	}

	user.Name = "张三"
	err = validator.ValidateWithGroups(user, "create", "update")
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}
