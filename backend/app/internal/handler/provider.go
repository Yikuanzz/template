package handler

import (
	userHandler "backend/app/internal/handler/user"

	"go.uber.org/fx"
)

// HandlerModule fx 处理器层模块
var HandlerModule = fx.Module("handler",
	fx.Provide(
		// User Handler
		fx.Annotate(
			userHandler.NewUserHandler,
		),
	),
)
