package validation

import (
	"testing"
)

// TestFieldMatch 测试字段匹配验证
func TestFieldMatch(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		Password        string `validate:"required,min=6"`
		ConfirmPassword string `validate:"fieldmatch=Password"`
	}

	user := User{Password: "password123", ConfirmPassword: "password123"}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.ConfirmPassword = "password456"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因密码不匹配而验证失败")
	}
}

// TestFieldNE 测试字段不等于验证
func TestFieldNE(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		OldPassword string `validate:"required"`
		NewPassword string `validate:"required,fieldne=OldPassword"`
	}

	user := User{OldPassword: "oldpass", NewPassword: "newpass"}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.NewPassword = "oldpass"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因新密码不能与旧密码相同而验证失败")
	}
}

// TestFieldGT 测试字段大于验证
func TestFieldGT(t *testing.T) {
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldgt=Min"`
	}

	r := Range{Min: 10, Max: 20}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 10
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须大于 Min 而验证失败")
	}
}

// TestFieldLT 测试字段小于验证
func TestFieldLT(t *testing.T) {
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldlt=Min"`
	}

	r := Range{Min: 20, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 20
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须小于 Min 而验证失败")
	}
}

// TestFieldGTE 测试字段大于等于验证
func TestFieldGTE(t *testing.T) {
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldgte=Min"`
	}

	r := Range{Min: 10, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 5
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须大于等于 Min 而验证失败")
	}
}

// TestFieldLTE 测试字段小于等于验证
func TestFieldLTE(t *testing.T) {
	validator := NewTagValidator()

	type Range struct {
		Min int `validate:"required"`
		Max int `validate:"required,fieldlte=Min"`
	}

	r := Range{Min: 10, Max: 10}
	err := validator.Validate(r)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	r.Max = 15
	err = validator.Validate(r)
	if err == nil {
		t.Error("预期因 Max 必须小于等于 Min 而验证失败")
	}
}

// TestWhenCondition 测试条件依赖验证
func TestWhenCondition(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		Type        string `validate:"required,oneof=personal business"`
		CompanyName string `validate:"when=Type==business:required;min=2"`
	}

	user := User{Type: "personal", CompanyName: ""}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Type = "business"
	user.CompanyName = ""
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因 Type 为 business 时 CompanyName 必填而验证失败")
	}

	user.CompanyName = "ABC Company"
	err = validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestWhenConditionWithAge 测试年龄条件验证
func TestWhenConditionWithAge(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		Age        int    `validate:"required,min=1"`
		ParentName string `validate:"when=Age<18:required"`
	}

	user := User{Age: 25, ParentName: ""}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Age = 15
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因 Age 小于 18 时 ParentName 必填而验证失败")
	}

	user.ParentName = "张三"
	err = validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestCombinedCrossField 测试组合跨字段验证
func TestCombinedCrossField(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		Password        string `validate:"required,min=6"`
		ConfirmPassword string `validate:"required,fieldmatch=Password"`
		Age             int    `validate:"required,min=1"`
		ParentName      string `validate:"when=Age<18:required"`
	}

	user := User{
		Password:        "password123",
		ConfirmPassword: "password123",
		Age:             25,
		ParentName:      "",
	}
	err := validator.Validate(user)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.ConfirmPassword = "password456"
	err = validator.Validate(user)
	if err == nil {
		t.Error("预期因密码不匹配而验证失败")
	}
}
