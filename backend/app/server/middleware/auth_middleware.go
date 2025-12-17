package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/app/types/consts"
	authError "backend/app/types/errorn"
	"backend/app/types/meta"
	"backend/utils/envx"
	"backend/utils/errorx"
	"backend/utils/handle"
	"backend/utils/logs"
	"backend/utils/trace"

	"github.com/gin-gonic/gin"

	"backend/utils/secret"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := trace.InjectSpan(c.Request.Context())

		jwt := getJWT()

		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logs.CtxWarnf(ctx, "Authorization header 为空: path=%s, method=%s", c.Request.URL.Path, c.Request.Method)
			err := errorx.New(authError.AuthErrTokenRequired)
			handle.HandleErrorWithContext(c, err, "JWT 认证", &handle.ErrorConfig{
				DefaultStatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 去除 "Bearer " 前缀
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" || tokenString == authHeader {
			logs.CtxWarnf(ctx, "Token 格式错误，缺少 Bearer 前缀: path=%s, method=%s", c.Request.URL.Path, c.Request.Method)
			err := errorx.New(authError.AuthErrTokenMalformed, errorx.K("reason", "缺少 Bearer 前缀"))
			handle.HandleErrorWithContext(c, err, "JWT 认证", &handle.ErrorConfig{
				DefaultStatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 验证并解析 JWT token
		userInfo, err := jwt.ParseToken(tokenString)
		if err != nil {
			logs.CtxWarnf(ctx, "JWT Token 验证失败: error=%s, path=%s, method=%s", err.Error(), c.Request.URL.Path, c.Request.Method)

			// 根据错误类型返回不同的错误码
			var authErr error
			errStr := err.Error()
			if strings.Contains(errStr, "expired") || strings.Contains(errStr, "exp") {
				authErr = errorx.New(authError.AuthErrTokenExpired)
			} else if strings.Contains(errStr, "malformed") || strings.Contains(errStr, "invalid character") {
				authErr = errorx.New(authError.AuthErrTokenMalformed, errorx.K("reason", err.Error()))
			} else if strings.Contains(errStr, "signature") {
				authErr = errorx.New(authError.AuthErrTokenSignature)
			} else {
				authErr = errorx.New(authError.AuthErrTokenInvalid, errorx.K("reason", err.Error()))
			}

			handle.HandleErrorWithContext(c, authErr, "JWT 认证", &handle.ErrorConfig{
				DefaultStatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		ctx = context.WithValue(ctx, meta.ContextKeyUserID, userInfo.UserID)
		ctx = context.WithValue(ctx, meta.ContextKeyAccessToken, tokenString)
		c.Request = c.Request.WithContext(ctx)

		// 继续执行下一个中间件或处理器
		c.Next()
	}
}

// getJWT 获取 JWT 实例
func getJWT() *secret.JWT {
	accessTokenExpire, err := envx.GetDuration(consts.AccessTokenExpire)
	if err != nil {
		panic(err)
	}
	refreshTokenExpire, err := envx.GetDuration(consts.RefreshTokenExpire)
	if err != nil {
		panic(err)
	}
	jwtSecret, err := envx.GetString(consts.JWTSecret)
	if err != nil {
		panic(err)
	}

	return secret.NewJWT(secret.TokenConfig{
		AccessTokenExpire:  accessTokenExpire,
		RefreshTokenExpire: refreshTokenExpire,
		Secret:             jwtSecret,
	})
}
