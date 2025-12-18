package user

import (
	"context"
	"errors"
	"strings"

	userModel "backend/app/model/user"
	"backend/app/types/consts"
	"backend/app/types/dto"
	authError "backend/app/types/errorn"
	"backend/app/types/meta"
	"backend/utils/envx"
	"backend/utils/errorx"
	"backend/utils/logs"
	"backend/utils/secret"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type UserRepo interface {
	GetUserByUsername(ctx context.Context, username string) (*userModel.User, error)
	GetUserByID(ctx context.Context, userID uint) (*userModel.User, error)
	UpdateUserInfo(ctx context.Context, userID uint, updates map[string]interface{}) error
}

type UserLogicParams struct {
	fx.In

	UserRepo UserRepo
}

type UserLogic struct {
	userRepo UserRepo
	jwt      *secret.JWT
}

func NewUserLogic(params UserLogicParams) *UserLogic {
	// 初始化 JWT 实例
	accessTokenExpire, err := envx.GetDuration(consts.AccessTokenExpire)
	if err != nil {
		logs.Error("获取 AccessTokenExpire 配置失败", "error", err.Error())
		panic(err)
	}
	refreshTokenExpire, err := envx.GetDuration(consts.RefreshTokenExpire)
	if err != nil {
		logs.Error("获取 RefreshTokenExpire 配置失败", "error", err.Error())
		panic(err)
	}
	jwtSecret, err := envx.GetString(consts.JWTSecret)
	if err != nil {
		logs.Error("获取 JWT_SECRET 配置失败", "error", err.Error())
		panic(err)
	}

	jwt := secret.NewJWT(secret.TokenConfig{
		AccessTokenExpire:  accessTokenExpire,
		RefreshTokenExpire: refreshTokenExpire,
		Secret:             jwtSecret,
	})

	return &UserLogic{
		userRepo: params.UserRepo,
		jwt:      jwt,
	}
}

func (l *UserLogic) Login(ctx context.Context, username string, password string) (*dto.UserDTO, *dto.TokenDTO, error) {
	// 查询用户
	user, err := l.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.CtxWarnf(ctx, "用户不存在: username=%s", username)
			return nil, nil, errorx.New(authError.AuthErrUserNotFound, errorx.K("user_uid", username))
		}
		logs.CtxErrorf(ctx, "查询用户失败: username=%s, error=%s", username, err.Error())
		return nil, nil, errorx.Wrap(err, authError.AuthErrUserNotFound, errorx.K("user_uid", username))
	}

	// 验证密码
	if !secret.VerifyPassword(password, user.PasswordHash) {
		logs.CtxWarnf(ctx, "密码错误: username=%s, user_id=%d", username, user.ID)
		return nil, nil, errorx.New(authError.AuthErrPasswordIncorrect)
	}

	// 生成 access token
	accessToken, _, err := l.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		logs.CtxErrorf(ctx, "生成 access token 失败: user_id=%d, error=%s", user.ID, err.Error())
		return nil, nil, errorx.Wrap(err, authError.AuthErrTokenInvalid)
	}

	// 生成 refresh token
	refreshToken, _, err := l.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		logs.CtxErrorf(ctx, "生成 refresh token 失败: user_id=%d, error=%s", user.ID, err.Error())
		return nil, nil, errorx.Wrap(err, authError.AuthErrTokenInvalid)
	}

	// 构建返回数据
	userDTO := &dto.UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		NickName: user.NickName,
		Avatar:   user.Avatar,
	}

	tokenDTO := &dto.TokenDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return userDTO, tokenDTO, nil
}

