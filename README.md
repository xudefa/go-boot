# go-boot

Go 语言轻量级应用开发框架，提供依赖注入(IoC)、面向切面编程(AOP)和数据访问层支持。

## 特性

- **IoC 容器** - 轻量级依赖注入容器，支持组件扫描和自动装配
- **AOP** - 完整的切面编程支持，包括切点、通知、连接点等
- **数据访问层** - 基于 GORM/XORM 的 Repository 模式实现

## 快速开始

### 安装

```bash
go get go-boot/core
```

### 基本使用

```go
package main

import (
	"fmt"
	"go-boot/core"
)

type HelloService struct {
	Message string `inject:"message"`
}

func (s *HelloService) Say() string {
	return s.Message
}

func main() {
	container := core.New()

	container.Register("message", core.Bean("Hello World"))
	container.Register("hello", core.Bean(&HelloService{}))

	svc := container.Get("hello").(*HelloService)
	fmt.Println(svc.Say())
}
```

## 模块

| 模块       | 路径                              | 说明                                   |
|----------|---------------------------------|--------------------------------------|
| core     | [core/](core/README.md)         | 核心 IoC 容器，提供依赖注入和组件管理                |
| aop      | [aop/](aop/README.md)           | 面向切面编程，提供切面、通知、代理等功能                 |
| data     | [data/](data/README.md)         | 数据访问层，提供 GORM/XORM 集成和 Repository 模式 |
| log      | [log/](log/README.md)           | 日志模块，提供统一日志接口                        |
| cache    | [cache/](cache/README.md)       | 缓存模块，支持多种缓存实现                        |
| ingress  | [ingress/](ingress/README.md)   | HTTP 入口模块                            |
| security | [security/](security/README.md) | 安全模块，提供认证授权功能                        |

## 架构

```
go-boot/
├── core/           # 依赖注入容器
│   ├── container   # 容器核心实现
│   ├── builder     # Bean 构建器
│   ├── component   # 组件标签
│   └── scanner     # 包扫描器
├── aop/            # 面向切面编程
│   ├── aspect      # 切面定义
│   ├── advice      # 通知类型
│   ├── pointcut    # 切点匹配
│   ├── joinpoint   # 连接点
│   ├── proxy       # 动态代理
│   └── weaver      # AOP 织入器
├── data/           # 数据访问层
│   ├── gorm/       # GORM 实现
│   └── xorm/       # XORM 实现
└── examples/       # 示例代码
```

## 示例

查看 [examples/](examples/) 目录获取完整示例：

- [examples/core/](examples/core/) - 核心 IoC 容器使用示例
- [examples/aop/](examples/aop/) - AOP 使用示例
- [examples/data/](examples/data/) - 数据访问层使用示例

## 开发

### 构建

```bash
make build
# 或
go build ./...
```

### 测试

```bash
go test ./...
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```