package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"backend/app/internal/handler/file"
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
	FileHandler *file.FileHandler
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

	// 静态文件服务（用于访问上传的文件）
	setupStaticFileServer(r)

	// API 路由
	router.SetupAPIRouter(r, params.UserHandler, params.FileHandler)

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

// setupStaticFileServer 设置静态文件服务器
func setupStaticFileServer(r *gin.Engine) {
	// 读取本地存储路径
	storageLocalPath := envx.GetStringOptional(consts.StorageLocalPath)
	if storageLocalPath == "" {
		storageLocalPath = "./uploads" // 默认路径
		logs.Info("未配置本地存储路径，使用默认值: ./uploads")
	}

	// 读取本地存储访问URL
	storageLocalBaseURL := envx.GetStringOptional(consts.StorageLocalBaseURL)
	if storageLocalBaseURL == "" {
		// 如果没有配置，使用默认路径
		storageLocalBaseURL = "/uploads"
		logs.Info("未配置本地存储访问URL，使用默认路径: /uploads")
	} else {
		// 从完整URL中提取路径部分
		parsedURL, err := url.Parse(storageLocalBaseURL)
		if err != nil {
			logs.Warn("解析 StorageLocalBaseURL 失败，使用默认路径", "error", err.Error(), "url", storageLocalBaseURL)
			storageLocalBaseURL = "/uploads"
		} else {
			storageLocalBaseURL = parsedURL.Path
			// 如果路径为空，使用默认路径
			if storageLocalBaseURL == "" {
				storageLocalBaseURL = "/uploads"
			}
		}
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(storageLocalBaseURL, "/") {
		storageLocalBaseURL = "/" + storageLocalBaseURL
	}

	// 移除路径末尾的 /
	storageLocalBaseURL = strings.TrimSuffix(storageLocalBaseURL, "/")

	// 转换为绝对路径
	absPath, err := filepath.Abs(storageLocalPath)
	if err != nil {
		logs.Error("获取绝对路径失败", "error", err.Error(), "path", storageLocalPath)
		absPath = storageLocalPath
	}

	// 设置静态文件服务
	// 使用 StaticFS 可以更好地控制文件访问
	r.StaticFS(storageLocalBaseURL, http.Dir(absPath))

	logs.Info("静态文件服务已配置", "url_path", storageLocalBaseURL, "file_path", absPath)
}
