/**
 * 用户详情页面
 */
import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { getUserInfo, updateUserInfo } from '@/api/userApi'
import { uploadFile } from '@/api/fileApi'
import { useUserStore } from '@/store/userStore'
import type { User } from '../../types/user'

function UserProfile() {
  const navigate = useNavigate()
  const { user, setUser, clearUser } = useUserStore()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [userInfo, setUserInfo] = useState<User | null>(null)
  const [nickName, setNickName] = useState('')
  const [avatar, setAvatar] = useState('')
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  // 加载用户信息
  useEffect(() => {
    const loadUserInfo = async () => {
      try {
        const info = await getUserInfo()
        setUserInfo(info)
        setUser(info)
        setNickName(info.nick_name)
        setAvatar(info.avatar)
      } catch (err) {
        setError(err instanceof Error ? err.message : '获取用户信息失败')
        // 如果获取失败，可能是 token 过期，跳转到登录页
        setTimeout(() => navigate('/login'), 2000)
      }
    }

    loadUserInfo()
  }, [navigate, setUser])

  // 处理头像上传
  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // 验证文件类型
    if (!file.type.startsWith('image/')) {
      setError('请选择图片文件')
      return
    }

    // 验证文件大小（限制 5MB）
    if (file.size > 5 * 1024 * 1024) {
      setError('图片大小不能超过 5MB')
      return
    }

    setUploading(true)
    setError('')

    try {
      const response = await uploadFile(file)
      setAvatar(response.file_url)
      setSuccess('头像上传成功！请点击保存更新')
    } catch (err) {
      setError(err instanceof Error ? err.message : '头像上传失败')
    } finally {
      setUploading(false)
    }
  }

  // 处理信息更新
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')

    if (!nickName.trim()) {
      setError('昵称不能为空')
      return
    }

    setLoading(true)

    try {
      const response = await updateUserInfo({
        nick_name: nickName,
        avatar: avatar,
      })

      // 更新本地状态
      if (userInfo) {
        const updatedUser = {
          ...userInfo,
          nick_name: response.nick_name,
          avatar: response.avatar,
        }
        setUserInfo(updatedUser)
        setUser(updatedUser)
      }

      setSuccess('信息更新成功！✨')
    } catch (err) {
      setError(err instanceof Error ? err.message : '更新失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  // 处理登出
  const handleLogout = () => {
    clearUser()
    navigate('/login')
  }

  if (!userInfo) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 text-lg">加载中...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-pink-100 p-4">
      <div className="mx-auto max-w-2xl space-y-6 py-8">
        {/* 头部 */}
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">个人中心</h1>
          <Button variant="outline" onClick={handleLogout}>
            退出登录
          </Button>
        </div>

        {/* 用户信息卡片 */}
        <Card>
          <CardHeader>
            <CardTitle>基本信息</CardTitle>
            <CardDescription>查看和编辑你的个人信息</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* 头像 */}
            <div className="flex flex-col items-center space-y-4">
              <Avatar className="size-24">
                <AvatarImage src={avatar} alt={nickName} />
                <AvatarFallback>{nickName.charAt(0).toUpperCase()}</AvatarFallback>
              </Avatar>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleAvatarChange}
                className="hidden"
              />
              <Button
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
              >
                {uploading ? '上传中...' : '更换头像'}
              </Button>
            </div>

            {/* 表单 */}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="user-id">用户 ID</Label>
                <Input id="user-id" value={userInfo.user_id} disabled />
              </div>

              <div className="space-y-2">
                <Label htmlFor="username">用户名</Label>
                <Input id="username" value={userInfo.username} disabled />
              </div>

              <div className="space-y-2">
                <Label htmlFor="nickname">昵称</Label>
                <Input
                  id="nickname"
                  type="text"
                  placeholder="请输入昵称"
                  value={nickName}
                  onChange={e => setNickName(e.target.value)}
                  disabled={loading}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="avatar-url">头像 URL</Label>
                <Input
                  id="avatar-url"
                  type="text"
                  placeholder="头像 URL"
                  value={avatar}
                  onChange={e => setAvatar(e.target.value)}
                  disabled={loading}
                />
              </div>

              {error && (
                <div className="rounded-md bg-red-50 p-3 text-sm text-red-600">{error}</div>
              )}

              {success && (
                <div className="rounded-md bg-green-50 p-3 text-sm text-green-600">{success}</div>
              )}

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? '保存中...' : '保存修改'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default UserProfile
