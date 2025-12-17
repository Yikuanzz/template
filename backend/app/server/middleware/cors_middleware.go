package middleware

import (
	"backend/utils/logs"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取请求的 Origin
		origin := c.GetHeader("Origin")

		// 设置允许的源（可以根据实际需求配置）
		// 这里允许所有源，生产环境建议配置具体的域名
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// 设置允许的请求方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

		// 设置允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

		// 设置允许暴露的响应头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// 设置是否允许携带凭证（Cookie等）
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// 设置预检请求的缓存时间（秒）
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			logs.CtxDebugf(ctx, "处理 OPTIONS 预检请求: origin=%s, path=%s", origin, c.Request.URL.Path)
			c.AbortWithStatus(204) // No Content
			return
		}

		// 继续处理请求
		c.Next()
	}
}
