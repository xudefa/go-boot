// Package environment 提供配置源（PropertySource）的多种实现。
//
// 支持的配置源类型：
//   - MapPropertySource: 基于 map 的内存配置源
//   - EnvPropertySource: 基于环境变量的配置源，支持前缀映射
//   - ArgsPropertySource: 基于命令行参数的配置源（最高优先级）
package environment

import (
	"os"
	"strings"
)

// 五级优先级常量定义，值越大优先级越高
type Priority int

const (
	// PriorityLowest 最低优先级，应用配置使用
	PriorityLowest Priority = iota
	// PriorityLow 低优先级
	PriorityLow
	// PriorityNormal 正常优先级
	PriorityNormal
	// PriorityHigh 高优先级，环境变量使用
	PriorityHigh
	// PriorityHighest 最高优先级，命令行使用
	PriorityHighest
)

// PropertySource 配置源接口
//
// 代表一个配置数据源，具有名称和优先级。
// 多个 PropertySource 组成 Environment 的分层配置体系。
type PropertySource interface {
	// Name 返回配置源名称
	Name() string

	// GetProperty 返回指定键的值
	GetProperty(key string) (any, bool)

	// Priority 返回优先级（值越大优先级越高）
	Priority() Priority

	// Contains 检查键是否存在
	Contains(key string) bool
}

// MapPropertySource 基于 map 的内存配置源.
//
// 存储键值对数据,支持自定义优先级.
type MapPropertySource struct {
	name     string
	data     map[string]any
	priority Priority
}

// NewMapPropertySource 创建 MapPropertySource
func NewMapPropertySource(name string, priority Priority, data map[string]any) *MapPropertySource {
	if data == nil {
		data = make(map[string]any)
	}
	return &MapPropertySource{name: name, data: data, priority: priority}
}

// NewDefaultPropertySource 创建最低优先级的默认配置源
//
// 使用 PriorityLowest 优先级，被所有其他配置源覆盖。
func NewDefaultPropertySource(name string, data map[string]any) *MapPropertySource {
	return NewMapPropertySource(name, PriorityLowest, data)
}

func (m *MapPropertySource) Name() string {
	return m.name
}

func (m *MapPropertySource) GetProperty(key string) (any, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *MapPropertySource) Priority() Priority {
	return m.priority
}

func (m *MapPropertySource) Contains(key string) bool {
	_, ok := m.data[key]
	return ok
}

// Keys 返回所有键
func (m *MapPropertySource) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// EnvPropertySource 基于环境变量的配置源.
//
// 支持前缀映射,例如 prefix="GO_BOOT" 时,
// 键 "server.port" 将映射为环境变量 "GO_BOOT_SERVER_PORT".
type EnvPropertySource struct {
	name     string
	prefix   string
	priority Priority
}

// NewEnvPropertySource 创建环境变量配置源
//
// prefix 是环境变量前缀，例如 "GO_BOOT"。
// 环境变量 GO_BOOT_SERVER_PORT=9090 将映射为键 "server.port"。
func NewEnvPropertySource(name, prefix string) *EnvPropertySource {
	return &EnvPropertySource{name: name, prefix: prefix, priority: PriorityHigh}
}

func (e *EnvPropertySource) Name() string {
	return e.name
}

func (e *EnvPropertySource) Priority() Priority {
	return e.priority
}

func (e *EnvPropertySource) Contains(key string) bool {
	_, ok := e.GetProperty(key)
	return ok
}

// GetProperty 从环境变量获取配置
//
// 键 "server.port" 将转换为环境变量 "GO_BOOT_SERVER_PORT"（如果 prefix="GO_BOOT"）。
func (e *EnvPropertySource) GetProperty(key string) (any, bool) {
	envKey := toEnvKey(key)
	if e.prefix != "" {
		envKey = e.prefix + "_" + envKey
	}
	val, ok := lookupEnv(envKey)
	if !ok {
		return nil, false
	}
	return val, true
}

// lookupEnv 可被测试替换
var lookupEnv = os.LookupEnv

// ArgsPropertySource 基于命令行参数的配置源.
//
// 优先级最高（PriorityHighest），解析 --key=value 格式的命令行参数.
type ArgsPropertySource struct {
	name     string
	args     map[string]string
	priority Priority
}

// NewArgsPropertySource 解析命令行参数并创建配置源
//
// 支持的格式：--key=value
func NewArgsPropertySource(name string, args []string) *ArgsPropertySource {
	data := make(map[string]string)
	for _, arg := range args {
		if len(arg) > 2 && arg[:2] == "--" {
			kv := arg[2:]
			if key, val, found := strings.Cut(kv, "="); found && key != "" {
				data[key] = val
			}
		}
	}
	return &ArgsPropertySource{name: name, args: data, priority: PriorityHighest}
}

func (a *ArgsPropertySource) Name() string {
	return a.name
}

func (a *ArgsPropertySource) Priority() Priority {
	return a.priority
}

func (a *ArgsPropertySource) Contains(key string) bool {
	_, ok := a.args[key]
	return ok
}

func (a *ArgsPropertySource) GetProperty(key string) (any, bool) {
	val, ok := a.args[key]
	return val, ok
}

// toEnvKey 将 "server.port" 转换为 "SERVER_PORT"
func toEnvKey(key string) string {
	result := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '.' || c == '-' {
			result = append(result, '_')
			continue
		}
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		result = append(result, c)
	}
	return string(result)
}
