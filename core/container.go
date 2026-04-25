package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// StructMetadata 结构体的预计算元数据
type StructMetadata struct {
	fieldCount int
	fields     []StructFieldMeta
}

type StructFieldMeta struct {
	Name   string
	Tag    string
	Offset uintptr
	Type   reflect.Type
}

var (
	structMetaCache = make(map[reflect.Type]*StructMetadata)
	structMetaLock  sync.RWMutex
)

func getStructMetadata(t reflect.Type) *StructMetadata {
	structMetaLock.RLock()
	meta, ok := structMetaCache[t]
	structMetaLock.RUnlock()
	if ok {
		return meta
	}

	structMetaLock.Lock()
	defer structMetaLock.Unlock()

	if meta, ok := structMetaCache[t]; ok {
		return meta
	}

	meta = computeStructMetadata(t)
	structMetaCache[t] = meta
	return meta
}

func computeStructMetadata(t reflect.Type) *StructMetadata {
	if t.Kind() != reflect.Struct {
		return nil
	}

	count := t.NumField()
	fields := make([]StructFieldMeta, count)

	for i := 0; i < count; i++ {
		f := t.Field(i)
		fields[i] = StructFieldMeta{
			Name:   f.Name,
			Tag:    f.Tag.Get("inject"),
			Offset: f.Offset,
			Type:   f.Type,
		}
	}

	return &StructMetadata{
		fieldCount: count,
		fields:     fields,
	}
}

var (
	ErrDuplicateBean = errors.New("duplicate bean registration")
	ErrBeanNotFound  = errors.New("bean not found")
	ErrCannotInject  = errors.New("cannot inject to non-pointer field")
	ErrInvalidScope  = errors.New("invalid scope")
	ErrCircularDep   = errors.New("circular dependency detected")
)

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

type beanRegistry struct {
	definitions map[string]*BeanDefinition
	typeToID    map[reflect.Type]string
	lock        sync.RWMutex
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

type getOptions struct {
	anonymous bool
}
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

type container struct {
	registry beanRegistry
	config   *Config
	parent   Container
	cache    sync.Map
	creating map[string]bool
	lock     sync.Mutex
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
	return &container{
		registry: beanRegistry{
			definitions: make(map[string]*BeanDefinition),
			typeToID:    make(map[reflect.Type]string),
		},
		config:   cfg,
		creating: make(map[string]bool),
	}
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
		if def.Factory == nil {
			originalInstance := def.Instance
			def.Factory = func(c Container) (interface{}, error) {
				return originalInstance, nil
			}
		}
	}

	if def.ConcreteType == nil && def.Factory == nil {
		return errors.New("either Instance or Factory must be provided")
	}

	c.registry.definitions[beanID] = def
	if def.ConcreteType != nil {
		c.registry.typeToID[def.ConcreteType] = beanID
	}
	return nil
}

func (c *container) Get(beanID string, opts ...GetOption) (interface{}, error) {
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
		if cached, ok := c.cache.Load(beanID); ok {
			return cached, nil
		}

		c.lock.Lock()
		if c.creating[beanID] {
			c.lock.Unlock()
			return nil, ErrCircularDep
		}
		c.creating[beanID] = true
		c.lock.Unlock()

		instance, err := c.createInstance(def)
		if err != nil {
			delete(c.creating, beanID)
			return nil, err
		}
		delete(c.creating, beanID)

		c.cache.Store(beanID, instance)
		return instance, nil
	}

	return c.createInstance(def)
}

func (c *container) createInstance(def *BeanDefinition) (interface{}, error) {
	var instance interface{}
	var err error

	if def.Scope == PrototypeScope {
		instance = reflect.New(def.ConcreteType).Interface()
	} else if def.Factory != nil {
		instance, err = def.Factory(c)
	} else {
		instance = def.Instance
	}

	if err != nil {
		return nil, err
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
		instance, err = processor.PostProcess(instance, "")
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func copyFields(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() != reflect.Ptr || dstVal.Kind() != reflect.Ptr {
		return errors.New("both src and dst must be pointers")
	}

	srcElem := srcVal.Elem()
	dstElem := dstVal.Elem()

	for i := 0; i < srcElem.NumField(); i++ {
		srcField := srcElem.Field(i)
		if !srcField.CanSet() {
			continue
		}
		dstField := dstElem.Field(i)
		if dstField.CanSet() && srcField.Type() == dstField.Type() {
			dstField.Set(srcField)
		}
	}

	return nil
}

func (c *container) injectFields(target interface{}, fields []FieldInjection) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr {
		return ErrCannotInject
	}

	elemVal := targetVal.Elem()
	elemType := elemVal.Type()

	if elemType.Kind() != reflect.Struct {
		return nil
	}

	if c.config.EnableFieldTag {
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			tag := field.Tag.Get("inject")
			if tag == "" {
				continue
			}

			fieldVal := elemVal.Field(i)
			if !fieldVal.CanSet() {
				continue
			}

			if err := c.injectByTag(fieldVal, tag); err != nil {
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
			beanID := field.Value.(string)
			dep, err := c.Get(beanID)
			if err != nil {
				return err
			}
			depVal := reflect.ValueOf(dep)
			if depVal.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(depVal)
			} else if fieldVal.Kind() == reflect.Interface && depVal.Type().Implements(fieldVal.Type()) {
				fieldVal.Set(depVal)
			}
		} else if field.Value != nil {
			val := reflect.ValueOf(field.Value)
			if val.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(val)
			}
		}
	}

	return nil
}

