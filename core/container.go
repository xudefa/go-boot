// Package core 提供了一个轻量级的依赖注入(DI)容器实现,灵感来自Spring Framework的IoC容器.
//
// # 核心功能
//
//   - Bean注册: 支持通过实例、工厂函数或类型注册bean
//   - 依赖注入: 支持字段注入(通过inject标签)和构造函数注入
//   - 作用域管理: 支持单例(Singleton)和原型(Prototype)作用域
//   - 组件扫描: 自动扫描并注册带有组件标签的结构体
//   - 生命周期管理: 支持初始化函数和Bean后置处理器
//   - 条件注册: 支持根据条件动态决定是否注册bean
//   - 泛型支持: 提供类型安全的泛型API
//
// # 快速开始
//
//	// 创建容器
//	container := core.New()
//
//	// 注册bean
//	container.Register("userService", core.Bean(&UserService{}))
//
//	// 获取bean
//	svc, err := container.Get("userService")
//
//	// 自动注入依赖
//	type Handler struct {
//	    Service *UserService `inject:"userService"`
//	}
//	var h Handler
//	container.Inject(&h)
//
// # Bean作用域
//
//   - SingletonScope: 单例模式,容器缓存实例,多次获取返回同一实例
//   - PrototypeScope: 原型模式,每次获取都创建新实例
//
// # 组件扫描
//
// 通过在结构体注释中添加@Component、@Service、@Repository或@Configuration标签,
// 可以使用ComponentScanner自动扫描并注册组件.
//
// # 相关包
//
//   - core.BuilderOption: 用于配置bean注册的构建器选项
//   - core.BeanDefinition: bean的元数据定义
//   - core.BeanPostProcessor: bean后置处理器接口
package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// StructFieldMeta 结构体字段元数据
//
// 记录结构体字段的名称、注入标签、偏移量和反射类型信息。
type StructFieldMeta struct {
	Name   string       // 字段名称
	Tag    string       // 字段的注入标签（inject）
	Offset uintptr      // 字段在结构体中的内存偏移量
	Type   reflect.Type // 字段的反射类型
}

var (
	ErrDuplicateBean = errors.New("duplicate bean registration")
	ErrBeanNotFound  = errors.New("bean not found")
	ErrCannotInject  = errors.New("cannot inject to non-pointer field")
	ErrInvalidScope  = errors.New("invalid scope")
	ErrCircularDep   = errors.New("circular dependency detected")
)

// BeanCreator Bean 创建器接口，用于解决循环导入问题
type BeanCreator interface {
	// CreateBean 创建 Bean 实例
	CreateBean(beanID string) (any, error)
}

