# go-boot 单模块迁移指南

## 概述

go-boot 已从多模块架构（Go workspace 管理 18 个模块）重构为单模块架构。核心框架仅保留 19 个子包，零外部依赖。所有集成模块已拆分为独立的 GitHub 仓库。

## 变更内容

### 删除的内容

- 所有集成模块子目录（`gin/`, `gorm/`, `redis/`, `viper/`, `zap/`, `zerolog/`, `etcd/`, `nacos/`, `consul/`, `grpc/`, `hertz/`, `fasthttp/`, `websocket/`, `jwt/`, `casbin/`, `email/`, `swagger/`, `xorm/`, `kitex/`）
- `go.work` 和 `go.work.sum` 文件
- 多模块管理脚本（`scripts/fix-deps.sh`, `scripts/tag-manager.sh`）
- `docs/modules/` 目录下的集成模块文档
- `docs/multi-module-management.md`

### 保留的内容

- 核心框架 19 个子包（`core/`, `aop/`, `boot/`, `context/`, `environment/`, `condition/`, `event/`, `life/`, `data/`, `cache/`, `config/`, `log/`, `net/`, `health/`, `metrics/`, `tracing/`, `actuator/`, `schedule/`, `center/`）
- `examples/` 目录（仅保留核心模块示例）
- `docs/` 目录下的核心模块文档

## 迁移步骤

### 1. 更新导入路径

**之前（多模块）：**
```go
import (
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot/gin"
    "github.com/xudefa/go-boot/gorm"
    "github.com/xudefa/go-boot/redis"
)
```

**之后（单模块）：**
```go
import (
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot-gin"
    "github.com/xudefa/go-boot-gorm"
    "github.com/xudefa/go-boot-redis"
)
```

### 2. 更新 go.mod

**之前：**
```go
module myapp

go 1.25.7

require (
    github.com/xudefa/go-boot v1.0.0
)
```

**之后：**
```go
module myapp

go 1.25.7

require (
    github.com/xudefa/go-boot v2.0.0
    github.com/xudefa/go-boot-gin v1.0.0
    github.com/xudefa/go-boot-gorm v1.0.0
    github.com/xudefa/go-boot-redis v1.0.0
)
```

### 3. 删除 go.work

如果你的项目使用了 `go.work` 文件，请删除它：

```bash
rm go.work go.work.sum
```

### 4. 更新代码

核心框架 API 保持不变，仅需更新集成模块的导入路径：

```go
// 核心框架用法不变
app, _ := boot.NewApplication(boot.WithAppName("my-app"))
app.Container().Register("service", core.Bean(&MyService{}))
app.Start()

// 集成模块使用新导入
import "github.com/xudefa/go-boot-gin"

server := gin.New(
    gin.WithContainer(app.Container()),
    gin.WithHost(":8080"),
)
```

## 集成模块仓库列表

| 模块 | 仓库 |
|------|------|
| Gin | https://github.com/xudefa/go-boot-gin |
| GORM | https://github.com/xudefa/go-boot-gorm |
| Redis | https://github.com/xudefa/go-boot-redis |
| Viper | https://github.com/xudefa/go-boot-viper |
| Zap | https://github.com/xudefa/go-boot-zap |
| Zerolog | https://github.com/xudefa/go-boot-zerolog |
| Etcd | https://github.com/xudefa/go-boot-etcd |
| Nacos | https://github.com/xudefa/go-boot-nacos |
| Consul | https://github.com/xudefa/go-boot-consul |
| gRPC | https://github.com/xudefa/go-boot-grpc |
| Hertz | https://github.com/xudefa/go-boot-hertz |
| Fasthttp | https://github.com/xudefa/go-boot-fasthttp |
| WebSocket | https://github.com/xudefa/go-boot-websocket |
| JWT | https://github.com/xudefa/go-boot-jwt |
| Casbin | https://github.com/xudefa/go-boot-casbin |
| Email | https://github.com/xudefa/go-boot-email |
| Swagger | https://github.com/xudefa/go-boot-swagger |
| XORM | https://github.com/xudefa/go-boot-xorm |
| Kitex | https://github.com/xudefa/go-boot-kitex |
| OpenTelemetry | https://github.com/xudefa/go-boot-opentelemetry |

## 常见问题

### Q: 核心框架 API 有变化吗？

A: 没有变化。核心框架的 API 完全保持不变，仅需更新集成模块的导入路径。

### Q: 我可以只使用核心框架吗？

A: 可以。核心框架零外部依赖，仅使用 Go 标准库。如果你只需要 IoC 容器、AOP、事件系统等核心功能，无需引入任何集成模块。

### Q: 集成模块的版本如何管理？

A: 每个集成模块有独立的版本控制。你可以根据需要单独升级某个集成模块，而不影响核心框架或其他集成模块。

### Q: 之前的 go.work 方式还能用吗？

A: 不能。go-boot 已不再使用 Go workspace 管理多模块。集成模块以独立仓库形式发布，通过 `go get` 安装。