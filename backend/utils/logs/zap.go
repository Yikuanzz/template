package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend/app/types/consts"
	"backend/utils/envx"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ANSI 颜色代码（参考 api_middleware.go）
const (
	colorReset = "\033[0m"
	colorWhite = "\033[37m"
	colorCyan  = "\033[36m"
	colorGray  = "\033[90m"

	// 背景色
	bgRed     = "\033[41m"
	bgGreen   = "\033[42m"
	bgYellow  = "\033[43m"
	bgBlue    = "\033[44m"
	bgCyan    = "\033[46m"
	bgMagenta = "\033[45m"
)

// zapLogger zap 实现的日志器
type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// GetLogger 返回底层 zap logger
func (z *zapLogger) GetLogger() interface{} {
	return z.logger
}

// Error 记录错误级别日志
func (z *zapLogger) Error(msg string, keyvals ...interface{}) {
	z.logger.Error(msg, z.parseKeyvals(keyvals...)...)
}

// Warn 记录警告级别日志
func (z *zapLogger) Warn(msg string, keyvals ...interface{}) {
	z.logger.Warn(msg, z.parseKeyvals(keyvals...)...)
}

// Info 记录信息级别日志
func (z *zapLogger) Info(msg string, keyvals ...interface{}) {
	z.logger.Info(msg, z.parseKeyvals(keyvals...)...)
}

// Debug 记录调试级别日志
func (z *zapLogger) Debug(msg string, keyvals ...interface{}) {
	z.logger.Debug(msg, z.parseKeyvals(keyvals...)...)
}

// CtxError 记录带上下文的错误级别日志
func (z *zapLogger) CtxError(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := append(extractTraceFields(ctx), z.parseKeyvals(keyvals...)...)
	z.logger.Error(msg, fields...)
}

// CtxWarn 记录带上下文的警告级别日志
func (z *zapLogger) CtxWarn(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := append(extractTraceFields(ctx), z.parseKeyvals(keyvals...)...)
	z.logger.Warn(msg, fields...)
}

// CtxInfo 记录带上下文的信息级别日志
func (z *zapLogger) CtxInfo(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := append(extractTraceFields(ctx), z.parseKeyvals(keyvals...)...)
	z.logger.Info(msg, fields...)
}

// CtxDebug 记录带上下文的调试级别日志
func (z *zapLogger) CtxDebug(ctx context.Context, msg string, keyvals ...interface{}) {
	fields := append(extractTraceFields(ctx), z.parseKeyvals(keyvals...)...)
	z.logger.Debug(msg, fields...)
}

// WithTraceFields 为 logger 添加追踪字段（用于降级处理）
func (z *zapLogger) WithTraceFields(ctx context.Context) *zap.SugaredLogger {
	fields := extractTraceFields(ctx)
	if len(fields) == 0 {
		return z.sugar
	}
	return z.logger.With(fields...).Sugar()
}

// parseKeyvals 将 key-value 对转换为 zap.Field
// 支持两种格式：
// 1. keyvals 是成对的 key-value: "key1", value1, "key2", value2
// 2. keyvals 是单个值: value
func (z *zapLogger) parseKeyvals(keyvals ...interface{}) []zap.Field {
	if len(keyvals) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keyvals)/2+1)

	// 如果是成对的 key-value
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			// 如果不是字符串 key，跳过这个键值对
			continue
		}
		value := keyvals[i+1]
		fields = append(fields, zap.Any(key, value))
	}

	// 如果 keyvals 是奇数个，最后一个值作为通用字段
	if len(keyvals)%2 == 1 {
		fields = append(fields, zap.Any("extra", keyvals[len(keyvals)-1]))
	}

	return fields
}

// contextKey 定义 context key 类型
type contextKey string

const (
	TraceIDContextKey      contextKey = "trace_id"
	SpanIDContextKey       contextKey = "span_id"
	ParentSpanIDContextKey contextKey = "parent_span_id"
)

