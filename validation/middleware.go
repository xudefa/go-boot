package validation

import (
	"errors"
)

// MiddlewareValidator 中间件验证器接口
type MiddlewareValidator interface {
	ValidateRequest(c interface{}, obj interface{}) error
	HandleValidationError(c interface{}, err error)
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	Validator    Validator
	Groups       []string
	ErrorHandler func(c interface{}, err error)
	SkipPaths    []string
}

// DefaultErrorHandler 默认错误处理器
func DefaultErrorHandler(c interface{}, err error) {
}

// ValidateMiddleware 通用验证中间件函数类型
type ValidateMiddleware func(c interface{}, obj interface{}, config *MiddlewareConfig) error

// NewValidateMiddleware 创建新的验证中间件
func NewValidateMiddleware(config *MiddlewareConfig) ValidateMiddleware {
	return func(c interface{}, obj interface{}, cfg *MiddlewareConfig) error {
		if cfg.Validator == nil {
			return errors.New("validator is required")
		}

		if shouldSkipPath(c, cfg.SkipPaths) {
			return nil
		}

		var err error
		if groupedValidator, ok := cfg.Validator.(*GroupedTagValidator); ok && len(cfg.Groups) > 0 {
			err = groupedValidator.ValidateWithGroups(obj, cfg.Groups...)
		} else {
			err = cfg.Validator.Validate(obj)
		}

		if err != nil {
			if cfg.ErrorHandler != nil {
				cfg.ErrorHandler(c, err)
			}
			return err
		}

		return nil
	}
}

// shouldSkipPath 检查是否应该跳过路径
func shouldSkipPath(c interface{}, skipPaths []string) bool {
	return false
}
