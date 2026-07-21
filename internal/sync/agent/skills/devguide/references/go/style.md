# Go 代码规范

## 1. 工具与检查

### 格式化

Go 代码必须使用 `goimports` 格式化。

### 静态检查

Go 代码必须使用 `golangci-lint` 做静态检查。

每次开发完成或构建前，必须运行 `golangci-lint run`。

## 2. 文件与注释

### 文件组织

不要默认创建 `types.go`。

类型应优先放在表达领域或职责的文件中。只有当一组类型确实属于同一个明确主题，且没有更清晰的领域文件名时，才允许使用聚合文件名。

### 注释

注释规范见 `references/writing.md`。Go 只使用 `//`，不使用 `/* ... */`。

## 3. 声明

### 导入

`import` 必须使用 `import (...)` 聚合声明，即使只有一个导入也必须使用大括号。

### 类型定义

避免无意义的 `type` 定义。

`struct` 和 `interface` 必须独立声明，不使用 `type (...)` 聚合。

```go
// Bad
type (
	User struct{}

	Store interface {
		Save(ctx context.Context, user User) error
	}
)

// Good
type User struct{}

type Store interface {
	Save(ctx context.Context, user User) error
}
```

### 常量与变量

`const` 和 `var` 必须按类别聚合声明。即使只有一个声明，也必须使用 `const (...)` 或 `var (...)`。

```go
// Bad
const defaultLimit = 20

// Good
const (
	defaultLimit = 20
)
```

## 4. 函数与初始化

### 函数与方法

每个函数参数都必须显式指定类型，不使用 `a, b string` 这类参数类型省略写法。

```go
// Bad
func Move(x, y int) {}

// Good
func Move(x int, y int) {}
```

### 错误检查

必须严格限制 `err` 的作用域。当一次调用产生的 `err` 仅用于紧随其后的错误判断，且没有其他返回值需要在 `if` 之后继续使用时，必须在 `if` 初始化语句中声明，不要先单独赋值再判断。

```go
// Bad
err := doSomething()
if err != nil {
	return err
}

// Good
if err := doSomething(); err != nil {
	return err
}
```

只有当 `err` 或同一调用的其他返回值需要在 `if` 之后继续使用时，才单独声明。

当同一作用域内存在多个需要独立保留或区分的错误变量时，当同一作用域内存在多个需要独立保留或区分的错误变量时，应根据对操作命名错误变量。

```go
user, loadErr := loadUser()
if loadErr != nil {
	return loadErr
}

order, createErr := createOrder(user)
if createErr != nil {
	return createErr
}
```

同一作用域内存在多个同类操作时，在动词后补充操作对象，如 `loadUserErr`、`loadOrderErr`。

### 初始化

复杂初始化优先使用 options 模式；必要参数直接传入，可选配置通过 `Option` 追加。

## 5. 库使用

本规范适用于基础设施级能力，例如日志、配置、链路追踪、错误处理、AWS 集成、worker pool、参数校验、分页和 JWT 等。

引入基础设施能力前，必须先检查 `uqpay-core-sdk/pkg/`。

- `uqpay-core-sdk` 已提供能力时，直接使用 SDK 包，不再封装一层。
- `uqpay-core-sdk` 未提供能力时，直接使用社区标准库或事实标准库，不在业务项目里新增本地 utility wrapper。
- 如果缺失能力属于可复用基础设施能力，应补到 `uqpay-core-sdk`，而不是沉淀在单个业务项目中。
