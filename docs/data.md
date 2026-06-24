# data 包 — 数据访问抽象层

## 概述

`data` 包提供数据库访问的抽象层，包含通用 CRUD 接口、事务操作接口、ORM 工厂注册机制以及数据库配置管理。核心设计思想是**面向接口编程**，允许不同的 ORM 实现（如 GORM、XORM）无缝替换而不影响业务代码。

主要组件：

- `Repository[T]` — 泛型 CRUD 接口
- `Transactor` / `Transaction` — 事务操作接口
- `DBFactory` — ORM 提供者注册与创建
- `DBConfig` / `DatabaseConfig` — 数据库连接配置
- `BaseModel` — 基础模型嵌入结构

---

## Repository[T] — 通用 CRUD 接口

`Repository` 是泛型数据访问接口，提供类型安全的 CRUD 操作。

```go
type Repository[T any] interface {
    Create(entity *T) error
    CreateBatch(entities []T) error
    Delete(id any) error
    DeleteByCondition(where any, args ...any) error
    Update(entity *T) error
    UpdateByCondition(where any, args ...any) (int64, error)
    FindByID(id any) (*T, error)
    FindOne(where any, args ...any) (*T, error)
    FindAll(where any, args ...any) ([]T, error)
    Count(where any, args ...any) (int64, error)
    Raw(sql string, args ...any) ([]T, error)
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Create` | 创建单条记录 |
| `CreateBatch` | 批量创建记录 |
| `Delete` | 根据主键删除 |
| `DeleteByCondition` | 根据条件删除 |
| `Update` | 更新单条记录 |
| `UpdateByCondition` | 根据条件更新，返回影响行数 |
| `FindByID` | 根据主键查询 |
| `FindOne` | 查询单条记录 |
| `FindAll` | 查询多条记录 |
| `Count` | 统计数量 |
| `Raw` | 执行原生 SQL |

### 使用示例

```go
type User struct {
    data.BaseModel
    Name  string
    Email string
}

// 获取 Repository 实例（由 ORM 实现提供）
var userRepo data.Repository[User]

// 创建
user := &User{Name: "张三", Email: "zhangsan@example.com"}
err := userRepo.Create(user)

// 查询
u, err := userRepo.FindByID(1)

// 条件查询
u, err := userRepo.FindOne("name = ?", "张三")

// 列表查询
users, err := userRepo.FindAll("age > ?", 18)

// 更新
u.Name = "李四"
err := userRepo.Update(u)

// 删除
err := userRepo.Delete(1)

// 统计
count, err := userRepo.Count("status = ?", "active")

// 原生 SQL
results, err := userRepo.Raw("SELECT * FROM users WHERE name LIKE ?", "%张%")
```

---

## RepositoryFactory — 仓库工厂

```go
type RepositoryFactory[T any] interface {
    NewRepository() Repository[T]
}
```

ORM 实现通过实现此接口来创建具体的 `Repository` 实例。

---

## Model 接口与 BaseModel

### Model 接口

实体实现此接口可自定义表名：

```go
type Model interface {
    TableName() string
}
```

### BaseModel 基础模型

提供通用字段，实体可直接嵌入：

```go
type BaseModel struct {
    ID        uint    `json:"id"`
    CreatedAt int64   `json:"created_at"`
    UpdatedAt int64   `json:"updated_at"`
    DeletedAt *int64  `json:"deleted_at"`
}

func (BaseModel) TableName() string {
    return ""
}
```

### 使用示例

```go
// 自定义表名
type Order struct {
    data.BaseModel
    OrderNo string
}

func (Order) TableName() string {
    return "t_order"
}

// 自动推导表名（User -> users）
type User struct {
    data.BaseModel
    Name string
}
```

---

## 工具函数

### GetTableName

获取实体对应的表名，优先使用 `Model.TableName()`，否则将类型名转换为蛇形命名：

```go
tableName := data.GetTableName[User]() // "users"
tableName := data.GetTableName[Order]() // "t_order"
```

### ToSnakeCase

将驼峰命名转换为蛇形命名：

