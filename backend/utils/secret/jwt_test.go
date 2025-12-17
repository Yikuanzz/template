package secret_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/utils/secret"
)

func TestNewJWT(t *testing.T) {
	config := secret.TokenConfig{
		AccessTokenExpire:  time.Hour,
		RefreshTokenExpire: 24 * time.Hour,
		Secret:             "test-secret-key",
	}

	jwtInstance := secret.NewJWT(config)
	require.NotNil(t, jwtInstance)
}

func TestGenerateAccessToken(t *testing.T) {
	config := secret.TokenConfig{
		AccessTokenExpire:  time.Hour,
		RefreshTokenExpire: 24 * time.Hour,
		Secret:             "test-secret-key",
	}
	jwtInstance := secret.NewJWT(config)

	t.Run("成功生成访问令牌", func(t *testing.T) {
		userID := uint(123)
		token, expireUnix, err := jwtInstance.GenerateAccessToken(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, expireUnix, time.Now().Unix())
		assert.Less(t, expireUnix, time.Now().Add(2*time.Hour).Unix())
	})

	t.Run("不同用户ID生成不同令牌", func(t *testing.T) {
		token1, _, err1 := jwtInstance.GenerateAccessToken(1)
		token2, _, err2 := jwtInstance.GenerateAccessToken(2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, token1, token2)
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	config := secret.TokenConfig{
		AccessTokenExpire:  time.Hour,
		RefreshTokenExpire: 24 * time.Hour,
		Secret:             "test-secret-key",
	}
	jwtInstance := secret.NewJWT(config)

	t.Run("成功生成刷新令牌", func(t *testing.T) {
		token, expireUnix, err := jwtInstance.GenerateRefreshToken(123)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, expireUnix, time.Now().Unix())
		assert.Less(t, expireUnix, time.Now().Add(25*time.Hour).Unix())
	})

	t.Run("刷新令牌过期时间比访问令牌长", func(t *testing.T) {
		accessToken, accessExpire, _ := jwtInstance.GenerateAccessToken(123)
		refreshToken, refreshExpire, _ := jwtInstance.GenerateRefreshToken(123)

		require.NotEmpty(t, accessToken)
		require.NotEmpty(t, refreshToken)
		assert.Greater(t, refreshExpire, accessExpire)
	})
}

func TestParseToken(t *testing.T) {
	config := secret.TokenConfig{
		AccessTokenExpire:  time.Hour,
		RefreshTokenExpire: 24 * time.Hour,
		Secret:             "test-secret-key",
	}
	jwtInstance := secret.NewJWT(config)

	t.Run("成功解析有效令牌", func(t *testing.T) {
		userID := uint(456)
		token, _, err := jwtInstance.GenerateAccessToken(userID)
		require.NoError(t, err)

		claims, err := jwtInstance.ParseToken(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
	})

	t.Run("解析无效令牌返回错误", func(t *testing.T) {
		invalidToken := "invalid.token.string"
		claims, err := jwtInstance.ParseToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("解析空令牌返回错误", func(t *testing.T) {
		claims, err := jwtInstance.ParseToken("")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("使用错误密钥解析令牌失败", func(t *testing.T) {
		// 使用一个密钥生成令牌
		config1 := secret.TokenConfig{
			AccessTokenExpire:  time.Hour,
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "secret-key-1",
		}
		jwt1 := secret.NewJWT(config1)
		token, _, err := jwt1.GenerateAccessToken(123)
		require.NoError(t, err)

		// 使用不同密钥解析
		config2 := secret.TokenConfig{
			AccessTokenExpire:  time.Hour,
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "secret-key-2",
		}
		jwt2 := secret.NewJWT(config2)
		claims, err := jwt2.ParseToken(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestIsTokenExpired(t *testing.T) {
	t.Run("未过期令牌返回false", func(t *testing.T) {
		config := secret.TokenConfig{
			AccessTokenExpire:  time.Hour,
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "test-secret-key",
		}
		jwtInstance := secret.NewJWT(config)

		token, _, err := jwtInstance.GenerateAccessToken(123)
		require.NoError(t, err)

		isExpired := jwtInstance.IsTokenExpired(token)
		assert.False(t, isExpired)
	})

	t.Run("已过期令牌返回true", func(t *testing.T) {
		config := secret.TokenConfig{
			AccessTokenExpire:  -time.Hour, // 设置为负值，立即过期
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "test-secret-key",
		}
		jwtInstance := secret.NewJWT(config)

		token, _, err := jwtInstance.GenerateAccessToken(123)
		require.NoError(t, err)

		// 等待一小段时间确保过期
		time.Sleep(100 * time.Millisecond)
		isExpired := jwtInstance.IsTokenExpired(token)
		assert.True(t, isExpired)
	})

	t.Run("无效令牌返回true", func(t *testing.T) {
		config := secret.TokenConfig{
			AccessTokenExpire:  time.Hour,
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "test-secret-key",
		}
		jwtInstance := secret.NewJWT(config)

		isExpired := jwtInstance.IsTokenExpired("invalid.token")
		assert.True(t, isExpired)
	})

	t.Run("空令牌返回true", func(t *testing.T) {
		config := secret.TokenConfig{
			AccessTokenExpire:  time.Hour,
			RefreshTokenExpire: 24 * time.Hour,
			Secret:             "test-secret-key",
		}
		jwtInstance := secret.NewJWT(config)

		isExpired := jwtInstance.IsTokenExpired("")
		assert.True(t, isExpired)
	})
}

func TestJWT_Integration(t *testing.T) {
	config := secret.TokenConfig{
		AccessTokenExpire:  time.Hour,
		RefreshTokenExpire: 24 * time.Hour,
		Secret:             "integration-test-secret",
	}
	jwtInstance := secret.NewJWT(config)

	t.Run("完整流程：生成、解析、验证", func(t *testing.T) {
		userID := uint(789)

		// 生成访问令牌
		accessToken, expireUnix, err := jwtInstance.GenerateAccessToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.Greater(t, expireUnix, time.Now().Unix())

		// 解析令牌
		claims, err := jwtInstance.ParseToken(accessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)

		// 验证未过期
		isExpired := jwtInstance.IsTokenExpired(accessToken)
		assert.False(t, isExpired)

		// 生成刷新令牌
		refreshToken, refreshExpire, err := jwtInstance.GenerateRefreshToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshToken)
		assert.Greater(t, refreshExpire, expireUnix)
	})
}
