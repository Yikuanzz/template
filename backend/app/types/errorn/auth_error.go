package errorn

import (
	"backend/utils/errorx"
)

const (
	// 认证错误码 (2000000-2000099)
	AuthErrTokenRequired     = int32(2000000) // Token 必填
	AuthErrTokenInvalid      = int32(2000001) // Token 无效
	AuthErrTokenExpired      = int32(2000002) // Token 已过期
	AuthErrTokenMalformed    = int32(2000003) // Token 格式错误
	AuthErrTokenSignature    = int32(2000004) // Token 签名错误
	AuthErrUserNotFound      = int32(2000005) // 用户不存在
	AuthErrJWTSecretMissing  = int32(2000006) // JWT 密钥未配置
	AuthErrJWTSecretInvalid  = int32(2000007) // JWT 密钥无效
	AuthErrPasswordIncorrect = int32(2000008) // 密码错误
	AuthErrCaptchaInvalid    = int32(2000009) // 验证码错误
	AuthErrCaptchaExpired    = int32(2000010) // 验证码已过期
	AuthErrUserInactive      = int32(2000011) // 用户已被禁用
	AuthErrUserCreateFailed  = int32(2000012) // 创建用户失败
	AuthErrUserDeleteFailed  = int32(2000013) // 删除用户失败
	AuthErrUserAlreadyExists = int32(2000014) // 用户已存在
	AuthErrUserLocked        = int32(2000015) // 账号已被锁定
	AuthErrUserUpdateFailed  = int32(2000016) // 更新用户信息失败
)

func init() {
	// 注册认证错误码
	errorx.RegisterBatch(map[int32]string{
		AuthErrTokenRequired:     "Token 不能为空",
		AuthErrTokenInvalid:      "Token 无效: {reason}",
		AuthErrTokenExpired:      "Token 已过期",
		AuthErrTokenMalformed:    "Token 格式错误: {reason}",
		AuthErrTokenSignature:    "Token 签名验证失败",
		AuthErrUserNotFound:      "用户不存在: {user_uid}",
		AuthErrJWTSecretMissing:  "JWT 密钥未配置，请设置环境变量 JWT_SECRET",
		AuthErrJWTSecretInvalid:  "JWT 密钥无效",
		AuthErrPasswordIncorrect: "密码错误",
		AuthErrCaptchaInvalid:    "验证码错误",
		AuthErrCaptchaExpired:    "验证码已过期",
		AuthErrUserInactive:      "用户已被禁用",
		AuthErrUserCreateFailed:  "创建用户失败: {reason}",
		AuthErrUserDeleteFailed:  "删除用户失败: {reason}",
		AuthErrUserAlreadyExists: "用户已存在: {username}",
		AuthErrUserLocked:        "账号已被锁定，请30分钟后再试",
		AuthErrUserUpdateFailed:  "更新用户信息失败: {reason}",
	})
}
