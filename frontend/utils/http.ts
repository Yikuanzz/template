/**
 * HTTP 客户端配置模块
 */
import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { env } from './env'

// 创建 axios 实例
export const http = axios.create({
  baseURL: `${env.apiUrl}${env.apiPrefix}`,
  timeout: env.apiTimeout,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
http.interceptors.response.use(
  response => {
    // 直接返回响应数据
    return response
  },
  async (error: AxiosError<ApiErrorResponse>) => {
    // 处理 401 未授权错误
    if (error.response?.status === 401) {
      // 清除本地存储的认证信息
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user_id')

      // 跳转到登录页
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

// API 错误响应类型
interface ApiErrorResponse {
  code: number
  message: string
}

// API 响应类型
export interface ApiResponse<T = unknown> {
  code: number
  message?: string
  data?: T
}
