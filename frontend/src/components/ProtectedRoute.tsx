/**
 * 受保护的路由组件
 */
import { Navigate } from 'react-router'
import { useUserStore } from '@/store/userStore'

interface ProtectedRouteProps {
  children: React.ReactNode
}

function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { accessToken } = useUserStore()
  
  // 如果没有 token，重定向到登录页
  if (!accessToken) {
    const token = localStorage.getItem('access_token')
    if (!token) {
      return <Navigate to="/login" replace />
    }
  }

  return <>{children}</>
}

export default ProtectedRoute

