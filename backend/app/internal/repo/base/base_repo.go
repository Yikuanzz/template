package base

import (
	"context"
	"errors"

	fileModel "backend/app/model/file"
	systemModel "backend/app/model/system"
	userModel "backend/app/model/user"
	"backend/utils/logs"
	"backend/utils/secret"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *userModel.User) error
}

type SysRepo interface {
	GetSystemConfig(ctx context.Context, key string) (string, error)
	SetSystemConfig(ctx context.Context, key string, value string) error
	CreateOrUpdateSystemConfig(ctx context.Context, key string, value string) error
}

type BaseRepoParams struct {
	fx.In

	UserRepo UserRepo
	SysRepo  SysRepo
	DB       *gorm.DB
}

type BaseRepo struct {
	userRepo UserRepo
	sysRepo  SysRepo
	db       *gorm.DB
}

// InitBaseData 初始化基础数据
// 包括：数据库表迁移、系统配置初始化、用户数据初始化
func InitBaseData(params BaseRepoParams) error {
	r := &BaseRepo{
		userRepo: params.UserRepo,
		sysRepo:  params.SysRepo,
		db:       params.DB,
	}

	// 1. 初始化数据库表
	if err := r.InitTables(); err != nil {
		return err
	}

	// 2. 初始化系统配置
	// 如果系统已初始化（配置存在且值为"ok"），则跳过用户数据初始化
	alreadyInitialized, err := r.InitSystemConfig()
	if err != nil {
		return err
	}

	if alreadyInitialized {
		logs.Info("系统配置已初始化，跳过数据初始化")
		return nil
	}

	// 3. 初始化用户数据（仅在首次启动时执行）
	if err := r.InitUsers(); err != nil {
		return err
	}

	return nil
}

// InitTables 初始化数据库表
// 使用 AutoMigrate 自动创建或更新表结构
func (r *BaseRepo) InitTables() error {
	logs.Info("初始化数据库表")
	err := r.db.AutoMigrate(
		&userModel.User{},
		&systemModel.SystemConfig{},
		&fileModel.File{},
	)
	if err != nil {
		logs.Error("初始化数据库表失败", "error", err.Error())
		return err
	}
	logs.Info("数据库表初始化完成")
	return nil
}

// InitSystemConfig 初始化系统配置
// 返回值：
//   - bool: true 表示系统已初始化（配置存在且值为"ok"），false 表示首次初始化
//   - error: 错误信息
func (r *BaseRepo) InitSystemConfig() (bool, error) {
	logs.Info("检查系统配置")

	initKey := "init"
	initValue := "ok"

	// 尝试获取系统配置
	value, err := r.sysRepo.GetSystemConfig(context.Background(), initKey)
	if err != nil {
		// 如果配置不存在（record not found），这是正常的首次启动情况
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.Info("系统配置不存在，开始初始化")
			// 创建系统配置
			if err := r.sysRepo.CreateOrUpdateSystemConfig(context.Background(), initKey, initValue); err != nil {
				logs.Error("创建系统配置失败", "error", err.Error())
				return false, err
			}
			logs.Info("系统配置初始化完成")
			return false, nil
		}
		// 其他错误（如数据库连接错误等）
		logs.Error("获取系统配置失败", "error", err.Error())
		return false, err
	}

	// 配置存在，检查值是否为"ok"
	if value == initValue {
		logs.Info("系统配置已初始化，跳过数据初始化")
		return true, nil
	}

	// 配置存在但值不是"ok"，更新为"ok"
	logs.Info("系统配置值不正确，更新配置", "current_value", value)
	if err := r.sysRepo.SetSystemConfig(context.Background(), initKey, initValue); err != nil {
		logs.Error("更新系统配置失败", "error", err.Error())
		return false, err
	}
	logs.Info("系统配置更新完成")
	return false, nil
}

// InitUsers 初始化用户数据
// 仅在首次启动时执行，创建默认管理员账户
func (r *BaseRepo) InitUsers() error {
	logs.Info("初始化用户数据")

	// 检查是否已存在用户
	var userCount int64
	if err := r.db.Model(&userModel.User{}).Count(&userCount).Error; err != nil {
		logs.Error("查询用户数量失败", "error", err.Error())
		return err
	}

	if userCount > 0 {
		logs.Info("用户数据已存在，跳过初始化", "user_count", userCount)
		return nil
	}

	// 创建默认管理员账户
	passwordHash, err := secret.HashPassword("12345678")
	if err != nil {
		logs.Error("生成密码哈希失败", "error", err.Error())
		return err
	}

	err = r.userRepo.CreateUser(context.Background(), &userModel.User{
		Username:     "admin",
		PasswordHash: passwordHash,
		NickName:     "admin",
		Avatar:       "https://avatar.iran.liara.run/public",
	})
	if err != nil {
		logs.Error("创建默认用户失败", "error", err.Error())
		return err
	}

	logs.Info("用户数据初始化完成", "username", "admin", "password", "12345678")
	return nil
}
