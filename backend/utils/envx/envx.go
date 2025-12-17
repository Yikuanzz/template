package envx

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetString 从环境变量读取字符串（必需）
// 如果环境变量不存在或为空，返回错误
func GetString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("环境变量 %s 未配置或为空", key)
	}
	return value, nil
}

// GetStringOptional 从环境变量读取字符串（可选）
// 如果环境变量不存在或为空，返回空字符串
func GetStringOptional(key string) string {
	return os.Getenv(key)
}

// GetInt 从环境变量读取整数（必需）
// 如果环境变量不存在、为空或解析失败，返回错误
func GetInt(key string) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("环境变量 %s 未配置或为空", key)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("解析环境变量 %s 失败: %w", key, err)
	}
	return value, nil
}

// GetIntWithDefault 从环境变量读取整数（可选，带默认值）
// 如果环境变量不存在或为空，返回默认值
// 如果解析失败，返回错误
func GetIntWithDefault(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("解析环境变量 %s 失败: %w", key, err)
	}
	return value, nil
}

// GetIntWithDefaultAndMin 从环境变量读取整数（可选，带默认值和最小值校验）
// 如果环境变量不存在或为空，返回默认值
// 如果解析失败或值小于最小值，返回错误
func GetIntWithDefaultAndMin(key string, defaultValue int, minValue int) (int, error) {
	value, err := GetIntWithDefault(key, defaultValue)
	if err != nil {
		return 0, err
	}

	if value < minValue {
		return 0, fmt.Errorf("环境变量 %s 的值 %d 小于最小值 %d", key, value, minValue)
	}
	return value, nil
}

// GetBool 从环境变量读取布尔值（可选，带默认值）
// 支持的值：true, 1, yes, on（不区分大小写）
// 如果环境变量不存在或为空，返回默认值
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// 支持多种布尔值表示方式
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes" || value == "on"
}

// GetDuration 从环境变量读取时间 Duration（必需）
// 支持格式：30s, 1m, 5m, 1h 等，也支持纯数字（作为秒数）
// 如果环境变量不存在、为空或解析失败，返回错误
func GetDuration(key string) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("环境变量 %s 未配置或为空", key)
	}

	// 先尝试作为 Duration 解析（支持 30s, 1m 等格式）
	duration, err := time.ParseDuration(valueStr)
	if err == nil {
		return duration, nil
	}

	// 如果解析失败，尝试作为秒数解析
	seconds, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("解析环境变量 %s 失败，期望格式为 Duration（如 30s, 1m）或秒数: %w", key, err)
	}

	return time.Duration(seconds) * time.Second, nil
}

// GetDurationWithDefault 从环境变量读取时间 Duration（可选，带默认值）
// 支持格式：30s, 1m, 5m, 1h 等，也支持纯数字（作为秒数）
// 如果环境变量不存在或为空，返回默认值
// 如果解析失败，返回错误
func GetDurationWithDefault(key string, defaultValue time.Duration) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	// 先尝试作为 Duration 解析（支持 30s, 1m 等格式）
	duration, err := time.ParseDuration(valueStr)
	if err == nil {
		return duration, nil
	}

	// 如果解析失败，尝试作为秒数解析
	seconds, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("解析环境变量 %s 失败，期望格式为 Duration（如 30s, 1m）或秒数: %w", key, err)
	}

	return time.Duration(seconds) * time.Second, nil
}

// GetDurationFromSeconds 从环境变量读取秒数并转换为 Duration（可选，带默认值）
// 如果环境变量不存在或为空，返回默认值
// 如果解析失败或值小于等于0，返回错误
func GetDurationFromSeconds(key string, defaultValue time.Duration) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	seconds, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("解析环境变量 %s 失败: %w", key, err)
	}

	if seconds <= 0 {
		return 0, fmt.Errorf("环境变量 %s 的值 %d 必须大于 0", key, seconds)
	}

	return time.Duration(seconds) * time.Second, nil
}

// GetStringSlice 从环境变量读取字符串切片（逗号分隔）
// 如果环境变量不存在或为空，返回空切片
// 会自动去除每个元素的前后空格
func GetStringSlice(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}

	// 去除所有空格后按逗号分割
	value = strings.ReplaceAll(value, " ", "")
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// GetStringSliceRequired 从环境变量读取字符串切片（必需，逗号分隔）
// 如果环境变量不存在、为空或解析后为空，返回错误
func GetStringSliceRequired(key string) ([]string, error) {
	value := os.Getenv(key)
	if value == "" {
		return nil, fmt.Errorf("环境变量 %s 未配置或为空", key)
	}

	// 去除所有空格后按逗号分割
	value = strings.ReplaceAll(value, " ", "")
	if value == "" {
		return nil, fmt.Errorf("环境变量 %s 格式错误", key)
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("环境变量 %s 解析后为空", key)
	}

	return result, nil
}
