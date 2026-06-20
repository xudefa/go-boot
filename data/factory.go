package data

import (
	"fmt"
	"sync"
)

// DBConfig 数据库工厂配置
//
// 指定 ORM 类型和数据库连接信息，用于 DBFactory 创建对应的数据访问层实现。
type DBConfig struct {
	ORM    string // ORM 类型： "gorm" | "xorm"
	Driver string // 数据库驱动： mysql | postgres | sqlite
	DSN    string // 数据库连接字符串
}

// DBFactory 数据库工厂
//
// 根据配置动态创建不同的 ORM 数据访问层实现。
// 通过 Register 注册 ORM 提供者，通过 Create 创建具体的 Transactor 实例。
type DBFactory struct {
	mu        sync.Mutex
	providers map[string]func(*DBConfig) (Transactor, error) // ORM 提供者映射
}

// NewDBFactory 创建数据库工厂实例，初始化空的 ORM 提供者映射
func NewDBFactory() *DBFactory {
	return &DBFactory{
		providers: make(map[string]func(*DBConfig) (Transactor, error)),
	}
}

// Register 注册 ORM 数据库提供者
//
// 参数:
//   - name: ORM 名称（如 "gorm", "xorm"）
//   - fn: 工厂函数，接收 DBConfig 返回 Transactor 实例
func (f *DBFactory) Register(name string, fn func(*DBConfig) (Transactor, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.providers == nil {
		f.providers = make(map[string]func(*DBConfig) (Transactor, error))
	}
	f.providers[name] = fn
}

// Create 创建数据库事务操作器实例
//
// 创建流程：
//  1. 校验配置不为空
//  2. 查找 ORM 类型对应的已注册提供者（默认 "gorm"）
//  3. 调用提供者工厂函数创建 Transactor 实例
func (f *DBFactory) Create(cfg *DBConfig) (Transactor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cfg is nil")
	}
	name := cfg.ORM
	if name == "" {
		name = "gorm"
	}
	f.mu.Lock()
	fn, ok := f.providers[name]
	f.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unsupported ORM: %s", name)
	}
	return fn(cfg)
}
