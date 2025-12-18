/**
 * App 根组件 - 路由配置
 */
import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import Login from '@/pages/Login'
import UserProfile from '@/pages/UserProfile'
import ProtectedRoute from '@/components/ProtectedRoute'
import { useUserStore } from '@/store/userStore'

function App() {
  const { init } = useUserStore()

  // 初始化：从 localStorage 恢复状态
  useEffect(() => {
    init()
  }, [init])

  return (
    <BrowserRouter>
      <Routes>
        {/* 根路径重定向到登录页 */}
        <Route path="/" element={<Navigate to="/login" replace />} />
        
        {/* 登录页 */}
        <Route path="/login" element={<Login />} />
        
        {/* 用户详情页（受保护） */}
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <UserProfile />
            </ProtectedRoute>
          }
        />
        
        {/* 404 页面 */}
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
