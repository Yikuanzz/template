# errorx 错误处理包

一个简洁、强大的 Go 错误处理库，提供基于错误码的错误管理、自动堆栈跟踪、错误包装等功能。

## 📋 设计理念

`errorx` 包采用以下设计原则：

1. **简洁易用**：API 设计简洁，学习成本低
2. **错误码驱动**：通过预定义错误码统一管理错误信息
3. **自动堆栈跟踪**：自动捕获并记录错误发生时的调用堆栈
4. **错误包装**：支持包装现有错误，保留错误链信息
5. **参数化消息**：支持在错误消息中使用占位符，动态填充参数
6. **标准库兼容**：完全兼容 Go 标准库的 `errors` 包

## 🚀 快速开始

### 1. 注册错误码

在应用初始化时注册错误码：

```go
import "bid_engine/utils/errorx"

const (
    ErrPermissionDenied = int32(1000000)
    ErrInvalidParam     = int32(1000001)
    ErrNotFound         = int32(1000002)
)

func init() {
    // 单个注册
    errorx.Register(ErrPermissionDenied, "unauthorized access: {reason}")
    errorx.Register(ErrInvalidParam, "invalid parameter: {param}")
    
    // 批量注册
    errorx.RegisterBatch(map[int32]string{
        ErrNotFound: "resource not found: {resource}",
    })
}
```

### 2. 创建错误

#### 基本用法

```go
// 使用注册的错误码
err := errorx.New(ErrPermissionDenied)
```

#### 使用键值对替换占位符

```go
// 消息模板: "unauthorized access: {reason}"
err := errorx.New(ErrPermissionDenied, errorx.K("reason", "insufficient permissions"))
// 结果: "unauthorized access: insufficient permissions"

// 使用格式化字符串
err := errorx.New(ErrPermissionDenied, errorx.Kf("reason", "user %s has no permission", "alice"))
```

#### 直接提供消息

```go
err := errorx.New(ErrInvalidParam, "参数不能为空")
```

### 3. 包装错误

#### 包装标准错误

```go
originalErr := errors.New("database connection failed")
err := errorx.Wrap(originalErr, ErrInvalidParam, errorx.K("param", "database_url"))
```

#### 使用格式化消息包装

```go
originalErr := errors.New("connection timeout")
err := errorx.Wrapf(originalErr, "failed to connect to %s", "localhost:8080")
```

### 4. 提取错误信息

```go
import (
    "errors"
    "bid_engine/utils/errorx"
)

var statusErr errorx.StatusError
if errors.As(err, &statusErr) {
    code := statusErr.Code()      // 获取错误码
    msg := statusErr.Msg()        // 获取错误消息
    cause := statusErr.Unwrap()   // 获取原始错误
}
```

### 5. 获取简洁的错误消息

```go
// 不包含堆栈信息的简洁消息
msg := errorx.ErrorWithoutStack(err)
// 格式: "code=1000000 message=unauthorized access: test"
```

## 📝 错误消息格式

完整的错误消息格式：

```text
code=<错误码> message=<错误消息>
cause=<原始错误信息>
stack=<堆栈跟踪信息>
```

使用 `ErrorWithoutStack()` 时，只返回 `code` 和 `message` 部分。

## 🔧 API 参考

### 核心函数

- `New(code int32, args ...interface{}) error`: 创建新错误
- `Wrap(err error, code int32, args ...interface{}) error`: 包装现有错误
- `Wrapf(err error, format string, args ...interface{}) error`: 使用格式化消息包装错误
- `ErrorWithoutStack(err error) string`: 获取不包含堆栈的错误消息

### 辅助函数

- `K(key, value string) KV`: 创建键值对
- `Kf(key, format string, args ...interface{}) KV`: 使用格式化字符串创建键值对

### 注册函数

- `Register(code int32, message string)`: 注册单个错误码
- `RegisterBatch(codes map[int32]string)`: 批量注册错误码
- `IsRegistered(code int32) bool`: 检查错误码是否已注册

### StatusError 接口

```go
type StatusError interface {
    error
    Code() int32      // 错误码
    Msg() string      // 错误消息
    Unwrap() error    // 返回被包装的原始错误
}
```

## 💡 最佳实践

1. **统一错误码管理**：在应用启动时集中注册所有错误码
2. **使用有意义的错误码**：建议使用分层错误码（如：1000000 表示权限相关错误）
3. **保留错误链**：使用 `Wrap()` 而不是直接创建新错误，保留原始错误信息
4. **使用占位符**：在错误消息模板中使用 `{key}` 占位符，提高灵活性
5. **标准库兼容**：充分利用 `errors.Is()`、`errors.As()`、`errors.Unwrap()` 等标准库功能

## 📦 包结构

```shell
errorx/
├── error.go      # 核心错误类型和 API
├── code.go       # 错误码注册
└── README.md     # 文档
```

## 🔍 特性说明

- **自动堆栈跟踪**：所有通过 `New()` 和 `Wrap()` 创建的错误都会自动包含堆栈信息
- **避免重复堆栈**：如果错误已经被包装过（已有堆栈），`Wrap()` 不会重复添加堆栈
- **标准错误兼容**：完全兼容 Go 标准库的 `errors` 包，支持 `errors.Is()`、`errors.As()`、`errors.Unwrap()`
- **默认错误消息**：如果使用未注册的错误码，会使用默认错误消息

## 📌 完整示例

```go
package main

import (
    "errors"
    "fmt"
    "bid_engine/utils/errorx"
)

const (
    ErrPermissionDenied = int32(1000000)
    ErrInvalidParam     = int32(1000001)
)

func init() {
    errorx.Register(ErrPermissionDenied, "unauthorized access: {reason}")
    errorx.Register(ErrInvalidParam, "invalid parameter: {param}")
}

func main() {
    // 创建错误
    err := errorx.New(ErrPermissionDenied, errorx.K("reason", "test"))
    fmt.Println(errorx.ErrorWithoutStack(err))
    
    // 包装错误
    originalErr := errors.New("database error")
    wrappedErr := errorx.Wrap(originalErr, ErrInvalidParam, errorx.K("param", "id"))
    
    // 提取错误信息
    var statusErr errorx.StatusError
    if errors.As(wrappedErr, &statusErr) {
        fmt.Printf("Code: %d\n", statusErr.Code())
        fmt.Printf("Message: %s\n", statusErr.Msg())
        fmt.Printf("Cause: %v\n", statusErr.Unwrap())
    }
}
```
