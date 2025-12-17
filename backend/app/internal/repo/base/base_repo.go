package base

import (
	"context"

	"go.uber.org/fx"
)

type UserRepo interface{}

type SysRepo interface {
	GetSystemConfig(ctx context.Context, key string) (string, error)
	SetSystemConfig(ctx context.Context, key string, value string) error
}

type BaseRepoParams struct {
	fx.In

	UserRepo UserRepo
	SysRepo  SysRepo
}

type BaseRepo struct {
	userRepo UserRepo
	sysRepo  SysRepo
}

func InitBaseData(params BaseRepoParams) error {
	r := &BaseRepo{
		userRepo: params.UserRepo,
		sysRepo:  params.SysRepo,
	}
	if err := r.InitTables(); err != nil {
		return err
	}
	if err := r.InitUsers(); err != nil {
		return err
	}
	return nil
}

func (r *BaseRepo) InitTables() error {
	return nil
}

func (r *BaseRepo) InitUsers() error {
	return nil
}
