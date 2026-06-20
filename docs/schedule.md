# schedule 包 — 定时任务调度框架

## 概述

`schedule` 包提供零外部依赖的 Spring 风格定时任务调度框架，分为四层：

| 层次 | 组件 | 说明 |
|------|------|------|
| 第 1 层 | `CronExpression` | 6 字段 Cron 表达式解析，位图编码，`Next()` 算法 |
| 第 2 层 | `Task` 接口 | 任务抽象：名称、Cron 表达式、执行逻辑 |
| 第 3 层 | `Scheduler` | 基于最小堆的调度器，支持动态注册/注销、并发池控制、优雅关闭 |
| 第 4 层 | `@Scheduled` 注解 | 通过 `go/ast` 解析源码注解，结合 IoC 容器自动注册任务 |

自动配置 `ScheduleAutoConfiguration` 在 `schedule.enabled=true` 时自动启用，并与 `ScheduleStarter` 协同管理调度器生命周期。

---

## 第 1 层 — CronExpression：Cron 表达式解析

### CronExpression 结构体

```go
type CronExpression struct {
    second        uint64
    minute        uint64
    hour          uint64
    dayOfMonth    uint64
    month         uint64
    dayOfWeek     uint64
    hasDayOfMonth bool
    hasDayOfWeek  bool
}
```

每个时间字段使用 `uint64` 位图存储有效值。例如 `second` 的第 0 位为 1 表示第 0 秒可触发，第 30 位为 1 表示第 30 秒可触发。

### Parse — 解析 6 字段 Cron 表达式

```go
func Parse(expr string) (*CronExpression, error)
```

格式：`秒 分 时 日 月 周`

| 字段 | 取值范围 | 支持名称 |
|------|----------|----------|
| 秒 | 0-59 | - |
| 分 | 0-59 | - |
| 时 | 0-23 | - |
| 日 | 1-31，`?` 表示不指定 | - |
| 月 | 1-12 或 `JAN`-`DEC` | `jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec` |
| 周 | 0-6 或 `SUN`-`SAT`，`?` 表示不指定 | `sun,mon,tue,wed,thu,fri,sat` |

日和周均支持 `?` 表示不指定。当两者同时指定时，满足任一条件即可触发。

#### 解析示例

```go
// 每天 12:00:00 执行
expr, _ := schedule.Parse("0 0 12 * * ?")

// 每小时的第 5 分 30 秒执行
expr, _ := schedule.Parse("30 5 * * * ?")

// 工作日（周一至周五）每天 9:30:00 执行
expr, _ := schedule.Parse("0 30 9 * * MON-FRI")

// 每月的第 1 天和第 15 天执行
expr, _ := schedule.Parse("0 0 0 1,15 * ?")
```

### Next — 计算下次触发时间

```go
func (ce *CronExpression) Next(after time.Time) time.Time
```

从 `after + 1 秒` 开始向前搜索，最大搜索范围为 4 年。如果 4 年内无法找到匹配时间，返回零值 `time.Time{}`。

算法流程：
1. 从 `after` 的下一秒开始
2. 推进月份直到匹配 `month` 字段
3. 检查日/周条件，不匹配则推进到次日
4. 检查小时、分钟、秒条件，不匹配则逐级递增
5. 所有字段匹配后返回该时间点

```go
expr, _ := schedule.Parse("0 0 12 * * ?")
next := expr.Next(time.Now())
fmt.Println("下次执行时间:", next)
```

### Cron 表达式语法

#### 通配符 `*`

匹配字段所有有效值。例如 `*` 在分钟字段表示每分钟。

#### 不指定 `?`

仅用于日和周字段，表示不指定该条件。例如 `0 0 12 ? * MON-FRI` 表示工作日每天中午 12 点。

#### 固定值

直接写数字。例如 `5` 在秒字段表示第 5 秒。

#### 列表 `,`

逗号分隔多个值。例如 `1,3,5` 在小时字段表示第 1、3、5 小时。

#### 范围 `-`

连接号表示连续范围。例如 `1-5` 在分钟字段表示第 1 到第 5 分钟。`MON-FRI` 表示周一至周五。

#### 步进 `/`

