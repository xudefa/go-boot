# validation 包 — 数据验证

## 概述

`validation` 包提供了灵活的数据验证能力，支持 HTTP 请求验证和结构体验证。支持多种验证类型，包括必填、字符串长度、数值范围、邮箱格式、正则表达式、枚举值等。

## 核心概念

### 验证规则

每个验证规则由 `ValidationRule` 定义：

```go
type ValidationRule struct {
    Field     string   // 字段名称
    Type      string   // 验证类型：required, string, number, email, regex, enum, min, max, length
    Value     string   // 验证值（用于 enum, regex 等）
    Min       *float64 // 最小值
    Max       *float64 // 最大值
    MinLength *int     // 最小长度
    MaxLength *int     // 最大长度
    Pattern   string   // 正则表达式
    Message   string   // 自定义错误消息
    In        []string // 枚举值
}
```

### 验证配置

```go
type ValidationConfig struct {
    Rules    []ValidationRule // 验证规则
    Source   string           // 验证来源：query, header, body
    FailFast bool             // 快速失败（遇到第一个错误就停止）
}
```

### 验证结果

```go
type RuleValidationResult struct {
    Valid  bool                // 是否通过验证
    Errors []RuleValidationError // 错误列表
}

type RuleValidationError struct {
    Field   string // 字段名称
    Message string // 错误消息
    Type    string // 错误类型
}
```

## 验证类型

### 1. required — 必填

验证字段不能为空。

```go
rule := validation.ValidationRule{
    Field: "name",
    Type:  "required",
}
```

### 2. string — 字符串验证

验证字符串长度。

```go
minLen := 2
maxLen := 50
rule := validation.ValidationRule{
    Field:     "name",
    Type:      "string",
    MinLength: &minLen,
    MaxLength: &maxLen,
}
```

### 3. number — 数值验证

验证数值范围。

```go
minVal := 1.0
maxVal := 100.0
rule := validation.ValidationRule{
    Field: "age",
    Type:  "number",
    Min:   &minVal,
    Max:   &maxVal,
}
```

### 4. email — 邮箱验证

验证邮箱格式。

```go
rule := validation.ValidationRule{
    Field: "email",
    Type:  "email",
}
```

### 5. regex — 正则表达式验证

使用正则表达式验证格式。

```go
rule := validation.ValidationRule{
    Field:   "phone",
    Type:    "regex",
    Pattern: `^\d{3}-\d{3}-\d{4}$`,
}
```

### 6. enum — 枚举值验证

验证值是否在允许的枚举值中。

```go
rule := validation.ValidationRule{
    Field: "status",
    Type:  "enum",
    In:    []string{"active", "inactive", "pending"},
}
```

### 7. min — 最小值验证

验证数值不小于最小值。

```go
minVal := 0.0
rule := validation.ValidationRule{
    Field: "price",
    Type:  "min",
    Min:   &minVal,
}
```

### 8. max — 最大值验证

验证数值不大于最大值。

```go
maxVal := 1000.0
rule := validation.ValidationRule{
    Field: "quantity",
    Type:  "max",
    Max:   &maxVal,
}
```

### 9. length — 长度验证

验证字符串长度范围。

```go
minLen := 6
maxLen := 20
rule := validation.ValidationRule{
    Field:     "password",
    Type:      "length",
    MinLength: &minLen,
    MaxLength: &maxLen,
}
```

## 使用方式

### 1. RequestValidator — 请求验证器

创建可复用的请求验证器：

```go
config := validation.ValidationConfig{
    Source: "query",
    Rules: []validation.ValidationRule{
        {Field: "page", Type: "required"},
        {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
    },
    FailFast: false,
}

validator, err := validation.NewRequestValidator(config)
if err != nil {
    // 处理配置错误（如无效的正则表达式）
}

// 验证请求
result := validator.Validate(req)
if !result.Valid {
    // 处理验证错误
    for _, e := range result.Errors {
        fmt.Printf("Field: %s, Error: %s\n", e.Field, e.Message)
    }
}
```

