package secret

import (
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptMaxPasswordLength bcrypt 支持的最大密码长度（72 字节）
	bcryptMaxPasswordLength = 72
)

// HashPassword 哈希密码
// 如果密码长度超过 72 字节，会先使用 SHA256 进行哈希，然后再使用 bcrypt
func HashPassword(password string) (string, error) {
	passwordBytes := []byte(password)

	// 如果密码超过 bcrypt 的限制，先进行 SHA256 哈希
	if len(passwordBytes) > bcryptMaxPasswordLength {
		hash := sha256.Sum256(passwordBytes)
		passwordBytes = hash[:]
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword 验证密码
// 如果密码长度超过 72 字节，会先使用 SHA256 进行哈希，然后再验证
func VerifyPassword(password string, hashedPassword string) bool {
	passwordBytes := []byte(password)

	// 如果密码超过 bcrypt 的限制，先进行 SHA256 哈希
	if len(passwordBytes) > bcryptMaxPasswordLength {
		hash := sha256.Sum256(passwordBytes)
		passwordBytes = hash[:]
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), passwordBytes) == nil
}