```go
data.ToSnakeCase("UserName")     // "user_name"
data.ToSnakeCase("UserID")       // "user_id"
data.ToSnakeCase("HTMLParser")   // "html_parser"
data.ToSnakeCase("ID")           // "id"
data.ToSnakeCase("UserIDString") // "user_id_string"
```

---

## Transactor / Transaction — 事务操作接口

### Transactor 接口

```go
type Transactor interface {
    Query(ctx context.Context, query string, args ...any) (Rows, error)
    QueryRow(ctx context.Context, query string, args ...any) Row
    Exec(ctx context.Context, query string, args ...any) (Result, error)
    Begin(ctx context.Context) (Transaction, error)
    Stats() DBStats
    Close() error
}
```

### Transaction 接口

扩展 Transactor，增加提交和回滚能力：

```go
type Transaction interface {
    Transactor
    Commit() error
    Rollback() error
}
```

### Rows / Row / Result 接口

```go
type Rows interface {
    Next() bool
    Scan(dest ...any) error
    Close() error
    Err() error
}

type Row interface {
    Scan(dest ...any) error
}

type Result interface {
    LastInsertId() (int64, error)
    RowsAffected() (int64, error)
}
```

### DBStats

```go
type DBStats struct {
    MaxOpenConnections int
    OpenConnections    int
    InUse              int
    Idle               int
}
```

### 使用示例

```go
var tx data.Transaction
tx, err := transactor.Begin(ctx)

// 在事务中执行操作
result, err := tx.Exec(ctx, "UPDATE users SET name = ? WHERE id = ?", "新名称", 1)

if err != nil {
    tx.Rollback()
    return err
}

return tx.Commit()
```

---

## DBFactory — ORM 工厂

DBFactory 允许运行时注册和创建不同 ORM 的实现：

```go
type DBFactory struct {
    providers map[string]func(*DBConfig) (Transactor, error)
}

func NewDBFactory() *DBFactory

// 注册 ORM 提供者
func (f *DBFactory) Register(name string, fn func(*DBConfig) (Transactor, error))

// 创建 Transactor 实例
func (f *DBFactory) Create(cfg *DBConfig) (Transactor, error)
```

### 使用示例

```go
factory := data.NewDBFactory()

// 注册 GORM 提供者
factory.Register("gorm", func(cfg *data.DBConfig) (data.Transactor, error) {
    // 创建并返回 GORM 实现的 Transactor
    return gormTransactor, nil
})

// 创建实例
transactor, err := factory.Create(&data.DBConfig{
    ORM:    "gorm",
    Driver: "mysql",
    DSN:    "user:password@tcp(127.0.0.1:3306)/dbname",
})
```

---

## DatabaseConfig — 数据库配置

### 函数式选项

```go
type DatabaseOption func(*DatabaseConfig)

func WithHost(host string) DatabaseOption
func WithPort(port int) DatabaseOption
func WithDriver(driver string) DatabaseOption
func WithCharset(charset string) DatabaseOption
func WithMaxOpenConns(n int) DatabaseOption
func WithMaxIdleConns(n int) DatabaseOption
func WithConnMaxLifetime(d time.Duration) DatabaseOption
func WithConnMaxIdleTime(d time.Duration) DatabaseOption
func WithDebug(debug bool) DatabaseOption
```

### DatabaseConfig 结构体

```go
type DatabaseConfig struct {
    Dsn         string
    Driver      string
    Host        string
    Port        int
    Username    string
    Password    string
    Name        string
    Charset     string
    MaxOpen     int
    MaxIdle     int
    MaxLife     time.Duration
    MaxIdleTime time.Duration
    Debug       bool
}
```

### DSN() 自动生成

根据驱动类型自动生成连接字符串，支持 MySQL、PostgreSQL、Oracle。

### 使用示例

```go
cfg := data.NewDefaultDatabaseConfig(
    "root", "password", "mydb",
    data.WithHost("192.168.1.100"),
    data.WithPort(3306),
    data.WithMaxIdleConns(10),
    data.WithDebug(true),
)
dsn := cfg.DSN()
// 输出: "root:password@tcp(192.168.1.100:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
```