// BeanInfo Bean 摘要信息
type BeanInfo struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Scope     string   `json:"scope"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// BeanScope 定义bean的作用域类型
type BeanScope string

const (
	// SingletonScope 单例作用域,容器只创建一个实例并缓存
	SingletonScope BeanScope = "singleton"
	// PrototypeScope 原型作用域,每次获取都创建新实例
	PrototypeScope BeanScope = "prototype"
)

// BeanDefinition 定义bean的元数据信息
//
// 字段说明:
//   - Instance: bean的实例(如果直接注册实例)
//   - OriginalInstance: 原始实例(用于复制)
//   - ConcreteType: bean的具体类型
//   - Scope: bean的作用域
//   - Factory: 工厂函数,用于创建bean实例
//   - Fields: 字段注入列表
//   - Init: 初始化函数
//   - DependsOn: 依赖的bean ID列表
//   - Condition: 条件创建函数,返回true时才创建bean
type BeanDefinition struct {
	Instance         any
	OriginalInstance any
	ConcreteType     reflect.Type
	Scope            BeanScope
	Factory          func(Container) (any, error)
	Fields           []FieldInjection
	Init             func(any) error
	DependsOn        []string
	Condition        func(Container) bool
	PostProcessors   []BeanPostProcessor
	RefreshScope     bool     // 是否支持刷新
	ConfigKeys       []string // Bean 依赖的配置键
}

// BeanPostProcessor bean后置处理器接口
//
// 在bean初始化之后调用,允许对bean进行修改或包装
type BeanPostProcessor interface {
	PostProcess(bean any, beanID string) (any, error)
}

// FieldInjection 定义字段注入配置
//
// 字段说明:
//   - Name: 目标字段名
//   - Value: 注入的值(可以是具体值或bean ID)
//   - IsRef: 是否为引用(true表示Value是bean ID)
type FieldInjection struct {
	Name  string
	Value any
	IsRef bool
}

// beanRegistry Bean 注册表内部实现
//
// 管理 Bean 定义的存储和索引，支持按 ID 和按类型查找。
type beanRegistry struct {
	definitions map[string]*BeanDefinition // Bean ID 到定义的映射
	typeToIDs   map[reflect.Type][]string  // 反射类型到 Bean ID 列表的映射
	lock        sync.RWMutex               // 保护并发访问的读写锁
}

// Container 是一个依赖注入容器,用于管理bean的注册、解析和生命周期
//
// # 功能概述
//
// - Bean注册: 支持通过实例、工厂函数或类型注册bean
// - 依赖注入: 支持字段注入(通过inject标签)和构造函数注入
// - 作用域: 支持单例(singleton)和原型(prototype)作用域
// - 方法注入: 支持通过Invoke调用函数并自动注入依赖
//
// # 使用示例
//
//	// 创建容器
//	container := core.New()
//
//	// 注册单例bean
//	container.Register("userService", core.Bean(&UserService{}))
//
//	// 注册带工厂函数的bean
//	container.Register("config", core.Factory(func(c core.Container) (interface{}, error) {
//	    return &Config{Path: "/etc/app"}, nil
//	}, reflect.TypeOf((*Config)(nil)).Elem()))
//
//	// 获取bean
//	userService, _ := container.Get("userService")
//
//	// 自动注入到结构体
//	var handler MyHandler
//	container.Inject(&handler)
//
//	// 调用函数并注入依赖
//	result, _ := container.Invoke(myFunc)
type Container interface {
	// Register 注册一个bean到容器中
	//
	// 参数:
	//   - beanID: bean的唯一标识符
	//   - builder: 可选的构建选项,如core.Bean(), core.Factory(), core.Singleton(), core.Prototype()等
	//
	// 返回值:
	//   - error: 注册失败时返回错误,可能为ErrDuplicateBean
	//
	// 示例:
	//
	//	// 注册实例bean
	//	container.Register("service", core.Bean(&MyService{}))
	//
	//	// 注册工厂bean
	//	container.Register("config", core.Factory(func(c core.Container) (interface{}, error) {
	//	    return loadConfig(), nil
	//	}, reflect.TypeOf((*Config)(nil)).Elem()))
	//
	//	// 注册单例(默认)
	//	container.Register("singleton", core.Bean(&Obj{}), core.Singleton())
	//
	//	// 注册原型
	//	container.Register("prototype", core.Bean(&Obj{}), core.Prototype())
	//
	//	// 带有依赖和初始化
	//	container.Register("service", core.Bean(&Service{}),
	//	    core.DependsOn("db", "logger"),
	//	    core.Init(func(s interface{}) error { return s.(*Service).Init() }))
	Register(beanID string, builder ...BuilderOption) error

	// Inject 自动注入目标结构体中的依赖字段
	//
	// 参数:
	//   - target: 目标结构体指针,字段需使用`inject`标签指定beanID
	//
	// 返回值:
	//   - error: 注入失败时返回错误
	//
	// 注意:
	//   - 只有使用`inject`标签的导出的可设置字段才会被注入
	//   - 支持通过父容器链向上查找依赖
	//
	// 示例:
	//
	//	type Handler struct {
	//	    Service *MyService `inject:"myService"`
	//	    Logger  Logger     `inject:"logger"`
	//	}
	//
	//	var h Handler
	//	container.Inject(&h)
	Inject(target any) error

	// Get 根据beanID获取bean实例
	//
	// 参数:
	//   - beanID: bean的唯一标识符
	//   - opts: 可选参数,如core.WithAnonymous()
	//
	// 返回值:
	//   - any: bean实例
	//   - error: 获取失败时返回错误,可能为ErrBeanNotFound或ErrCircularDep
	//
	// 注意:
	//   - 单例bean会被缓存,多次获取返回同一实例
	//   - 原型bean每次调用都会创建新实例
	//   - 支持通过父容器链向上查找
	//
	// 示例:
	//
	//	svc, err := container.Get("myService")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	//	svc.(*MyService).DoSomething()
	Get(beanID string, opts ...GetOption) (any, error)

	// GetAll 获取指定接口类型的所有实现bean
	//
	// 参数:
	//   - beanType: 接口类型
	//
	// 返回值:
	//   - []any: 所有实现该接口的bean实例数组
	//   - error: 获取失败时返回错误
	//
	// 注意:
	//   - beanType必须是接口类型
	//   - 只返回实现了该接口的bean
	//
	// 示例:
	//
	//	// 获取所有实现了Plugin接口的bean
	//	plugins, _ := container.GetAll((*Plugin)(nil))
	//	for _, p := range plugins {
	//	    p.(Plugin).Init()
	//	}
	GetAll(beanType any) ([]any, error)

	// Invoke 自动调用函数并注入依赖
	//
	// 参数:
	//   - fn: 要调用的函数,参数会自动从容器中获取
	//   - opts: 可选参数(预留)
	//
	// 返回值:
	//   - []any: 函数的返回值数组
	//   - error: 调用失败时返回错误
	//
	// 注意:
	//   - 函数的每个参数都会尝试从容器中获取对应类型的bean
	//   - 如果找不到对应类型,则传入nil
	//   - 只支持函数类型
	//
	// 示例:
	//
	//	// 假设容器中有*UserService和*Logger
	//	result, err := container.Invoke(func(s *UserService, l Logger) error {
	//	    return s.DoSomething(l)
	//	})
	Invoke(fn any, opts ...InvokeOption) ([]any, error)

	// ListBeans 返回所有已注册 bean 的摘要信息列表
	//
	// 返回值:
	//   - []BeanInfo: bean 摘要信息列表，包含 ID、类型、作用域和依赖
	//
	// 示例:
	//
	//	for _, info := range container.ListBeans() {
	//	    fmt.Printf("bean: %s (%s)\n", info.ID, info.Type)
	//	}
	ListBeans() []BeanInfo

	// Has 检查容器中是否存在指定ID的bean
	//
	// 参数:
	//   - beanID: bean的唯一标识符
	//
	// 返回值:
	//   - bool: 存在返回true,否则返回false
	//
	// 注意:
	//   - 会检查当前容器和父容器
	Has(beanID string) bool

	// Remove 从容器中移除指定ID的bean
	//
	// 参数:
	//   - beanID: bean的唯一标识符
	//
	// 返回值:
	//   - error: 移除失败时返回错误,如bean不存在
	//
	// 注意:
	//   - 也会清除单例bean的缓存
	Remove(beanID string) error

	// Close 关闭容器,清空所有缓存
	//
	// 返回值:
	//   - error: 关闭时发生错误(当前实现总是返回nil)
	Close() error

	BeanCreator
}

// BuilderOption bean注册构建选项函数
//
// 用于配置BeanDefinition的各个属性
type BuilderOption func(*BeanDefinition) error

// GetOption 获取bean时的选项函数
//
// 用于配置Get方法的行为,如获取匿名bean
type GetOption func(*getOptions)

// InvokeOption 调用函数时的选项函数(预留)
type InvokeOption func(*invokeOptions)

// getOptions 获取 Bean 时的选项配置
type getOptions struct {
	anonymous bool // 是否获取匿名 Bean（预留功能）
}

// invokeOptions 调用函数时的选项配置（预留）
type invokeOptions struct{}

// WithAnonymous 获取匿名bean的选项
//
// 返回值:
//   - GetOption: 用于Get方法的选项
//
// 注意:
//   - 匿名bean是指没有显式ID的bean(预留功能)
func WithAnonymous() GetOption {
	return func(o *getOptions) {
		o.anonymous = true
	}
}

// container IoC 容器的内部实现
//
// 组合了 Bean 注册表、缓存、并发控制和父容器链。
type container struct {
	registry      beanRegistry       // Bean 定义注册表
	config        *Config            // 容器配置选项
	parent        Container          // 父容器（用于容器链查找）
	cache         sync.Map           // 单例 Bean 实例缓存
	creationGuard *BeanCreationGuard // Bean 创建守卫，管理并发安全
}

// Config 容器配置选项
//
// 字段说明:
//   - EnableFieldTag: 是否启用inject标签注入(默认true)
//   - ScanPackages: 要扫描的包路径列表(预留)
type Config struct {
	EnableFieldTag bool
	ScanPackages   []string
}

func New(opts ...Option) Container {
	cfg := &Config{
		EnableFieldTag: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	c := &container{
		registry: beanRegistry{
			definitions: make(map[string]*BeanDefinition),
			typeToIDs:   make(map[reflect.Type][]string),
		},
		config: cfg,
	}
	c.creationGuard = NewBeanCreationGuard(c)
	return c
}

// Option 容器配置函数选项
//
// 用于配置Container的行为,如启用/禁用标签注入、扫描包等
type Option func(*Config)

// EnableFieldTag 设置是否启用inject标签注入
//
// 参数:
//   - enable: 是否启用(默认true)
//
// 返回值:
//   - Option: 可传递给New的选项
//
// 示例:
//
//	container := core.New(core.EnableFieldTag(true))
func EnableFieldTag(enable bool) Option {
	return func(c *Config) {
		c.EnableFieldTag = enable
	}
}

// Scan 设置要扫描的包路径(预留功能)
//
// 参数:
//   - packages: 包路径列表
//
// 返回值:
//   - Option: 可传递给New的选项
func Scan(packages ...string) Option {
	return func(c *Config) {
		c.ScanPackages = packages
	}
}

func (c *container) Register(beanID string, builders ...BuilderOption) error {
	c.registry.lock.Lock()
	defer c.registry.lock.Unlock()

	if _, ok := c.registry.definitions[beanID]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicateBean, beanID)
	}

	def := &BeanDefinition{
		Scope:     SingletonScope,
		DependsOn: []string{},
	}

	for _, builder := range builders {
		if err := builder(def); err != nil {
			return err
		}
	}

	if def.Instance != nil {
		if def.ConcreteType == nil {
			def.ConcreteType = reflect.TypeOf(def.Instance)
		}
		// 原型模式: 只记录 Instance 和 ConcreteType,不设置 Factory,
		// createInstance 中通过反射克隆来创建新实例
		if def.Factory == nil && def.Scope != PrototypeScope {
			originalInstance := def.Instance
			def.Factory = func(c Container) (any, error) {
				return originalInstance, nil
			}
		}
	}

	if def.ConcreteType == nil && def.Factory == nil {
		return errors.New("either Instance or Factory must be provided")
	}

	c.registry.definitions[beanID] = def
	if def.ConcreteType != nil {
		c.registry.typeToIDs[def.ConcreteType] = append(c.registry.typeToIDs[def.ConcreteType], beanID)
	}
	return nil
}

// Get 从容器中获取指定 ID 的 Bean 实例
//
// 获取流程：
//  1. 查找 Bean 定义，未找到时尝试从父容器获取
//  2. 检查条件函数 Condition，不满足时返回 ErrBeanNotFound
//  3. 单例模式：先查缓存，缓存未命中则在双重检查锁定下创建
//  4. 并发控制：检测到循环依赖时等待其他协程完成或返回 ErrCircularDep
//  5. 原型模式：每次调用都创建新实例
func (c *container) Get(beanID string, opts ...GetOption) (any, error) {
	getOpts := &getOptions{}
	for _, opt := range opts {
		opt(getOpts)
	}

	c.registry.lock.RLock()
	def, ok := c.registry.definitions[beanID]
	c.registry.lock.RUnlock()

	if !ok {
		if c.parent != nil {
			return c.parent.Get(beanID)
		}
		return nil, fmt.Errorf("%w: %s", ErrBeanNotFound, beanID)
	}

	if def.Condition != nil && !def.Condition(c) {
		return nil, fmt.Errorf("%w: %s (condition not met)", ErrBeanNotFound, beanID)
	}

	if def.Scope == SingletonScope {
		return c.getSingleton(beanID, def)
	}

	return c.createInstance(beanID, def)
}

// getSingleton 获取单例 Bean 实例
//
// 获取流程：
//  1. 先查缓存，命中则直接返回
//  2. 双重检查锁定，避免重复创建
//  3. 检测并发创建等待和循环依赖
//  4. 创建成功后缓存实例
func (c *container) getSingleton(beanID string, def *BeanDefinition) (any, error) {
	if cached, ok := c.cache.Load(beanID); ok {
		return cached, nil
	}

	return c.creationGuard.GetOrCompute(beanID, func() (any, error) {
		instance, err := c.createInstance(beanID, def)
		if err != nil {
			return nil, err
		}

		c.cache.Store(beanID, instance)
		return instance, nil
	})
}

// createInstance 创建 Bean 实例
//
// 创建流程：
//  1. 根据作用域（单例/原型）和定义类型（实例/工厂）选择创建策略
//  2. 原型模式：通过反射克隆或工厂函数创建新实例
//  3. 单例模式：通过工厂函数创建或直接使用注册的实例
//  4. 注入配置的字段依赖
//  5. 调用初始化函数 Init
//  6. 执行所有后置处理器 PostProcessors
func (c *container) createInstance(beanID string, def *BeanDefinition) (any, error) {
	var instance any
	var err error

	if def.Scope == PrototypeScope {
		if def.ConcreteType == nil {
			return nil, fmt.Errorf("ConcreteType is nil for prototype bean")
		}
		// 优先使用 Factory 创建（Bean() + Prototype() 时不会设置 Factory）
		if def.Factory != nil {
			instance, err = def.Factory(c)
			if err != nil {
				return nil, err
			}
			if instance == nil {
				return nil, fmt.Errorf("factory returned nil instance")
			}
		} else {
			// 反射创建新实例,并深拷贝初始值（避免 sync.Mutex 等不可复制类型的共享问题）
			var t reflect.Type
			if def.ConcreteType.Kind() == reflect.Pointer {
				t = def.ConcreteType.Elem()
			} else {
				t = def.ConcreteType
			}
			instance = reflect.New(t).Interface()
			if def.Instance != nil {
				srcVal := reflect.ValueOf(def.Instance)
				dstVal := reflect.ValueOf(instance).Elem()
				if srcVal.Kind() == reflect.Pointer {
					dstVal.Set(deepCopy(srcVal.Elem()))
				} else {
					dstVal.Set(deepCopy(srcVal))
				}
			}
		}
	} else if def.Factory != nil {
		instance, err = def.Factory(c)
		if err != nil {
			return nil, err
		}
		if instance == nil {
			return nil, fmt.Errorf("factory returned nil instance")
		}
		// 如果 Factory 注册时未指定 ConcreteType（如 ComponentScanner 懒加载路径），
		// 在首次创建后根据实际实例类型补齐，确保 typeToIDs 索引可用。
		if def.ConcreteType == nil {
			def.ConcreteType = reflect.TypeOf(instance)
			if def.ConcreteType != nil {
				c.registry.lock.Lock()
				c.registry.typeToIDs[def.ConcreteType] = append(c.registry.typeToIDs[def.ConcreteType], beanID)
				c.registry.lock.Unlock()
			}
		}
	} else {
		if def.Instance == nil {
			return nil, fmt.Errorf("instance is nil for singleton bean")
		}
		instance = def.Instance
	}

	if err := c.injectFields(instance, def.Fields); err != nil {
		return nil, err
	}

	if def.Init != nil {
		if err := def.Init(instance); err != nil {
			return nil, err
		}
	}

	for _, processor := range def.PostProcessors {
		if processor == nil {
			continue
		}
		instance, err = processor.PostProcess(instance, beanID)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

// injectFields 向目标结构体注入依赖字段
//
// 注入流程：
//  1. 校验目标为可设置的指针类型
//  2. 如果启用了字段标签注入（EnableFieldTag），遍历字段处理 inject 标签
//  3. 处理配置的 FieldInjection 列表（支持值注入和引用注入）
//  4. 引用注入时从容器获取依赖 Bean，进行类型匹配后设置字段值
func (c *container) injectFields(target any, fields []FieldInjection) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer {
		return ErrCannotInject
	}

	elemVal := targetVal.Elem()
	elemType := elemVal.Type()

	if elemType.Kind() != reflect.Struct {
		return nil
	}

	injector := NewFieldInjector(c)

	if c.config.EnableFieldTag {
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			tag := field.Tag.Get("inject")
			if tag == "" {
				continue
			}

			if !elemVal.Field(i).CanSet() {
				continue
			}

			if err := injector.InjectField(elemVal, field, tag, "tag"); err != nil {
				return err
			}
		}
	}

	for _, field := range fields {
		structField := findFieldByName(elemType, field.Name)
		if structField == nil {
			continue
		}

		fieldVal := elemVal.FieldByIndex(structField.Index)
		if !fieldVal.CanSet() {
			continue
		}

		if field.IsRef {
			if err := injector.InjectField(elemVal, *structField, field.Value.(string), "ref"); err != nil {
				return err
			}
		} else if field.Value != nil {
			val := reflect.ValueOf(field.Value)
			if val.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(val)
			} else {
				return fmt.Errorf("type mismatch for field %s: cannot assign %s to %s",
					field.Name, val.Type(), fieldVal.Type())
			}
		}
	}

	return nil
}

func findFieldByName(t reflect.Type, name string) *reflect.StructField {
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (c *container) Inject(target any) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer {
		return ErrCannotInject
	}

	elemVal := targetVal.Elem()
	if !elemVal.IsValid() {
		return fmt.Errorf("cannot inject into nil pointer")
	}
	elemType := elemVal.Type()

	if c.config.EnableFieldTag {
		injector := NewFieldInjector(c)
		for i := range elemType.NumField() {
			field := elemType.Field(i)
			tag := field.Tag.Get("inject")
			if tag == "" {
				continue
			}

			if !elemVal.Field(i).CanSet() {
				continue
			}

			if err := injector.InjectField(elemVal, field, tag, "tag"); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAll 获取所有实现了指定接口类型的 Bean
//
// 获取流程：
//  1. 校验 beanType 为接口类型
//  2. 遍历注册表中所有 Bean 定义，筛选实现该接口的类型
//  3. 逐个获取筛选出的 Bean 实例（跳过获取失败的）
//  4. 返回所有成功获取的实例列表
func (c *container) GetAll(beanType any) ([]any, error) {
	t := reflect.TypeOf(beanType)
	if t == nil {
		return nil, errors.New("beanType must not be nil")
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Interface {
		return nil, errors.New("GetAll requires interface type")
	}

	// 复制定义 ID 列表,释放读锁后逐个获取,避免死锁
	c.registry.lock.RLock()
	ids := make([]string, 0)
	for id, def := range c.registry.definitions {
		if def.ConcreteType != nil && (def.ConcreteType.Implements(t) || implementsInterface(def.ConcreteType, t)) {
			ids = append(ids, id)
		}
	}
	c.registry.lock.RUnlock()

	var results []any
	for _, id := range ids {
		instance, err := c.Get(id)
		if err != nil {
			continue
		}
		results = append(results, instance)
	}

	return results, nil
}

func implementsInterface(t reflect.Type, i reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Implements(i)
}

// Invoke 自动调用函数并注入参数
//
// 调用流程：
//  1. 校验 fn 为有效的函数类型
//  2. 解析函数参数类型列表
//  3. 遍历参数类型，在注册表中查找匹配的 Bean：
//     - 接口参数：查找实现了该接口的 Bean
//     - 具体类型参数：查找类型匹配的 Bean
//  4. 未找到匹配参数时传入零值
//  5. 调用函数并返回结果
func (c *container) Invoke(fn any, opts ...InvokeOption) ([]any, error) {
	fnVal := reflect.ValueOf(fn)
	if !fnVal.IsValid() {
		return nil, errors.New("fn must not be nil")
	}
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return nil, errors.New("fn must be a function")
	}

	argTypes := make([]reflect.Type, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		argTypes[i] = fnType.In(i)
	}

	// 先收集每个参数类型对应的 beanID，不调用 Get()（可能获取写锁导致死锁）
	beanIDs := make([]string, len(argTypes))
	foundMap := make([]bool, len(argTypes))

	c.registry.lock.RLock()
	for i, argType := range argTypes {
		for _, def := range c.registry.definitions {
			if def.ConcreteType != nil {
				if argType.Kind() == reflect.Interface {
					if def.ConcreteType.Implements(argType) {
						beanIDs[i] = c.getBeanIDByType(def.ConcreteType)
						foundMap[i] = true
						break
					}
				} else if def.ConcreteType == argType {
					beanIDs[i] = c.getBeanIDByType(def.ConcreteType)
					foundMap[i] = true
					break
				} else if argType.Kind() == reflect.Pointer && def.ConcreteType == argType.Elem() {
					// argType 是指针且 ConcreteType 是指针元素类型（如 *Database vs Database）
					beanIDs[i] = c.getBeanIDByType(def.ConcreteType)
					foundMap[i] = true
					break
				} else if def.ConcreteType.Kind() == reflect.Pointer && def.ConcreteType.Elem() == argType {
					// ConcreteType 是指针且 argType 是指针元素类型（如 Database vs *Database）
					beanIDs[i] = c.getBeanIDByType(def.ConcreteType)
					foundMap[i] = true
					break
				}
			}
		}
	}
	c.registry.lock.RUnlock()

	var args []reflect.Value
	for i, argType := range argTypes {
		if foundMap[i] {
			instance, err := c.Get(beanIDs[i])
			if err != nil {
				args = append(args, reflect.Zero(argType))
			} else {
				args = append(args, reflect.ValueOf(instance))
			}
		} else {
			args = append(args, reflect.Zero(argType))
		}
	}

	retVals := fnVal.Call(args)
	results := make([]any, len(retVals))
	for i, v := range retVals {
		results[i] = v.Interface()
	}

	return results, nil
}

func (c *container) getBeanIDByType(t reflect.Type) string {
	ids, ok := c.registry.typeToIDs[t]
	if ok && len(ids) > 0 {
		return ids[0]
	}
	return ""
}

func (c *container) ListBeans() []BeanInfo {
	c.registry.lock.RLock()
	defer c.registry.lock.RUnlock()

	infos := make([]BeanInfo, 0, len(c.registry.definitions))
	for id, def := range c.registry.definitions {
		info := BeanInfo{
			ID:        id,
			Scope:     string(def.Scope),
			DependsOn: def.DependsOn,
		}
		if def.ConcreteType != nil {
			info.Type = def.ConcreteType.String()
		}
		infos = append(infos, info)
	}
	return infos
}

func (c *container) Has(beanID string) bool {
	c.registry.lock.RLock()
	_, ok := c.registry.definitions[beanID]
	c.registry.lock.RUnlock()
	if ok {
		return true
	}
	if c.parent != nil {
		return c.parent.Has(beanID)
	}
	return false
}

func (c *container) Remove(beanID string) error {
	c.registry.lock.Lock()
	defer c.registry.lock.Unlock()

	def, ok := c.registry.definitions[beanID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrBeanNotFound, beanID)
	}

	delete(c.registry.definitions, beanID)
	c.cache.Delete(beanID)

	if def.ConcreteType != nil {
		ids := c.registry.typeToIDs[def.ConcreteType]
		for i, id := range ids {
			if id == beanID {
				c.registry.typeToIDs[def.ConcreteType] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (c *container) Close() error {
	c.cache.Clear()
	return nil
}

// CreateBean 实现 BeanCreator 接口，用于创建 Bean 实例
func (c *container) CreateBean(beanID string) (any, error) {
	c.registry.lock.RLock()
	def, ok := c.registry.definitions[beanID]
	c.registry.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrBeanNotFound, beanID)
	}

	return c.createInstance(beanID, def)
}
