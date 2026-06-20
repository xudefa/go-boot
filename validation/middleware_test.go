package validation

import (
	"errors"
	"testing"
)

// TestMiddlewareConfig 测试中间件配置
func TestMiddlewareConfig(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	if config.Validator == nil {
		t.Error("Validator 不能为 nil")
	}

	if len(config.Groups) != 1 {
		t.Errorf("预期 1 个组，得到 %d", len(config.Groups))
	}
}

// TestNewValidateMiddleware 测试创建验证中间件
func TestNewValidateMiddleware(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	middleware := NewValidateMiddleware(config)
	if middleware == nil {
		t.Fatal("NewValidateMiddleware 返回 nil")
	}
}

// TestValidateMiddlewareWithNilValidator 测试 nil 验证器
func TestValidateMiddlewareWithNilValidator(t *testing.T) {
	config := &MiddlewareConfig{
		Validator: nil,
	}

	middleware := NewValidateMiddleware(config)
	err := middleware(nil, nil, config)
	if err == nil {
		t.Error("预期因验证器为 nil 而返回错误")
	}
}

// TestValidateMiddlewareWithValidData 测试有效数据
func TestValidateMiddlewareWithValidData(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"create:required,min=2"`
	}

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张三"}
	err := middleware(nil, user, config)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestValidateMiddlewareWithInvalidData 测试无效数据
func TestValidateMiddlewareWithInvalidData(t *testing.T) {
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"create:required,min=2"`
	}

	errorCalled := false
	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
		ErrorHandler: func(c interface{}, err error) {
			errorCalled = true
		},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张"}
	err := middleware(nil, user, config)
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}

	if !errorCalled {
		t.Error("错误处理器应该被调用")
	}
}

// TestValidateMiddlewareWithStandardValidator 测试标准验证器
func TestValidateMiddlewareWithStandardValidator(t *testing.T) {
	validator := NewTagValidator()

	type User struct {
		Name string `validate:"required,min=2"`
	}

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张三"}
	err := middleware(nil, user, config)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Name = "张"
	err = middleware(nil, user, config)
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}
}

// TestDefaultErrorHandler 测试默认错误处理器
func TestDefaultErrorHandler(t *testing.T) {
	err := errors.New("test error")
	DefaultErrorHandler(nil, err)
}

// TestShouldSkipPath 测试跳过路径
func TestShouldSkipPath(t *testing.T) {
	result := shouldSkipPath(nil, []string{"/skip"})
	if result {
		t.Error("shouldSkipPath 应该返回 false（未实现）")
	}
}
