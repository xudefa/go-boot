# condition 包 — 条件判断（@Conditional 机制）

## 概述

`condition` 包提供条件判断机制，参考 Spring Boot 的 `@Conditional` 注解体系设计。用于在自动配置（AutoConfiguration）中控制 Bean 或配置是否生效，实现"按需加载"。

### 核心接口

```go
type Condition interface {
    Matches(ctx ConditionContext) bool  // 判断条件是否匹配
    String() string                     // 条件描述（用于日志和调试）
}

type ConditionContext interface {
    Environment() interface{ GetProperty(key string) (any, bool) }  // 环境配置
    Container() interface{ Has(id string) bool }                     // IoC 容器
    ClassLoader() interface{ HasClass(name string) bool }            // 类路径检查
    GetBean(beanID string) (any, bool)                               // 获取 Bean
    HasProperty(key string) bool                                     // 属性是否存在
    GetProperty(key string) (any, bool)                              // 获取属性
}
```

### 使用场景

```go
// 在自动配置中使用条件
boot.RegisterAutoConfig(
    &GinAutoConfiguration{},
    condition.OnProperty("gin.enabled", "true"),     // 属性条件
    condition.OnClass("github.com/gin-gonic/gin"),   // 类路径条件
    condition.All(
        condition.OnProperty("db.enabled", "true"),
        condition.OnBean("dataSource"),
    ),
)
```

---

## 内置条件

### OnProperty — 属性条件

当配置属性存在且匹配预期值时通过。参考 Spring 的 `@ConditionalOnProperty`：

```go
// 条件：属性存在且值不为空
cond := condition.OnProperty("gin.enabled")

// 条件：属性存在且等于指定值
cond := condition.OnProperty("gin.enabled", "true")

// 使用示例
boot.RegisterAutoConfig(&GinAutoConfig{}, condition.OnProperty("gin.enabled", "true"))
```

匹配逻辑：
1. 从环境获取属性值
2. 仅传 key 时：存在且为有效字符串（非空）则匹配
3. 传 key+value 时：存在且值等于预期值则匹配

### OnMissingProperty — 属性缺失条件

当配置属性不存在时通过：

```go
cond := condition.OnMissingProperty("gin.enabled")
// 属性不存在时返回 true，存在时返回 false
```

### OnBean — Bean 存在条件

当 IoC 容器中存在指定 ID 的 Bean 时通过。参考 Spring 的 `@ConditionalOnBean`：

```go
cond := condition.OnBean("redisClient")
// 容器中有 "redisClient" 时匹配
```

### OnMissingBean — Bean 缺失条件

当 IoC 容器中不存在指定 ID 的 Bean 时通过。参考 Spring 的 `@ConditionalOnMissingBean`：

```go
cond := condition.OnMissingBean("customCache")
// 容器中没有 "customCache" 时匹配
```

### OnProfile — Profile 条件

当当前激活的 Profile 匹配时通过。参考 Spring 的 `@Profile`：

```go
cond := condition.OnProfile("dev")

// 支持否定前缀
cond := condition.OnProfile("!dev") // 非 dev 环境时匹配

// 匹配逻辑：委托给 Environment.AcceptsProfile()
```

匹配流程：
1. 检查 Environment 是否实现了 `AcceptsProfile(profile string) bool` 接口
2. 是则委托给该方法（支持 ! 否定前缀）
3. 否则仅检查否定前缀（默认不匹配）

### OnClass — 类路径条件

当类路径中存在指定类时通过。参考 Spring 的 `@ConditionalOnClass`：

```go
// 扩展："类"可以是 Go 包的导入路径或任意模块名
cond := condition.OnClass("github.com/gin-gonic/gin")
```

### OnMissingClass — 类路径缺失条件

当类路径中不存在指定类时通过：

```go
cond := condition.OnMissingClass("github.com/xx/xx")
```

### OnPropertyPrefix — 配置前缀条件

当存在指定前缀的配置属性时通过：

```go
cond := condition.OnPropertyPrefix("server")
// 当任意配置源中存在以 "server" 为前缀的键时匹配
```

实现机制：通过 `propertySourceLister` 接口遍历所有配置源，使用 `keyLister` 接口获取键列表，检查是否存在指定前缀。

---

## 组合条件

| 函数 | 逻辑 | 说明 |
|------|------|------|
| `All(conditions...)` | AND | 所有子条件都匹配时通过（短路） |
| `Any(conditions...)` | OR | 任一子条件匹配时通过（短路） |
| `Not(condition)` | NOT | 对子条件结果取反 |

### All — 逻辑与

参考 Spring 的 `@Conditional` 多个条件同时生效：

```go
cond := condition.All(
    condition.OnProperty("db.enabled", "true"),
    condition.OnBean("dataSource"),
    condition.OnProfile("!test"),
)
// 三个条件全部满足时才匹配，第一个不满足的条件会短路后续判断
```

