package db

import (
	"context"

	"backend/app/types/consts"
	"backend/pkg/sqlite"
	"backend/utils/envx"
	"backend/utils/logs"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// ProvideDatabaseParams 定义 Database 的依赖
type ProvideDatabaseParams struct {
	fx.In

	Lifecycle fx.Lifecycle
}

// ProvideDatabase 提供数据库实例
func ProvideDatabase(params ProvideDatabaseParams) (*gorm.DB, error) {
	// 读取数据库文件路径（可选，带默认值）
	dbPath := envx.GetStringOptional(consts.SQLiteDBPath)
	if dbPath == "" {
		dbPath = "data.db"
		logs.Info("使用默认数据库路径", "path", dbPath)
	}

	// 读取最大空闲连接数（可选，带默认值）
	maxIdleConns, err := envx.GetIntWithDefault(consts.SQLiteMaxIdleConns, 10)
	if err != nil {
		return nil, err
	}

	// 读取最大打开连接数（可选，带默认值）
	maxOpenConns, err := envx.GetIntWithDefault(consts.SQLiteMaxOpenConns, 100)
	if err != nil {
		return nil, err
	}

	// 读取连接最大生存时间（可选，带默认值）
	connMaxLifetimeMin, err := envx.GetIntWithDefault(consts.SQLiteConnMaxLifetimeMin, 60)
	if err != nil {
		return nil, err
	}

	// 读取连接最大空闲时间（可选，带默认值）
	connMaxIdleTimeMin, err := envx.GetIntWithDefault(consts.SQLiteConnMaxIdleTimeMin, 10)
	if err != nil {
		return nil, err
	}

	// 读取是否启用慢查询日志（可选，带默认值）
	enableSlowQueryLog := envx.GetBool(consts.SQLiteEnableSlowQueryLog, false)

	// 读取慢查询阈值（可选，带默认值）
	slowQueryThreshold, err := envx.GetIntWithDefault(consts.SQLiteSlowQueryThreshold, 200)
	if err != nil {
		return nil, err
	}

	// 构建 SQLite 配置
	config := &sqlite.SQLiteConfig{
		DBPath:             dbPath,
		MaxIdleConns:       maxIdleConns,
		MaxOpenConns:       maxOpenConns,
		ConnMaxLifetimeMin: connMaxLifetimeMin,
		ConnMaxIdleTimeMin: connMaxIdleTimeMin,
		EnableSlowQueryLog: enableSlowQueryLog,
		SlowQueryThreshold: slowQueryThreshold,
	}

	// 创建数据库连接
	db, err := sqlite.NewSQLite(config)
	if err != nil {
		return nil, err
	}

	// 注册生命周期钩子，在应用关闭时关闭数据库连接
	params.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logs.Info("正在关闭数据库连接", "path", dbPath)
			sqlDB, err := db.DB()
			if err != nil {
				logs.Error("获取底层数据库对象失败", "error", err.Error())
				return err
			}
			if err := sqlDB.Close(); err != nil {
				logs.Error("数据库连接关闭失败", "error", err.Error(), "path", dbPath)
				return err
			}
			logs.Info("数据库连接已关闭", "path", dbPath)
			return nil
		},
	})

	logs.Info("数据库连接已创建", "path", dbPath)
	return db, nil
}
