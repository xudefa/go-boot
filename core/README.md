# go-boot/core

go-boot 核心依赖注入容器模块，提供类型安全的依赖注入和 bean 生命周期管理。

## 功能特性

- **Bean 注册** - 支持实例、工厂函数、类型注册
- **依赖注入** - 通过 `inject` 标签自动注入依赖
- **作用域管理** - 单例(singleton)和原型(prototype)作用域
- **条件创建** - 支持条件判断创建 bean
- **后置处理** - 支持 BeanPostProcessor 扩展
- **方法注入** - 支持通过 Invoke 调用函数并自动注入依赖

## 快速开始

```go
package main

import (
    "fmt"
    "go-boot/core"
)

func main() {
    container := core.New()

    // 注册 bean
    container.Register("message", core.Bean("Hello World"))
    container.Register("hello", core.Bean(&HelloService{}))

    // 获取 bean (单例,会缓存)
    svc, _ := container.Get("hello")
    fmt.Println(svc.(*HelloService).Say())
}

type HelloService struct {
    Message string `inject:"message"`
}

func (s *HelloService) Say() string {
    return s.Message
}
```

## 安装

```bash
go get go-boot/core
```

## 使用指南

### 创建容器

```go
container := core.New()
```

### 注册 Bean

**实例注册** - 直接注册实例：

```go
container.Register("service", core.Bean(&MyService{}))
```

**工厂函数** - 使用函数创建实例：

```go
container.Register("config", core.Factory(func(c core.Container) (interface{}, error) {
    return loadConfig(), nil
}, reflect.TypeOf((*Config)(nil)).Elem()))
```

### 设置作用域

**单例** (默认) - 容器只创建一次：

```go
container.Register("service", core.Bean(&Service{}), core.Singleton())
```

**原型** - 每次获取创建新实例：

```go
container.Register("service", core.Bean(&Service{}), core.Prototype())
```

### 字段注入

**使用 inject 标签**：

```go
type UserService struct {
    DB *Database `inject:"database"`
}
```

**使用 Ref 选项**：

```go
container.Register("userService", core.Bean(&UserService{}), 
    core.Ref("database"))
```

### 初始化函数

```go
container.Register("service", core.Bean(&Service{}),
    core.Init(func(s interface{}) error {
        return s.(*Service).Connect()
    }))
```

### 条件创建

```go
container.Register("devLogger", core.Bean(&DevLogger{}),
    core.Condition(func(c core.Container) bool {
        return os.Getenv("ENV") == "development"
    }))
```

### 自动注入到结构体

```go
type Handler struct {
    Service *MyService `inject:"myService"`
    Logger  Logger     `inject:"logger"`
}

var h Handler
container.Inject(&h)
```

### 调用函数并注入依赖

```go
result, err := container.Invoke(func(s *UserService, l Logger) error {
    return s.DoSomething(l)
})
```

## API 参考

### Container 接口

| 方法                                                        | 说明           |
|-----------------------------------------------------------|--------------|
| `Register(beanID string, builder ...BuilderOption) error` | 注册 bean      |
| `Get(beanID string) (interface{}, error)`                 | 获取 bean 实例   |
| `Inject(target interface{}) error`                        | 注入依赖到结构体     |
| `Invoke(fn interface{}) ([]interface{}, error)`           | 调用函数并注入依赖    |
| `GetAll(beanType interface{}) ([]interface{}, error)`     | 获取接口类型的所有实现  |
| `Has(beanID string) bool`                                 | 检查 bean 是否存在 |
| `Remove(beanID string) error`                             | 移除 bean      |

### BuilderOption 函数

| 函数                          | 说明         |
|-----------------------------|------------|
| `Bean(bean interface{})`    | 注册 bean 实例 |
| `Factory(fn, concreteType)` | 使用工厂函数     |
| `Type(t reflect.Type)`      | 设置类型       |
| `Singleton()`               | 单例作用域      |
| `Prototype()`               | 原型作用域      |
| `Field(name, value)`        | 设置字段值      |
| `Ref(beanID)`               | 引用其他 bean  |
| `DependsOn(beanIDs...)`     | 设置依赖顺序     |
| `Init(fn)`                  | 初始化函数      |
| `Condition(fn)`             | 条件创建       |
| `PostProcessor(p...)`       | 后置处理器      |

## 模块结构

```
core/
├── container.go    # 容器接口和实现
├── builder.go   # Bean 构建器选项
├── component.go # 组件标签定义
├── scanner.go  # 包扫描器
└── README.md  # 本文档
```

## 完整示例

查看 `examples/core/main.go` 获取完整示例: