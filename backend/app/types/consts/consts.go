package consts

// HTTP 配置环境变量名
const (
	// HTTPPort HTTP 端口
	HTTPPort = "HTTP_PORT"

	// GINMode GIN 模式
	GINMode = "GIN_MODE"
)

// JWT 配置环境变量名
const (
	// JWTSecret JWT Secret
	JWTSecret = "JWT_SECRET"
	// AccessTokenExpire AccessToken Expire
	AccessTokenExpire = "ACCESS_TOKEN_EXPIRE"
	// RefreshTokenExpire RefreshToken Expire
	RefreshTokenExpire = "REFRESH_TOKEN_EXPIRE"
)

// 日志相关环境变量
const (
	// EnvLogLevel 日志级别环境变量名
	// 可选值: trace, debug, info, notice, warn, error, fatal
	// 默认值: info
	EnvLogLevel = "LOG_LEVEL"

	// EnvLogOutput 日志输出方式环境变量名
	// 可选值: console, json
	// 默认值: console
	EnvLogOutput = "LOG_OUTPUT"

	// EnvLogDevelopment 是否为开发模式环境变量名
	// 可选值: true, false
	// 默认值: false
	EnvLogDevelopment = "LOG_DEVELOPMENT"

	// EnvLogFile 日志文件路径环境变量名
	// 如果设置，日志会同时输出到 stdout 和该文件
	// 如果不设置，只输出到 stdout（适合容器化部署）
	// 默认值: 空（只输出到 stdout）
	EnvLogFile = "LOG_FILE"

	// EnvLogMaxSize 单个日志文件的最大大小（MB）
	// 当日志文件达到此大小时，会自动轮转
	// 默认值: 100 (100MB)
	EnvLogMaxSize = "LOG_MAX_SIZE"

	// EnvLogMaxBackups 保留的旧日志文件数量
	// 默认值: 7 (保留7个旧文件)
	EnvLogMaxBackups = "LOG_MAX_BACKUPS"

	// EnvLogMaxAge 日志文件保留天数
	// 超过此天数的旧日志文件会被自动删除
	// 默认值: 30 (30天)
	EnvLogMaxAge = "LOG_MAX_AGE"

	// EnvLogCompress 是否压缩轮转后的旧日志文件
	// 可选值: true, false
	// 默认值: true
	EnvLogCompress = "LOG_COMPRESS"
)

// SQLite 数据库配置环境变量名
const (
	// SQLiteDBPath SQLite 数据库文件路径
	// 默认值: data.db
	SQLiteDBPath = "SQLITE_DB_PATH"

	// SQLiteMaxIdleConns 最大空闲连接数
	// 默认值: 10
	SQLiteMaxIdleConns = "SQLITE_MAX_IDLE_CONNS"

	// SQLiteMaxOpenConns 最大打开连接数
	// 默认值: 100
	SQLiteMaxOpenConns = "SQLITE_MAX_OPEN_CONNS"

	// SQLiteConnMaxLifetimeMin 连接最大生存时间（分钟）
	// 默认值: 60
	SQLiteConnMaxLifetimeMin = "SQLITE_CONN_MAX_LIFETIME_MIN"

	// SQLiteConnMaxIdleTimeMin 连接最大空闲时间（分钟）
	// 默认值: 10
	SQLiteConnMaxIdleTimeMin = "SQLITE_CONN_MAX_IDLE_TIME_MIN"

	// SQLiteEnableSlowQueryLog 是否启用慢查询日志
	// 可选值: true, false
	// 默认值: false
	SQLiteEnableSlowQueryLog = "SQLITE_ENABLE_SLOW_QUERY_LOG"

	// SQLiteSlowQueryThreshold 慢查询阈值（毫秒）
	// 默认值: 200
	SQLiteSlowQueryThreshold = "SQLITE_SLOW_QUERY_THRESHOLD"
)
