/**
 * 环境变量配置模块
 */

export const env = {
  // API 配置
  apiUrl: import.meta.env.BACKEND_SERVER_API_URL || 'http://localhost:6512',
  apiPrefix: import.meta.env.BACKEND_SERVER_API_PREFIX || '/api',
  apiTimeout: Number(import.meta.env.BACKEND_SERVER_API_TIMEOUT) || 100000,
}
