// Package sse 提供了简洁的 SSE 断点续传功能，支持任务管理、数据缓存和断线重连
package sse

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"backend/utils/safego"
)

var (
	// ErrTaskNotFound 任务不存在
	ErrTaskNotFound = errors.New("task not found")
	// ErrTaskNotRunning 任务不在运行状态
	ErrTaskNotRunning = errors.New("task is not running")
	// ErrTaskExpired 任务已过期
	ErrTaskExpired = errors.New("task expired")

	// defaultManager 默认的 SSE 管理器，使用包级别函数时会自动初始化
	defaultManager     *SSEManager
	defaultManagerOnce sync.Once
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusRunning   TaskStatus = "running"   // 运行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
)

// TaskInfo 任务信息
type TaskInfo struct {
	TaskID          string                      // 任务ID
	ResumeKey       string                      // 断点续传标识
	Status          TaskStatus                  // 任务状态
	Progress        interface{}                 // 当前进度
	CachedData      []interface{}               // 缓存的数据（断线期间）
	CreatedAt       time.Time                   // 创建时间
	UpdatedAt       time.Time                   // 更新时间
	ExpiresAt       time.Time                   // 过期时间
	DataChannel     chan interface{}            // 实时数据通道
	Subscribers     map[string]chan interface{} // 订阅者列表（key: 订阅者ID）
	mu              sync.RWMutex                // 保护并发访问
	listenerStarted bool                        // 数据监听器是否已启动
}

// AsyncTaskFunc 异步任务执行函数
// ctx: 独立的 context，不受 HTTP 请求断开影响
// taskID: 任务ID
// updateProgress: 更新进度的函数，可以在任务中调用
type AsyncTaskFunc func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error

// SSEManager SSE 管理器
type SSEManager struct {
	tasks       map[string]*TaskInfo // 内存任务缓存
	mu          sync.RWMutex         // 保护 tasks map
	defaultTTL  time.Duration        // 默认任务过期时间
	cleanupTick *time.Ticker         // 清理过期任务的定时器
	stopCleanup chan struct{}        // 停止清理的信号
}

// NewSSEManager 创建 SSE 管理器
// defaultTTL: 默认任务过期时间，过期任务无法续传
func NewSSEManager(defaultTTL time.Duration) *SSEManager {
	if defaultTTL <= 0 {
		defaultTTL = 1 * time.Hour // 默认1小时
	}

	m := &SSEManager{
		tasks:       make(map[string]*TaskInfo),
		defaultTTL:  defaultTTL,
		stopCleanup: make(chan struct{}),
	}

	// 启动清理过期任务的 goroutine
	m.cleanupTick = time.NewTicker(5 * time.Minute)
	safego.Go(context.Background(), func() {
		m.cleanupExpiredTasks()
	})

	return m
}

// cleanupExpiredTasks 定期清理过期任务
func (m *SSEManager) cleanupExpiredTasks() {
	for {
		select {
		case <-m.cleanupTick.C:
			m.mu.Lock()
			now := time.Now()
			for taskID, task := range m.tasks {
				task.mu.RLock()
				expired := task.ExpiresAt.Before(now)
				status := task.Status
				task.mu.RUnlock()

				if expired || (status != TaskStatusRunning) {
					delete(m.tasks, taskID)
					// 关闭通道
					task.mu.Lock()
					// 安全关闭数据通道
					func() {
						defer func() {
							if r := recover(); r != nil {
								// channel 已经被关闭，忽略 panic
							}
						}()
						if task.DataChannel != nil {
							close(task.DataChannel)
						}
					}()
					// 安全关闭所有订阅者通道
					for _, subChan := range task.Subscribers {
						func() {
							defer func() {
								if r := recover(); r != nil {
									// channel 已经被关闭，忽略 panic
								}
							}()
							close(subChan)
						}()
					}
					task.mu.Unlock()
				}
			}
			m.mu.Unlock()
		case <-m.stopCleanup:
			return
		}
	}
}

