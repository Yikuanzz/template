package logs

import (
	"context"
	"fmt"
)

// Logger 基础日志接口
type Logger interface {
	// 获取底层实现（用于类型断言）
	GetLogger() interface{}
}

// StructuredLogger 结构化日志接口
// 支持 key-value 格式的结构化日志
type StructuredLogger interface {
	Logger
	Error(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
}

// CtxStructuredLogger 带上下文的结构化日志接口
// 支持从 context 中提取 trace_id 和 span_id
type CtxStructuredLogger interface {
	StructuredLogger
	CtxError(ctx context.Context, msg string, keyvals ...interface{})
	CtxWarn(ctx context.Context, msg string, keyvals ...interface{})
	CtxInfo(ctx context.Context, msg string, keyvals ...interface{})
	CtxDebug(ctx context.Context, msg string, keyvals ...interface{})
}

var (
	defaultLogger Logger
)

// Init 初始化默认 logger
// 如果 logger 为 nil，会使用 zap 创建一个默认 logger
func Init(logger Logger) {
	if logger == nil {
		defaultLogger = newZapLogger()
		return
	}
	defaultLogger = logger
}

// GetDefaultLogger 获取默认 logger
func GetDefaultLogger() Logger {
	if defaultLogger == nil {
		defaultLogger = newZapLogger()
	}
	return defaultLogger
}

// 包级别的日志方法（兼容性接口）

// Error 记录错误级别日志
func Error(args ...interface{}) {
	logger := GetDefaultLogger()
	if structuredLogger, ok := logger.(StructuredLogger); ok {
		msg, keyvals := parseArgs(args...)
		structuredLogger.Error(msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.sugar.Error(args...)
	}
}

// Warn 记录警告级别日志
func Warn(args ...interface{}) {
	logger := GetDefaultLogger()
	if structuredLogger, ok := logger.(StructuredLogger); ok {
		msg, keyvals := parseArgs(args...)
		structuredLogger.Warn(msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.sugar.Warn(args...)
	}
}

// Info 记录信息级别日志
func Info(args ...interface{}) {
	logger := GetDefaultLogger()
	if structuredLogger, ok := logger.(StructuredLogger); ok {
		msg, keyvals := parseArgs(args...)
		structuredLogger.Info(msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.sugar.Info(args...)
	}
}

// Debug 记录调试级别日志
func Debug(args ...interface{}) {
	logger := GetDefaultLogger()
	if structuredLogger, ok := logger.(StructuredLogger); ok {
		msg, keyvals := parseArgs(args...)
		structuredLogger.Debug(msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.sugar.Debug(args...)
	}
}

// CtxErrorf 记录带上下文的错误级别日志（格式化）
func CtxErrorf(ctx context.Context, format string, args ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		msg := fmt.Sprintf(format, args...)
		ctxLogger.CtxError(ctx, msg)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.WithTraceFields(ctx).Errorf(format, args...)
	}
}

// CtxWarnf 记录带上下文的警告级别日志（格式化）
func CtxWarnf(ctx context.Context, format string, args ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		msg := fmt.Sprintf(format, args...)
		ctxLogger.CtxWarn(ctx, msg)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.WithTraceFields(ctx).Warnf(format, args...)
	}
}

// CtxInfof 记录带上下文的信息级别日志（格式化）
func CtxInfof(ctx context.Context, format string, args ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		msg := fmt.Sprintf(format, args...)
		ctxLogger.CtxInfo(ctx, msg)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.WithTraceFields(ctx).Infof(format, args...)
	}
}

// CtxDebugf 记录带上下文的调试级别日志（格式化）
func CtxDebugf(ctx context.Context, format string, args ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		msg := fmt.Sprintf(format, args...)
		ctxLogger.CtxDebug(ctx, msg)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		zapLogger.WithTraceFields(ctx).Debugf(format, args...)
	}
}

// CtxError 记录带上下文的错误级别日志（结构化）
func CtxError(ctx context.Context, msg string, keyvals ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		ctxLogger.CtxError(ctx, msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		fields := append(extractTraceFields(ctx), zapLogger.parseKeyvals(keyvals...)...)
		zapLogger.logger.Error(msg, fields...)
	}
}

// CtxWarn 记录带上下文的警告级别日志（结构化）
func CtxWarn(ctx context.Context, msg string, keyvals ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		ctxLogger.CtxWarn(ctx, msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		fields := append(extractTraceFields(ctx), zapLogger.parseKeyvals(keyvals...)...)
		zapLogger.logger.Warn(msg, fields...)
	}
}

// CtxInfo 记录带上下文的信息级别日志（结构化）
func CtxInfo(ctx context.Context, msg string, keyvals ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		ctxLogger.CtxInfo(ctx, msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		fields := append(extractTraceFields(ctx), zapLogger.parseKeyvals(keyvals...)...)
		zapLogger.logger.Info(msg, fields...)
	}
}

// CtxDebug 记录带上下文的调试级别日志（结构化）
func CtxDebug(ctx context.Context, msg string, keyvals ...interface{}) {
	logger := GetDefaultLogger()
	if ctxLogger, ok := logger.(CtxStructuredLogger); ok {
		ctxLogger.CtxDebug(ctx, msg, keyvals...)
		return
	}
	// 降级处理
	if zapLogger, ok := logger.(*zapLogger); ok {
		fields := append(extractTraceFields(ctx), zapLogger.parseKeyvals(keyvals...)...)
		zapLogger.logger.Debug(msg, fields...)
	}
}

// parseArgs 解析参数，将第一个参数作为 msg，其余作为 keyvals
func parseArgs(args ...interface{}) (msg string, keyvals []interface{}) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 1 {
		if msgStr, ok := args[0].(string); ok {
			return msgStr, nil
		}
		return "", args
	}
	if msgStr, ok := args[0].(string); ok {
		return msgStr, args[1:]
	}
	return "", args
}
