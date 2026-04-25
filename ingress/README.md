# go-boot-ingress

HTTP 服务器统一入口模块，提供多种 HTTP 框架的抽象接口和集成支持。

## 功能

- **统一接口** - 定义标准 HTTP Server 接口，支持不同框架
- **Gin 集成** - 基于 Gin 框架的 HTTP 服务器实现
- **Hertz 集成** - 基于 CloudWeGo Hertz 框架的 HTTP 服务器实现
- **中间件支持** - 统一的中间件注册和管理
- **依赖注入** - 与 go-boot 容器系统无缝集成

## 使用方法

### 基本使用

```go
import (
    "github.com/xudefa/go-boot/ingress"
    "github.com/xudefa/go-boot/ingress/gin"
)
```

### 启动 Gin 服务器

```go
func main() {
    gin.Starter()
}
```

### 启动 Hertz 服务器

```go
func main() {
    hertz.Starter()
}
```

### 手动注册服务器

```go
import (
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot/ingress/gin"
    "github.com/xudefa/go-boot/log"
)

func main() {
    container := core.New()
    logger := log.DefaultLogger()
    
    config := &gin.GinConfig{
        HostPorts: ":8080",
    }
    
    server, _ := gin.Register(container, config, logger)
    
    server.Use(func(c *gin.Context) {
        // 自定义中间件
        c.Next()
    })
    
    server.Run()
}
```

### 注册路由处理器

```go
server.Register(func(container core.Container) error {
    router := container.Get("gin").(*gin.Engine)
    router.GET("/hello", func(c *gin.Context) {
        c.String(200, "Hello World")
    })
    return nil
})
```

## 配置

### Gin 配置 (server.gin)

```yaml
server:
  gin:
    hostPorts: ":8080"
    readTimeout: "10s"
    writeTimeout: "10s"
    idleTimeout: "60s"
    mode: "debug"
```

### Hertz 配置 (server.hertz)

```yaml
server:
  hertz:
    hostPorts: "0.0.0.0:8080"
    readTimeout: "10s"
    writeTimeout: "10s"
    idleTimeout: "60s"
    keepAlive: true
    bathPath: "/"
```

## 结构

- `ingress.go` - Server 接口定义
- `gin/gin.go` - Gin 服务器实现
- `gin/starter.go` - Gin 启动器
- `hertz/hertz.go` - Hertz 服务器实现
- `hertz/starter.go` - Hertz 启动器