// Stop 停止管理器，清理资源
func (m *SSEManager) Stop() {
	if m.cleanupTick != nil {
		m.cleanupTick.Stop()
	}
	close(m.stopCleanup)
}

// ExecuteWithSSE 执行带有 SSE 的任务，自动处理断线重连、任务创建、数据缓存等
//
// 参数:
//   - ctx: HTTP 请求的 context（用于检测客户端断线）
//   - resumeKey: 断点续传标识，如果提供则尝试恢复已有任务，为空则创建新任务
//   - subscriberID: 订阅者ID，用于标识不同的客户端连接
//   - asyncFunc: 异步任务执行函数，会在独立的 context 中执行
//   - asyncTimeout: 异步任务超时时间
//
// 返回:
//   - dataChan: 数据通道，用于接收任务进度数据
//   - taskID: 任务ID，可用于后续的断点续传
//   - error: 错误信息
func (m *SSEManager) ExecuteWithSSE(
	ctx context.Context,
	resumeKey string,
	subscriberID string,
	asyncFunc AsyncTaskFunc,
	asyncTimeout time.Duration,
) (<-chan interface{}, string, error) {
	// 1. 检查是否需要恢复任务
	var task *TaskInfo
	var taskID string
	isNewTask := false

	if resumeKey != "" {
		// 尝试恢复已有任务
		m.mu.RLock()
		for id, t := range m.tasks {
			t.mu.RLock()
			if t.ResumeKey == resumeKey {
				task = t
				taskID = id
				t.mu.RUnlock()
				break
			}
			t.mu.RUnlock()
		}
		m.mu.RUnlock()

		if task != nil {
			task.mu.RLock()
			status := task.Status
			expired := task.ExpiresAt.Before(time.Now())
			task.mu.RUnlock()

			if expired {
				return nil, "", ErrTaskExpired
			}
			if status != TaskStatusRunning {
				return nil, "", ErrTaskNotRunning
			}
		}
	}

	// 2. 创建新任务（如果不存在）
	if task == nil {
		isNewTask = true
		taskID = fmt.Sprintf("task_%d", time.Now().UnixNano())
		resumeKey = fmt.Sprintf("resume_%d", time.Now().UnixNano())

		task = &TaskInfo{
			TaskID:      taskID,
			ResumeKey:   resumeKey,
			Status:      TaskStatusRunning,
			CachedData:  make([]interface{}, 0),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExpiresAt:   time.Now().Add(m.defaultTTL),
			DataChannel: make(chan interface{}, 100),
			Subscribers: make(map[string]chan interface{}),
		}

		m.mu.Lock()
		m.tasks[taskID] = task
		m.mu.Unlock()
	}

	// 3. 创建订阅者通道
	subChan := make(chan interface{}, 100)
	task.mu.Lock()
	task.Subscribers[subscriberID] = subChan

	// 4. 如果是重连，发送缓存的历史数据
	outputChan := make(chan interface{}, 100)
	if len(task.CachedData) > 0 {
		// 先发送缓存的数据
		for _, cached := range task.CachedData {
			select {
			case subChan <- cached:
			case <-ctx.Done():
				task.mu.Unlock()
				return nil, "", ctx.Err()
			default:
			}
		}
		// 清空缓存
		task.CachedData = make([]interface{}, 0)
	}
	task.mu.Unlock()

	// 5. 如果是新任务，启动异步任务执行和数据监听器
	if isNewTask {
		// 创建独立的 context（不受 HTTP 请求断开影响）
		asyncCtx, cancel := context.WithTimeout(context.Background(), asyncTimeout)
		if asyncTimeout <= 0 {
			asyncCtx, cancel = context.WithCancel(context.Background())
		}

		// 定义更新进度的函数
		updateProgress := func(data interface{}) error {
			return m.UpdateProgress(ctx, taskID, data)
		}

		// 使用 safego 安全执行异步任务
		safego.Go(ctx, func() {
			defer cancel()
			err := asyncFunc(asyncCtx, taskID, updateProgress)
			if err != nil {
				m.CompleteTask(ctx, taskID, TaskStatusFailed)
			} else {
				m.CompleteTask(ctx, taskID, TaskStatusCompleted)
			}
		})

		// 启动数据监听 goroutine（从任务数据通道转发到订阅者）
		// 这个监听器只在任务创建时启动一次
		task.mu.Lock()
		if !task.listenerStarted {
			task.listenerStarted = true
			task.mu.Unlock()

			safego.Go(ctx, func() {
				for {
					select {
					case data, ok := <-task.DataChannel:
						if !ok {
							return
						}

						task.mu.RLock()
						hasSubscribers := len(task.Subscribers) > 0
						subscribers := make(map[string]chan interface{})
						for k, v := range task.Subscribers {
							subscribers[k] = v
						}
						task.mu.RUnlock()

						if hasSubscribers {
							// 有订阅者，直接发送
							for _, subChan := range subscribers {
								select {
								case subChan <- data:
								default:
									// 订阅者通道已满，跳过
								}
							}
						} else {
							// 无订阅者，缓存数据（断线期间）
							task.mu.Lock()
							task.CachedData = append(task.CachedData, data)
							task.mu.Unlock()
						}
					case <-asyncCtx.Done():
						return
					}
				}
			})
		} else {
			task.mu.Unlock()
		}
	}

	// 6. 启动数据转发 goroutine（从订阅者通道转发到输出通道）
	safego.Go(ctx, func() {
		defer close(outputChan)
		defer func() {
			// 清理订阅者，使用 recover 防止重复关闭 channel
			task.mu.Lock()
			defer task.mu.Unlock()
			// 检查订阅者是否还存在（可能已被 CompleteTask 清理）
			if _, exists := task.Subscribers[subscriberID]; exists {
				delete(task.Subscribers, subscriberID)
				// 安全关闭 channel，防止重复关闭
				func() {
					defer func() {
						if r := recover(); r != nil {
							// channel 已经被关闭，忽略 panic
						}
					}()
					close(subChan)
				}()
			}
		}()

		// 先发送缓存的数据（如果有）
		task.mu.RLock()
		cachedData := make([]interface{}, len(task.CachedData))
		copy(cachedData, task.CachedData)
		task.mu.RUnlock()

		for _, data := range cachedData {
			select {
			case outputChan <- data:
			case <-ctx.Done():
				return
			}
		}

		// 然后转发实时数据
		for {
			select {
			case data, ok := <-subChan:
				if !ok {
					return
				}
				select {
				case outputChan <- data:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})

	return outputChan, taskID, nil
}

// UpdateProgress 更新任务进度，自动处理数据缓存和转发
//
// 参数:
//   - ctx: 上下文
//   - taskID: 任务ID
//   - data: 进度数据
//
// 返回: error
func (m *SSEManager) UpdateProgress(ctx context.Context, taskID string, data interface{}) error {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	m.mu.RUnlock()

	if !exists {
		return ErrTaskNotFound
	}

	task.mu.RLock()
	status := task.Status
	task.mu.RUnlock()

	if status != TaskStatusRunning {
		return ErrTaskNotRunning
	}

	// 更新任务信息
	task.mu.Lock()
	task.Progress = data
	task.UpdatedAt = time.Now()
	task.mu.Unlock()

	// 发送数据到任务通道
	select {
	case task.DataChannel <- data:
	case <-ctx.Done():
		return ctx.Err()
	default:
		// 通道已满，跳过
	}

	return nil
}

// CompleteTask 标记任务完成
//
// 参数:
//   - ctx: 上下文
//   - taskID: 任务ID
//   - status: 最终状态（completed 或 failed）
func (m *SSEManager) CompleteTask(ctx context.Context, taskID string, status TaskStatus) {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	task.mu.Lock()
	task.Status = status
	task.UpdatedAt = time.Now()
	// 清空缓存
	task.CachedData = make([]interface{}, 0)
	// 安全关闭数据通道
	if task.DataChannel != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// channel 已经被关闭，忽略 panic
				}
			}()
			close(task.DataChannel)
		}()
		task.DataChannel = nil
	}
	task.mu.Unlock()

	// 关闭所有订阅者通道
	task.mu.Lock()
	subscribers := make(map[string]chan interface{})
	for k, v := range task.Subscribers {
		subscribers[k] = v
	}
	// 清空订阅者列表，防止数据转发 goroutine 的 defer 重复关闭
	task.Subscribers = make(map[string]chan interface{})
	task.mu.Unlock()

	// 安全关闭所有订阅者通道
	for _, subChan := range subscribers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// channel 已经被关闭，忽略 panic
				}
			}()
			close(subChan)
		}()
	}
}