执行逻辑：
1. 按顺序遍历所有子条件
2. 遇到第一个不匹配的立即返回 `false`（短路）
3. 全部匹配返回 `true`

### Any — 逻辑或

```go
cond := condition.Any(
    condition.OnProfile("dev"),
    condition.OnProfile("test"),
)
// 任一个条件满足即匹配，第一个满足的条件会短路后续判断
```

执行逻辑：
1. 按顺序遍历所有子条件
2. 遇到第一个匹配的立即返回 `true`（短路）
3. 全部不匹配返回 `false`

### Not — 逻辑非

```go
cond := condition.Not(condition.OnProfile("prod"))
// 非 prod 环境时匹配
```

### 嵌套组合

条件可以任意嵌套组合：

```go
cond := condition.All(
    condition.OnProperty("app.enabled", "true"),
    condition.Any(
        condition.OnProfile("dev"),
        condition.OnProfile("staging"),  
    ),
    condition.Not(
        condition.OnMissingBean("customHandler"),
    ),
)
```

---

## Custom — 自定义条件

```go
// 方式一：内联函数
cond := condition.Custom("hasCustomDB", func(ctx condition.ConditionContext) bool {
    val, ok := ctx.GetProperty("db.type")
    return ok && val == "postgres"
})

// 方式二：实现 Condition 接口
type MyCondition struct{}

func (m *MyCondition) Matches(ctx condition.ConditionContext) bool {
    bean, ok := ctx.GetBean("userService")
    if !ok {
        return false
    }
    return bean != nil
}

func (m *MyCondition) String() string {
    return "MyCondition"
}

// 使用
condition.All(
    condition.OnProperty("my.feature.enabled", "true"),
    &MyCondition{},
)
```

---

## 在启动器中使用条件

`boot.Starter` 接口通过 `GetCondition()` 方法支持条件控制：

```go
type MyStarter struct{}

func (s *MyStarter) Name() string                     { return "my-starter" }
func (s *MyStarter) Dependencies() []string           { return nil }
func (s *MyStarter) Configure(ctx boot.ApplicationContext) error { return nil }
func (s *MyStarter) Start(ctx boot.ApplicationContext) error     { return nil }
func (s *MyStarter) Stop(ctx boot.ApplicationContext) error      { return nil }
func (s *MyStarter) GetCondition() condition.Condition {
    return condition.All(
        condition.OnProperty("my.enabled", "true"),
        condition.OnClass("github.com/some/lib"),
    )
}

boot.RegisterStarter(&MyStarter{})
```

---

## 在自动配置中使用条件

```go
func init() {
    boot.RegisterAutoConfig(
        &RedisAutoConfiguration{},
        condition.OnProperty("redis.enabled", "true"),
        condition.All(
            condition.OnProperty("redis.host"),
            condition.OnProperty("redis.port"),
        ),
    )
}
```

---

## 与 Spring Boot @Conditional 对照

| Spring Boot | go-boot | 说明 |
|-------------|---------|------|
| `@ConditionalOnProperty` | `condition.OnProperty()` | 配置属性条件 |
| `@ConditionalOnMissingProperty` | `condition.OnMissingProperty()` | 配置属性缺失 |
| `@ConditionalOnBean` | `condition.OnBean()` | Bean 存在条件 |
| `@ConditionalOnMissingBean` | `condition.OnMissingBean()` | Bean 缺失条件 |
| `@Profile` | `condition.OnProfile()` | Profile 条件 |
| `@ConditionalOnClass` | `condition.OnClass()` | 类路径条件 |
| `@ConditionalOnMissingClass` | `condition.OnMissingClass()` | 类路径缺失条件 |
| `@ConditionalOnExpression` | `condition.Custom()` | 自定义条件 |
| `@Conditional` (多个) | `condition.All()` / `condition.Any()` | 组合条件 |

---

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/condition"
)

// 自定义条件：仅当应用版本 >= 2.0 时生效
var versionCondition = condition.Custom("version>=2.0", func(ctx condition.ConditionContext) bool {
    val, ok := ctx.GetProperty("app.version")
    if !ok {
        return false
    }
    return val == "2.0" || val == "2.1" || val == "3.0"
})

func main() {
    // 条件评估示例
    fmt.Println(condition.OnProperty("gin.enabled", "true").String())
    // 输出: OnProperty(gin.enabled=true)

    fmt.Println(condition.OnMissingBean("redis").String())
    // 输出: OnMissingBean(redis)

    fmt.Println(condition.All(
        condition.OnProperty("db.enabled", "true"),
        condition.Any(
            condition.OnProfile("dev"),
            condition.OnProfile("staging"),
        ),
    ).String())
    // 输出: All(OnProperty(db.enabled=true), Any(OnProfile(dev), OnProfile(staging)))
}
```
