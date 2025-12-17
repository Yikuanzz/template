package secret

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenConfig 令牌配置
type TokenConfig struct {
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration
	Secret             string
}

// Claims JWT声明
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type JWT struct {
	tokenConfig TokenConfig
}

func NewJWT(tokenConfig TokenConfig) *JWT {
	return &JWT{
		tokenConfig: tokenConfig,
	}
}

// GenerateAccessToken 生成访问令牌
func (j *JWT) GenerateAccessToken(userID uint) (string, int64, error) {
	expireTime := time.Now().Add(j.tokenConfig.AccessTokenExpire)
	expireUnix := expireTime.Unix()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.tokenConfig.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireUnix, nil
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWT) GenerateRefreshToken(userID uint) (string, int64, error) {
	expireTime := time.Now().Add(j.tokenConfig.RefreshTokenExpire)
	expireUnix := expireTime.Unix()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.tokenConfig.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireUnix, nil
}

// ParseToken 解析令牌
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.tokenConfig.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// IsTokenExpired 检查令牌是否过期
func (j *JWT) IsTokenExpired(tokenString string) bool {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return true
	}
	return time.Now().Unix() > claims.ExpiresAt.Unix()
}