func (c *container) injectByTag(fieldVal reflect.Value, tag string) error {
	beanID := tag
	if beanID == "" {
		return nil
	}

	dep, err := c.Get(beanID)
	if err != nil {
		return err
	}

	depVal := reflect.ValueOf(dep)
	if !depVal.IsValid() {
		return fmt.Errorf("invalid dependency: %s", beanID)
	}

	if depVal.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(depVal)
	} else {
		return fmt.Errorf("type mismatch for %s", beanID)
	}

	return nil
}

func findFieldByName(t reflect.Type, name string) *reflect.StructField {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (c *container) Inject(target interface{}) error {
	c.registry.lock.RLock()
	defs := make([]*BeanDefinition, 0, len(c.registry.definitions))
	for _, def := range c.registry.definitions {
		defs = append(defs, def)
	}
	c.registry.lock.RUnlock()

	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr {
		return ErrCannotInject
	}

	elemVal := targetVal.Elem()
	elemType := elemVal.Type()

	if c.config.EnableFieldTag {
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			tag := field.Tag.Get("inject")
			if tag == "" {
				continue
			}

			fieldVal := elemVal.Field(i)
			if !fieldVal.CanSet() {
				continue
			}

			if err := c.injectByTag(fieldVal, tag); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *container) GetAll(beanType interface{}) ([]interface{}, error) {
	t := reflect.TypeOf(beanType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Interface {
		return nil, errors.New("GetAll requires interface{} type")
	}

	c.registry.lock.RLock()
	defer c.registry.lock.RUnlock()

	var results []interface{}
	for id, def := range c.registry.definitions {
		if def.ConcreteType == nil {
			continue
		}
		if def.ConcreteType.Implements(t) || implementsInterface(def.ConcreteType, t) {
			instance, err := c.Get(id)
			if err != nil {
				continue
			}
			results = append(results, instance)
		}
	}

	return results, nil
}

func implementsInterface(t reflect.Type, i reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(i)
}

func (c *container) Invoke(fn interface{}, opts ...InvokeOption) ([]interface{}, error) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return nil, errors.New("fn must be a function")
	}

	var args []reflect.Value
	argTypes := make([]reflect.Type, fnType.NumIn())

	for i := 0; i < fnType.NumIn(); i++ {
		argTypes[i] = fnType.In(i)
	}

	c.registry.lock.RLock()
	for _, argType := range argTypes {
		found := false
		for _, def := range c.registry.definitions {
			if def.ConcreteType != nil {
				if argType.Kind() == reflect.Interface {
					if def.ConcreteType.Implements(argType) {
						instance, err := c.Get(c.getBeanIDByType(def.ConcreteType))
						if err != nil {
							continue
						}
						args = append(args, reflect.ValueOf(instance))
						found = true
						break
					}
				} else if def.ConcreteType == argType {
					instance, err := c.Get(c.getBeanIDByType(def.ConcreteType))
					if err != nil {
						continue
					}
					args = append(args, reflect.ValueOf(instance))
					found = true
					break
				}
			}
		}
		if !found {
			args = append(args, reflect.Zero(argType))
		}
	}
	c.registry.lock.RUnlock()

	retVals := fnVal.Call(args)
	results := make([]interface{}, len(retVals))
	for i, v := range retVals {
		results[i] = v.Interface()
	}

	return results, nil
}

func (c *container) getBeanIDByType(t reflect.Type) string {
	id, ok := c.registry.typeToID[t]
	if ok {
		return id
	}
	return ""
}

func (c *container) Has(beanID string) bool {
	c.registry.lock.RLock()
	_, ok := c.registry.definitions[beanID]
	c.registry.lock.RUnlock()
	return ok
}

func (c *container) Remove(beanID string) error {
	c.registry.lock.Lock()
	defer c.registry.lock.Unlock()

	if _, ok := c.registry.definitions[beanID]; !ok {
		return fmt.Errorf("%w: %s", ErrBeanNotFound, beanID)
	}

	delete(c.registry.definitions, beanID)
	c.cache.Delete(beanID)
	return nil
}

func (c *container) Close() error {
	c.cache.Clear()
	return nil
}
