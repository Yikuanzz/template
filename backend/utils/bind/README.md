# bind 包 - Gin 请求绑定和验证工具

提供统一的 Gin 请求绑定和验证错误处理功能。

## 功能特性

- ✅ 统一的错误处理：将 gin binding 错误转换为 errorx 错误
- ✅ 灵活的配置：支持自定义错误码映射
- ✅ 友好的错误消息：自动生成中文错误提示
- ✅ 多种绑定方式：支持 JSON、Query、URI 等

## 快速开始

### 1. 定义错误码配置

```go
import (
    usererror "bid_engine/app/interface/types/error"
    "bid_engine/utils/bind"
)

// 创建错误码配置
var userBindConfig = bind.FieldErrorConfig{
    InvalidParamCode: usererror.UserErrInvalidParam,
    RequiredCode:     usererror.UserErrParamRequired,
    FieldErrorCodes: map[string]int32{
        "Email":    usererror.UserErrParamEmailInvalid,
        "Mobile":   usererror.UserErrParamMobileInvalid,
        "Username": usererror.UserErrParamUsernameInvalid,
    },
    FieldLabels: map[string]string{
        "Email":    "邮箱",
        "Mobile":   "手机号",
        "Username": "用户名",
    },
}
```

### 2. 在 Handler 中使用

```go
func (h *UserHandler) SendRegisterEmailCode(c *gin.Context) {
    var req SendEmailCodeReq
    
    // 使用 ShouldBindJSON 绑定并验证
    if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
        h.handleError(c, err, "发送注册邮箱验证码")
        return
    }
    
    // 处理业务逻辑...
}
```

## API 参考

### FieldErrorConfig

错误码配置结构体：

```go
type FieldErrorConfig struct {
    InvalidParamCode int32            // 通用参数无效错误码
    RequiredCode     int32            // 参数必填错误码
    FieldErrorCodes  map[string]int32 // 字段名到错误码的映射
    FieldLabels      map[string]string // 字段名到中文标签的映射
}
```

### HandleBindingError

处理 gin binding 验证错误：

```go
func HandleBindingError(config FieldErrorConfig, err error) error
```

### ShouldBindJSON

绑定并验证 JSON 请求体：

```go
func ShouldBindJSON(c *gin.Context, obj interface{}, config FieldErrorConfig) error
```

### ShouldBindQuery

绑定并验证 Query 参数：

```go
func ShouldBindQuery(c *gin.Context, obj interface{}, config FieldErrorConfig) error
```

### ShouldBindURI

绑定并验证 URI 参数：

```go
func ShouldBindURI(c *gin.Context, obj interface{}, config FieldErrorConfig) error
```

### ShouldBind

自动识别 Content-Type 并绑定：

```go
func ShouldBind(c *gin.Context, obj interface{}, config FieldErrorConfig) error
```

## 支持的验证标签

- `required` - 必填
- `email` - 邮箱格式
- `len=n` - 固定长度
- `min=n` - 最小长度
- `max=n` - 最大长度
- `gte=n` - 大于等于
- `lte=n` - 小于等于
- `gt=n` - 大于
- `lt=n` - 小于
- `oneof=val1 val2` - 枚举值
- `regexp=pattern` - 正则表达式

## 示例

### 完整示例

```go
package user

import (
    "bid_engine/utils/bind"
    usererror "bid_engine/app/interface/types/error"
)

// 定义配置
var userBindConfig = bind.FieldErrorConfig{
    InvalidParamCode: usererror.UserErrInvalidParam,
    RequiredCode:     usererror.UserErrParamRequired,
    FieldErrorCodes: map[string]int32{
        "Email":    usererror.UserErrParamEmailInvalid,
        "Mobile":   usererror.UserErrParamMobileInvalid,
        "Username": usererror.UserErrParamUsernameInvalid,
    },
    FieldLabels: map[string]string{
        "Email":    "邮箱",
        "Mobile":   "手机号",
        "Username": "用户名",
    },
}

func (h *UserHandler) SendRegisterEmailCode(c *gin.Context) {
    var req SendEmailCodeReq
    
    if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
        h.handleError(c, err, "发送注册邮箱验证码")
        return
    }
    
    // 业务逻辑...
}
```

## 注意事项

1. **错误码配置**：确保所有错误码已在 errorx 中注册
2. **字段名匹配**：FieldErrorCodes 和 FieldLabels 中的 key 必须与结构体字段名完全匹配（区分大小写）
3. **可选配置**：InvalidParamCode 和 RequiredCode 可以为 0，此时会使用默认错误消息
