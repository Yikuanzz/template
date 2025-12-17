package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// GenerateUID 生成一个唯一的短UID
// 返回格式：时间戳(压缩) + 随机数，使用 base64 URL 编码，长度约 14 字符
func GenerateUID() (string, error) {
	// 获取当前时间戳（毫秒级）
	timestamp := time.Now().UnixMilli()

	// 生成 6 字节随机数
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// 组合时间戳（4字节）和随机数（6字节）= 10字节
	data := make([]byte, 10)
	// 时间戳转换为4字节（大端序）
	data[0] = byte(timestamp >> 24)
	data[1] = byte(timestamp >> 16)
	data[2] = byte(timestamp >> 8)
	data[3] = byte(timestamp)
	// 复制随机数
	copy(data[4:], randomBytes)

	// 使用 base64 URL 编码（无填充，更短）
	uid := base64.RawURLEncoding.EncodeToString(data)

	return uid, nil
}

// MustGenerateUID 生成一个唯一的短UID，如果失败会 panic
// 适用于确定不会失败的场景
func MustGenerateUID() string {
	uid, err := GenerateUID()
	if err != nil {
		panic(fmt.Sprintf("生成UID失败: %v", err))
	}
	return uid
}

// MustGenerateUIDWithPrefix 生成一个唯一的短UID，如果失败会 panic
func MustGenerateUIDWithPrefix(prefix string) string {
	return prefix + MustGenerateUID()
}

// GenTraceID 生成一个唯一的 trace_id
func GenTraceID() string {
	return MustGenerateUIDWithPrefix("trace_")
}

// GenSpanID 生成一个唯一的 span_id
func GenSpanID() string {
	return MustGenerateUIDWithPrefix("span_")
}
