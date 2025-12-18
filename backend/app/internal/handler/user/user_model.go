package user

// LoginReq 用户名密码登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=32" label:"用户名" example:"alice123"`
	Password string `json:"password" binding:"required,min=8,max=16" label:"密码" example:"password123"`
}

// LoginResp 登录响应
type LoginResp struct {
	UserID       uint   `json:"user_id" example:"1"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenReq 刷新令牌请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required" label:"刷新令牌" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResp 刷新令牌响应
type RefreshTokenResp struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// GetUserInfoResp 获取用户的响应信息
type GetUserInfoResp struct {
	// 用户基本信息
	UserID   uint   `json:"user_id" example:"1"`
	Username string `json:"username" example:"alice123"`
	NickName string `json:"nick_name" example:"爱丽丝"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
}

// UpateUserInfoReq 更新用户信息请求
type UpateUserInfoReq struct {
	NickName *string `json:"nick_name" binding:"omitempty" label:"昵称" example:"爱丽丝"`
	Avatar   *string `json:"avatar" binding:"omitempty" label:"头像" example:"https://example.com/avatar.jpg"`
}

// UpateUserInfoResp 更新用户信息响应
type UpateUserInfoResp struct {
	UserID   uint   `json:"user_id" example:"1"`
	NickName string `json:"nick_name" example:"爱丽丝"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
}
