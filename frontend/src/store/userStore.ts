/**
 * 用户状态管理
 */
import { create } from 'zustand'
import type { User } from '../../types/user'

interface UserState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null

  // Actions
  setUser: (user: User) => void
  setTokens: (accessToken: string, refreshToken: string) => void
  clearUser: () => void

  // 初始化：从 localStorage 恢复状态
  init: () => void
}

export const useUserStore = create<UserState>(set => ({
  user: null,
  accessToken: null,
  refreshToken: null,

  setUser: user => set({ user }),

  setTokens: (accessToken, refreshToken) => {
    // 保存到 localStorage
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)

    set({ accessToken, refreshToken })
  },

  clearUser: () => {
    // 清除 localStorage
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user_id')

    set({ user: null, accessToken: null, refreshToken: null })
  },

  init: () => {
    // 从 localStorage 恢复状态
    const accessToken = localStorage.getItem('access_token')
    const refreshToken = localStorage.getItem('refresh_token')

    if (accessToken && refreshToken) {
      set({ accessToken, refreshToken })
    }
  },
}))
