# go-boot-aop

面向切面编程(AOP)模块,提供切面、通知、代理等功能。

## 功能

- **切面(Aspect)** - 定义切点和通知
- **通知(Advice)** - 前置、后置、环绕通知
- **切点(Pointcut)** - 匹配方法规则
- **连接点(JoinPoint)** - 方法调用信息
- **代理(Proxy)** - 动态代理生成

## 使用方法

### 定义切面

```go
import "go-boot/aop"

type LoggingAspect struct {
    aop.Aspect
}

func (a *LoggingAspect) Pointcut() aop.Pointcut {
    return aop.NewMethodPointcut("*.Say")
}

func (a *LoggingAspect) Before(jp aop.JoinPoint) interface{} {
    fmt.Println("Before:", jp.Method().Name)
    return nil
}
```

### 注册切面

```go
func main() {
    weaver := aop.NewWeaver()
    weaver.RegisterAspect(&LoggingAspect{})
    weaver.Weave()
}
```

## 结构

- `aspect.go` - 切面定义
- `advice.go` - 通知类型
- `pointcut.go` - 切点匹配
- `joinpoint.go` - 连接点
- `proxy.go` - 代理生成
- `weaver.go` - AOP 织入器