### 2. ValidateQuery — 快速验证查询参数

```go
rules := []validation.ValidationRule{
    {Field: "page", Type: "required"},
    {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
}

result := validation.ValidateQuery(req, rules)
if !result.Valid {
    // 处理验证错误
}
```

### 3. ValidateHeaders — 快速验证请求头

```go
rules := []validation.ValidationRule{
    {Field: "X-Api-Key", Type: "required"},
    {Field: "X-Request-Id", Type: "required"},
}

result := validation.ValidateHeaders(req, rules)
if !result.Valid {
    // 处理验证错误
}
```

### 4. ValidateJSONBody — 验证 JSON Body

```go
rules := []validation.ValidationRule{
    {Field: "name", Type: "required"},
    {Field: "email", Type: "email"},
    {Field: "age", Type: "number", Min: floatPtr(18), Max: floatPtr(100)},
}

body := []byte(`{"name":"test","email":"test@example.com","age":25}`)
result := validation.ValidateJSONBody(body, rules)
if !result.Valid {
    // 处理验证错误
}
```

## 自定义错误消息

```go
rule := validation.ValidationRule{
    Field:   "email",
    Type:    "email",
    Message: "请输入有效的邮箱地址",
}
```

## 快速失败模式

```go
config := validation.ValidationConfig{
    Source: "query",
    Rules: []validation.ValidationRule{
        {Field: "name", Type: "required"},
        {Field: "email", Type: "email"},
        {Field: "age", Type: "number"},
    },
    FailFast: true, // 遇到第一个错误就停止验证
}
```

## 完整示例

### HTTP 处理器中的验证

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // 验证请求头
    headerRules := []validation.ValidationRule{
        {Field: "Content-Type", Type: "required"},
        {Field: "X-Request-Id", Type: "required"},
    }
    
    headerResult := validation.ValidateHeaders(r, headerRules)
    if !headerResult.Valid {
        http.Error(w, fmt.Sprintf("Header validation failed: %v", headerResult.Errors), http.StatusBadRequest)
        return
    }
    
    // 读取并验证 body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    
    bodyRules := []validation.ValidationRule{
        {Field: "name", Type: "required", Message: "Name is required"},
        {Field: "email", Type: "email", Message: "Valid email is required"},
        {Field: "age", Type: "number", Min: floatPtr(18), Max: floatPtr(100)},
    }
    
    bodyResult := validation.ValidateJSONBody(body, bodyRules)
    if !bodyResult.Valid {
        http.Error(w, fmt.Sprintf("Body validation failed: %v", bodyResult.Errors), http.StatusBadRequest)
        return
    }
    
    // 处理创建用户逻辑
    // ...
}
```

### 中间件集成

```go
func ValidationMiddleware(rules []validation.ValidationRule) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            result := validation.ValidateQuery(r, rules)
            if !result.Valid {
                http.Error(w, fmt.Sprintf("Validation failed: %v", result.Errors), http.StatusBadRequest)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 使用中间件
rules := []validation.ValidationRule{
    {Field: "page", Type: "required"},
    {Field: "size", Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
}

handler := ValidationMiddleware(rules)(myHandler)
```

## 错误处理

验证器返回 `RuleValidationResult`，包含所有验证错误（除非启用 FailFast）：

```go
result := validator.Validate(req)
if !result.Valid {
    for _, err := range result.Errors {
        log.Printf("Field: %s, Type: %s, Message: %s", 
            err.Field, err.Type, err.Message)
    }
}
```

## 注意事项

1. **空值处理**: 除 `required` 类型外，其他验证类型在值为空时会跳过验证
2. **正则表达式**: 建议在创建验证器时预编译正则表达式，提高性能
3. **JSON Body**: `ValidateJSONBody` 会自动处理 JSON 中的数字类型（float64）
4. **线程安全**: `RequestValidator` 创建后是只读的，可以安全地在多个 goroutine 中使用
5. **自定义消息**: 使用 `Message` 字段可以提供更友好的错误提示