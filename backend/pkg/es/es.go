package es

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticSearchConfig struct {
	Addresses            []string      // ES 服务器地址列表，如 ["http://localhost:9200"]
	Username             string        // 用户名（可选）
	Password             string        // 密码（可选）
	CloudID              string        // Elastic Cloud ID（可选）
	APIKey               string        // API Key（可选）
	MaxRetries           int           // 最大重试次数
	EnableRetryOnTimeout bool          // 是否在超时时重试
	EnableCompression    bool          // 是否启用压缩
	DisableMetaHeader    bool          // 是否禁用元数据头
	RequestTimeout       time.Duration // 请求超时时间
	PingTimeout          time.Duration // Ping 超时时间
}

func NewElasticSearch(config *ElasticSearchConfig) (*elasticsearch.Client, error) {
	// 设置默认值
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 10 * time.Second
	}
	if config.PingTimeout == 0 {
		config.PingTimeout = 5 * time.Second
	}

	// 构建 ES 客户端配置
	esConfig := elasticsearch.Config{
		Addresses:  config.Addresses,
		MaxRetries: config.MaxRetries,
	}

	// 如果需要超时重试，使用 RetryOnError 回调
	if config.EnableRetryOnTimeout {
		esConfig.RetryOnError = func(req *http.Request, err error) bool {
			// 在超时错误时重试
			return err != nil
		}
	}

	// 设置认证方式（优先级：APIKey > Username/Password > CloudID）
	if config.APIKey != "" {
		esConfig.APIKey = config.APIKey
		log.Printf("🔑 使用 API Key 认证")
	} else if config.Username != "" && config.Password != "" {
		esConfig.Username = config.Username
		esConfig.Password = config.Password
		log.Printf("🔑 使用用户名密码认证")
	} else if config.CloudID != "" {
		esConfig.CloudID = config.CloudID
		log.Printf("🔑 使用 Cloud ID 认证")
	} else {
		log.Printf("⚠️  未配置认证信息，使用匿名连接")
	}

	// 创建 ES 客户端
	esClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 ElasticSearch 客户端失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.PingTimeout)
	defer cancel()

	res, err := esClient.Info(esClient.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("ElasticSearch 连接测试失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ElasticSearch 连接测试失败: %s", res.String())
	}

	// 输出连接信息
	log.Printf("✅ ElasticSearch 连接成功")
	log.Printf("📊 ElasticSearch 配置信息:")
	log.Printf("   - 服务器地址: %s", strings.Join(config.Addresses, ", "))
	log.Printf("   - 最大重试次数: %d", config.MaxRetries)
	log.Printf("   - 请求超时时间: %v", config.RequestTimeout)
	log.Printf("   - Ping 超时时间: %v", config.PingTimeout)
	log.Printf("   - 启用压缩: %v", config.EnableCompression)
	log.Printf("   - 超时重试: %v", config.EnableRetryOnTimeout)

	return esClient, nil
}
