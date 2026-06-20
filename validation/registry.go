package validation

import (
	"reflect"
	"sync"
)

// CustomValidator 自定义验证器接口
type CustomValidator interface {
	Validate(field reflect.Value, param string) (bool, string)
}

// ValidatorRegistry 验证器注册表，支持并发安全地注册和获取自定义验证器
type ValidatorRegistry struct {
	validators     map[string]CustomValidator
	funcValidators map[string]func(reflect.Value, string) (bool, string)
	mu             sync.RWMutex
}

// NewValidatorRegistry 创建新的验证器注册表
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators:     make(map[string]CustomValidator),
		funcValidators: make(map[string]func(reflect.Value, string) (bool, string)),
	}
}

// Register 注册结构体验证器
func (r *ValidatorRegistry) Register(name string, validator CustomValidator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators[name] = validator
}

// RegisterFunc 注册函数式验证器
func (r *ValidatorRegistry) RegisterFunc(name string, validator func(reflect.Value, string) (bool, string)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funcValidators[name] = validator
}

// Get 获取结构体验证器
func (r *ValidatorRegistry) Get(name string) (CustomValidator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	validator, ok := r.validators[name]
	return validator, ok
}

// GetFunc 获取函数式验证器
func (r *ValidatorRegistry) GetFunc(name string) (func(reflect.Value, string) (bool, string), bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	validator, ok := r.funcValidators[name]
	return validator, ok
}

// Unregister 注销验证器
func (r *ValidatorRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.validators, name)
	delete(r.funcValidators, name)
}