// GetTaskInfo 获取任务信息（用于查询任务状态）
func (m *SSEManager) GetTaskInfo(taskID string) (*TaskInfo, error) {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrTaskNotFound
	}

	task.mu.RLock()
	defer task.mu.RUnlock()

	// 返回副本，避免并发修改
	info := &TaskInfo{
		TaskID:    task.TaskID,
		ResumeKey: task.ResumeKey,
		Status:    task.Status,
		Progress:  task.Progress,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		ExpiresAt: task.ExpiresAt,
	}

	return info, nil
}

// getDefaultManager 获取默认管理器，如果不存在则创建
func getDefaultManager() *SSEManager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewSSEManager(1 * time.Hour)
	})
	return defaultManager
}

// Init 初始化默认管理器（可选）
// 如果不调用此函数，会在第一次使用包级别函数时自动初始化（默认TTL为1小时）
//
// 参数:
//   - defaultTTL: 默认任务过期时间，过期任务无法续传
func Init(defaultTTL time.Duration) {
	defaultManagerOnce.Do(func() {
		defaultManager = NewSSEManager(defaultTTL)
	})
}

// ExecuteWithSSE 使用默认管理器执行带有 SSE 的任务
// 这是包级别的便捷函数，直接调用即可，无需创建管理器对象
//
// 参数:
//   - ctx: HTTP 请求的 context（用于检测客户端断线）
//   - resumeKey: 断点续传标识，如果提供则尝试恢复已有任务，为空则创建新任务
//   - subscriberID: 订阅者ID，用于标识不同的客户端连接
//   - asyncFunc: 异步任务执行函数，会在独立的 context 中执行
//   - asyncTimeout: 异步任务超时时间
//
// 返回:
//   - dataChan: 数据通道，用于接收任务进度数据
//   - taskID: 任务ID，可用于后续的断点续传
//   - error: 错误信息
func ExecuteWithSSE(
	ctx context.Context,
	resumeKey string,
	subscriberID string,
	asyncFunc AsyncTaskFunc,
	asyncTimeout time.Duration,
) (<-chan interface{}, string, error) {
	return getDefaultManager().ExecuteWithSSE(ctx, resumeKey, subscriberID, asyncFunc, asyncTimeout)
}

// UpdateProgress 使用默认管理器更新任务进度
// 这是包级别的便捷函数，直接调用即可
//
// 参数:
//   - ctx: 上下文
//   - taskID: 任务ID
//   - data: 进度数据
//
// 返回: error
func UpdateProgress(ctx context.Context, taskID string, data interface{}) error {
	return getDefaultManager().UpdateProgress(ctx, taskID, data)
}

// CompleteTask 使用默认管理器标记任务完成
// 这是包级别的便捷函数，直接调用即可
//
// 参数:
//   - ctx: 上下文
//   - taskID: 任务ID
//   - status: 最终状态（TaskStatusCompleted 或 TaskStatusFailed）
func CompleteTask(ctx context.Context, taskID string, status TaskStatus) {
	getDefaultManager().CompleteTask(ctx, taskID, status)
}

// GetTaskInfo 使用默认管理器获取任务信息
// 这是包级别的便捷函数，直接调用即可
//
// 参数:
//   - taskID: 任务ID
//
// 返回:
//   - *TaskInfo: 任务信息
//   - error: 错误信息
func GetTaskInfo(taskID string) (*TaskInfo, error) {
	return getDefaultManager().GetTaskInfo(taskID)
}
