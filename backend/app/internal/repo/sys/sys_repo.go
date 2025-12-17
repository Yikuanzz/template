package sys

import (
	"context"

	sysModel "backend/app/model/system"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type SysRepoParams struct {
	fx.In

	DB *gorm.DB
}

type SysRepo struct {
	db *gorm.DB
}

func NewSysRepo(params SysRepoParams) *SysRepo {
	return &SysRepo{
		db: params.DB,
	}
}

func (r *SysRepo) GetSystemConfig(ctx context.Context, key string) (string, error) {
	var systemConfig sysModel.SystemConfig
	if err := r.db.WithContext(ctx).Where("k = ?", key).First(&systemConfig).Error; err != nil {
		return "", err
	}
	return systemConfig.V, nil
}

func (r *SysRepo) SetSystemConfig(ctx context.Context, key string, value string) error {
	return r.db.WithContext(ctx).Model(&sysModel.SystemConfig{}).Where("k = ?", key).Update("v", value).Error
}
