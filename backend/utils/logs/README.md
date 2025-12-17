# 结构化日志系统

简洁好用的结构化日志系统，基于 `zap` 实现，支持日志追踪检索。

## ✨ 特性

- 🎯 **结构化日志**：支持 key-value 格式的结构化日志
- 🔍 **追踪支持**：自动从 context 中提取 `trace_id`、`span_id` 和 `parent_span_id`
- 🌳 **层级追踪**：支持通过 `parent_span_id` 追踪调用层级关系
- 📊 **Loki 集成**：输出格式符合 Loki/Promtail 要求
- 🎨 **简洁 API**：提供简洁明了的日志方法
- ⚙️ **环境配置**：支持通过环境变量配置日志级别和输出格式

## 🚀 快速开始

### 基本使用

```go
import "bid_engine/utils/logs"

// 简单日志
logs.Info("用户登录成功")
logs.Error("数据库连接失败")

// 结构化日志（推荐）
logs.Info("用户登录成功", 
    "user_id", 12345,
    "ip", "192.168.1.1",
    "user_agent", "Mozilla/5.0",
)

logs.Error("订单创建失败",
    "order_id", "ORD-001",
    "error_code", 5001,
    "error_msg", "库存不足",
)
```

### 带上下文的日志（追踪支持）

```go
import (
    "context"
    "bid_engine/utils/logs"
)

// 在 context 中设置 trace_id 和 span_id
ctx := context.WithValue(context.Background(), "trace_id", "trace-12345")
ctx = context.WithValue(ctx, "span_id", "span-67890")

// 使用格式化日志（兼容旧代码）
logs.CtxInfof(ctx, "处理请求: %s", "/api/users")
logs.CtxErrorf(ctx, "处理失败: %v", err)

// 使用结构化日志（推荐）- 包级别函数，更方便
logs.CtxInfo(ctx, "用户操作",
    "action", "create_order",
    "user_id", 12345,
    "order_id", "ORD-001",
)

// 或者使用接口方式（功能相同）
logger := logs.GetDefaultLogger()
if ctxLogger, ok := logger.(logs.CtxStructuredLogger); ok {
    ctxLogger.CtxInfo(ctx, "用户操作",
        "action", "create_order",
        "user_id", 12345,
        "order_id", "ORD-001",
    )
}
```

### 追踪字段说明

日志系统支持以下追踪字段，用于分布式追踪和日志关联：

- **trace_id**: 追踪 ID，标识整个请求链路
- **span_id**: Span ID，标识当前操作
- **parent_span_id**: 父 Span ID，标识调用层级关系

这些字段会自动从 `context.Context` 中提取并添加到日志中。

### 在 Gin 中间件中使用

```go
func TraceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 生成 trace_id 和 span_id
        traceID := generateTraceID()
        spanID := generateSpanID()
        
        // 设置到 context
        ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
        ctx = context.WithValue(ctx, "span_id", spanID)
        c.Request = c.Request.WithContext(ctx)
        
        // 记录请求日志（推荐使用结构化日志）
        logs.CtxInfo(ctx, "收到请求",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "ip", c.ClientIP(),
        )
        
        // 或者使用格式化日志（兼容旧代码）
        logs.CtxInfof(ctx, "收到请求: %s %s", c.Request.Method, c.Request.URL.Path)
        
        c.Next()
    }
}
```

### 创建子 Span（追踪层级调用关系）

当需要追踪嵌套调用时，可以使用 `parent_span_id` 来建立调用层级关系：

```go
// 方式 1: 手动设置 parent_span_id（不推荐）
func callDownstreamService(ctx context.Context) {
    // 获取当前的 span_id 作为 parent_span_id
    currentSpanID := ctx.Value("span_id").(string)
    newSpanID := generateSpanID()
    
    // 创建新的 context，设置新的 span_id 和 parent_span_id
    newCtx := context.WithValue(ctx, "span_id", newSpanID)
    newCtx = context.WithValue(newCtx, "parent_span_id", currentSpanID)
    
    logs.CtxInfo(newCtx, "调用下游服务",
        "service", "payment-service",
        "endpoint", "/api/payments/process",
    )
}

// 方式 2: 使用辅助函数（推荐）
func callDownstreamService(ctx context.Context) {
    // 创建新的 span，自动将当前的 span_id 设置为 parent_span_id
    newSpanID := generateSpanID()
    newCtx := withNewSpan(ctx, newSpanID)  // 自动处理 parent_span_id
    
    logs.CtxInfo(newCtx, "调用下游服务",
        "service", "payment-service",
        "endpoint", "/api/payments/process",
    )
}

// withNewSpan 辅助函数示例
func withNewSpan(ctx context.Context, newSpanID string) context.Context {
    // 获取当前的 span_id 作为 parent_span_id
    var parentSpanID string
    if currentSpanID := ctx.Value("span_id"); currentSpanID != nil {
        if spanIDStr, ok := currentSpanID.(string); ok {
            parentSpanID = spanIDStr
        }
    }
    
    // 设置新的 span_id
    ctx = context.WithValue(ctx, "span_id", newSpanID)
    
    // 如果有 parent_span_id，则设置它
    if parentSpanID != "" {
        ctx = context.WithValue(ctx, "parent_span_id", parentSpanID)
    }
    
    return ctx
}
```

