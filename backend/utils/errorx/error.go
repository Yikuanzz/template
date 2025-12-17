package errorx

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// StatusError 表示带状态码的错误
type StatusError interface {
	error
	Code() int32   // 错误码
	Msg() string   // 错误消息
	Unwrap() error // 返回被包装的原始错误
}

// statusError 实现 StatusError 接口
type statusError struct {
	code    int32
	msg     string
	cause   error
	stack   []uintptr
	callers []string
}

// Error 实现 error 接口
func (e *statusError) Error() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("code=%d", e.code))
	parts = append(parts, fmt.Sprintf("message=%s", e.msg))

	if e.cause != nil {
		parts = append(parts, fmt.Sprintf("cause=%s", e.cause.Error()))
	}

	if len(e.callers) > 0 {
		parts = append(parts, fmt.Sprintf("stack=%s", strings.Join(e.callers, "\n")))
	}

	return strings.Join(parts, " ")
}

// Code 返回错误码
func (e *statusError) Code() int32 {
	return e.code
}

// Msg 返回错误消息
func (e *statusError) Msg() string {
	return e.msg
}

// Unwrap 返回被包装的原始错误
func (e *statusError) Unwrap() error {
	return e.cause
}

// New 创建新的错误
// code: 错误码
// args: 可选参数，支持以下类型：
//   - string: 直接作为错误消息
//   - KV: 键值对，用于替换消息模板中的占位符
//   - error: 被包装的原始错误
func New(code int32, args ...interface{}) error {
	err := &statusError{
		code: code,
	}

	// 解析参数
	var msg string
	var cause error
	var kvs map[string]string

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			msg = v
		case KV:
			if kvs == nil {
				kvs = make(map[string]string)
			}
			kvs[v.Key] = v.Value
		case error:
			cause = v
		}
	}

	// 获取错误消息
	if msg == "" {
		msg = getMessage(code, kvs)
	} else if len(kvs) > 0 {
		// 如果提供了消息和键值对，使用键值对替换消息中的占位符
		msg = replacePlaceholders(msg, kvs)
	}

	err.msg = msg
	err.cause = cause

	// 捕获堆栈（如果错误还没有堆栈）
	if !hasStack(cause) {
		err.stack = captureStack(2)
		err.callers = formatStack(err.stack)
	}

	return err
}

// Wrap 包装现有错误
// err: 要包装的错误
// code: 错误码
// args: 可选参数，同 New 函数
func Wrap(err error, code int32, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// 如果已经是 StatusError，直接返回
	if _, ok := err.(StatusError); ok {
		return err
	}

	// 创建新的错误并包装原始错误
	args = append(args, err)
	return New(code, args...)
}

// Wrapf 使用格式化消息包装错误
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	return Wrap(err, 0, msg)
}

// KV 键值对，用于替换消息模板中的占位符
type KV struct {
	Key   string
	Value string
}

// K 创建键值对
func K(key, value string) KV {
	return KV{Key: key, Value: value}
}

// Kf 使用格式化字符串创建键值对
func Kf(key, format string, args ...interface{}) KV {
	return KV{Key: key, Value: fmt.Sprintf(format, args...)}
}

// ErrorWithoutStack 返回不包含堆栈信息的错误消息
func ErrorWithoutStack(err error) string {
	var statusErr StatusError
	if errors.As(err, &statusErr) {
		return fmt.Sprintf("code=%d message=%s", statusErr.Code(), statusErr.Msg())
	}
	return err.Error()
}

// ExtractRPCDesc 从 RPC 错误中提取 desc 部分
// RPC 错误格式: "rpc error: code = InvalidArgument desc = 用户名格式错误: 用户名包含敏感词"
// 返回 desc 后面的内容，如果无法提取则返回原始错误消息
func ExtractRPCDesc(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	// 查找 "desc = " 的位置
	descPrefix := "desc = "
	descIndex := strings.Index(errStr, descPrefix)
	if descIndex == -1 {
		// 如果没有找到 desc，返回原始错误消息
		return errStr
	}

	// 提取 desc 后面的内容
	descStart := descIndex + len(descPrefix)
	desc := strings.TrimSpace(errStr[descStart:])
	return desc
}

// getMessage 获取错误消息
func getMessage(code int32, kvs map[string]string) string {
	msg := getRegisteredMessage(code)
	if msg == "" {
		return fmt.Sprintf("unknown error code: %d", code)
	}

	// 替换占位符
	if len(kvs) > 0 {
		msg = replacePlaceholders(msg, kvs)
	}

	return msg
}

// replacePlaceholders 替换消息中的占位符 {key}
func replacePlaceholders(msg string, kvs map[string]string) string {
	result := msg
	for key, value := range kvs {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// hasStack 检查错误是否已经有堆栈信息
func hasStack(err error) bool {
	if err == nil {
		return false
	}
	var statusErr *statusError
	return errors.As(err, &statusErr) && len(statusErr.stack) > 0
}

// captureStack 捕获堆栈信息
func captureStack(skip int) []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	return pcs[0:n]
}

// formatStack 格式化堆栈信息
func formatStack(pcs []uintptr) []string {
	if len(pcs) == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs)
	var callers []string

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		callers = append(callers, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
	}

	return callers
}
