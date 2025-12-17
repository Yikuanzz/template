// Package taskgroup 提供了一个并发任务组管理工具，支持并发控制和错误处理策略
package taskgroup

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"backend/utils/logs"
)

// TaskGroup 定义了任务组的接口，用于管理并发任务的执行
type TaskGroup interface {
	// Go 添加一个任务到任务组中异步执行
	// f 是要执行的任务函数，返回 error 表示任务执行结果
	Go(f func() error)
	// Wait 等待所有任务执行完成，并返回第一个遇到的错误（如果有）
	Wait() error
}

// taskGroup 是 TaskGroup 接口的实现
type taskGroup struct {
	errGroup    *errgroup.Group // 底层的 errgroup，用于管理并发任务
	ctx         context.Context // 上下文，用于任务取消和超时控制
	execAllTask atomic.Bool     // 是否执行所有任务的标志（即使有任务失败）
}

// NewTaskGroup 创建一个可中断的任务组
// 当某个任务返回错误时，会取消 context，其他未执行的任务将不会执行，正在执行的任务会收到取消信号
//
// 参数:
//   - ctx: 上下文，用于控制任务的生命周期
//   - concurrentCount: 最大并发数，限制同时执行的任务数量
//
// 返回: TaskGroup 实例
func NewTaskGroup(ctx context.Context, concurrentCount int) TaskGroup {
	t := &taskGroup{}
	t.errGroup, t.ctx = errgroup.WithContext(ctx)
	t.errGroup.SetLimit(concurrentCount)
	t.execAllTask.Store(false) // 设置为 false，表示遇到错误时中断其他任务

	return t
}

// NewUninterruptibleTaskGroup 创建一个不可中断的任务组
// 当某个任务返回错误时，其他任务会继续执行，直到所有任务完成
//
// 参数:
//   - ctx: 上下文，用于控制任务的生命周期
//   - concurrentCount: 最大并发数，限制同时执行的任务数量
//
// 返回: TaskGroup 实例
func NewUninterruptibleTaskGroup(ctx context.Context, concurrentCount int) TaskGroup {
	t := &taskGroup{}
	t.errGroup, t.ctx = errgroup.WithContext(ctx)
	t.errGroup.SetLimit(concurrentCount)
	t.execAllTask.Store(true) // 设置为 true，表示即使有任务失败也继续执行其他任务

	return t
}

// Go 将任务添加到任务组中异步执行
// 任务执行时会自动捕获 panic，并在日志中记录
// 对于可中断的任务组，如果 context 已被取消，任务会立即返回错误
func (t *taskGroup) Go(f func() error) {
	t.errGroup.Go(func() (err error) {
		// 捕获 panic，防止单个任务崩溃影响整个程序
		defer func() {
			if r := recover(); r != nil {
				logs.CtxErrorf(t.ctx, "[TaskGroup] exec panic recover:%+v", r)
				// 将 panic 转换为 error，确保 errgroup 能够正确处理
				err = fmt.Errorf("task panic: %v", r)
			}
		}()

		// 如果是可中断模式，检查 context 是否已取消
		// 如果已取消，直接返回错误，不执行任务
		if !t.execAllTask.Load() {
			select {
			case <-t.ctx.Done():
				return t.ctx.Err()
			default:
			}
		}

		// 执行用户提供的任务函数
		return f()
	})
}

// Wait 等待所有任务执行完成
// 返回第一个遇到的错误（如果有），如果所有任务都成功完成则返回 nil
func (t *taskGroup) Wait() error {
	return t.errGroup.Wait()
}
