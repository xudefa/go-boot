// Package config 定义配置管理的核心接口和模型。
//
// 提供统一的配置访问方式（Config 接口）、加载器链（LoaderChain）、
// 验证器（Validator）和热重载（WatchManager）等功能。
package config

import (
	"sort"
)

// Loader 配置加载器接口.
//
// 定义配置加载的统一方式,用于支持多种配置源.
// 优先级越低的加载器越先加载,后加载的会覆盖先加载的值.
type Loader interface {
	// Load 加载配置.
	//
	// 参数:
	//   - opts: 加载选项
	//
	// 返回:
	//   - Config: 配置实例
	//   - error: 加载错误
	Load(opts ...LoaderOption) (Config, error)

	// Priority 获取加载器优先级.
	//
	// 返回:
	//   - int: 优先级,数值越小优先级越高
	Priority() int

	// Name 获取加载器名称.
	//
	// 返回:
	//   - string: 加载器名称
	Name() string

	// SupportsWatch 是否支持热重载.
	//
	// 返回:
	//   - bool: 是否支持热重载
	SupportsWatch() bool
}

// LoaderOption 配置加载器选项函数类型.
//
// 用于在加载配置时设置加载选项.
type LoaderOption func(*LoaderModel) error

// LoaderModel 配置加载器模型结构体.
//
// 存储配置加载相关的参数.
type LoaderModel struct {
	Paths      []string
	FileName   string
	FileType   string
	Env        string
	Prefix     string
	RemoteType string
	Endpoints  []string
	Key        string
}

// LoaderChain 配置加载器链.
//
// 用于管理多个加载器,并按优先级排序.
type LoaderChain struct {
	loaders []Loader
}

// Add 添加加载器到链中
//
// 参数:
//   - loader: 配置加载器实例
func (c *LoaderChain) Add(loader Loader) {
	c.loaders = append(c.loaders, loader)
}

// Len 返回加载器数量
//
// 实现 sort.Interface 接口，用于按优先级排序。
//
// 返回:
//   - int: 加载器数量
func (c *LoaderChain) Len() int {
	return len(c.loaders)
}

// Less 比较两个加载器的优先级
//
// 实现 sort.Interface 接口，优先级数值越小的加载器排在越前面。
//
// 参数:
//   - i: 第一个加载器索引
//   - j: 第二个加载器索引
//
// 返回:
//   - bool: i 的优先级是否低于 j
func (c *LoaderChain) Less(i, j int) bool {
	if c.loaders[i] == nil || c.loaders[j] == nil {
		return false
	}
	return c.loaders[i].Priority() < c.loaders[j].Priority()
}

// Swap 交换两个加载器位置
//
// 实现 sort.Interface 接口。
//
// 参数:
//   - i: 第一个加载器索引
//   - j: 第二个加载器索引
func (c *LoaderChain) Swap(i, j int) {
	c.loaders[i], c.loaders[j] = c.loaders[j], c.loaders[i]
}

// Sorted 返回按优先级排序的加载器列表.
func (c *LoaderChain) Sorted() []Loader {
	sorted := make([]Loader, len(c.loaders))
	copy(sorted, c.loaders)
	sc := &LoaderChain{loaders: sorted}
	sort.Sort(sc)
	return sorted
}

// WithPaths 设置配置搜索路径选项.
//
// 参数:
//   - paths: 搜索路径列表
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithPaths(paths ...string) LoaderOption {
	return func(m *LoaderModel) error {
		m.Paths = paths
		return nil
	}
}

// WithFileName 设置文件名选项.
//
// 参数:
//   - name: 文件名
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithFileName(name string) LoaderOption {
	return func(m *LoaderModel) error {
		m.FileName = name
		return nil
	}
}

// WithLoaderFileType 设置文件类型选项.
//
// 参数:
//   - fileType: 文件类型,如 json
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithLoaderFileType(fileType string) LoaderOption {
	return func(m *LoaderModel) error {
		m.FileType = fileType
		return nil
	}
}

// WithLoaderEnv 设置环境选项.
//
// 参数:
//   - env: 环境名称
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithLoaderEnv(env string) LoaderOption {
	return func(m *LoaderModel) error {
		m.Env = env
		return nil
	}
}

// WithPrefix 设置环境变量前缀选项.
//
// 参数:
//   - prefix: 环境变量前缀
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithPrefix(prefix string) LoaderOption {
	return func(m *LoaderModel) error {
		m.Prefix = prefix
		return nil
	}
}

// WithRemoteType 设置远程配置类型选项.
//
// 参数:
//   - remoteType: 远程配置类型,如 etcd, consul, apollo
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithRemoteType(remoteType string) LoaderOption {
	return func(m *LoaderModel) error {
		m.RemoteType = remoteType
		return nil
	}
}

// WithEndpoints 设置远程配置端点选项.
//
// 参数:
//   - endpoints: 端点地址列表
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithEndpoints(endpoints []string) LoaderOption {
	return func(m *LoaderModel) error {
		m.Endpoints = endpoints
		return nil
	}
}

// WithLoaderKey 设置配置键选项.
//
// 参数:
//   - key: 配置键
//
// 返回:
//   - LoaderOption: 加载器选项函数
func WithLoaderKey(key string) LoaderOption {
	return func(m *LoaderModel) error {
		m.Key = key
		return nil
	}
}
