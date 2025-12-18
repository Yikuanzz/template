/**
 * 用户相关 API 接口
 */
import { http, type ApiResponse } from '../../utils/http'
import type {
  LoginRequest,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  User,
  UpdateUserInfoRequest,
  UpdateUserInfoResponse,
} from '../../types/user'

/**
 * 用户登录
 */
export const login = async (data: LoginRequest): Promise<LoginResponse> => {
  const response = await http.post<ApiResponse<LoginResponse>>('/user/login', data)
  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '登录失败')
}

/**
 * 刷新令牌
 */
export const refreshToken = async (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
  const response = await http.post<ApiResponse<RefreshTokenResponse>>('/user/refresh-token', data)
  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '刷新令牌失败')
}

/**
 * 获取用户信息
 */
export const getUserInfo = async (): Promise<User> => {
  const response = await http.get<ApiResponse<User>>('/user/info')
  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '获取用户信息失败')
}

/**
 * 更新用户信息
 */
export const updateUserInfo = async (
  data: UpdateUserInfoRequest
): Promise<UpdateUserInfoResponse> => {
  const response = await http.put<ApiResponse<UpdateUserInfoResponse>>('/user/info', data)
  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '更新用户信息失败')
}
