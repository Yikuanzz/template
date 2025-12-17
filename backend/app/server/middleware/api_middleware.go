package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ANSI 颜色代码
const (
	colorReset = "\033[0m"
	colorWhite = "\033[37m"

	// 背景色
	bgRed    = "\033[41m"
	bgGreen  = "\033[42m"
	bgYellow = "\033[43m"
	bgBlue   = "\033[44m"
	bgCyan   = "\033[46m"
)

// APILoggerConfig API 日志中间件配置
type APILoggerConfig struct {
	SkipPaths []string // 跳过的路径
}

// APILoggerMiddleware 创建自定义 API 日志中间件
func APILoggerMiddleware(config ...APILoggerConfig) gin.HandlerFunc {
	var cfg APILoggerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// 跳过指定路径
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		latencyStr := formatLatency(latency)

		// 获取状态码
		statusCode := c.Writer.Status()

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 如果有查询参数，添加到路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 格式化时间
		timestamp := time.Now().Format("2006/01/02 - 15:04:05")

		// 构建日志输出
		logLine := fmt.Sprintf("%s | %s | %s | %s | %s     \"%s\"",
			timestamp,
			colorizeStatusCode(statusCode),
			latencyStr,
			clientIP,
			colorizeMethod(method),
			path,
		)

		// 输出日志
		fmt.Println(logLine)
	}
}

// colorizeStatusCode 根据状态码返回带颜色的字符串
func colorizeStatusCode(code int) string {
	var bgColor string
	switch {
	case code >= 200 && code < 300:
		bgColor = bgGreen
	case code >= 300 && code < 400:
		bgColor = bgCyan
	case code >= 400 && code < 500:
		bgColor = bgYellow
	case code >= 500:
		bgColor = bgRed
	default:
		bgColor = ""
	}

	return fmt.Sprintf("%s%s %3d %s", bgColor, colorWhite, code, colorReset)
}

// colorizeMethod 根据 HTTP 方法返回带颜色的字符串
func colorizeMethod(method string) string {
	var bgColor string
	switch method {
	case "GET":
		bgColor = bgBlue
	case "POST":
		bgColor = bgGreen
	case "PUT":
		bgColor = bgYellow
	case "DELETE":
		bgColor = bgRed
	case "PATCH":
		bgColor = bgCyan
	default:
		bgColor = ""
	}

	return fmt.Sprintf("%s%s %s %s", bgColor, colorWhite, method, colorReset)
}

// formatLatency 格式化耗时
func formatLatency(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%8.4fns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%8.4fµs", float64(d.Nanoseconds())/1000.0)
	} else {
		return fmt.Sprintf("%8.4fms", float64(d.Nanoseconds())/1000000.0)
	}
}