格式 `<起始>/<步长>`，起始可以是 `*`、范围或固定值。例如 `*/5` 在秒字段表示每 5 秒，`1-10/2` 表示从 1 到 10 每 2 个值（1,3,5,7,9）。

---

## 第 2 层 — Task：任务接口

```go
type Task interface {
    Name() string                       // 任务名称，用于注册和查找
    Cron() string                       // Cron 表达式（6 字段 Spring 风格）
    Execute(ctx context.Context) error  // 任务执行逻辑
}
```

### NewTask — 创建基于函数的任务

```go
func NewTask(name, cronExpr string, fn func(context.Context) error) Task
```

```go
task := schedule.NewTask(
    "cacheCleanup",
    "0 0 3 * * ?", // 每天凌晨 3 点
    func(ctx context.Context) error {
        return cache.Clear()
    },
)
```

---

## 第 3 层 — Scheduler：基于最小堆的调度器

### Scheduler 接口

```go
type Scheduler interface {
    Start(ctx context.Context) error              // 启动调度器
    Shutdown(ctx context.Context) error           // 优雅关闭
    Register(task Task) error                     // 注册任务
    Unregister(name string) bool                  // 注销任务
    IsRunning() bool                              // 是否运行中
    RegisteredTasks() []Task                      // 所有已注册任务
}
```

### 创建调度器

```go
func NewScheduler(opts ...SchedulerOption) Scheduler
```

#### SchedulerOption

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithPoolSize(n)` | 任务执行并发池大小 | 10 |
| `WithErrorHandler(h)` | 执行错误处理函数 | 空函数 |

```go
scheduler := schedule.NewScheduler(
    schedule.WithPoolSize(5),
    schedule.WithErrorHandler(func(task schedule.Task, err error) {
        log.Printf("任务 %s 执行失败: %v", task.Name(), err)
    }),
)
```

### Register — 注册任务

解析任务的 Cron 表达式并计算下次执行时间，加入最小堆。如果调度器已运行，注册后立即唤醒主循环重新计算。

```go
err := scheduler.Register(schedule.NewTask(
    "cleanup",
    "0 */5 * * * ?",
    func(ctx context.Context) error {
        return cleanup(ctx)
    },
))
```

任务名必须唯一，重复注册返回错误。

### Start — 启动调度器

创建内部上下文并启动主循环 goroutine。重复启动返回错误。

```go
if err := scheduler.Start(ctx); err != nil {
    log.Fatal(err)
}
```

### Shutdown — 优雅关闭

1. 标记停止，阻止新任务派发
2. 取消内部上下文
3. 等待所有正在执行的任务完成（受外部 `ctx` 超时控制）

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := scheduler.Shutdown(shutdownCtx); err != nil {
    log.Printf("调度器关闭超时: %v", err)
}
```

### 调度器内部原理

调度器主循环 `loop()` 运行在独立 goroutine 中：

1. 查看堆顶任务（最近需执行的任务）
2. 若堆空，等待唤醒信号
3. 若未到期，等待定时器或唤醒信号
4. 到期后，从堆顶取出任务并执行
5. 执行重新计算下次执行时间，将任务放回堆中
6. 若永不到期，从堆和映射中移除

并发控制通过信号量 `sem`（channel）限制同时执行的任务数，`WaitGroup` 追踪所有执行中的任务确保优雅关闭。

### 完整使用示例

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/xudefa/go-boot/schedule"
)

func main() {
    scheduler := schedule.NewScheduler(
        schedule.WithPoolSize(3),
        schedule.WithErrorHandler(func(task schedule.Task, err error) {
            log.Printf("任务出错: %s - %v", task.Name(), err)
        }),
    )

    scheduler.Register(schedule.NewTask("task1", "0 */1 * * * ?",
        func(ctx context.Context) error {
            log.Println("task1 执行中...")
            return nil
        },
    ))

    scheduler.Register(schedule.NewTask("task2", "0/30 * * * * ?",
        func(ctx context.Context) error {
            log.Println("task2 执行中...")
            return nil
        },
    ))

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := scheduler.Start(ctx); err != nil {
        log.Fatal(err)
    }

    time.Sleep(5 * time.Minute)

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()
    scheduler.Shutdown(shutdownCtx)
}
```

---

## 第 4 层 — @Scheduled 注解扫描

通过 `go/ast` 解析 Go 源码中的注释标签，结合 IoC 容器自动注册定时任务。

### 注解语法

在结构体的方法上使用 Go 注释作为注解：

```go
// @Service("myService")
type ReportService struct{}

