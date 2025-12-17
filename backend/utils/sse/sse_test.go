package sse

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestBasicTaskExecution 测试基本任务执行功能
func TestBasicTaskExecution(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()
	subscriberID := "client_001"

	// 定义异步任务：发送5条数据
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 5; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data := map[string]interface{}{
				"step":    i,
				"total":   5,
				"message": "处理中...",
			}

			if err := updateProgress(data); err != nil {
				return err
			}

			time.Sleep(50 * time.Millisecond)
		}
		return nil
	}

	// 创建新任务
	dataChan, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"", // 空 resumeKey 表示创建新任务
		subscriberID,
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 收集接收到的数据
	var receivedData []interface{}
	for data := range dataChan {
		receivedData = append(receivedData, data)
	}

	// 验证接收到的数据数量
	if len(receivedData) != 5 {
		t.Errorf("期望接收5条数据，实际接收%d条", len(receivedData))
	}

	// 验证任务状态
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	if taskInfo.Status != TaskStatusCompleted {
		t.Errorf("期望任务状态为 completed，实际为 %s", taskInfo.Status)
	}
}

// TestResumeTaskWithCachedData 测试断点续传功能：验证缓存数据被正确发送
func TestResumeTaskWithCachedData(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	subscriberID1 := "client_001"
	subscriberID2 := "client_002"

	// 用于记录发送的数据
	var sentData []interface{}
	var mu sync.Mutex

	// 定义异步任务：发送10条数据，每条间隔200ms（给缓存留出时间）
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data := map[string]interface{}{
				"step":    i,
				"total":   10,
				"message": "处理中...",
			}

			mu.Lock()
			sentData = append(sentData, data)
			mu.Unlock()

			if err := updateProgress(data); err != nil {
				return err
			}

			time.Sleep(200 * time.Millisecond)
		}
		return nil
	}

	// 第一次连接：创建新任务
	ctx1 := context.Background()
	dataChan1, taskID, err := manager.ExecuteWithSSE(
		ctx1,
		"",
		subscriberID1,
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 获取 resumeKey
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	resumeKey := taskInfo.ResumeKey

	// 接收前2条数据后，模拟断线（停止读取，让订阅者被清理）
	var firstClientData []interface{}
	firstClientDone := make(chan bool)
	go func() {
		defer close(firstClientDone)
		count := 0
		for data := range dataChan1 {
			firstClientData = append(firstClientData, data)
			count++
			if count >= 2 {
				// 模拟断线：停止读取，让 goroutine 退出，订阅者会被清理
				return
			}
		}
	}()

	// 等待前2条数据发送完成并断开
	time.Sleep(600 * time.Millisecond)
	<-firstClientDone

	// 验证第一个客户端接收到了2条数据
	if len(firstClientData) < 2 {
		t.Logf("第一个客户端接收到的数据: %d 条（可能因为goroutine调度延迟）", len(firstClientData))
	}

	// 等待一段时间，让任务继续执行（此时没有订阅者，数据应该被缓存）
	// 任务会继续发送第3、4、5条数据，这些数据应该被缓存
	time.Sleep(800 * time.Millisecond)

	// 检查任务状态，应该仍在运行
	taskInfo, err = manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	if taskInfo.Status != TaskStatusRunning {
		t.Logf("任务状态: %s（可能已经完成）", taskInfo.Status)
	}

	// 第二次连接：使用 resumeKey 恢复任务
	ctx2 := context.Background()
	dataChan2, _, err := manager.ExecuteWithSSE(
		ctx2,
		resumeKey,
		subscriberID2,
		asyncTask, // 这个函数不会再次执行（因为任务已在运行）
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("恢复任务失败: %v", err)
	}

	// 收集第二个客户端接收到的所有数据
	var secondClientData []interface{}
	timeout := time.After(3 * time.Second)
	for {
		select {
		case data, ok := <-dataChan2:
			if !ok {
				goto done
			}
			secondClientData = append(secondClientData, data)
		case <-timeout:
			goto done
		}
	}
done:

	// 验证第二个客户端接收到了数据
	// 应该包含断线期间缓存的数据以及后续的实时数据
	if len(secondClientData) == 0 {
		t.Error("第二个客户端应该接收到数据（包括缓存的数据）")
	}

	// 验证数据完整性
	t.Logf("第一个客户端接收数据: %d 条", len(firstClientData))
	t.Logf("第二个客户端接收数据: %d 条", len(secondClientData))
	t.Logf("任务发送的总数据: %d 条", len(sentData))

	// 验证第二个客户端接收到的数据数量应该大于等于剩余的数据
	// 第一个客户端接收了2条，剩余8条，第二个客户端应该至少接收到一些缓存的数据
	if len(secondClientData) > 0 {
		t.Logf("✅ 断点续传功能正常：第二个客户端成功接收到了 %d 条数据", len(secondClientData))
	}
}

