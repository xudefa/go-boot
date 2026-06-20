package refresh

import (
	"log/slog"
	"time"
)

// RefreshConfig 刷新配置
//
// 控制刷新功能的行为，包括是否启用、刷新延迟和重试策略。
type RefreshConfig struct {
	Enabled            bool          // 是否启用刷新功能
	RefreshDelay       time.Duration // 刷新延迟，避免频繁刷新
	MaxRefreshAttempts int           // 最大刷新尝试次数
	Logger             *slog.Logger  // 日志记录器
}

// RefreshOption 刷新配置选项函数
type RefreshOption func(*RefreshConfig)

// WithRefreshEnabled 设置是否启用刷新功能
func WithRefreshEnabled(enabled bool) RefreshOption {
	return func(c *RefreshConfig) {
		c.Enabled = enabled
	}
}

// WithRefreshDelay 设置刷新延迟时间
func WithRefreshDelay(delay time.Duration) RefreshOption {
	return func(c *RefreshConfig) {
		c.RefreshDelay = delay
	}
}

// WithMaxRefreshAttempts 设置最大刷新尝试次数
func WithMaxRefreshAttempts(attempts int) RefreshOption {
	return func(c *RefreshConfig) {
		c.MaxRefreshAttempts = attempts
	}
}

// WithRefreshLogger 设置日志记录器
func WithRefreshLogger(logger *slog.Logger) RefreshOption {
	return func(c *RefreshConfig) {
		c.Logger = logger
	}
}

// DefaultRefreshConfig 返回默认刷新配置
//
// 默认值：
//   - Enabled: true
//   - RefreshDelay: 100ms
//   - MaxRefreshAttempts: 3
//   - Logger: slog.Default()
func DefaultRefreshConfig() *RefreshConfig {
	return &RefreshConfig{
		Enabled:            true,
		RefreshDelay:       100 * time.Millisecond,
		MaxRefreshAttempts: 3,
		Logger:             slog.Default(),
	}
}

// ApplyOptions 应用配置选项列表
func (c *RefreshConfig) ApplyOptions(opts []RefreshOption) {
	for _, opt := range opts {
		opt(c)
	}
}
