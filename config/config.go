// Package config 提供配置管理功能
//
// 基于 Viper 库的配置文件加载和读取支持.
// 支持多种配置格式(YAML, JSON, TOML, INI, HCL, ENV, Properties),
// 环境变量覆盖, 以及分层配置.
//
// 功能特点:
//
//   - 支持从文件、环境变量、命令行读取配置
//   - 自动根据环境加载不同配置文件(如 config.dev.yaml, config.prod.yaml)
//   - 支持配置热重载和默认值
//
// 使用示例:
//
//	// 基本用法
//	cfg, err := config.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	dbHost := cfg.GetString("database.host")
//	dbPort := cfg.GetInt("database.port")
//
//	// 带选项创建
//	cfg, err := config.New(
//	    config.WithConfigName("app"),
//	    config.WithConfigPath("./config", "/etc/app"),
//	    config.WithEnvironment("dev"),
//	)
//
//	// 使用环境变量默认
//	cfg, err := config.New(config.WithDefaultEnv("development"))
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ConfigOption 配置选项函数
//
// 用于在创建配置时设置各种选项.
type ConfigOption func(*ViperConfig) error

// ViperConfig Viper配置包装器
//
// 包装 viper.Viper 并提供类型安全的配置访问方法.
//
// 字段说明:
//
//   - v: 底层 Viper 实例
//   - env: 当前环境名称
//   - configName: 配置文件名(不含扩展名)
//   - configPaths: 配置搜索路径列表
//   - optionName: 环境变量前缀
//   - configFile: 配置文件完整路径
type ViperConfig struct {
	v           *viper.Viper
	env         string
	configName  string
	configPaths []string
	configType  string
	optionName  string
	configFile  string
}

