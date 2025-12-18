/**
 * 用户相关类型定义
 */

// 用户信息
export interface User {
  user_id: number
  username: string
  nick_name: string
  avatar: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  user_id: number
  access_token: string
  refresh_token: string
}

// 刷新令牌请求
export interface RefreshTokenRequest {
  refresh_token: string
}

// 刷新令牌响应
export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
}

// 更新用户信息请求
export interface UpdateUserInfoRequest {
  nick_name?: string
  avatar?: string
}

// 更新用户信息响应
export interface UpdateUserInfoResponse {
  user_id: number
  nick_name: string
  avatar: string
}
