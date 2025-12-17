package server

import (
	"backend/app/server/http"

	"go.uber.org/fx"
)

// ServerModule fx 服务器模块
var ServerModule = fx.Module("server",
	fx.Invoke(
		// 启动 HTTP 服务器
		http.HTTPServer,
	),
)
