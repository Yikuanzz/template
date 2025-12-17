package logs_test

import (
	"context"
	"fmt"

	"backend/utils/logs"
)

func Example_basic() {
	// 基本日志
	logs.Info("应用启动成功")
	logs.Error("连接失败")
}

func Example_structured() {
	// 结构化日志
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
}

func Example_context() {
	// 创建带追踪信息的 context
	ctx := context.WithValue(context.Background(), "trace_id", "trace-12345")
	ctx = context.WithValue(ctx, "span_id", "span-67890")

	// 使用格式化日志（兼容旧代码）
	logs.CtxInfof(ctx, "处理请求: %s", "/api/users")
	logs.CtxErrorf(ctx, "处理失败: %v", fmt.Errorf("some error"))

	// 使用结构化日志（推荐）
	logger := logs.GetDefaultLogger()
	if ctxLogger, ok := logger.(logs.CtxStructuredLogger); ok {
		ctxLogger.CtxInfo(ctx, "用户操作",
			"action", "create_order",
			"user_id", 12345,
			"order_id", "ORD-001",
		)
	}
}

func Example_interface() {
	// 使用接口进行结构化日志
	logger := logs.GetDefaultLogger()
	if structuredLogger, ok := logger.(logs.StructuredLogger); ok {
		structuredLogger.Info("操作成功",
			"operation", "update_user",
			"user_id", 12345,
		)
	}
}
