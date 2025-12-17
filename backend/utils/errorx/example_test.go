package errorx_test

import (
	"errors"
	"fmt"
	"testing"

	"backend/utils/errorx"
)

// 定义错误码常量
const (
	ErrPermissionDenied = int32(1000000)
	ErrInvalidParam     = int32(1000001)
	ErrNotFound         = int32(1000002)
)

func init() {
	// 注册错误码
	errorx.Register(ErrPermissionDenied, "unauthorized access: {reason}")
	errorx.Register(ErrInvalidParam, "invalid parameter: {param}")
	errorx.Register(ErrNotFound, "resource not found: {resource}")
}

func ExampleNew() {
	// 基本用法：使用注册的错误码
	err := errorx.New(ErrPermissionDenied)
	fmt.Println(errorx.ErrorWithoutStack(err))
	// Output: code=1000000 message=unauthorized access: {reason}
}

func ExampleNew_withKeyValue() {
	// 使用键值对替换占位符
	err := errorx.New(ErrPermissionDenied, errorx.K("reason", "insufficient permissions"))
	fmt.Println(errorx.ErrorWithoutStack(err))
	// Output: code=1000000 message=unauthorized access: insufficient permissions
}

func ExampleNew_withMessage() {
	// 直接提供消息
	err := errorx.New(ErrInvalidParam, "参数不能为空")
	fmt.Println(errorx.ErrorWithoutStack(err))
	// Output: code=1000001 message=参数不能为空
}

func ExampleWrap() {
	// 包装标准错误
	originalErr := errors.New("database connection failed")
	err := errorx.Wrap(originalErr, ErrInvalidParam, errorx.K("param", "database_url"))

	var statusErr errorx.StatusError
	if errors.As(err, &statusErr) {
		fmt.Printf("Code: %d\n", statusErr.Code())
		fmt.Printf("Message: %s\n", statusErr.Msg())
		fmt.Printf("Cause: %v\n", statusErr.Unwrap())
	}
}

func ExampleWrapf() {
	// 使用格式化消息包装错误
	originalErr := errors.New("connection timeout")
	err := errorx.Wrapf(originalErr, "failed to connect to %s", "localhost:8080")

	fmt.Println(err.Error())
}

func TestNew(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := errorx.New(ErrPermissionDenied)
		var statusErr errorx.StatusError
		if !errors.As(err, &statusErr) {
			t.Fatal("expected StatusError")
		}
		if statusErr.Code() != ErrPermissionDenied {
			t.Errorf("expected code %d, got %d", ErrPermissionDenied, statusErr.Code())
		}
	})

	t.Run("with key-value", func(t *testing.T) {
		err := errorx.New(ErrPermissionDenied, errorx.K("reason", "test"))
		var statusErr errorx.StatusError
		if !errors.As(err, &statusErr) {
			t.Fatal("expected StatusError")
		}
		expected := "unauthorized access: test"
		if statusErr.Msg() != expected {
			t.Errorf("expected message %q, got %q", expected, statusErr.Msg())
		}
	})

	t.Run("with direct message", func(t *testing.T) {
		err := errorx.New(ErrInvalidParam, "custom message")
		var statusErr errorx.StatusError
		if !errors.As(err, &statusErr) {
			t.Fatal("expected StatusError")
		}
		if statusErr.Msg() != "custom message" {
			t.Errorf("expected message %q, got %q", "custom message", statusErr.Msg())
		}
	})
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := errorx.Wrap(originalErr, ErrNotFound, errorx.K("resource", "user"))

	var statusErr errorx.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatal("expected StatusError")
	}

	if statusErr.Unwrap() != originalErr {
		t.Error("unwrap should return original error")
	}

	if !errors.Is(err, originalErr) {
		t.Error("errors.Is should work")
	}
}

func TestErrorWithoutStack(t *testing.T) {
	err := errorx.New(ErrPermissionDenied, errorx.K("reason", "test"))
	msg := errorx.ErrorWithoutStack(err)
	expected := "code=1000000 message=unauthorized access: test"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}
