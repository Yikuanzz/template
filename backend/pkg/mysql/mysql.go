package mysql

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	MaxIdleConns       int  // 最大空闲连接数
	MaxOpenConns       int  // 最大打开连接数
	ConnMaxLifetimeMin int  // 连接最大生存时间（分钟）
	ConnMaxIdleTimeMin int  // 连接最大空闲时间（分钟）
	EnableSlowQueryLog bool // 是否启用慢查询日志
	SlowQueryThreshold int  // 慢查询阈值（毫秒）
}

func NewMySQL(config *MySQLConfig) (*gorm.DB, error) {
	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// 配置 GORM 日志
	gormConfig := &gorm.Config{}

	// 根据配置启用慢查询日志
	if config.EnableSlowQueryLog {
		slowThreshold := time.Duration(config.SlowQueryThreshold) * time.Millisecond
		gormConfig.Logger = logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             slowThreshold, // 慢查询阈值
				LogLevel:                  logger.Warn,   // 日志级别
				IgnoreRecordNotFoundError: true,          // 忽略 ErrRecordNotFound 错误
				Colorful:                  true,          // 启用彩色输出
			},
		)
		log.Printf("✅ 慢查询日志已启用，阈值: %dms", config.SlowQueryThreshold)
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 获取底层的 sql.DB 对象来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库对象失败: %w", err)
	}

	// 配置连接池（使用配置文件中的参数）
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetimeMin) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTimeMin) * time.Minute)

	// 输出连接池配置信息
	// log.Printf("📊 数据库连接池配置:")
	// log.Printf("   - 最大空闲连接数: %d", config.MaxIdleConns)
	// log.Printf("   - 最大打开连接数: %d", config.MaxOpenConns)
	// log.Printf("   - 连接最大生存时间: %d 分钟", config.ConnMaxLifetimeMin)
	// log.Printf("   - 连接最大空闲时间: %d 分钟", config.ConnMaxIdleTimeMin)

	// 检查数据库连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}