// @Scheduled(cron="0 0 8 * * ?")
func (s *ReportService) GenerateDailyReport(ctx context.Context) error {
    // 每天上午 8 点执行
    return nil
}

// @Scheduled(cron="0 0 0 1 * ?")
func (s *ReportService) GenerateMonthlyReport(ctx context.Context) error {
    // 每月 1 号执行
    return nil
}
```

### ScanScheduledTasks — 扫描并构建任务列表

```go
func ScanScheduledTasks(container core.Container, scanDir string) ([]Task, error)
```

扫描流程：
1. 递归遍历 `scanDir` 下所有 Go 源文件（排除 `_test.go`）
2. `go/parser` 解析 AST，收集结构体名称
3. 匹配接收者方法上的 `@Scheduled(cron="...")` 注释
4. 从 IoC 容器中按结构体名（首字母小写）查找 Bean
5. 通过反射包装为 `Task`，方法签名必须为 `func(context.Context) error`

```go
container := core.New()
container.Register("reportService", core.Bean(&ReportService{}))

tasks, err := schedule.ScanScheduledTasks(container, ".")
if err != nil {
    log.Fatal(err)
}

scheduler := schedule.NewScheduler()
for _, t := range tasks {
    scheduler.Register(t)
}
```

### 注解解析细节

- 支持单引号和双引号包裹 cron 值：`cron="0 0 * * * ?"` 或 `cron='0 0 * * * ?'`
- 支持额外参数（被忽略）：`@Scheduled(cron="0 * * * * ?", zone="UTC")`
- 方法查找支持指针和值接收者
- Bean ID 默认为结构体类型名首字母小写（`MyService` → `myService`）

---

## 自动配置

`ScheduleAutoConfiguration` 在 `init()` 中注册，条件为 `schedule.enabled=true`。

### 配置项

| 配置键 | 说明 | 默认值 |
|--------|------|--------|
| `schedule.enabled` | 是否启用定时任务 | false |
| `schedule.pool-size` | 并发池大小 | 10 |
| `schedule.scan-annotations` | 是否扫描 @Scheduled 注解 | true |

### 自动配置流程

1. 从 `Environment` 读取配置
2. 创建 `Scheduler` Bean（Bean ID: `scheduleScheduler`）并注册到 IoC 容器
3. 若 `scan-annotations=true`，扫描当前目录源码，自动注册 @Scheduled 任务
4. `ScheduleStarter` 在应用启动时调用 `Scheduler.Start()`，关闭时调用 `Scheduler.Shutdown()`

### 完整自动配置示例

```go
// main.go
package main

import (
    _ "github.com/xudefa/go-boot/schedule"
    "github.com/xudefa/go-boot/boot"
)

type MyTask struct{}

// @Scheduled(cron="0/30 * * * * ?")
func (m *MyTask) Run(ctx context.Context) error {
    log.Println("定时任务执行")
    return nil
}

func main() {
    app := boot.NewApplication(
        boot.WithProperty("schedule.enabled", "true"),
        boot.WithProperty("schedule.pool-size", "5"),
    )
    app.Container().Register("myTask", core.Bean(&MyTask{}), core.Singleton())
    app.Run()
}
```

---

## 完整 API 参考

| 函数/类型 | 说明 |
|-----------|------|
| `Parse(expr string)` | 解析 6 字段 Cron 表达式 |
| `CronExpression.Next(time.Time)` | 计算下次触发时间 |
| `Task` | 任务接口（Name/Cron/Execute） |
| `NewTask(name, cron, fn)` | 创建基于函数的任务 |
| `Scheduler` | 调度器接口 |
| `NewScheduler(opts...)` | 创建调度器 |
| `WithPoolSize(n)` | 设置并发池大小 |
| `WithErrorHandler(h)` | 设置错误处理函数 |
| `ScanScheduledTasks(container, scanDir)` | 扫描 @Scheduled 注解 |
| `ScheduleAutoConfiguration` | 自动配置 |
| `ScheduleStarter` | 调度器启动器 |