// extractTraceFields 从 context 中提取追踪字段
// 支持从 context 中提取 trace_id、span_id 和 parent_span_id
// 同时支持类型化的 key 和字符串 key（向后兼容）
func extractTraceFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := make([]zap.Field, 0, 3)

	// 尝试从 context 中提取 trace_id（优先使用类型化的 key）
	var traceID interface{}
	if traceID = ctx.Value(TraceIDContextKey); traceID == nil {
		// 向后兼容：尝试字符串 key
		traceID = ctx.Value("trace_id")
	}
	if traceID != nil {
		if traceIDStr, ok := traceID.(string); ok && traceIDStr != "" {
			fields = append(fields, zap.String("trace_id", traceIDStr))
		}
	}

	// 尝试从 context 中提取 span_id（优先使用类型化的 key）
	var spanID interface{}
	if spanID = ctx.Value(SpanIDContextKey); spanID == nil {
		// 向后兼容：尝试字符串 key
		spanID = ctx.Value("span_id")
	}
	if spanID != nil {
		if spanIDStr, ok := spanID.(string); ok && spanIDStr != "" {
			fields = append(fields, zap.String("span_id", spanIDStr))
		}
	}

	// 尝试从 context 中提取 parent_span_id（优先使用类型化的 key）
	var parentSpanID interface{}
	if parentSpanID = ctx.Value(ParentSpanIDContextKey); parentSpanID == nil {
		// 向后兼容：尝试字符串 key
		parentSpanID = ctx.Value("parent_span_id")
	}
	if parentSpanID != nil {
		if parentSpanIDStr, ok := parentSpanID.(string); ok && parentSpanIDStr != "" {
			fields = append(fields, zap.String("parent_span_id", parentSpanIDStr))
		}
	}

	return fields
}

// newZapLogger 创建新的 zap logger
func newZapLogger() *zapLogger {
	var logger *zap.Logger

	// 读取配置
	logLevel := envx.GetStringOptional(consts.EnvLogLevel)
	logOutput := envx.GetStringOptional(consts.EnvLogOutput)
	logDevelopment := envx.GetBool(consts.EnvLogDevelopment, false)
	logFile := envx.GetStringOptional(consts.EnvLogFile)

	// 设置日志级别
	level := parseLogLevel(logLevel)
	zapLevel := zapcore.Level(level)

	// 如果设置了日志文件，使用 lumberjack 进行日志轮转
	var fileWriter zapcore.WriteSyncer
	if logFile != "" {
		// 确保日志目录存在
		if mkdirErr := os.MkdirAll(filepath.Dir(logFile), 0o755); mkdirErr == nil {
			// 配置日志轮转
			lumberjackLogger := &lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    getLogMaxSize(),    // 单个文件最大大小（MB）
				MaxBackups: getLogMaxBackups(), // 保留的旧文件数量
				MaxAge:     getLogMaxAge(),     // 保留天数
				Compress:   getLogCompress(),   // 是否压缩旧文件
			}
			fileWriter = zapcore.AddSync(lumberjackLogger)
		}
	}

	// 创建 logger
	var cores []zapcore.Core

	// 根据输出格式选择编码器
	var encoder zapcore.Encoder
	encoderConfig := getEncoderConfig(logDevelopment, getEncoding(logOutput) == "console")
	if getEncoding(logOutput) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建 stdout core
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapLevel)
	cores = append(cores, stdoutCore)

	// 如果配置了文件输出，创建文件 core（文件输出使用 JSON 格式，不使用颜色）
	if fileWriter != nil {
		fileEncoderConfig := getEncoderConfig(logDevelopment, false)
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, zapLevel)
		cores = append(cores, fileCore)
	}

	// 合并所有 cores
	core := zapcore.NewTee(cores...)

	// 构建 logger
	var options []zap.Option
	options = append(options, zap.AddCaller(), zap.AddCallerSkip(2))
	if logDevelopment {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	logger = zap.New(core, options...)

	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) int8 {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "debug":
		return int8(zapcore.DebugLevel)
	case "info":
		return int8(zapcore.InfoLevel)
	case "warn", "warning":
		return int8(zapcore.WarnLevel)
	case "error":
		return int8(zapcore.ErrorLevel)
	case "fatal":
		return int8(zapcore.FatalLevel)
	case "panic":
		return int8(zapcore.PanicLevel)
	default:
		return int8(zapcore.InfoLevel) // 默认 info
	}
}

