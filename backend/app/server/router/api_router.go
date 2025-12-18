package router

import (
	"backend/app/internal/handler/file"
	"backend/app/internal/handler/user"
	"backend/app/server/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAPIRouter 设置 API 路由
// userHandler: User 处理器
// fileHandler: File 处理器
func SetupAPIRouter(r *gin.Engine, userHandler *user.UserHandler, fileHandler *file.FileHandler) {
	api := r.Group("/api")

	// 用户相关路由
	{
		userGroup := api.Group("/user")
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/refresh-token", userHandler.RefreshToken)
		// 需要认证的路由
		userGroupAuth := userGroup.Group("")
		userGroupAuth.Use(middleware.AuthMiddleware())
		userGroupAuth.GET("/info", userHandler.GetUserInfo)
		userGroupAuth.PUT("/info", userHandler.UpateUserInfo)
	}

	// 文件相关路由
	{
		fileGroup := api.Group("/file")
		fileGroup.POST("/upload", fileHandler.UploadFile)
	}
}
