# handle 包 - Gin 错误处理和响应工具

提供统一的 Gin 错误处理和成功响应功能。

## 功能特性

- ✅ 统一的错误处理：自动识别 errorx.StatusError 类型
- ✅ 灵活的配置：支持自定义 HTTP 状态码和错误码
- ✅ 自动日志记录：根据错误类型自动记录日志
- ✅ 统一的响应格式：标准化的 JSON 响应结构
- ✅ 成功响应封装：便捷的成功响应方法

## 快速开始

### 1. 基本错误处理

```go
import (
    "bid_engine/utils/handle"
    "bid_engine/utils/errorx"
)

func (h *UserHandler) SendRegisterEmailCode(c *gin.Context) {
    var req SendEmailCodeReq
    if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
        handle.HandleError(c, err, "发送注册邮箱验证码", nil)
        return
    }
    
    // 处理业务逻辑...
    handle.Success(c, SendEmailCodeResp{})
}
```

### 2. 使用自定义配置

```go
import (
    usererror "bid_engine/app/interface/types/error"
    "bid_engine/utils/handle"
)

// 定义错误处理配置
var errorConfig = &handle.ErrorConfig{
    DefaultStatusCode: http.StatusBadRequest,
    DefaultErrorCode:  usererror.UserErrInvalidParam,
    LogLevel:          "warn",
}

func (h *UserHandler) MyMethod(c *gin.Context) {
    // ...
    if err != nil {
        handle.HandleError(c, err, "操作名称", errorConfig)
        return
    }
}
```

### 3. 带上下文的错误处理

```go
func (h *UserHandler) SendRegisterEmailCode(c *gin.Context) {
    ctx := c.Request.Context()
    
    if err := h.smsLogic.SendEmailCode(ctx, req.Email, 1); err != nil {
        // 使用带上下文的错误处理，可以从 context 中提取追踪信息
        handle.HandleErrorWithContext(c, err, "发送注册邮箱验证码", nil)
        return
    }
}
```

## API 参考

### ErrorConfig

错误处理配置结构体：

```go
type ErrorConfig struct {
    DefaultStatusCode int    // 默认 HTTP 状态码（当错误不是 StatusError 时使用）
    DefaultErrorCode  int32  // 默认错误码（当错误不是 StatusError 时使用）
    LogLevel          string // 日志级别: "warn", "error", "info", "debug"
}
```

### HandleError

统一处理错误并返回响应：

```go
func HandleError(c *gin.Context, err error, operation string, config *ErrorConfig)
```

**参数说明：**

- `c`: gin.Context
- `err`: 错误对象
- `operation`: 操作名称（用于日志记录）
- `config`: 错误处理配置（可选，如果为 nil 则使用默认配置）

**响应格式：**

当错误是 `errorx.StatusError` 类型时：

```json
{
    "code": 1000200,
    "message": "参数无效: 邮箱格式不正确"
}
```

当错误是普通错误时：

```json
{
    "code": 1000200,  // 如果配置了 DefaultErrorCode
    "message": "错误消息"
}
```

### HandleErrorWithContext

带上下文的错误处理：

```go
func HandleErrorWithContext(c *gin.Context, err error, operation string, config *ErrorConfig)
```

与 `HandleError` 相同，但会从 context 中提取额外信息记录日志。

### Success

返回成功响应：

```go
func Success(c *gin.Context, data interface{})
```

**响应格式：**

```json
{
    "code": 0,
    "data": { ... }
}
```

### SuccessWithMessage

返回带消息的成功响应：

```go
func SuccessWithMessage(c *gin.Context, message string, data interface{})
```

**响应格式：**

```json
{
    "code": 0,
    "message": "操作成功",
    "data": { ... }
}
```

## 完整示例

### 示例 1: 基本使用

```go
package user

import (
    "bid_engine/utils/bind"
    "bid_engine/utils/handle"
    usererror "bid_engine/app/interface/types/error"
    "bid_engine/utils/errorx"
)

func (h *UserHandler) SendRegisterEmailCode(c *gin.Context) {
    var req SendEmailCodeReq
    
    // 绑定和验证
    if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
        handle.HandleError(c, err, "发送注册邮箱验证码", nil)
        return
    }
    
    // 业务逻辑
    ctx := c.Request.Context()
    if err := h.smsLogic.SendEmailCode(ctx, req.Email, 1); err != nil {
        wrappedErr := errorx.Wrap(err, usererror.UserErrSendEmailCodeFailed, errorx.K("reason", err.Error()))
        handle.HandleErrorWithContext(c, wrappedErr, "发送注册邮箱验证码", nil)
        return
    }
    
    // 成功响应
    handle.Success(c, SendEmailCodeResp{})
}
```

### 示例 2: 使用自定义配置

```go
package user

import (
    "net/http"
    "bid_engine/utils/handle"
    usererror "bid_engine/app/interface/types/error"
)

// 定义模块级别的错误配置
var errorConfig = &handle.ErrorConfig{
    DefaultStatusCode: http.StatusBadRequest,
    DefaultErrorCode:  usererror.UserErrInvalidParam,
    LogLevel:          "warn",
}

func (h *UserHandler) MyMethod(c *gin.Context) {
    // ...
    if err != nil {
        handle.HandleError(c, err, "操作名称", errorConfig)
        return
    }
    
    handle.SuccessWithMessage(c, "操作成功", result)
}
```

### 示例 3: 不同 HTTP 状态码

```go
// 客户端错误（400）
clientErrorConfig := &handle.ErrorConfig{
    DefaultStatusCode: http.StatusBadRequest,
    LogLevel:          "warn",
}

// 服务器错误（500）
serverErrorConfig := &handle.ErrorConfig{
    DefaultStatusCode: http.StatusInternalServerError,
    LogLevel:          "error",
}

func (h *UserHandler) HandleRequest(c *gin.Context) {
    if validationErr != nil {
        handle.HandleError(c, validationErr, "参数验证", clientErrorConfig)
        return
    }
    
    if serverErr != nil {
        handle.HandleError(c, serverErr, "服务器错误", serverErrorConfig)
        return
    }
}
```

## 响应格式说明

### 错误响应

**StatusError 类型：**

```json
{
    "code": 1000200,
    "message": "参数无效: 邮箱格式不正确"
}
```

**普通错误类型：**

```json
{
    "code": 1000200,  // 如果配置了 DefaultErrorCode
    "message": "错误消息"
}
```

### 成功响应

**Success：**

```json
{
    "code": 0,
    "data": { ... }
}
```

**SuccessWithMessage：**

```json
{
    "code": 0,
    "message": "操作成功",
    "data": { ... }
}
```

## 注意事项

1. **错误码配置**：确保所有错误码已在 errorx 中注册
2. **日志级别**：根据错误严重程度选择合适的日志级别
3. **HTTP 状态码**：客户端错误使用 400，服务器错误使用 500
4. **Context 使用**：需要追踪信息时使用 `HandleErrorWithContext`
5. **配置复用**：建议在模块级别定义配置，避免重复创建

## 与 bind 包配合使用

`handle` 包通常与 `bind` 包配合使用：

```go
// 1. 使用 bind 进行参数验证
if err := bind.ShouldBindJSON(c, &req, bindConfig); err != nil {
    handle.HandleError(c, err, "操作名称", nil)
    return
}

// 2. 处理业务逻辑
if err := businessLogic(); err != nil {
    handle.HandleError(c, err, "操作名称", nil)
    return
}

// 3. 返回成功响应
handle.Success(c, result)
```