### 追踪层级示例

```go
// 第一层：HTTP 请求
traceID := "trace-123"
spanID := "span-001"
ctx := context.WithValue(context.Background(), "trace_id", traceID)
ctx = context.WithValue(ctx, "span_id", spanID)
logs.CtxInfo(ctx, "收到订单请求")  // trace_id: trace-123, span_id: span-001

// 第二层：订单处理
orderSpanID := "span-002"
orderCtx := withNewSpan(ctx, orderSpanID)
logs.CtxInfo(orderCtx, "处理订单")  // trace_id: trace-123, span_id: span-002, parent_span_id: span-001

// 第三层：支付处理
paymentSpanID := "span-003"
paymentCtx := withNewSpan(orderCtx, paymentSpanID)
logs.CtxInfo(paymentCtx, "处理支付")  // trace_id: trace-123, span_id: span-003, parent_span_id: span-002
```

这样可以在 Grafana 中通过 `parent_span_id` 查询整个调用链：

```logql
{app="bid_engine"} | json | trace_id="trace-123"
```

## 📝 日志级别

支持以下日志级别：

- `Debug`：调试信息
- `Info`：一般信息
- `Warn`：警告信息
- `Error`：错误信息

```go
// 不带上下文的日志
logs.Debug("调试信息", "key", "value")
logs.Info("一般信息", "key", "value")
logs.Warn("警告信息", "key", "value")
logs.Error("错误信息", "key", "value")

// 带上下文的日志（推荐，自动提取 trace_id 和 span_id）
logs.CtxDebug(ctx, "调试信息", "key", "value")
logs.CtxInfo(ctx, "一般信息", "key", "value")
logs.CtxWarn(ctx, "警告信息", "key", "value")
logs.CtxError(ctx, "错误信息", "key", "value")
```

## 🔧 环境变量配置

通过环境变量配置日志行为：

| 环境变量 | 说明 | 可选值 | 默认值 |
|---------|------|--------|--------|
| `LOG_LEVEL` | 日志级别 | debug, info, warn, error, fatal | info |
| `LOG_OUTPUT` | 输出格式 | console, json | 自动检测（容器中为 json） |
| `LOG_DEVELOPMENT` | 开发模式 | true, false | false |
| `LOG_FILE` | 日志文件路径 | 文件路径 | 空（只输出到 stdout） |
| `LOG_MAX_SIZE` | 单个日志文件最大大小（MB） | 正整数 | 100 |
| `LOG_MAX_BACKUPS` | 保留的旧日志文件数量 | 非负整数 | 7 |
| `LOG_MAX_AGE` | 日志文件保留天数 | 正整数 | 30 |
| `LOG_COMPRESS` | 是否压缩旧日志文件 | true, false | true |

### 日志轮转配置

当日志文件达到 `LOG_MAX_SIZE` 时，会自动轮转：

- 当前日志文件会被重命名为 `app.log.1`、`app.log.2` 等
- 超过 `LOG_MAX_BACKUPS` 数量的旧文件会被删除
- 超过 `LOG_MAX_AGE` 天的旧文件会被删除
- 如果 `LOG_COMPRESS=true`，旧文件会被压缩为 `.gz` 格式

### 配置示例

```bash
# 基本配置
export LOG_OUTPUT=json
export LOG_LEVEL=info
export LOG_FILE=./logs/app.log

# 日志轮转配置
export LOG_MAX_SIZE=100        # 单个文件最大 100MB
export LOG_MAX_BACKUPS=7       # 保留 7 个旧文件
export LOG_MAX_AGE=30          # 保留 30 天
export LOG_COMPRESS=true       # 压缩旧文件
```

### 示例

```bash
# 设置日志级别为 debug
export LOG_LEVEL=debug

# 设置输出格式为 JSON（用于 Loki）
export LOG_OUTPUT=json

# 启用开发模式（更详细的日志）
export LOG_DEVELOPMENT=true
```

## 📊 Loki 集成

日志系统会自动输出符合 Loki/Promtail 要求的 JSON 格式，包含以下字段：

- `ts`：Unix 时间戳
- `level`：日志级别（小写）
- `msg`：日志消息
- `caller`：调用位置
- `trace_id`：追踪 ID（如果 context 中存在）
- `span_id`：Span ID（如果 context 中存在）
- 其他自定义字段

