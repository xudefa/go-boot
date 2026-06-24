# go-boot 核心框架说明

## 概述

go-boot 是一个参考 Spring Boot 设计哲学、保留 Go 语言特性的工程化应用框架。核心使用 Go 标准库实现（零外部依赖），提供依赖注入（IoC）、面向切面编程（AOP）、数据访问抽象、缓存抽象、日志抽象、配置管理、健康检查、指标收集、服务注册发现、熔断器、负载均衡、数据验证、定时任务调度等开箱即用的能力。

## 核心模块

| 模块 | 说明 |
|------|------|
| core | IoC 容器，依赖注入核心 |
| aop | 面向切面编程框架 |
| boot | 应用启动器和自动配置 |
| context | 应用上下文聚合 |
| environment | 环境配置管理 |
| condition | 条件装配机制 |
| event | 事件驱动系统 |
| life | 生命周期管理 |
| data | 数据访问抽象 |
| cache | 缓存抽象接口 |
| config | 配置加载与管理 |
| log | 日志抽象接口 |
| net | HTTP 服务器/客户端接口 |
| health | 健康检查指标 |
| metrics | 指标收集器 |
| actuator | 运维端点 |
| schedule | 定时任务调度 |
| center | 服务注册与发现 |
| circuit | 熔断器 |
| loadbalancer | 负载均衡器 |
| validation | 数据验证 |
| security | 安全框架 |
| exception | 异常处理 |
| refresh | 配置热刷新 |

## 使用方式

```bash
go get github.com/xudefa/go-boot@latest
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/core"
)

type MyService struct {
    Message string `inject:"message"`
}

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    app.Container().Register("message", core.Bean("Hello from go-boot!"))
    app.Container().Register("myService", core.Bean(&MyService{}))

    app.Start()

    svc := app.Container().Get("myService").(*MyService)
    fmt.Println(svc.Message)

    app.WaitForSignal()
}
```

## 文档

详细文档请参考 [docs/](docs/) 目录下的各模块文档。