// getEncoding 获取编码格式
func getEncoding(output string) string {
	output = strings.ToLower(strings.TrimSpace(output))
	switch output {
	case "json":
		return "json"
	case "console":
		return "console"
	default:
		// 如果环境变量未设置，检查是否在容器中运行（通过检查 /proc/1/cgroup）
		// 如果在容器中，默认使用 json；否则使用 console
		if isRunningInContainer() {
			return "json"
		}
		return "console"
	}
}

// isRunningInContainer 检查是否在容器中运行
func isRunningInContainer() bool {
	// 检查常见的容器标识文件
	containerFiles := []string{
		"/.dockerenv",
		"/proc/1/cgroup",
	}
	for _, file := range containerFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	return false
}

// getLogMaxSize 获取日志文件最大大小（MB）
func getLogMaxSize() int {
	maxSize := envx.GetStringOptional(consts.EnvLogMaxSize)
	if maxSize == "" {
		return 100 // 默认 100MB
	}
	if size, err := strconv.Atoi(maxSize); err == nil && size > 0 {
		return size
	}
	return 100
}

// getLogMaxBackups 获取保留的旧日志文件数量
func getLogMaxBackups() int {
	maxBackups := envx.GetStringOptional(consts.EnvLogMaxBackups)
	if maxBackups == "" {
		return 7 // 默认保留 7 个文件
	}
	if backups, err := strconv.Atoi(maxBackups); err == nil && backups >= 0 {
		return backups
	}
	return 7
}

// getLogMaxAge 获取日志文件保留天数
func getLogMaxAge() int {
	maxAge := envx.GetStringOptional(consts.EnvLogMaxAge)
	if maxAge == "" {
		return 30 // 默认保留 30 天
	}
	if age, err := strconv.Atoi(maxAge); err == nil && age > 0 {
		return age
	}
	return 30
}

// getLogCompress 获取是否压缩旧日志文件
func getLogCompress() bool {
	return envx.GetBool(consts.EnvLogCompress, true) // 默认压缩
}

// getEncoderConfig 获取编码器配置
func getEncoderConfig(development bool, isConsole bool) zapcore.EncoderConfig {
	if isConsole {
		// 控制台模式：使用优化的显示格式，带颜色和色块
		return zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    encodeLevelWithColorBlock, // 自定义：带色块的级别编码
			EncodeTime:     encodeTimeReadable,        // 自定义：可读时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   encodeCallerWithColor, // 自定义：带颜色的调用位置
		}
	}

	if development {
		// 开发模式：使用更详细的配置
		return zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	// 生产模式：使用 JSON 格式，符合 Loki 要求
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写，符合 Loki 要求
		EncodeTime:     zapcore.EpochTimeEncoder,      // Unix 时间戳（秒），符合 Promtail 配置
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// encodeTimeReadable 编码时间为可读格式（参考 api_middleware.go）
func encodeTimeReadable(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	timestamp := t.Format("2006/01/02 - 15:04:05")
	enc.AppendString(fmt.Sprintf("%s%s %s %s", bgCyan, colorWhite, timestamp, colorReset))
}

// encodeLevelWithColorBlock 编码日志级别，带色块（参考 api_middleware.go）
func encodeLevelWithColorBlock(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var bgColor string
	var levelText string

	switch level {
	case zapcore.DebugLevel:
		bgColor = bgCyan
		levelText = "DEBUG"
	case zapcore.InfoLevel:
		bgColor = bgGreen
		levelText = "INFO"
	case zapcore.WarnLevel:
		bgColor = bgYellow
		levelText = "WARN"
	case zapcore.ErrorLevel:
		bgColor = bgRed
		levelText = "ERROR"
	case zapcore.FatalLevel, zapcore.PanicLevel:
		bgColor = bgRed
		levelText = level.CapitalString()
	default:
		bgColor = ""
		levelText = level.CapitalString()
	}

	if bgColor != "" {
		enc.AppendString(fmt.Sprintf("%s%s %s %s", bgColor, colorWhite, levelText, colorReset))
	} else {
		enc.AppendString(levelText)
	}
}

// encodeCallerWithColor 编码调用位置，带颜色
func encodeCallerWithColor(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	callerStr := caller.String()
	enc.AppendString(fmt.Sprintf("%s%s%s", colorGray, callerStr, colorReset))
}
