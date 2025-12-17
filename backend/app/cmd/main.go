package main

import (
	"flag"
	"fmt"

	"backend/app/internal/handler"
	"backend/app/internal/logic"
	"backend/app/internal/repo"
	"backend/app/plugins"
	"backend/app/server"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// @title Backend API
// @version 1.0
// @description 这是一个基于 Go 和 Gin 框架的 API 服务
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @schemes http https
func main() {
	envFile := flag.String("env", ".env", "环境变量文件")
	flag.Parse()

	if err := godotenv.Load(*envFile); err != nil {
		fmt.Println("加载环境变量失败", err.Error())
		return
	}

	app := fx.New(
		// fx.NopLogger,
		// 基础设施模块
		plugins.PluginsModule,

		// 应用层模块
		repo.RepoModule,
		logic.LogicModule,
		handler.HandlerModule,

		// 服务器模块
		server.ServerModule,
	)

	app.Run()
}
