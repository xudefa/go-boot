# net 包 — HTTP 服务器/客户端抽象层

## 概述

`net` 包提供 HTTP 服务器和客户端的统一接口抽象，支持不同的框架实现（如 Gin、Hertz 等）无缝替换。内置默认客户端实现 `NetClient`，支持连接池、超时、认证、中间件链等特性。

主要组件：

- `Server` — HTTP 服务器抽象接口
- `HttpClient` — RESTful 客户端接口
- `NetClient` — 默认 HTTP 客户端实现
- `MiddlewareFunc` / `HandlerContext` — 服务端中间件
- `ClientMiddlewareFunc` — 客户端中间件

---

## Server — HTTP 服务器接口

```go
type Server interface {
    Start() error
    Use(m any)
    Stop(ctx context.Context) error
    SetConfig(opts ...ServerOption)
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Start` | 启动 HTTP 服务器并开始监听请求 |
| `Use` | 注册一个中间件 |
| `Stop` | 优雅停止服务器，等待正在处理的请求完成 |
| `SetConfig` | 设置服务器配置参数 |

### ServerConfig

```go
type ServerConfig struct {
    Host         string
    Mode         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}
```

### ServerOption

```go
type ServerOption func(*ServerConfig)

func WithHost(host string) ServerOption
func WithMode(mode string) ServerOption
func WithReadTimeout(timeout time.Duration) ServerOption
func WithWriteTimeout(timeout time.Duration) ServerOption
func WithIdleTimeout(timeout time.Duration) ServerOption
```

### 使用示例

```go
// 创建 Gin 服务器（由 gin 集成模块提供）
server := gin.New(
    gin.WithContainer(container),
    gin.WithHost(":8080"),
    gin.WithMode("release"),
)

// 注册中间件
server.Use(middleware.Logger(), middleware.Recovery())

// 启动
if err := server.Start(); err != nil {
    log.Fatal(err)
}

// 优雅停止
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Stop(ctx)
```

---

## HandlerContext — 请求处理上下文

```go
type HandlerContext interface {
    RequestMethod() string
    RequestURI() string
    Header(key string) string
    SetStatusCode(code int)
    SetHeader(key, value string)
    AbortWithStatus(code int)
    Next()
    IsAborted() bool
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `RequestMethod` | 获取请求方法（GET、POST 等） |
| `RequestURI` | 获取请求 URI |
| `Header` | 获取请求头 |
| `SetStatusCode` | 设置响应状态码 |
| `SetHeader` | 设置响应头 |
| `AbortWithStatus` | 中止请求处理并设置状态码 |
| `Next` | 调用下一个中间件 |
| `IsAborted` | 检查请求是否已被中止 |

---

## MiddlewareFunc — 中间件函数

```go
type MiddlewareFunc func(HandlerContext)
```

### 使用示例

```go
func LoggerMiddleware() net.MiddlewareFunc {
    return func(ctx net.HandlerContext) {
        start := time.Now()
        ctx.Next()
        duration := time.Since(start)
        fmt.Printf("%s %s %d %v\n",
            ctx.RequestMethod(), ctx.RequestURI(), 200, duration)
    }
}

func AuthMiddleware() net.MiddlewareFunc {
    return func(ctx net.HandlerContext) {
        token := ctx.Header("Authorization")
        if token == "" {
            ctx.AbortWithStatus(401)
            return
        }
        ctx.Next()
    }
}
```

---

## HttpClient — HTTP 客户端接口

```go
type HttpClient interface {
    Get(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)
    Head(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)
    Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)
    Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)
    Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)
    Delete(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)
    Options(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)
    Do(ctx context.Context, req any) (*HttpResponse, error)
    Close() error
}
```

---

## HttpRequest / HttpResponse

### HttpRequest

```go
type HttpRequest struct {
    Header      http.Header
    Query       url.Values
    Timeout     time.Duration
    AuthToken   string
    ContentType string
    BasicAuth   BasicAuth
}
```

### HttpResponse

```go
type HttpResponse struct {
    StatusCode int
    Header     http.Header
    Body       []byte
}

// 判断响应类型
func (r *HttpResponse) IsSuccess() bool       // 2xx
func (r *HttpResponse) IsRedirect() bool      // 3xx
func (r *HttpResponse) IsClientError() bool   // 4xx
func (r *HttpResponse) IsServerError() bool   // 5xx

// 解析响应
func (r *HttpResponse) Bind(v any) error       // JSON 反序列化
func (r *HttpResponse) String() string         // 获取原始字符串
```

---

## RequestOption — 请求选项

```go
type RequestOption func(*HttpRequest)

func WithHeader(key, value string) RequestOption              // 设置请求头
func WithQuery(key, value string) RequestOption               // 设置查询参数
func WithTimeout(timeout time.Duration) RequestOption          // 设置超时
func WithAuthToken(token string) RequestOption                 // 设置 Bearer Token
func WithContentType(contentType string) RequestOption         // 设置 Content-Type
func WithBasicAuth(username, password string) RequestOption    // 设置基本认证
```

---

## ClientOption — 客户端配置

```go
type ClientOption func(*NetClient)

func WithClientTimeout(timeout time.Duration) ClientOption     // 设置全局超时
func WithHeaders(headers http.Header) ClientOption             // 设置默认请求头
```

---

## ClientMiddlewareFunc — 客户端中间件

```go
type ClientMiddlewareFunc func(*http.Request, *HttpResponse) error
```

可以访问原始请求和响应，用于日志记录、重试、监控等场景。

---

## NetClient — 默认客户端实现

### 创建

```go
client := net.NewClient("http://localhost:8080",
    net.WithClientTimeout(10*time.Second),
    net.WithHeaders(headers),
)

// 添加中间件
client.WithMiddleware(func(req *http.Request, resp *net.HttpResponse) error {
    log.Printf("%s %s -> %d", req.Method, req.URL, resp.StatusCode)
    return nil
})
```

### 请求示例

```go
client := net.NewClient("https://api.example.com")

// GET 请求
resp, err := client.Get(ctx, "/users",
    net.WithQuery("page", "1"),
    net.WithQuery("size", "20"),
)

// POST 请求（JSON body）
data := map[string]any{"name": "张三", "email": "zhangsan@example.com"}
resp, err := client.Post(ctx, "/users", data)

// PUT 请求
resp, err := client.Put(ctx, "/users/1", data)

// DELETE 请求
resp, err := client.Delete(ctx, "/users/1")

// 带认证和超时的请求
resp, err := client.Get(ctx, "/users/profile",
    net.WithAuthToken("eyJhbGciOiJIUzI1NiIs..."),
    net.WithTimeout(5*time.Second),
)

// 响应处理
if resp.IsSuccess() {
    var users []User
    resp.Bind(&users)
}

if resp.IsClientError() {
    log.Printf("客户端错误: %s", resp.String())
}
```

### body 类型自动识别

`Post` / `Put` / `Patch` 的 body 参数支持多种类型，客户端自动设置对应的 Content-Type：

| 类型 | Content-Type |
|------|-------------|
| `string` | `text/plain` |
| `[]byte` | `application/octet-stream` |
| `url.Values` | `application/x-www-form-urlencoded` |
| 其他（json.Marshal） | `application/json` |

### 与依赖注入集成

```go
container.Register("httpClient",
    core.Bean(net.NewClient("http://localhost:8080")),
    core.Singleton(),
)

type UserService struct {
    Client net.HttpClient `inject:"httpClient"`
}
```
