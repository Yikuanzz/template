package secret_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/utils/secret"
)

func TestHashPassword(t *testing.T) {
	t.Run("成功哈希密码", func(t *testing.T) {
		password := "mySecurePassword123"
		hashed, err := secret.HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)
	})

	t.Run("相同密码生成不同哈希值", func(t *testing.T) {
		password := "samePassword"
		hashed1, err1 := secret.HashPassword(password)
		hashed2, err2 := secret.HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		// bcrypt 每次生成的哈希都不同（因为包含随机盐）
		assert.NotEqual(t, hashed1, hashed2)
	})

	t.Run("空密码也能哈希", func(t *testing.T) {
		hashed, err := secret.HashPassword("")

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
	})

	t.Run("长密码也能哈希", func(t *testing.T) {
		longPassword := make([]byte, 1000)
		for i := range longPassword {
			longPassword[i] = byte('a' + (i % 26))
		}
		hashed, err := secret.HashPassword(string(longPassword))

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("正确密码验证成功", func(t *testing.T) {
		password := "correctPassword123"
		hashed, err := secret.HashPassword(password)
		require.NoError(t, err)

		isValid := secret.VerifyPassword(password, hashed)
		assert.True(t, isValid)
	})

	t.Run("错误密码验证失败", func(t *testing.T) {
		password := "correctPassword123"
		wrongPassword := "wrongPassword456"
		hashed, err := secret.HashPassword(password)
		require.NoError(t, err)

		isValid := secret.VerifyPassword(wrongPassword, hashed)
		assert.False(t, isValid)
	})

	t.Run("空密码验证", func(t *testing.T) {
		hashed, err := secret.HashPassword("")
		require.NoError(t, err)

		isValid := secret.VerifyPassword("", hashed)
		assert.True(t, isValid)

		isValidWrong := secret.VerifyPassword("not-empty", hashed)
		assert.False(t, isValidWrong)
	})

	t.Run("大小写敏感", func(t *testing.T) {
		password := "Password123"
		hashed, err := secret.HashPassword(password)
		require.NoError(t, err)

		isValid := secret.VerifyPassword("password123", hashed)
		assert.False(t, isValid)

		isValidCorrect := secret.VerifyPassword("Password123", hashed)
		assert.True(t, isValidCorrect)
	})

	t.Run("特殊字符密码", func(t *testing.T) {
		password := "P@ssw0rd!#$%^&*()"
		hashed, err := secret.HashPassword(password)
		require.NoError(t, err)

		isValid := secret.VerifyPassword(password, hashed)
		assert.True(t, isValid)

		isValidWrong := secret.VerifyPassword("P@ssw0rd!#$%^&*()_", hashed)
		assert.False(t, isValidWrong)
	})

	t.Run("无效哈希值验证失败", func(t *testing.T) {
		password := "anyPassword"
		invalidHash := "invalid-hash-string"

		isValid := secret.VerifyPassword(password, invalidHash)
		assert.False(t, isValid)
	})

	t.Run("空哈希值验证失败", func(t *testing.T) {
		isValid := secret.VerifyPassword("anyPassword", "")
		assert.False(t, isValid)
	})
}

func TestHashPassword_Integration(t *testing.T) {
	t.Run("完整流程：哈希和验证", func(t *testing.T) {
		passwords := []string{
			"simple",
			"complexP@ssw0rd!",
			"123456",
			"   spaces   ",
			"中文密码",
			"emoji🔐password",
		}

		for _, password := range passwords {
			t.Run("密码: "+password, func(t *testing.T) {
				hashed, err := secret.HashPassword(password)
				require.NoError(t, err)

				// 正确密码应该验证成功
				isValid := secret.VerifyPassword(password, hashed)
				assert.True(t, isValid, "密码 %s 应该验证成功", password)

				// 错误密码应该验证失败
				wrongPassword := password + "wrong"
				isValidWrong := secret.VerifyPassword(wrongPassword, hashed)
				assert.False(t, isValidWrong, "错误密码应该验证失败")
			})
		}
	})
}