func (l *UserLogic) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenDTO, error) {
	// 解析 refresh token
	claims, err := l.jwt.ParseToken(refreshToken)
	if err != nil {
		logs.CtxWarnf(ctx, "解析 refresh token 失败: error=%s", err.Error())
		// 根据错误类型返回不同的错误码
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "expired") || strings.Contains(errStr, "exp") {
			return nil, errorx.New(authError.AuthErrTokenExpired)
		} else if strings.Contains(errStr, "malformed") || strings.Contains(errStr, "invalid character") {
			return nil, errorx.New(authError.AuthErrTokenMalformed, errorx.K("reason", err.Error()))
		} else if strings.Contains(errStr, "signature") {
			return nil, errorx.New(authError.AuthErrTokenSignature)
		} else {
			return nil, errorx.New(authError.AuthErrTokenInvalid, errorx.K("reason", err.Error()))
		}
	}

	// 验证用户是否存在
	user, err := l.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.CtxWarnf(ctx, "用户不存在: user_id=%d", claims.UserID)
			return nil, errorx.New(authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", claims.UserID))
		}
		logs.CtxErrorf(ctx, "查询用户失败: user_id=%d, error=%s", claims.UserID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", claims.UserID))
	}

	// 生成新的 access token
	accessToken, _, err := l.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		logs.CtxErrorf(ctx, "生成 access token 失败: user_id=%d, error=%s", user.ID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrTokenInvalid)
	}

	// 生成新的 refresh token
	newRefreshToken, _, err := l.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		logs.CtxErrorf(ctx, "生成 refresh token 失败: user_id=%d, error=%s", user.ID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrTokenInvalid)
	}

	tokenDTO := &dto.TokenDTO{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}

	return tokenDTO, nil
}

func (l *UserLogic) GetUserInfo(ctx context.Context) (*dto.UserDTO, error) {
	// 从 context 中获取用户ID
	userIDValue := ctx.Value(meta.ContextKeyUserID)
	if userIDValue == nil {
		logs.CtxWarnf(ctx, "context 中未找到 user_id")
		return nil, errorx.New(authError.AuthErrTokenRequired)
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		logs.CtxWarnf(ctx, "context 中的 user_id 类型错误")
		return nil, errorx.New(authError.AuthErrTokenInvalid)
	}

	// 查询用户信息
	user, err := l.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.CtxWarnf(ctx, "用户不存在: user_id=%d", userID)
			return nil, errorx.New(authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", userID))
		}
		logs.CtxErrorf(ctx, "查询用户失败: user_id=%d, error=%s", userID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", userID))
	}

	// 构建返回数据
	userDTO := &dto.UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		NickName: user.NickName,
		Avatar:   user.Avatar,
	}

	return userDTO, nil
}

func (l *UserLogic) UpdateUserInfo(ctx context.Context, nickName *string, avatar *string) (*dto.UserDTO, error) {
	// 从 context 中获取用户ID
	userIDValue := ctx.Value(meta.ContextKeyUserID)
	if userIDValue == nil {
		logs.CtxWarnf(ctx, "context 中未找到 user_id")
		return nil, errorx.New(authError.AuthErrTokenRequired)
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		logs.CtxWarnf(ctx, "context 中的 user_id 类型错误")
		return nil, errorx.New(authError.AuthErrTokenInvalid)
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if nickName != nil {
		updates["nick_name"] = *nickName
	}
	if avatar != nil {
		updates["avatar"] = *avatar
	}

	// 如果没有需要更新的字段，直接返回当前用户信息
	if len(updates) == 0 {
		return l.GetUserInfo(ctx)
	}

	// 更新用户信息
	err := l.userRepo.UpdateUserInfo(ctx, userID, updates)
	if err != nil {
		logs.CtxErrorf(ctx, "更新用户信息失败: user_id=%d, error=%s", userID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrUserUpdateFailed)
	}

	// 重新查询用户信息
	user, err := l.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.CtxWarnf(ctx, "用户不存在: user_id=%d", userID)
			return nil, errorx.New(authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", userID))
		}
		logs.CtxErrorf(ctx, "查询用户失败: user_id=%d, error=%s", userID, err.Error())
		return nil, errorx.Wrap(err, authError.AuthErrUserNotFound, errorx.Kf("user_uid", "%d", userID))
	}

	// 构建返回数据
	userDTO := &dto.UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		NickName: user.NickName,
		Avatar:   user.Avatar,
	}

	return userDTO, nil
}
