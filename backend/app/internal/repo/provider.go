package repo

import (
	userLogic "backend/app/internal/logic/user"
	baseRepo "backend/app/internal/repo/base"
	sysRepo "backend/app/internal/repo/sys"
	userRepo "backend/app/internal/repo/user"

	"go.uber.org/fx"
)

var RepoModule = fx.Module("repo",
	fx.Provide(
		// User Repo
		fx.Annotate(
			userRepo.NewUserRepo,
			fx.As(new(userLogic.UserRepo)),
			fx.As(new(baseRepo.UserRepo)),
		),
		// Sys Repo
		fx.Annotate(
			sysRepo.NewSysRepo,
			fx.As(new(baseRepo.SysRepo)),
		),
	),
	// 初始化基础数据
	fx.Invoke(baseRepo.InitBaseData),
)