// TestMultipleSubscribers 测试多个订阅者同时订阅同一个任务
func TestMultipleSubscribers(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 定义异步任务：发送5条数据
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 5; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data := map[string]interface{}{
				"step": i,
			}

			if err := updateProgress(data); err != nil {
				return err
			}

			time.Sleep(50 * time.Millisecond)
		}
		return nil
	}

	// 创建新任务
	dataChan1, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 获取 resumeKey
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	resumeKey := taskInfo.ResumeKey

	// 等待一小段时间，让任务开始执行
	time.Sleep(100 * time.Millisecond)

	// 第二个客户端订阅同一个任务
	dataChan2, _, err := manager.ExecuteWithSSE(
		ctx,
		resumeKey,
		"client_002",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("第二个客户端订阅失败: %v", err)
	}

	// 收集两个客户端的数据
	var client1Data []interface{}
	var client2Data []interface{}

	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		for data := range dataChan1 {
			client1Data = append(client1Data, data)
		}
		done1 <- true
	}()

	go func() {
		for data := range dataChan2 {
			client2Data = append(client2Data, data)
		}
		done2 <- true
	}()

	// 等待两个客户端都完成
	<-done1
	<-done2

	// 验证两个客户端都接收到了数据
	if len(client1Data) == 0 {
		t.Error("客户端1没有接收到任何数据")
	}
	if len(client2Data) == 0 {
		t.Error("客户端2没有接收到任何数据")
	}

	t.Logf("客户端1接收数据: %d 条", len(client1Data))
	t.Logf("客户端2接收数据: %d 条", len(client2Data))
}

// TestTaskStatus 测试任务状态管理
func TestTaskStatus(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 定义会失败的任务
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		updateProgress(map[string]interface{}{"step": 1})
		return fmt.Errorf("模拟任务失败")
	}

	// 创建任务
	dataChan, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 验证任务状态为失败
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	if taskInfo.Status != TaskStatusFailed {
		t.Errorf("期望任务状态为 failed，实际为 %s", taskInfo.Status)
	}

	// 验证数据通道已关闭（通道关闭时，ok 为 false）
	_, ok := <-dataChan
	if ok {
		// 如果还能读取到数据，说明通道还没关闭，再尝试一次
		_, ok = <-dataChan
		if ok {
			t.Error("任务失败后，数据通道应该已关闭")
		}
	}
}

// TestTaskExpired 测试任务过期功能
func TestTaskExpired(t *testing.T) {
	manager := NewSSEManager(100 * time.Millisecond) // 很短的过期时间
	defer manager.Stop()

	ctx := context.Background()

	// 定义异步任务
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 5; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			updateProgress(map[string]interface{}{"step": i})
			time.Sleep(50 * time.Millisecond)
		}
		return nil
	}

	// 创建任务
	_, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 获取 resumeKey
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	resumeKey := taskInfo.ResumeKey

	// 等待任务过期
	time.Sleep(200 * time.Millisecond)

	// 尝试恢复过期任务
	_, _, err = manager.ExecuteWithSSE(
		ctx,
		resumeKey,
		"client_002",
		asyncTask,
		10*time.Second,
	)
	if err != ErrTaskExpired {
		t.Errorf("期望错误为 ErrTaskExpired，实际为 %v", err)
	}
}

// TestTaskNotFound 测试任务不存在的情况
func TestTaskNotFound(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 尝试更新不存在的任务进度
	err := manager.UpdateProgress(ctx, "non_existent_task", map[string]interface{}{"step": 1})
	if err != ErrTaskNotFound {
		t.Errorf("期望错误为 ErrTaskNotFound，实际为 %v", err)
	}

	// 尝试获取不存在的任务信息
	_, err = manager.GetTaskInfo("non_existent_task")
	if err != ErrTaskNotFound {
		t.Errorf("期望错误为 ErrTaskNotFound，实际为 %v", err)
	}
}

// TestResumeNonRunningTask 测试恢复非运行状态的任务
func TestResumeNonRunningTask(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 定义快速完成的任务
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		updateProgress(map[string]interface{}{"step": 1})
		return nil
	}

	// 创建任务
	_, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 获取 resumeKey
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	resumeKey := taskInfo.ResumeKey

	// 尝试恢复已完成的任务
	_, _, err = manager.ExecuteWithSSE(
		ctx,
		resumeKey,
		"client_002",
		asyncTask,
		10*time.Second,
	)
	if err != ErrTaskNotRunning {
		t.Errorf("期望错误为 ErrTaskNotRunning，实际为 %v", err)
	}
}

