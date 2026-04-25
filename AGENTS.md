# AGENTS.md - Go 项目开发规范

## 构建命令

- 构建所有二进制文件：`make build` 或 `go build ./...`
- 构建当前模块：`go build -o /dev/null ./...`

## 测试命令

- 运行所有测试：`go test ./...`
- 运行单个测试（按名称匹配）：`go test -run <TestName> ./path/to/package`
- 运行单个测试文件：`go test -v ./path/to/package/file_test.go`
- 带覆盖率：`go test -cover ./...`

## Lint 与格式化

- 格式化代码：`go fmt ./...` 或 `gofumpt -w .`（如配置）
- 运行 linter：`golangci-lint run`
- 自动修复：`golangci-lint run --fix`

## 代码风格指南

### 导入规范

- 使用标准库分组 → 第三方包 → 本地包，每组之间用空白行分隔。
- 禁止相对导入（如 `../foo`），使用模块路径完整导入。

### 命名约定

- 包名：小写、单个单词，无下划线。例如 `userservice`
- 导出标识符：大写驼峰 (`UserID`)
- 非导出标识符：小写驼峰 (`userID`)
- 常量：使用驼峰，而非全大写加下划线

### 错误处理

- 不忽略方法、函数的返回错误，始终处理错误。
- 始终处理错误，避免 `_` 忽略。
- 使用 `fmt.Errorf` 或 `errors.New`，必要时用 `%w` 包装。
- 自定义错误类型时实现 `Error()` 方法。

### 类型与函数

- 接受接口，返回具体类型。
- 结构体字段顺序：先 `sync.Mutex` 等互斥体，再其他字段。
- 函数长度尽量不超过 50 行，文件不超过 500 行。

### 测试规范

- 表驱动测试（table-driven tests）优先。
- 测试函数命名：`TestFunctionName_Condition_ExpectedBehavior`。
- 使用 `t.Parallel()` 进行并行测试。

## 项目特定约定

- 主入口在 `cmd/server/main.go`
- 内部包放 `internal/` 目录下，避免外部导入
- 使用 `context.Context` 作为第一个参数传递请求上下文