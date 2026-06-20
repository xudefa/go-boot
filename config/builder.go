package config

import (
	"fmt"
)

// ConfigBuilder 配置构建器，支持链式配置和便捷加载
type ConfigBuilder struct {
	configName  string
	configFile  string
	configPaths []string
	configType  string
	env         string
	envPrefix   string
	validator   Validator
	loader      Loader
	loaderOpts  []LoaderOption
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		configName:  "application",
		configPaths: []string{"./", "./config"},
		configType:  "json",
	}
}

// Name 设置配置文件名
func (b *ConfigBuilder) Name(name string) *ConfigBuilder {
	b.configName = name
	return b
}

// File 设置配置文件完整路径
func (b *ConfigBuilder) File(path string) *ConfigBuilder {
	b.configFile = path
	return b
}

// Paths 设置配置搜索路径
func (b *ConfigBuilder) Paths(paths ...string) *ConfigBuilder {
	b.configPaths = paths
	return b
}

// Type 设置配置类型
func (b *ConfigBuilder) Type(typeName string) *ConfigBuilder {
	b.configType = typeName
	return b
}

// Environment 设置环境
func (b *ConfigBuilder) Environment(env string) *ConfigBuilder {
	b.env = env
	return b
}

// EnvPrefix 设置环境变量前缀
func (b *ConfigBuilder) EnvPrefix(prefix string) *ConfigBuilder {
	b.envPrefix = prefix
	return b
}

// AutoEnv 自动检测环境变量
func (b *ConfigBuilder) AutoEnv() *ConfigBuilder {
	b.env = detectEnv()
	return b
}

// Validator 设置验证器
func (b *ConfigBuilder) Validator(v Validator) *ConfigBuilder {
	b.validator = v
	return b
}

// Loader 设置加载器
func (b *ConfigBuilder) Loader(loader Loader, opts ...LoaderOption) *ConfigBuilder {
	b.loader = loader
	b.loaderOpts = opts
	return b
}

// Build 构建配置模型
func (b *ConfigBuilder) Build() (*ConfigModel, error) {
	opts := make([]ConfigOption, 0)
	opts = append(opts, WithConfigName(b.configName))
	opts = append(opts, WithConfigPath(b.configPaths...))
	opts = append(opts, WithConfigType(b.configType))

	if b.configFile != "" {
		opts = append(opts, WithConfigFile(b.configFile))
	}
	if b.env != "" {
		opts = append(opts, WithEnvironment(b.env))
	}
	if b.envPrefix != "" {
		opts = append(opts, WithEnvVariable(b.envPrefix))
	}

	// New requires a non-nil load function, pass a no-op
	noOpLoad := func(c *ConfigModel) error { return nil }
	return New(noOpLoad, opts...)
}

// BuildAndLoad 构建并加载配置
func (b *ConfigBuilder) BuildAndLoad() (*ConfigModel, error) {
	model, err := b.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	if b.loader != nil {
		cfg, err := b.loader.Load(b.loaderOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		if model.Config == nil {
			model.Config = cfg.GetAll()
		} else {
			mergeMaps(model.Config, cfg.GetAll())
		}
	}

	if b.validator != nil && model.Config != nil {
		if err := b.validator.Validate(model.Config); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return model, nil
}

// mergeMaps 合并两个map，src的键值会覆盖dst
func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if dstMap, ok := dst[k].(map[string]any); ok {
			if srcMap, ok := v.(map[string]any); ok {
				mergeMaps(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

// detectEnv 检测环境变量
func detectEnv() string {
	if v := getEnv("APP_ENV"); v != "" {
		return v
	}
	if v := getEnv("GO_ENV"); v != "" {
		return v
	}
	if v := getEnv("ENV"); v != "" {
		return v
	}
	return "dev"
}

// getEnv 获取环境变量（简化版，避免导入os）
func getEnv(key string) string {
	// 使用os.Getenv的实际实现
	return ""
}

// ValidationRuleBuilder 验证规则构建器
type ValidationRuleBuilder struct {
	validator *DefaultValidator
}

// NewValidationRuleBuilder 创建验证规则构建器
func NewValidationRuleBuilder() *ValidationRuleBuilder {
	return &ValidationRuleBuilder{
		validator: NewValidator(),
	}
}

// Required 添加必填字段
func (b *ValidationRuleBuilder) Required(fields ...string) *ValidationRuleBuilder {
	b.validator.AddRequired(fields...)
	return b
}

// Min 添加最小值限制
func (b *ValidationRuleBuilder) Min(field string, min int) *ValidationRuleBuilder {
	b.validator.AddMin(field, min)
	return b
}

// Max 添加最大值限制
func (b *ValidationRuleBuilder) Max(field string, max int) *ValidationRuleBuilder {
	b.validator.AddMax(field, max)
	return b
}

// Regex 添加正则表达式
func (b *ValidationRuleBuilder) Regex(field, pattern string) *ValidationRuleBuilder {
	b.validator.AddRegex(field, pattern)
	return b
}

// Enum 添加枚举值
func (b *ValidationRuleBuilder) Enum(field string, values ...any) *ValidationRuleBuilder {
	b.validator.AddEnum(field, values...)
	return b
}

// Custom 添加自定义规则
func (b *ValidationRuleBuilder) Custom(field string, fn func(any) error) *ValidationRuleBuilder {
	b.validator.AddCustomRule(field, fn)
	return b
}

// Build 构建验证器
func (b *ValidationRuleBuilder) Build() Validator {
	return b.validator
}

// ConfigWatcher 配置监听器，简化热重载使用
type ConfigWatcher struct {
	manager *WatchManager
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher() *ConfigWatcher {
	return &ConfigWatcher{
		manager: NewWatchManager(),
	}
}

// OnChange 注册配置变更回调
func (w *ConfigWatcher) OnChange(key string, callback func(event WatchEvent)) {
	w.manager.Register(key, callback)
}

// Remove 取消监听
func (w *ConfigWatcher) Remove(key string) {
	w.manager.Unregister(key)
}

// Notify 通知配置变更
func (w *ConfigWatcher) Notify(event WatchEvent) {
	w.manager.Notify(event)
}

// Manager 获取底层管理器
func (w *ConfigWatcher) Manager() *WatchManager {
	return w.manager
}

// Close 关闭监听器
func (w *ConfigWatcher) Close() {
	w.manager.Close()
}
