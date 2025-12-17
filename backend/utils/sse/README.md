# SSE 断点续传功能

一个简洁的 Server-Sent Events (SSE) 断点续传实现，支持任务管理、数据缓存和断线重连 ✨

## 🎯 核心特性

- ✅ **断线重连**：客户端断线后，任务继续在后台执行
- ✅ **数据缓存**：断线期间的数据自动缓存，重连后补发
- ✅ **独立 Context**：异步任务使用独立的 context，不受 HTTP 请求断开影响
- ✅ **任务状态管理**：完整的任务生命周期管理（创建、运行、完成、失败）
- ✅ **自动清理**：定期清理过期任务，防止内存泄漏
- ✅ **并发安全**：所有操作都是线程安全的

## 📦 依赖

- `bid_engine/utils/safego` - 安全的 goroutine 执行（panic 恢复）

## 🚀 快速开始

### 方式一：使用包级别函数（推荐，最简单）✨

无需创建管理器对象，直接调用包级别函数即可！

```go
import "bid_engine/utils/sse"

// 可选：初始化默认管理器（如果不调用，会在第一次使用时自动初始化，默认TTL为1小时）
// sse.Init(1 * time.Hour)
```

### 方式二：创建自定义管理器

如果需要多个独立的管理器实例，可以创建自定义管理器：

```go
import "bid_engine/utils/sse"

// 创建管理器，任务默认1小时过期
manager := sse.NewSSEManager(1 * time.Hour)
defer manager.Stop() // 程序退出时停止管理器
```

### 定义异步任务

```go
asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
    // 执行长时间运行的任务
    for i := 1; i <= 10; i++ {
        // 检查 context 是否取消
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // 更新进度
        progress := map[string]interface{}{
            "step":    i,
            "total":   10,
            "message": fmt.Sprintf("处理中... %d/10", i),
        }
        
        if err := updateProgress(progress); err != nil {
            return err
        }

        // 模拟处理时间
        time.Sleep(500 * time.Millisecond)
    }

    return nil
}
```

### 执行任务（使用包级别函数）

```go
ctx := context.Background()
subscriberID := "client_001" // 订阅者ID，用于标识不同的客户端连接

// 第一次连接 - 创建新任务（直接调用包级别函数，无需创建管理器）
dataChan, taskID, err := sse.ExecuteWithSSE(
    ctx,
    "",              // 空 resumeKey 表示创建新任务
    subscriberID,
    asyncTask,
    10*time.Minute,  // 任务超时时间
)
if err != nil {
    log.Fatal(err)
}

// 接收数据流
for data := range dataChan {
    // 处理接收到的数据
    fmt.Printf("收到进度: %+v\n", data)
}
```

### 断点续传（使用包级别函数）

```go
// 假设这是第一次连接时获取的 resumeKey（可以从 TaskInfo 中获取）
resumeKey := "resume_1234567890"

// 第二次连接 - 使用 resumeKey 恢复任务（直接调用包级别函数）
dataChan, _, err := sse.ExecuteWithSSE(
    ctx,
    resumeKey,       // 使用 resumeKey 恢复任务
    "client_002",    // 新的订阅者ID
    asyncTask,       // 这个函数不会再次执行（因为任务已在运行）
    10*time.Minute,
)
if err != nil {
    log.Fatal(err)
}

// 接收数据流（包括断线期间缓存的数据）
for data := range dataChan {
    fmt.Printf("收到数据: %+v\n", data)
}
```

## 📖 API 文档

### 包级别函数（推荐使用）✨

#### Init

```go
func Init(defaultTTL time.Duration)
```

初始化默认管理器（可选）。如果不调用此函数，会在第一次使用包级别函数时自动初始化（默认TTL为1小时）。

**参数：**

- `defaultTTL`: 默认任务过期时间，过期任务无法续传

#### ExecuteWithSSE（包级别函数）

```go
func ExecuteWithSSE(
    ctx context.Context,
    resumeKey string,
    subscriberID string,
    asyncFunc AsyncTaskFunc,
    asyncTimeout time.Duration,
) (<-chan interface{}, string, error)
```

