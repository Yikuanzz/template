package user

import (
	"context"

	"backend/app/types/dto"
	authError "backend/app/types/errorn"
	"backend/utils/bind"
	"backend/utils/handle"
	"backend/utils/logs"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type UserLogic interface {
	Login(ctx context.Context, username string, password string) (*dto.UserDTO, *dto.TokenDTO, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenDTO, error)
	GetUserInfo(ctx context.Context) (*dto.UserDTO, error)
	UpdateUserInfo(ctx context.Context, nickName *string, avatar *string) (*dto.UserDTO, error)
}

type UserHandlerParams struct {
	fx.In

	UserLogic UserLogic
}

type UserHandler struct {
	userLogic UserLogic
}

func NewUserHandler(params UserHandlerParams) *UserHandler {
	return &UserHandler{
		userLogic: params.UserLogic,
	}
}

var userBindConfig = bind.FieldErrorConfig{
	InvalidParamCode: authError.AuthErrTokenInvalid,
	RequiredCode:     authError.AuthErrTokenRequired,
	FieldLabels: map[string]string{
		"username":      "用户名",
		"password":      "密码",
		"refresh_token": "刷新令牌",
	},
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名、密码和验证码进行登录，返回访问令牌和刷新令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginReq true "登录请求"
// @Success 200 {object} handle.Response{data=LoginResp} "成功"
// @Failure 400 {object} handle.Response "请求参数错误"
// @Failure 401 {object} handle.Response "用户名或密码错误"
// @Failure 500 {object} handle.Response "服务器内部错误"
// @Router /api/user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginReq
	if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
		handle.HandleErrorWithContext(c, err, "登录", nil)
		return
	}

	u, t, err := h.userLogic.Login(ctx, req.Username, req.Password)
	if err != nil {
		handle.HandleErrorWithContext(c, err, "登录", nil)
		return
	}

	logs.CtxInfof(ctx, "用户登录成功: user_id=%d, username=%s", u.UserID, u.Username)
	handle.Success(c, LoginResp{
		UserID:       u.UserID,
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	})
}

// RefreshToken 刷新 Token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenReq true "刷新令牌请求"
// @Success 200 {object} handle.Response{data=RefreshTokenResp} "成功"
// @Failure 400 {object} handle.Response "请求参数错误"
// @Failure 401 {object} handle.Response "刷新令牌无效或已过期"
// @Failure 500 {object} handle.Response "服务器内部错误"
// @Router /api/user/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req RefreshTokenReq
	if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
		handle.HandleErrorWithContext(c, err, "刷新Token", nil)
		return
	}

	result, err := h.userLogic.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		handle.HandleErrorWithContext(c, err, "刷新Token", nil)
		return
	}

	logs.CtxInfof(ctx, "Token 刷新成功")
	handle.Success(c, RefreshTokenResp{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的基本信息和菜单列表
// @Tags 用户认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handle.Response{data=GetUserInfoResp} "成功"
// @Failure 401 {object} handle.Response "未授权"
// @Failure 500 {object} handle.Response "服务器内部错误"
// @Router /api/user/info [get]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()

	u, err := h.userLogic.GetUserInfo(ctx)
	if err != nil {
		handle.HandleErrorWithContext(c, err, "获取用户信息", nil)
		return
	}

	logs.CtxInfof(ctx, "获取用户信息成功: user_id=%d", u.UserID)
	handle.Success(c, GetUserInfoResp{
		UserID:   u.UserID,
		Username: u.Username,
		NickName: u.NickName,
		Avatar:   u.Avatar,
	})
}

// UpateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的基本信息和菜单列表
// @Tags 用户认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpateUserInfoReq true "更新用户信息请求"
// @Success 200 {object} handle.Response{data=UpateUserInfoResp} "成功"
// @Failure 401 {object} handle.Response "未授权"
// @Failure 500 {object} handle.Response "服务器内部错误"
// @Router /api/user/info [put]
func (h *UserHandler) UpateUserInfo(c *gin.Context) {
	ctx := c.Request.Context()

	var req UpateUserInfoReq
	if err := bind.ShouldBindJSON(c, &req, userBindConfig); err != nil {
		handle.HandleErrorWithContext(c, err, "更新用户信息", nil)
		return
	}

	result, err := h.userLogic.UpdateUserInfo(ctx, req.NickName, req.Avatar)
	if err != nil {
		handle.HandleErrorWithContext(c, err, "更新用户信息", nil)
		return
	}

	logs.CtxInfof(ctx, "更新用户信息成功: user_id=%d", result.UserID)
	handle.Success(c, UpateUserInfoResp{
		UserID:   result.UserID,
		NickName: result.NickName,
		Avatar:   result.Avatar,
	})
}