// New 创建新的配置实例
//
// 根据提供的选项加载配置.
//
// 参数:
//   - opts: 可变数量的配置选项,如 WithConfigName, WithConfigPath, WithEnvironment 等
//
// 返回值:
//   - *ViperConfig: 配置实例
//   - error: 加载失败时返回错误
//
// 示例:
//
//	cfg, err := config.New(config.WithConfigName("config"))
func New(opts ...ConfigOption) (*ViperConfig, error) {
	cfg := &ViperConfig{
		v:           viper.New(),
		configPaths: []string{"./", "./config"},
		configName:  "config",
		configType:  "yaml",
		env:         "dev",
		optionName:  "",
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ViperConfig) load() error {
	if c.configFile != "" {
		c.v.SetConfigFile(c.configFile)
	} else {
		if c.env != "" {
			c.configName = fmt.Sprintf("%s.%s", c.configName, c.env)
		}
		c.v.SetConfigType(c.configType)
		c.v.SetConfigName(c.configName)
		for _, p := range c.configPaths {
			c.v.AddConfigPath(p)
		}
	}
	return c.v.ReadInConfig()
}

// Viper 获取底层 Viper 实例
//
// 返回值:
//   - *viper.Viper: Viper 实例,可用于高级操作
func (c *ViperConfig) Viper() *viper.Viper {
	return c.v
}

// Get 获取配置值
//
// 参数:
//   - key: 配置键,支持点号分隔的嵌套键,如 "database.host"
//
// 返回值:
//   - any: 配置值,如果不存在返回 nil
func (c *ViperConfig) Get(key string) any {
	return c.v.Get(key)
}

// GetString 获取字符串配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - string: 配置值,如果不存在返回空字符串
func (c *ViperConfig) GetString(key string) string {
	return c.v.GetString(key)
}

// GetBool 获取布尔配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - bool: 配置值,如果不存在返回 false
func (c *ViperConfig) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetInt 获取整数配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - int: 配置值,如果不存在返回 0
func (c *ViperConfig) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetInt64 获取64位整数配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - int64: 配置值,如果不存在返回 0
func (c *ViperConfig) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// GetFloat64 获取浮点数配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - float64: 配置值,如果不存在返回 0
func (c *ViperConfig) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - []string: 配置值,如果不存在返回 nil
func (c *ViperConfig) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// GetStringMap 获取字符串到任意值的映射
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - map[string]any: 配置值,如果不存在返回 nil
func (c *ViperConfig) GetStringMap(key string) map[string]any {
	return c.v.GetStringMap(key)
}

// GetStringMapString 获取字符串到字符串的映射
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - map[string]string: 配置值,如果不存在返回 nil
func (c *ViperConfig) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

// Sub 获取子配置
//
// 返回指定键下的子配置树.
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - *ViperConfig: 子配置,如果键不存在返回 nil
func (c *ViperConfig) Sub(key string) *ViperConfig {
	subV := c.v.Sub(key)
	if subV == nil {
		return nil
	}
	return &ViperConfig{v: subV, env: c.env}
}

// AllSettings 获取所有配置
//
// 返回值:
//   - map[string]any: 所有配置键值对
func (c *ViperConfig) AllSettings() map[string]any {
	return c.v.AllSettings()
}

// IsSet 检查配置是否已设置
//
// 参数:
//   - key: 配置键
//
// 返回值:
//   - bool: 如果已设置返回 true
func (c *ViperConfig) IsSet(key string) bool {
	return c.v.IsSet(key)
}

// Environment 获取当前环境
//
// 返回值:
//   - string: 环境名称,如 "dev", "prod"
func (c *ViperConfig) Environment() string {
	return c.env
}

// FileName 获取配置文件名
//
// 返回值:
//   - string: 配置名(不含扩展名)
func (c *ViperConfig) FileName() string {
	return c.configName
}

// ConfigFile 获取使用的配置文件路径
//
// 返回值:
//   - string: 配置文件完整路径
func (c *ViperConfig) ConfigFile() string {
	return c.v.ConfigFileUsed()
}

// WithConfigName 设置配置文件名
//
// 参数:
//   - name: 配置名(不含扩展名),如 "config", "app"
//
// 返回值:
//   - ConfigOption: 配置选项
//
// 示例:
//
//	config.New(config.WithConfigName("app"))
func WithConfigName(name string) ConfigOption {
	return func(c *ViperConfig) error {
		if name != "" {
			c.configName = name
		} else {
			log.Print("The WithConfigName parameter is zero, which will use the default configuration [config].")
		}
		return nil
	}
}

// WithConfigPath 设置配置搜索路径
//
// 参数:
//   - paths: 可变数量的搜索路径
//
// 返回值:
//   - ConfigOption: 配置选项
//
// 示例:
//
//	config.New(config.WithConfigPath("./config", "/etc/app"))
func WithConfigPath(paths ...string) ConfigOption {
	return func(c *ViperConfig) error {
		c.configPaths = paths
		return nil
	}
}

// WithConfigFile 直接指定配置文件
//
// 优先级高于 WithConfigName 和 WithConfigPath.
//
// 参数:
//   - path: 配置文件完整路径
//
// 返回值:
//   - ConfigOption: 配置选项
//
// 示例:
//
//	config.New(config.WithConfigFile("/etc/app/config.yaml"))
func WithConfigFile(path string) ConfigOption {
	return func(c *ViperConfig) error {
		c.configFile = path
		return nil
	}
}

// WithEnvironment 设置环境
//
// 环境会附加到配置文件名,如 "config.dev.yaml".
//
// 参数:
//   - env: 环境名称,如 "dev", "prod", "test"
//
// 返回值:
//   - ConfigOption: 配置选项
//
// 示例:
//
//	config.New(config.WithEnvironment("dev"))
func WithEnvironment(env string) ConfigOption {
	return func(c *ViperConfig) error {
		c.env = env
		return nil
	}
}

// WithEnvVariable 设置环境变量前缀
//
// 用于从环境变量读取配置时的默认前缀.
//
// 参数:
//   - name: 环境变量名
//
// 返回值:
//   - ConfigOption: 配置选项
//
// 示例:
//
//	config.New(config.WithEnvVariable("APP"))
func WithEnvVariable(name string) ConfigOption {
	return func(c *ViperConfig) error {
		c.optionName = name
		return nil
	}
}

// WithConfigType 显式指定配置类型
//
// 通常可以自动检测,某些情况下需要显式指定.
//
// 参数:
//   - typeName: 配置类型,如 "yaml", "json", "toml"
//
// 返回值:
//   - ConfigOption: 配置选项
func WithConfigType(typeName string) ConfigOption {
	return func(c *ViperConfig) error {
		if typeName != "" {
			c.configType = typeName
		} else {
			log.Print("The WithConfigName parameter is zero, which will use the default type [yaml].")
		}
		return nil
	}
}

// WithEnvReplacer 设置环境变量替换器
//
// 用于转换环境变量名.
//
// 参数:
//   - replacer: 字符串替换器
//
// 返回值:
//   - ConfigOption: 配置选项
func WithEnvReplacer(replacer *strings.Replacer) ConfigOption {
	return func(c *ViperConfig) error {
		c.v.SetEnvKeyReplacer(replacer)
		return nil
	}
}

// WithEnvPrefix 设置环境变量前缀
//
// 所有环境变量会加上此前缀.
//
// 参数:
//   - prefix: 前缀字符串
//
// 返回值:
//   - ConfigOption: 配置选项
func WithEnvPrefix(prefix string) ConfigOption {
	return func(c *ViperConfig) error {
		c.v.SetEnvPrefix(prefix)
		return nil
	}
}

// WithDefaultEnv 从环境变量自动检测环境
//
// 依次检查 APP_ENV, GO_ENV, ENV 环境变量.
// 如果都未设置,使用 defaultEnv.
//
// 参数:
//   - defaultEnv: 默认环境名称
//
// 返回值:
//   - ConfigOption: 配置选项
func WithDefaultEnv() ConfigOption {
	return func(c *ViperConfig) error {
		env := ""
		if v := os.Getenv("APP_ENV"); v != "" {
			env = v
		}
		if v := os.Getenv("GO_ENV"); v != "" {
			env = v
		}
		if v := os.Getenv("ENV"); v != "" {
			env = v
		}
		if env == "" {
			env = "dev"
		}
		c.env = env
		return nil
	}
}