使用默认管理器执行带有 SSE 的任务。这是包级别的便捷函数，直接调用即可，无需创建管理器对象。

**参数：**

- `ctx`: HTTP 请求的 context（用于检测客户端断线）
- `resumeKey`: 断点续传标识，如果提供则尝试恢复已有任务，为空则创建新任务
- `subscriberID`: 订阅者ID，用于标识不同的客户端连接
- `asyncFunc`: 异步任务执行函数，会在独立的 context 中执行
- `asyncTimeout`: 异步任务超时时间

**返回：**

- `dataChan`: 数据通道，用于接收任务进度数据
- `taskID`: 任务ID，可用于后续的断点续传
- `error`: 错误信息

#### UpdateProgress（包级别函数）

```go
func UpdateProgress(ctx context.Context, taskID string, data interface{}) error
```

使用默认管理器更新任务进度。这是包级别的便捷函数，直接调用即可。

**参数：**

- `ctx`: 上下文
- `taskID`: 任务ID
- `data`: 进度数据

**返回：** error

#### CompleteTask（包级别函数）

```go
func CompleteTask(ctx context.Context, taskID string, status TaskStatus)
```

使用默认管理器标记任务完成。这是包级别的便捷函数，直接调用即可。

**参数：**

- `ctx`: 上下文
- `taskID`: 任务ID
- `status`: 最终状态（`TaskStatusCompleted` 或 `TaskStatusFailed`）

#### GetTaskInfo（包级别函数）

```go
func GetTaskInfo(taskID string) (*TaskInfo, error)
```

使用默认管理器获取任务信息。这是包级别的便捷函数，直接调用即可。

**参数：**

- `taskID`: 任务ID

**返回：**

- `*TaskInfo`: 任务信息
- `error`: 错误信息

### 管理器方法（如果需要自定义管理器）

#### NewSSEManager

```go
func NewSSEManager(defaultTTL time.Duration) *SSEManager
```

创建 SSE 管理器。

**参数：**

- `defaultTTL`: 默认任务过期时间，过期任务无法续传

**返回：** SSE 管理器实例

#### ExecuteWithSSE（管理器方法）

```go
func (m *SSEManager) ExecuteWithSSE(
    ctx context.Context,
    resumeKey string,
    subscriberID string,
    asyncFunc AsyncTaskFunc,
    asyncTimeout time.Duration,
) (<-chan interface{}, string, error)
```

执行带有 SSE 的任务，自动处理断线重连、任务创建、数据缓存等。

**参数：**

- `ctx`: HTTP 请求的 context（用于检测客户端断线）
- `resumeKey`: 断点续传标识，如果提供则尝试恢复已有任务，为空则创建新任务
- `subscriberID`: 订阅者ID，用于标识不同的客户端连接
- `asyncFunc`: 异步任务执行函数，会在独立的 context 中执行
- `asyncTimeout`: 异步任务超时时间

**返回：**

- `dataChan`: 数据通道，用于接收任务进度数据
- `taskID`: 任务ID，可用于后续的断点续传
- `error`: 错误信息

### UpdateProgress

```go
func (m *SSEManager) UpdateProgress(ctx context.Context, taskID string, data interface{}) error
```

更新任务进度，自动处理数据缓存和转发。

**参数：**

- `ctx`: 上下文
- `taskID`: 任务ID
- `data`: 进度数据

**返回：** error

### CompleteTask

```go
func (m *SSEManager) CompleteTask(ctx context.Context, taskID string, status TaskStatus)
```

标记任务完成。

**参数：**

- `ctx`: 上下文
- `taskID`: 任务ID
- `status`: 最终状态（`TaskStatusCompleted` 或 `TaskStatusFailed`）

### GetTaskInfo

```go
func (m *SSEManager) GetTaskInfo(taskID string) (*TaskInfo, error)
```

获取任务信息（用于查询任务状态）。

**参数：**

- `taskID`: 任务ID

**返回：**

- `*TaskInfo`: 任务信息
- `error`: 错误信息

## 🔄 断线重连机制

### 工作流程

1. **第一次连接**：
   - 创建新任务
   - 启动异步任务执行
   - 数据直接发送给订阅者，不缓存

