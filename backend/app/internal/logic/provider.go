package logic

import (
	userHandler "backend/app/internal/handler/user"
	userLogic "backend/app/internal/logic/user"

	"go.uber.org/fx"
)

// LogicModule fx 业务逻辑层模块
var LogicModule = fx.Module("logic",
	fx.Provide(
		// User Logic
		fx.Annotate(
			userLogic.NewUserLogic,
			fx.As(new(userHandler.UserLogic)),
		),
	),
)