### 在 Grafana 中查询日志

```logql
# 查询所有日志
{app="goall-test"}

# 查询错误日志
{app="goall-test", level="error"}

# 查询特定 trace_id 的日志
{app="goall-test"} | json | trace_id="trace-12345"

# 查询包含特定字段的日志
{app="goall-test"} | json | user_id="12345"
```

## 🎯 最佳实践

### 1. 使用结构化日志

✅ **推荐**：

```go
logs.Info("订单创建成功",
    "order_id", orderID,
    "user_id", userID,
    "amount", amount,
)
```

❌ **不推荐**：

```go
logs.Info(fmt.Sprintf("订单创建成功: order_id=%s, user_id=%d, amount=%.2f", orderID, userID, amount))
```

### 2. 在 context 中传递 trace_id

```go
// 在请求入口处生成 trace_id
func TraceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := generateTraceID()
        spanID := generateSpanID()
        ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
        ctx = context.WithValue(ctx, "span_id", spanID)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// 在业务代码中使用带上下文的日志（自动包含 trace_id 和 span_id）
func CreateOrder(ctx context.Context, userID int, orderID string) error {
    logs.CtxInfo(ctx, "开始创建订单",
        "user_id", userID,
        "order_id", orderID,
    )
    // ... 业务逻辑 ...
    return nil
}
```

### 3. 记录错误时包含上下文信息

```go
// 不带上下文的错误日志
logs.Error("数据库查询失败",
    "error", err.Error(),
    "query", sqlQuery,
    "params", params,
    "user_id", userID,
)

// 带上下文的错误日志（推荐，自动包含 trace_id）
logs.CtxError(ctx, "数据库查询失败",
    "error", err.Error(),
    "query", sqlQuery,
    "params", params,
    "user_id", userID,
)
```

## 🔍 追踪检索

日志系统支持通过 `trace_id` 和 `span_id` 进行追踪检索：

1. **设置追踪信息**：在 context 中设置 `trace_id` 和 `span_id`
2. **自动提取**：日志系统会自动从 context 中提取这些字段
3. **在 Grafana 中查询**：使用 LogQL 查询特定追踪的所有日志

```logql
# 查询特定 trace_id 的所有日志
{app="goall-test"} | json | trace_id="trace-12345"
```

## 📚 API 参考

### 包级别方法

#### 不带上下文的日志方法

- `logs.Error(args ...interface{})`：记录错误日志
- `logs.Warn(args ...interface{})`：记录警告日志
- `logs.Info(args ...interface{})`：记录信息日志
- `logs.Debug(args ...interface{})`：记录调试日志

#### 带上下文的日志方法（推荐）

- `logs.CtxError(ctx context.Context, msg string, keyvals ...interface{})`：记录带上下文的错误日志（结构化）⭐
- `logs.CtxWarn(ctx context.Context, msg string, keyvals ...interface{})`：记录带上下文的警告日志（结构化）⭐
- `logs.CtxInfo(ctx context.Context, msg string, keyvals ...interface{})`：记录带上下文的信息日志（结构化）⭐
- `logs.CtxDebug(ctx context.Context, msg string, keyvals ...interface{})`：记录带上下文的调试日志（结构化）⭐

#### 带上下文的格式化日志方法（兼容旧代码）

- `logs.CtxErrorf(ctx context.Context, format string, args ...interface{})`：记录带上下文的错误日志（格式化）
- `logs.CtxWarnf(ctx context.Context, format string, args ...interface{})`：记录带上下文的警告日志（格式化）
- `logs.CtxInfof(ctx context.Context, format string, args ...interface{})`：记录带上下文的信息日志（格式化）
- `logs.CtxDebugf(ctx context.Context, format string, args ...interface{})`：记录带上下文的调试日志（格式化）

#### 其他方法

- `logs.GetDefaultLogger()`：获取默认 logger

### 接口

- `StructuredLogger`：结构化日志接口
- `CtxStructuredLogger`：带上下文的结构化日志接口

## 🐛 故障排查

### 日志没有输出

1. 检查日志级别设置是否正确
2. 检查环境变量 `LOG_LEVEL` 是否设置过高

### trace_id 没有出现在日志中

1. 确保在 context 中设置了 `trace_id`：`ctx = context.WithValue(ctx, "trace_id", "xxx")`
2. 确保使用带上下文的日志方法：
   - 推荐：`logs.CtxInfo(ctx, "消息", "key", "value")`（结构化日志）
   - 兼容：`logs.CtxInfof(ctx, "消息: %s", value)`（格式化日志）

### JSON 格式不正确

1. 确保设置了 `LOG_OUTPUT=json`
2. 或者在容器中运行（会自动使用 JSON 格式）