2. **客户端断线**：
   - 订阅者被移除
   - 任务继续在后台执行
   - 数据自动缓存到 `CachedData`

3. **客户端重连**：
   - 使用 `resumeKey` 恢复任务
   - 先发送所有历史缓存数据
   - 清空缓存
   - 继续接收实时数据流

### 数据缓存策略

- **有订阅者时**：数据直接发送给订阅者，不缓存
- **无订阅者时**：数据自动缓存到 `CachedData`，等待重连
- **重连时**：先发送缓存数据，然后清空缓存
- **任务完成时**：自动清空缓存

## 💡 使用示例

### 在 HTTP Handler 中使用（使用包级别函数）

```go
import "bid_engine/utils/sse"

func (h *Handler) GenContentSSE(c *gin.Context) {
    ctx := c.Request.Context()
    subscriberID := c.ClientIP() // 使用客户端IP作为订阅者ID
    
    // 从请求中获取 resumeKey（如果有）
    resumeKey := c.Query("resume_key")

    // 定义异步任务
    asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
        // 执行业务逻辑
        return h.service.ProcessContent(ctx, taskID, updateProgress)
    }

    // 直接调用包级别函数，无需创建管理器
    dataChan, taskID, err := sse.ExecuteWithSSE(
        ctx,
        resumeKey,
        subscriberID,
        asyncTask,
        10*time.Minute,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // 设置 SSE 响应头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 发送初始信息（包含 taskID 和 resumeKey）
    taskInfo, _ := sse.GetTaskInfo(taskID)
    if taskInfo != nil {
        fmt.Fprintf(c.Writer, "data: %s\n\n", 
            fmt.Sprintf(`{"task_id":"%s","resume_key":"%s"}`, taskID, taskInfo.ResumeKey))
        c.Writer.Flush()
    }

    // 发送数据流
    for data := range dataChan {
        jsonData, _ := json.Marshal(data)
        fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonData))
        c.Writer.Flush()
    }
}
```

### 在业务逻辑中更新进度（使用包级别函数）

```go
import "bid_engine/utils/sse"

func (s *Service) ProcessContent(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
    // 步骤1
    updateProgress(map[string]interface{}{
        "step":   1,
        "message": "开始处理...",
    })

    // 执行步骤1的逻辑
    // ...

    // 步骤2
    updateProgress(map[string]interface{}{
        "step":   2,
        "message": "处理中...",
    })

    // 执行步骤2的逻辑
    // ...

    // 或者直接使用包级别函数更新进度
    sse.UpdateProgress(ctx, taskID, map[string]interface{}{
        "step":   3,
        "message": "继续处理...",
    })

    return nil
}
```

## ⚠️ 注意事项

1. **Context 管理**：
   - 异步任务使用独立的 context（detached）
   - HTTP 请求的 context 仅用于检测客户端断线
   - 任务不会因 HTTP 断开而取消

2. **任务过期**：
   - 任务默认1小时过期（可配置）
   - 过期任务无法续传
   - 定期清理过期任务

3. **并发安全**：
   - 所有操作都是线程安全的
   - 使用 `sync.RWMutex` 保护并发访问

4. **资源清理**：
   - 程序退出时调用 `manager.Stop()` 停止管理器
   - 任务完成或过期时自动清理资源

5. **订阅者ID**：
   - 每个客户端连接应该使用唯一的订阅者ID
   - 可以使用客户端IP、Session ID 等作为订阅者ID

## 🎨 设计优势

1. **简洁易用**：API 设计简洁，易于集成
2. **自动处理**：断线重连、数据缓存等自动处理
3. **类型安全**：支持任意类型的数据
4. **资源管理**：自动清理过期任务，防止内存泄漏
5. **并发安全**：所有操作都是线程安全的

## 📝 任务状态

- `TaskStatusRunning`: 运行中
- `TaskStatusCompleted`: 已完成
- `TaskStatusFailed`: 失败
- `TaskStatusCancelled`: 已取消

## 🔍 错误处理

- `ErrTaskNotFound`: 任务不存在
- `ErrTaskNotRunning`: 任务不在运行状态
- `ErrTaskExpired`: 任务已过期
