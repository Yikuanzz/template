package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"backend/app/internal/handler/user"
	"backend/app/server/middleware"
	"backend/app/server/router"
	"backend/app/types/consts"
	"backend/utils/envx"
	"backend/utils/logs"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// HTTPServerParams 定义 HTTP 服务器的依赖
type HTTPServerParams struct {
	fx.In
	Lifecycle   fx.Lifecycle
	UserHandler *user.UserHandler
}

// HTTPServer 创建 HTTP 服务器
func HTTPServer(params HTTPServerParams) *http.Server {
	// 设置 Gin Mode
	mode := envx.GetStringOptional(consts.GINMode)
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	// 禁用 Gin 框架的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	// 创建 Gin Engine
	r := gin.New()

	// 添加中间件（按顺序）
	// 1. CORS 中间件：处理跨域
	r.Use(middleware.CORSMiddleware())
	// 2. API Logger 中间件：记录请求日志
	r.Use(middleware.APILoggerMiddleware())
	// 3. Recovery 中间件：恢复 panic
	r.Use(gin.Recovery())

	// 设置路由

	// API 路由
	router.SetupAPIRouter(r, params.UserHandler)

	// Swagger 路由
	router.SetupSwaggerRouter(r)

	// 获取端口配置
	port := envx.GetStringOptional(consts.HTTPPort)
	if port == "" {
		port = "8080"
	}

	// 验证端口格式
	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("%s 配置错误: %v", consts.HTTPPort, err))
	}
	if portInt < 1 || portInt > 65535 {
		panic(fmt.Sprintf("%s 必须在 1-65535 之间，当前值: %d", consts.HTTPPort, portInt))
	}

	// 构建监听地址
	addr := fmt.Sprintf(":%s", port)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 注册生命周期钩子
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logs.Info("HTTP 服务器启动", "port", port, "mode", mode, "addr", addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logs.Error("HTTP 服务器启动失败", "error", err.Error(), "port", port)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logs.Info("正在关闭 HTTP 服务器", "port", port)
			if err := srv.Shutdown(ctx); err != nil {
				logs.Error("HTTP 服务器关闭失败", "error", err.Error(), "port", port)
				return err
			}
			logs.Info("HTTP 服务器已关闭", "port", port)
			return nil
		},
	})

	return srv
}
