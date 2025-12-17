package handle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"backend/utils/errorx"
	"backend/utils/logs"

	"github.com/gin-gonic/gin"
)

// ErrorConfig 错误处理配置
type ErrorConfig struct {
	// DefaultStatusCode 默认 HTTP 状态码（当错误不是 StatusError 时使用）
	DefaultStatusCode int
	// DefaultErrorCode 默认错误码（当错误不是 StatusError 时使用）
	DefaultErrorCode int32
	// LogLevel 日志级别，可选值: "warn", "error", "info", "debug"
	// 默认为 "warn"
	LogLevel string
}

// Response 统一响应结构体（用于 Swagger 文档）
type Response struct {
	Code    int32       `json:"code" example:"0"`                 // 响应码，0 表示成功
	Message string      `json:"message,omitempty" example:"操作成功"` // 响应消息（可选）
	Data    interface{} `json:"data,omitempty"`                   // 响应数据（可选）
}

// HandleError 统一处理错误并返回响应
// c: gin.Context
// err: 错误对象
// operation: 操作名称（用于日志记录）
// config: 错误处理配置（可选，如果为 nil 则使用默认配置）
func HandleError(c *gin.Context, err error, operation string, config *ErrorConfig) {
	if err == nil {
		return
	}

	// 使用默认配置
	if config == nil {
		config = &ErrorConfig{
			DefaultStatusCode: http.StatusBadRequest,
			DefaultErrorCode:  0,
			LogLevel:          "warn",
		}
	}

	// 检查是否是 StatusError 类型
	var statusErr errorx.StatusError
	if errors.As(err, &statusErr) {
		// 使用结构化日志记录
		logStructured(config.LogLevel, operation+"失败",
			"error_code", statusErr.Code(),
			"error_msg", statusErr.Msg(),
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// 使用配置的状态码，如果没有配置则使用默认的 BadRequest
		statusCode := config.DefaultStatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}

		// 返回 JSON 响应
		c.JSON(statusCode, gin.H{
			"code":    statusErr.Code(),
			"message": statusErr.Msg(),
		})
		return
	}

	// 处理普通错误
	logStructured(config.LogLevel, operation+"失败",
		"error", err.Error(),
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)

	statusCode := config.DefaultStatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	response := gin.H{
		"message": err.Error(),
	}
	if config.DefaultErrorCode > 0 {
		response["code"] = config.DefaultErrorCode
	}

	c.JSON(statusCode, response)
}

