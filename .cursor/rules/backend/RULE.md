---
alwaysApply: true
---

# Backend 开发规范

## 项目架构

### 分层架构

项目采用经典的三层架构模式：

1. **Handler 层** (`app/internal/handler/`)
   - 处理 HTTP 请求和响应
   - 参数绑定和验证
   - 调用 Logic 层处理业务逻辑
   - 返回统一格式的响应

2. **Logic 层** (`app/internal/logic/`)
   - 实现核心业务逻辑
   - 调用 Repo 层进行数据访问
   - 处理业务规则和验证
   - 不直接处理 HTTP 相关逻辑

3. **Repo 层** (`app/internal/repo/`)
   - 数据访问层
   - 使用 GORM 进行数据库操作
   - 封装数据库查询逻辑

### 依赖注入

使用 `uber/fx` 进行依赖注入：

- 每个模块通过 `provider.go` 文件组织依赖
- 使用 `fx.In` 和 `fx.Out` 定义依赖关系
- 使用 `fx.Annotate` 和 `fx.As` 实现接口注入

### 模块组织

每个功能模块包含以下文件：

- `*_handler.go`: Handler 层实现
- `*_model.go`: Handler 层的请求/响应模型
- `*_logic.go`: Logic 层实现
- `*_repo.go`: Repo 层实现
- `provider.go`: 模块依赖注入配置

## 代码风格

### 命名规范

- **包名**: 小写，简短，有意义
- **结构体**: 大驼峰命名，如 `UserHandler`, `UserLogic`
- **接口**: 大驼峰命名，如 `UserLogic`, `UserRepo`
- **函数**: 大驼峰命名，公开函数首字母大写
- **变量**: 小驼峰命名，如 `userID`, `accessToken`
- **常量**: 大驼峰或全大写，如 `ContextKeyUserID`, `AuthErrTokenRequired`

### 文件组织

- 每个包一个目录
- 相关功能放在同一包下
- 使用 `provider.go` 组织模块依赖

### 接口定义

- Handler 层定义 Logic 接口（在 Handler 包中）
- Logic 层定义 Repo 接口（在 Logic 包中）
- 使用接口实现依赖倒置，便于测试和解耦

### 错误处理

- 使用 `errorx` 包统一处理错误
- 定义错误码常量（在 `app/types/errorn/` 目录下）
- 使用 `errorx.New()` 创建错误，支持错误消息模板
- 使用 `errorx.Wrap()` 包装底层错误
- 错误消息支持占位符替换，如 `{user_uid}`, `{reason}`

### 响应处理

- 使用 `handle.Success()` 返回成功响应
- 使用 `handle.HandleErrorWithContext()` 处理错误
- 统一响应格式：`{code: 0, data: {...}}` 或 `{code: xxx, message: "..."}`

### 参数验证

- 使用 `bind.ShouldBindJSON()` 绑定和验证 JSON 请求体
- 使用 `gin` 的 `binding` 标签进行验证
- 使用 `label` 标签定义字段中文名称
- 配置 `FieldErrorConfig` 定义字段错误码映射

### 日志记录

- 使用 `logs` 包进行结构化日志记录
- 使用 `logs.CtxInfof()`, `logs.CtxWarnf()`, `logs.CtxErrorf()` 记录带上下文的日志
- 日志包含关键信息：用户ID、操作类型、错误详情等

### Context 使用

- 所有函数接收 `context.Context` 作为第一个参数
- 使用 `c.Request.Context()` 获取请求上下文
- 通过 Context 传递用户ID、Token 等信息（使用 `meta.ContextKey*`）

## 编程习惯

### API 文档

- 所有 Handler 方法必须添加 Swagger 注释
- 注释包括：`@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`
- 请求/响应模型使用 `example` 标签提供示例值

### 数据库操作

- 使用 GORM 进行数据库操作
- 所有查询必须使用 `WithContext(ctx)` 传递上下文
- 使用 `gorm.DeletedAt` 实现软删除
- 模型定义使用 `gorm` 标签指定列信息

### 环境变量

- 使用 `envx` 包读取环境变量
- 环境变量常量定义在 `app/types/consts/` 目录
- 提供默认值，避免配置缺失导致程序崩溃

### 中间件

- 认证中间件：`middleware.AuthMiddleware()`
- CORS 中间件：`middleware.CORSMiddleware()`
- API 日志中间件：`middleware.APILoggerMiddleware()`
- 在路由中按需使用中间件

### 生命周期管理

- 使用 `fx.Lifecycle` 管理资源生命周期
- 在 `OnStart` 中启动服务
- 在 `OnStop` 中优雅关闭资源（数据库连接、HTTP 服务器等）

### DTO 模式

- 使用 DTO（Data Transfer Object）进行数据传输
- DTO 定义在 `app/types/dto/` 目录
- Handler 层使用 `*_model.go` 定义请求/响应模型
- Logic 层返回 DTO，不直接返回数据库模型

### 代码示例

#### Handler 层示例

```go
func (h *UserHandler) Login(c *gin.Context) {
    ctx := c.Request.Context()
    
    var req LoginReq
    if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
        handle.HandleErrorWithContext(c, err, "登录", nil)
        return
    }
    
    u, t, err := h.userLogic.Login(ctx, req.Username, req.Password)
    if err != nil {
        handle.HandleErrorWithContext(c, err, "登录", nil)
        return
    }
    
    logs.CtxInfof(ctx, "用户登录成功: user_id=%d", u.UserID)
    handle.Success(c, LoginResp{
        UserID:       u.UserID,
        AccessToken:  t.AccessToken,
        RefreshToken: t.RefreshToken,
    })
}
```

#### Logic 层示例

```go
func (l *UserLogic) Login(ctx context.Context, username string, password string) (*dto.UserDTO, *dto.TokenDTO, error) {
    user, err := l.userRepo.GetUserByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil, errorx.New(authError.AuthErrUserNotFound, errorx.K("user_uid", username))
        }
        return nil, nil, errorx.Wrap(err, authError.AuthErrUserNotFound, errorx.K("user_uid", username))
    }
    
    // 业务逻辑处理...
    
    return userDTO, tokenDTO, nil
}
```

#### Repo 层示例

```go
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*userModel.User, error) {
    var user userModel.User
    if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

## 注意事项

1. **不要**在 Handler 层直接访问数据库
2. **不要**在 Logic 层直接处理 HTTP 相关逻辑
3. **不要**在 Repo 层实现业务逻辑
4. **必须**在所有数据库操作中传递 `context.Context`
5. **必须**使用统一的错误处理和响应格式
6. **必须**为所有公开的 API 添加 Swagger 注释
7. **必须**使用结构化日志记录关键操作
8. **必须**通过接口定义依赖，而不是直接依赖具体实现