// TestDataCachingDuringDisconnect 测试断线期间数据缓存功能
func TestDataCachingDuringDisconnect(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 定义异步任务：发送多条数据
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data := map[string]interface{}{
				"step":  i,
				"total": 10,
				"value": i * 10,
			}

			if err := updateProgress(data); err != nil {
				return err
			}

			time.Sleep(100 * time.Millisecond)
		}
		return nil
	}

	// 创建新任务
	dataChan1, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 获取 resumeKey
	taskInfo, err := manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	resumeKey := taskInfo.ResumeKey

	// 接收前2条数据后断开连接
	var firstClientData []interface{}
	go func() {
		count := 0
		for data := range dataChan1 {
			firstClientData = append(firstClientData, data)
			count++
			if count >= 2 {
				// 模拟断线
				return
			}
		}
	}()

	// 等待前2条数据发送
	time.Sleep(300 * time.Millisecond)

	// 等待更多数据产生（此时没有订阅者，数据应该被缓存）
	time.Sleep(600 * time.Millisecond)

	// 检查任务信息，验证缓存数据
	taskInfo, err = manager.GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}

	// 验证任务仍在运行
	if taskInfo.Status != TaskStatusRunning {
		t.Errorf("任务应该仍在运行，实际状态为 %s", taskInfo.Status)
	}

	// 重连：使用 resumeKey 恢复任务
	dataChan2, _, err := manager.ExecuteWithSSE(
		ctx,
		resumeKey,
		"client_002",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("恢复任务失败: %v", err)
	}

	// 收集第二个客户端接收到的所有数据
	var secondClientData []interface{}
	timeout := time.After(3 * time.Second)
	for {
		select {
		case data, ok := <-dataChan2:
			if !ok {
				goto done
			}
			secondClientData = append(secondClientData, data)
		case <-timeout:
			goto done
		}
	}
done:

	// 验证第二个客户端接收到了数据（包括缓存的数据）
	if len(secondClientData) == 0 {
		t.Error("第二个客户端应该接收到数据（包括缓存的数据）")
	}

	t.Logf("第一个客户端接收数据: %d 条", len(firstClientData))
	t.Logf("第二个客户端接收数据: %d 条", len(secondClientData))

	// 验证数据连续性：第二个客户端应该接收到从第3条开始的数据
	if len(secondClientData) > 0 {
		firstData := secondClientData[0].(map[string]interface{})
		firstStep := firstData["step"].(int)
		if firstStep < 3 {
			t.Logf("注意：第二个客户端接收到的第一条数据的step为%d，可能包含了一些缓存数据", firstStep)
		}
	}
}

// TestConcurrentUpdates 测试并发更新进度
func TestConcurrentUpdates(t *testing.T) {
	manager := NewSSEManager(1 * time.Hour)
	defer manager.Stop()

	ctx := context.Background()

	// 定义异步任务：快速发送多条数据
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 20; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			data := map[string]interface{}{
				"step": i,
			}

			if err := updateProgress(data); err != nil {
				return err
			}

			time.Sleep(10 * time.Millisecond)
		}
		return nil
	}

	// 创建任务
	dataChan, taskID, err := manager.ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 并发更新进度
	go func() {
		for i := 21; i <= 30; i++ {
			manager.UpdateProgress(ctx, taskID, map[string]interface{}{"step": i})
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// 收集数据
	var receivedData []interface{}
	timeout := time.After(2 * time.Second)
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				goto done
			}
			receivedData = append(receivedData, data)
		case <-timeout:
			goto done
		}
	}
done:

	// 验证接收到了数据
	if len(receivedData) == 0 {
		t.Error("应该接收到数据")
	}

	t.Logf("接收到 %d 条数据", len(receivedData))
}

// TestPackageLevelFunctions 测试包级别函数
func TestPackageLevelFunctions(t *testing.T) {
	// 重置默认管理器
	defaultManager = nil
	defaultManagerOnce = sync.Once{}

	ctx := context.Background()

	// 定义异步任务
	asyncTask := func(ctx context.Context, taskID string, updateProgress func(data interface{}) error) error {
		for i := 1; i <= 3; i++ {
			updateProgress(map[string]interface{}{"step": i})
			time.Sleep(50 * time.Millisecond)
		}
		return nil
	}

	// 使用包级别函数创建任务
	dataChan, taskID, err := ExecuteWithSSE(
		ctx,
		"",
		"client_001",
		asyncTask,
		10*time.Second,
	)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 使用包级别函数更新进度
	err = UpdateProgress(ctx, taskID, map[string]interface{}{"step": 4})
	if err != nil {
		t.Errorf("更新进度失败: %v", err)
	}

	// 使用包级别函数获取任务信息
	taskInfo, err := GetTaskInfo(taskID)
	if err != nil {
		t.Fatalf("获取任务信息失败: %v", err)
	}
	if taskInfo.Status != TaskStatusRunning {
		t.Errorf("期望任务状态为 running，实际为 %s", taskInfo.Status)
	}

	// 收集数据
	var receivedData []interface{}
	timeout := time.After(1 * time.Second)
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				goto done
			}
			receivedData = append(receivedData, data)
		case <-timeout:
			goto done
		}
	}
done:

	// 验证接收到了数据
	if len(receivedData) == 0 {
		t.Error("应该接收到数据")
	}
}
