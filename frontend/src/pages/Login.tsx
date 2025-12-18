/**
 * 登录页面
 */
import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { login } from '@/api/userApi'
import { useUserStore } from '@/store/userStore'

function Login() {
  const navigate = useNavigate()
  const { setTokens } = useUserStore()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // 表单验证
    if (!username || username.length < 3 || username.length > 32) {
      setError('用户名长度必须在 3-32 个字符之间')
      return
    }

    if (!password || password.length < 8 || password.length > 16) {
      setError('密码长度必须在 8-16 个字符之间')
      return
    }

    setLoading(true)

    try {
      const response = await login({ username, password })
      
      // 保存 tokens
      setTokens(response.access_token, response.refresh_token)
      
      // 保存 user_id
      localStorage.setItem('user_id', String(response.user_id))

      // 跳转到用户详情页
      navigate('/profile')
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-center text-2xl font-bold">欢迎回来 👋</CardTitle>
          <CardDescription className="text-center">
            请输入您的账号和密码登录
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                type="text"
                placeholder="请输入用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
                required
                minLength={3}
                maxLength={32}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input
                id="password"
                type="password"
                placeholder="请输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                required
                minLength={8}
                maxLength={16}
              />
            </div>

            {error && (
              <div className="rounded-md bg-red-50 p-3 text-sm text-red-600">
                {error}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? '登录中...' : '登录'}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm text-muted-foreground">
            <p>测试账号：alice123</p>
            <p>测试密码：password123</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default Login

