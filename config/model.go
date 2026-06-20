package config

import "time"

// ConfigModel 配置模型
//
// 存储配置构建的所有参数和最终配置数据。
type ConfigModel struct {
	ConfigName  string
	ConfigFile  string
	ConfigPaths []string
	ConfigType  string
	Env         string
	OptionName  string
	Config      map[string]any
}

// ConfigOption 配置选项函数类型
type ConfigOption func(*ConfigModel)

// WithConfigName 设置配置名称
func WithConfigName(name string) ConfigOption {
	return func(m *ConfigModel) {
		m.ConfigName = name
	}
}

// WithConfigFile 设置配置文件路径
func WithConfigFile(path string) ConfigOption {
	return func(m *ConfigModel) {
		m.ConfigFile = path
	}
}

// WithConfigPath 设置配置搜索路径
func WithConfigPath(paths ...string) ConfigOption {
	return func(m *ConfigModel) {
		m.ConfigPaths = paths
	}
}

// WithConfigType 设置配置类型
func WithConfigType(typeName string) ConfigOption {
	return func(m *ConfigModel) {
		m.ConfigType = typeName
	}
}

// WithEnvironment 设置环境
func WithEnvironment(env string) ConfigOption {
	return func(m *ConfigModel) {
		m.Env = env
	}
}

// WithEnvVariable 设置环境变量前缀
func WithEnvVariable(prefix string) ConfigOption {
	return func(m *ConfigModel) {
		m.OptionName = prefix
	}
}

// New 创建配置模型
func New(load func(*ConfigModel) error, opts ...ConfigOption) (*ConfigModel, error) {
	model := &ConfigModel{
		Config: make(map[string]any),
	}

	for _, opt := range opts {
		opt(model)
	}

	if load != nil {
		if err := load(model); err != nil {
			return nil, err
		}
	}

	return model, nil
}

// WatchEvent 配置变更事件
type WatchEvent struct {
	Type      string    // modify, delete, create
	Key       string    // 配置键
	Value     any       // 配置值
	Timestamp time.Time // 事件时间
	Source    string    // 事件来源
}

// NewWatchEvent 创建配置变更事件
func NewWatchEvent(eventType, key string, value any, source string) WatchEvent {
	return WatchEvent{
		Type:      eventType,
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Source:    source,
	}
}