// HandleErrorWithContext 带上下文的错误处理
// 与 HandleError 相同，但会从 context 中提取额外信息记录日志
func HandleErrorWithContext(c *gin.Context, err error, operation string, config *ErrorConfig) {
	if err == nil {
		return
	}

	// 使用默认配置
	if config == nil {
		config = &ErrorConfig{
			DefaultStatusCode: http.StatusBadRequest,
			DefaultErrorCode:  0,
			LogLevel:          "warn",
		}
	}

	ctx := c.Request.Context()

	// 检查是否是 StatusError 类型
	var statusErr errorx.StatusError
	if errors.As(err, &statusErr) {
		// 使用结构化日志记录（带上下文）
		logStructuredWithContext(ctx, config.LogLevel, operation+"失败",
			"error_code", statusErr.Code(),
			"error_msg", statusErr.Msg(),
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// 使用配置的状态码，如果没有配置则使用默认的 BadRequest
		statusCode := config.DefaultStatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}

		// 返回 JSON 响应
		c.JSON(statusCode, gin.H{
			"code":    statusErr.Code(),
			"message": statusErr.Msg(),
		})
		return
	}

	// 处理普通错误
	logStructuredWithContext(ctx, config.LogLevel, operation+"失败",
		"error", err.Error(),
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)

	statusCode := config.DefaultStatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	response := gin.H{
		"message": err.Error(),
	}
	if config.DefaultErrorCode > 0 {
		response["code"] = config.DefaultErrorCode
	}

	c.JSON(statusCode, response)
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	response := gin.H{
		"code":    0,
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	c.JSON(http.StatusOK, response)
}

// logStructured 根据日志级别记录结构化日志
func logStructured(level string, msg string, keyvals ...interface{}) {
	logger := logs.GetDefaultLogger()
	if structuredLogger, ok := logger.(logs.StructuredLogger); ok {
		switch level {
		case "error":
			structuredLogger.Error(msg, keyvals...)
		case "info":
			structuredLogger.Info(msg, keyvals...)
		case "debug":
			structuredLogger.Debug(msg, keyvals...)
		default: // warn
			structuredLogger.Warn(msg, keyvals...)
		}
		return
	}
	// 降级到旧接口（兼容性）
	var parts []interface{}
	parts = append(parts, msg)
	if len(keyvals) > 0 {
		parts = append(parts, keyvals...)
	}
	switch level {
	case "error":
		logs.Error(parts...)
	case "info":
		logs.Info(parts...)
	case "debug":
		logs.Debug(parts...)
	default: // warn
		logs.Warn(parts...)
	}
}

// logStructuredWithContext 根据日志级别记录带上下文的结构化日志
func logStructuredWithContext(ctx context.Context, level string, msg string, keyvals ...interface{}) {
	logger := logs.GetDefaultLogger()
	if ctxLogger, ok := logger.(logs.CtxStructuredLogger); ok {
		switch level {
		case "error":
			ctxLogger.CtxError(ctx, msg, keyvals...)
		case "info":
			ctxLogger.CtxInfo(ctx, msg, keyvals...)
		case "debug":
			ctxLogger.CtxDebug(ctx, msg, keyvals...)
		default: // warn
			ctxLogger.CtxWarn(ctx, msg, keyvals...)
		}
		return
	}
	// 降级到格式化日志（兼容性）
	switch level {
	case "error":
		logs.CtxErrorf(ctx, msg+": %v", keyvals...)
	case "info":
		logs.CtxInfof(ctx, msg+": %v", keyvals...)
	case "debug":
		logs.CtxDebugf(ctx, msg+": %v", keyvals...)
	default: // warn
		logs.CtxWarnf(ctx, msg+": %v", keyvals...)
	}
}

// SSEConfig SSE 配置选项
type SSEConfig struct {
	EventName     string        // 事件名称，默认为 "message"
	EnablePing    bool          // 是否启用心跳，默认 true
	PingInterval  time.Duration // 心跳间隔，默认 30 秒
	RetryInterval int           // 客户端重连间隔（毫秒），默认 3000
	OnConnect     func()        // 连接建立时的回调
	OnDisconnect  func()        // 连接断开时的回调
	OnError       func(error)   // 发生错误时的回调
}

// DefaultSSEConfig 默认 SSE 配置
func DefaultSSEConfig() SSEConfig {
	return SSEConfig{
		EventName:     "message",
		EnablePing:    true,
		PingInterval:  30 * time.Second,
		RetryInterval: 3000,
	}
}

// StreamSSE 通用 SSE 流处理函数
// T 是数据类型，数据发送完成后会自动发送 "done" 事件
func StreamSSE[T any](c *gin.Context, dataChan <-chan T, config ...SSEConfig) {
	// 合并配置
	cfg := DefaultSSEConfig()
	if len(config) > 0 {
		cfg = config[0]
		if cfg.EventName == "" {
			cfg.EventName = "message"
		}
		if cfg.PingInterval == 0 {
			cfg.PingInterval = 30 * time.Second
		}
		if cfg.RetryInterval == 0 {
			cfg.RetryInterval = 3000
		}
	}

	// 设置 SSE 响应头（符合 SSE 规范）
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")          // 禁用 nginx 缓冲
	c.Header("Access-Control-Allow-Origin", "*") // 允许跨域（根据需要调整）

	// 发送重连间隔（SSE 规范：retry: milliseconds）
	fmt.Fprintf(c.Writer, "retry: %d\n\n", cfg.RetryInterval)
	c.Writer.Flush()

	// 连接建立回调
	if cfg.OnConnect != nil {
		cfg.OnConnect()
	}

	// 创建心跳 ticker
	var pingTicker *time.Ticker
	if cfg.EnablePing {
		pingTicker = time.NewTicker(cfg.PingInterval)
		defer pingTicker.Stop()
	}

	// 获取客户端断开信号
	ctx := c.Request.Context()
	clientGone := ctx.Done()
	notify := c.Writer.CloseNotify()

	// 发送 SSE 事件的辅助函数
	sendEvent := func(eventName, data string) bool {
		_, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventName, data)
		if err != nil {
			if cfg.OnError != nil {
				cfg.OnError(err)
			}
			return false
		}
		c.Writer.Flush()
		return true
	}

	// 发送心跳的辅助函数
	sendPing := func() bool {
		_, err := fmt.Fprintf(c.Writer, ": ping\n\n")
		if err != nil {
			if cfg.OnError != nil {
				cfg.OnError(err)
			}
			return false
		}
		c.Writer.Flush()
		return true
	}

	// 清理并返回的辅助函数
	cleanup := func() {
		if cfg.OnDisconnect != nil {
			cfg.OnDisconnect()
		}
	}

	// 主循环
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				// 通道已关闭，发送 done 事件后结束
				sendEvent("done", `{"status":"completed"}`)
				cleanup()
				return
			}

			// 序列化数据
			jsonData, err := json.Marshal(data)
			if err != nil {
				if cfg.OnError != nil {
					cfg.OnError(err)
				}
				continue
			}

			// 发送 SSE 事件（SSE 规范：event: name\ndata: data\n\n）
			if !sendEvent(cfg.EventName, string(jsonData)) {
				cleanup()
				return
			}

		case <-pingTicker.C:
			// 发送心跳（SSE 规范：注释消息用于心跳）
			if !sendPing() {
				cleanup()
				return
			}

		case <-clientGone:
			// 客户端断开连接
			cleanup()
			return

		case <-notify:
			// 连接被关闭
			cleanup()
			return
		}
	}
}

// SSE 简化版本，使用默认配置
func SSE[T any](c *gin.Context, dataChan <-chan T, eventName string) {
	cfg := DefaultSSEConfig()
	cfg.EventName = eventName
	StreamSSE(c, dataChan, cfg)
}
