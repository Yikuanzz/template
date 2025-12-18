package logic

import (
	fileHandler "backend/app/internal/handler/file"
	userHandler "backend/app/internal/handler/user"
	fileLogic "backend/app/internal/logic/file"
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
		// File Logic
		fx.Annotate(
			fileLogic.NewFileLogic,
			fx.As(new(fileHandler.FileLogic)),
		),
	),
)
