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

// GetSystemConfig 获取系统配置
// 如果配置不存在，返回 gorm.ErrRecordNotFound 错误
func (r *SysRepo) GetSystemConfig(ctx context.Context, key string) (string, error) {
	var systemConfig sysModel.SystemConfig
	if err := r.db.WithContext(ctx).Where("k = ?", key).First(&systemConfig).Error; err != nil {
		return "", err
	}
	return systemConfig.V, nil
}

// SetSystemConfig 设置系统配置（更新已存在的配置）
// 如果配置不存在，不会创建新记录
func (r *SysRepo) SetSystemConfig(ctx context.Context, key string, value string) error {
	return r.db.WithContext(ctx).Model(&sysModel.SystemConfig{}).Where("k = ?", key).Update("v", value).Error
}

// CreateOrUpdateSystemConfig 创建或更新系统配置
// 如果配置不存在则创建，存在则更新
func (r *SysRepo) CreateOrUpdateSystemConfig(ctx context.Context, key string, value string) error {
	var systemConfig sysModel.SystemConfig
	systemConfig.K = key
	systemConfig.V = value

	// 使用 FirstOrCreate 或 Save 来实现创建或更新
	// 这里使用 Save，它会根据主键或唯一索引来决定是创建还是更新
	// 由于 SystemConfig 没有主键，我们使用 FirstOrCreate
	return r.db.WithContext(ctx).Where("k = ?", key).Assign(sysModel.SystemConfig{V: value}).FirstOrCreate(&systemConfig).Error
}
