package container

import (
	"reflect"
	"sync"

	"github.com/xudefa/go-boot/core/errors"
)

// Scope 定义 Bean 作用域
type Scope string

const (
	// Singleton 单例作用域
	Singleton Scope = "singleton"
	// Prototype 原型作用域
	Prototype Scope = "prototype"
)

// Container IoC 容器接口
type Container interface {
	// Register 注册 Bean 定义
	Register(id string, def BeanDefinition) error

	// Get 获取 Bean 实例
	Get(id string) any

	// Contains 检查 Bean 是否存在
	Contains(id string) bool

	// GetDefinition 获取 Bean 定义
	GetDefinition(id string) (BeanDefinition, error)

	// ListDefinitions 列出所有 Bean 定义 ID
	ListDefinitions() []string

	// Start 启动容器（实例化所有 Singleton Bean）
	Start() error

	// Stop 停止容器（调用所有 Destroy 回调）
	Stop() error

	// Invoke 方法注入，自动解析参数并调用函数
	Invoke(fn any) error
}

// GetT 类型安全获取 Bean（辅助函数）
func GetT[T any](c Container, id string) (*T, error) {
	bean := c.Get(id)
	if result, ok := bean.(*T); ok {
		return result, nil
	}
	return nil, errors.ErrInvalidBeanType
}

// impl Container 接口的实现
type impl struct {
	mu          sync.RWMutex
	definitions map[string]BeanDefinition
	singletons  map[string]any
	prototypes  map[string]any
	started     bool
	stopped     bool
}

// New 创建新容器
func New() Container {
	return &impl{
		definitions: make(map[string]BeanDefinition),
		singletons:  make(map[string]any),
		prototypes:  make(map[string]any),
	}
}

// Register 实现 Container.Register
func (c *impl) Register(id string, def BeanDefinition) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.definitions[id]; exists {
		return errors.ErrDuplicateBean
	}

	def.ID = id
	c.definitions[id] = def
	return nil
}

// Get 实现 Container.Get
func (c *impl) Get(id string) any {
	c.mu.RLock()
	def, exists := c.definitions[id]
	c.mu.RUnlock()

	if !exists {
		panic(errors.ErrBeanNotFound)
	}

	return c.resolve(id, def)
}

// Contains 实现 Container.Contains
func (c *impl) Contains(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.definitions[id]
	return exists
}

// GetDefinition 实现 Container.GetDefinition
func (c *impl) GetDefinition(id string) (BeanDefinition, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	def, exists := c.definitions[id]
	if !exists {
		return BeanDefinition{}, errors.ErrBeanNotFound
	}
	return def, nil
}

// ListDefinitions 实现 Container.ListDefinitions
func (c *impl) ListDefinitions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.definitions))
	for id := range c.definitions {
		ids = append(ids, id)
	}
	return ids
}

// Start 实现 Container.Start
func (c *impl) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	// 实例化所有 Singleton Bean
	for id, def := range c.definitions {
		if def.Scope == Singleton && !def.Lazy {
			c.singletons[id] = c.createBean(def)
		}
	}

	c.started = true
	return nil
}

// Stop 实现 Container.Stop
func (c *impl) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return nil
	}

	// 调用所有 Destroy 回调
	for id, def := range c.definitions {
		if def.DestroyFunc != nil {
			if bean, exists := c.singletons[id]; exists {
				def.DestroyFunc(bean)
			}
		}
	}

	c.stopped = true
	return nil
}

// Invoke 实现 Container.Invoke
func (c *impl) Invoke(fn any) error {
	// TODO: 实现方法注入
	return nil
}

// resolve 解析 Bean 实例
func (c *impl) resolve(id string, def BeanDefinition) any {
	if def.Scope == Singleton {
		if bean, exists := c.singletons[id]; exists {
			return bean
		}
		bean := c.createBean(def)
		c.singletons[id] = bean
		return bean
	}

	// Prototype 每次创建新实例
	return c.createBean(def)
}

// createBean 创建 Bean 实例
func (c *impl) createBean(def BeanDefinition) any {
	var bean any

	if def.Factory != nil {
		var err error
		bean, err = def.Factory()
		if err != nil {
			panic(errors.NewError("BEAN_CREATE").
				Message("factory creation failed").
				Cause(err).
				Context("bean_id", def.ID).
				Build())
		}
	} else {
		bean = reflect.New(def.Type.Elem()).Interface()
	}

	// 执行 Init 回调
	if def.InitFunc != nil {
		if err := def.InitFunc(bean); err != nil {
			panic(errors.NewError("BEAN_INIT").
				Message("init callback failed").
				Cause(err).
				Context("bean_id", def.ID).
				Build())
		}
	}

	return bean
}
