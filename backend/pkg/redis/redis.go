package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

func NewRedis(config *RedisConfig) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port), // Redis服务器地址
		Password:     config.Password,                                // Redis密码
		DB:           config.DB,                                      // 使用指定数据库
		PoolSize:     config.PoolSize,                                // 连接池大小
		MinIdleConns: config.MinIdleConns,                            // 最小空闲连接数
		MaxIdleConns: config.MaxIdleConns,                            // 最大空闲连接数
		MaxRetries:   config.MaxRetries,                              // 最大重试次数
		DialTimeout:  config.DialTimeout,                             // 连接超时时间
		ReadTimeout:  config.ReadTimeout,                             // 读取超时时间
		WriteTimeout: config.WriteTimeout,                            // 写入超时时间
		PoolTimeout:  config.PoolTimeout,                             // 连接池超时时间
	})

	// 检查Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return redisClient, nil
}
