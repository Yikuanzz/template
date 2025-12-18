---
alwaysApply: true
---

# Frontend 开发规范

## 技术栈

- **框架**: React 19.2.0 + TypeScript 5.9.3
- **构建工具**: Vite 7.2.4
- **路由**: React Router 7.11.0
- **状态管理**: Zustand 5.0.9
- **HTTP 客户端**: Axios 1.13.2
- **样式方案**: Tailwind CSS 4.1.18
- **组件库**: shadcn/ui（优先使用）
- **代码规范**: ESLint 9 + Prettier 3

## 项目架构

### 目录结构

```shell
frontend/
├── src/
│   ├── api/          # API 接口定义和调用
│   ├── components/   # 可复用组件
│   │   ├── ui/       # shadcn/ui 组件（通过 CLI 安装）
│   │   └── ...       # 自定义业务组件（如 ProtectedRoute）
│   ├── pages/        # 页面组件
│   ├── store/        # Zustand 状态管理
│   ├── lib/          # 工具函数（shadcn/ui 的 utils.ts）
│   ├── assets/       # 静态资源
│   ├── App.tsx       # 根组件（路由配置）
│   ├── main.tsx      # 应用入口
│   └── index.css     # 全局样式（包含 shadcn/ui CSS 变量）
├── types/            # TypeScript 类型定义（按模块组织）
│   ├── user.ts       # 用户相关类型
│   └── file.ts       # 文件相关类型
├── utils/            # 工具模块
│   ├── http.ts       # HTTP 客户端配置（Axios 实例和拦截器）
│   └── env.ts        # 环境变量配置
├── components.json   # shadcn/ui 配置文件
└── public/           # 公共静态资源
```

## 代码风格规范

### TypeScript

- **严格模式**: 启用 TypeScript 严格类型检查
- **类型定义**: 所有函数、组件、变量必须明确类型
- **禁止使用 `any`**: 除非特殊情况，必须使用具体类型
- **类型文件**: 类型定义统一放在根目录 `types/` 目录，按功能模块组织
- **路径别名**: 使用 `@/` 作为 `src/` 的别名，在 `vite.config.ts` 和 `tsconfig.app.json` 中配置

### 路径别名

项目配置了路径别名，方便导入：

- **`@/`**: 指向 `src/` 目录
- **配置位置**:
  - `vite.config.ts`: 配置 Vite 的路径解析
  - `tsconfig.app.json`: 配置 TypeScript 的路径映射

```typescript
// ✅ 正确示例 - 使用路径别名
import { Button } from '@/components/ui/button'
import { useUserStore } from '@/store/userStore'
import { login } from '@/api/userApi'
import type { User } from '@/types/user'

// ❌ 错误示例 - 使用相对路径
import { Button } from '../../components/ui/button'
```

### React 组件

- **函数式组件**: 统一使用函数式组件，不使用类组件
- **Hooks 优先**: 使用 React Hooks 进行状态管理和副作用处理
- **组件命名**: 使用 PascalCase，文件名与组件名保持一致
- **Props 类型**: 使用 TypeScript interface 定义组件 Props
- **JSDoc 注释**: 使用 JSDoc 风格注释（`/** */`）为组件和函数添加文档

```typescript
// ✅ 正确示例
/**
 * 按钮组件
 */
interface ButtonProps {
  label: string
  onClick: () => void
  disabled?: boolean
}

function Button({ label, onClick, disabled = false }: ButtonProps) {
  return (
    <button onClick={onClick} disabled={disabled}>
      {label}
    </button>
  )
}
```

### 代码格式化

遵循 Prettier 配置：

- **分号**: 不使用分号
- **引号**: 使用单引号
- **缩进**: 2 个空格
- **行宽**: 100 字符
- **尾随逗号**: ES5 风格
- **箭头函数**: 单参数时避免括号

```typescript
// ✅ 正确示例
const fetchData = async (id: string) => {
  const response = await api.get(`/data/${id}`)
  return response.data
}
```

### 样式规范

- **Tailwind CSS**: 优先使用 Tailwind 工具类
- **类名顺序**: 按照功能分组（布局、间距、颜色、字体等）
- **响应式**: 使用 Tailwind 响应式前缀（sm:, md:, lg: 等）
- **自定义样式**: 仅在必要时使用 CSS 模块或内联样式
- **CSS 变量**: shadcn/ui 使用 CSS 变量定义主题，在 `src/index.css` 中配置

```tsx
// ✅ 正确示例 - 使用 Tailwind 类名
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow-md">
  <h2 className="text-xl font-bold text-gray-800">标题</h2>
</div>

// ✅ 正确示例 - 使用 shadcn/ui CSS 变量
<div className="bg-background text-foreground border-border">
  <p className="text-muted-foreground">提示文本</p>
</div>
```

### CSS 变量（shadcn/ui 主题）

shadcn/ui 使用 CSS 变量定义主题颜色，在 `src/index.css` 中配置：

- **`:root`**: 定义浅色主题变量
- **`.dark`**: 定义深色主题变量
- **变量命名**: 使用 `--background`, `--foreground`, `--primary` 等语义化名称
- **使用方式**: 通过 Tailwind 类名使用，如 `bg-background`, `text-foreground`

## 组件库规范

### shadcn/ui

**优先使用 shadcn/ui 组件库**，这是一个基于开放代码的组件系统，不是传统的 NPM 包。

#### 核心理念

- **开放代码**: 组件代码直接复制到项目中，可以完全自定义和修改
- **组合式设计**: 所有组件使用统一的组合式接口，易于扩展
- **美观默认**: 精心设计的默认样式，开箱即用
- **AI 友好**: 开放代码便于 AI 工具理解和改进

#### 安装和使用

- **CLI 安装**: 使用 `npx shadcn@latest add [component]` 安装组件
- **组件位置**: 所有 shadcn/ui 组件统一放在 `src/components/ui/` 目录
- **配置文件**: 通过 `components.json` 管理组件配置和主题
- **直接修改**: 组件代码在项目中，可以直接编辑和自定义

```bash
# ✅ 安装组件示例
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add dialog
```

#### 组件使用规范

- **优先使用**: 需要 UI 组件时，优先查看 shadcn/ui 是否有对应组件
- **直接导入**: 从 `@/components/ui` 导入组件，无需通过 NPM 包
- **工具函数**: 使用 `@/lib/utils` 中的 `cn` 函数合并 Tailwind 类名
- **自由定制**: 可以根据需求直接修改组件代码
- **组合使用**: 利用 shadcn/ui 的组合式设计，组合多个组件构建复杂 UI

```tsx
// ✅ 正确示例 - 使用 shadcn/ui 组件
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Dialog, DialogContent, DialogTrigger } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

function UserCard({ className }: { className?: string }) {
  return (
    <Card className={cn('w-full', className)}>
      <CardHeader>
        <CardTitle>用户信息</CardTitle>
      </CardHeader>
      <CardContent>
        <Dialog>
          <DialogTrigger asChild>
            <Button>查看详情</Button>
          </DialogTrigger>
          <DialogContent>
            <p>用户详情内容</p>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  )
}
```

#### 工具函数

- **cn 函数**: shadcn/ui 提供的类名合并工具，位于 `src/lib/utils.ts`
- **使用场景**: 用于条件性地合并 Tailwind 类名，避免类名冲突

```typescript
// ✅ 正确示例 - 使用 cn 函数
import { cn } from '@/lib/utils'

function Button({ className, variant }: ButtonProps) {
  return (
    <button
      className={cn(
        'base-button-styles',
        variant === 'primary' && 'bg-blue-500',
        className
      )}
    >
      Click me
    </button>
  )
}
```

#### 组件自定义

- **直接编辑**: 组件代码在 `src/components/ui/` 目录，可以直接修改
- **保持一致性**: 自定义时保持与 shadcn/ui 设计系统的一致性
- **类型安全**: 所有组件都有完整的 TypeScript 类型定义
- **主题定制**: 通过 Tailwind CSS 配置和 `components.json` 定制主题

```tsx
// ✅ 正确示例 - 自定义 shadcn/ui 组件
// 可以直接修改 src/components/ui/button.tsx 来定制按钮样式和行为
import { Button } from '@/components/ui/button'

// 使用自定义的按钮
function CustomButton() {
  return <Button variant="custom">自定义按钮</Button>
}
```

#### 组件目录组织

- **ui 目录**: `src/components/ui/` 存放所有 shadcn/ui 组件
- **业务组件**: `src/components/` 其他目录存放自定义业务组件
- **命名规范**: shadcn/ui 组件保持原文件名，业务组件使用 PascalCase

```shell
src/components/
├── ui/              # shadcn/ui 组件（不要手动创建，通过 CLI 安装）
│   ├── button.tsx
│   ├── card.tsx
│   └── dialog.tsx
├── forms/           # 自定义表单组件
│   └── UserForm.tsx
└── layout/          # 自定义布局组件
    └── Header.tsx
```

#### 更新和维护

- **上游更新**: 使用 `npx shadcn@latest diff` 查看组件更新
- **选择性更新**: 只更新需要的组件，避免覆盖自定义修改
- **版本控制**: 组件代码纳入版本控制，便于团队协作
- **文档参考**: 参考 [shadcn/ui 官方文档](https://ui.shadcn.com/docs) 了解组件用法

## 状态管理

### Zustand Store

- **Store 位置**: 统一放在 `src/store/` 目录
- **命名规范**: 使用 camelCase，文件名以 `store.ts` 结尾
- **类型定义**: Store 状态和操作必须定义明确的类型
- **持久化**: 使用 localStorage 进行状态持久化，在 Store 中实现 `init` 方法恢复状态
- **JSDoc 注释**: 为 Store 添加 JSDoc 注释说明用途

```typescript
// ✅ 正确示例
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

export const useUserStore = create<UserState>((set) => ({
  user: null,
  accessToken: null,
  refreshToken: null,

  setUser: (user) => set({ user }),

  setTokens: (accessToken, refreshToken) => {
    // 保存到 localStorage
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
    set({ accessToken, refreshToken })
  },

  clearUser: () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    set({ user: null, accessToken: null, refreshToken: null })
  },

  init: () => {
    const accessToken = localStorage.getItem('access_token')
    const refreshToken = localStorage.getItem('refresh_token')
    if (accessToken && refreshToken) {
      set({ accessToken, refreshToken })
    }
  },
}))
```

## API 调用

### HTTP 客户端

- **使用 Axios**: 统一使用 Axios 进行 HTTP 请求
- **配置位置**: HTTP 客户端配置放在 `utils/http.ts`
- **环境变量**: API 配置通过 `utils/env.ts` 模块访问环境变量
- **类型安全**: API 请求和响应必须定义 TypeScript 类型
- **拦截器**: 使用请求拦截器自动添加 Authorization token，使用响应拦截器处理 401 错误

```typescript
// ✅ 正确示例 - HTTP 客户端配置
import axios from 'axios'
import { env } from './env'

export const http = axios.create({
  baseURL: `${env.apiUrl}${env.apiPrefix}`,
  timeout: env.apiTimeout,
})

// 请求拦截器：自动添加 token
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：处理 401 错误
http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

### API 接口组织

- **目录结构**: API 接口按功能模块组织在 `src/api/` 目录
- **命名规范**: 使用 camelCase，文件名以 `Api.ts` 结尾（如 `userApi.ts`）
- **错误处理**: 统一处理 API 错误，检查响应中的 `code` 字段
- **类型定义**: 请求和响应类型定义在 `types/` 目录对应的模块文件中
- **JSDoc 注释**: 为每个 API 函数添加 JSDoc 注释

```typescript
// ✅ 正确示例 - API 接口定义
/**
 * 用户相关 API 接口
 */
import { http, type ApiResponse } from '../../utils/http'
import type { LoginRequest, LoginResponse, User } from '../../types/user'

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
 * 获取用户信息
 */
export const getUserInfo = async (): Promise<User> => {
  const response = await http.get<ApiResponse<User>>('/user/info')
  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '获取用户信息失败')
}
```

## 路由管理

### React Router

- **路由配置**: 使用 React Router 7 的 `BrowserRouter` 和 `Routes`
- **路由定义**: 路由配置在 `App.tsx` 中统一管理
- **受保护路由**: 使用 `ProtectedRoute` 组件包装需要认证的路由
- **懒加载**: 页面组件使用 React.lazy 进行代码分割（可选）

```typescript
// ✅ 正确示例 - 路由配置
/**
 * App 根组件 - 路由配置
 */
import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import Login from '@/pages/Login'
import UserProfile from '@/pages/UserProfile'
import ProtectedRoute from '@/components/ProtectedRoute'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<Login />} />
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <UserProfile />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
```

### 受保护路由

- **ProtectedRoute 组件**: 位于 `src/components/ProtectedRoute.tsx`
- **功能**: 检查用户认证状态，未登录时重定向到登录页
- **实现**: 使用 Zustand store 检查 token，或从 localStorage 读取

```typescript
// ✅ 正确示例 - 受保护路由组件
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
  
  if (!accessToken) {
    const token = localStorage.getItem('access_token')
    if (!token) {
      return <Navigate to="/login" replace />
    }
  }

  return <>{children}</>
}
```

## 环境变量

### 配置管理

- **环境变量文件**: 使用 `.env` 文件管理环境变量（参考 `.example.env`）
- **变量命名**: 使用大写字母和下划线，如 `BACKEND_SERVER_API_URL`
- **访问方式**: 通过 `utils/env.ts` 模块统一访问环境变量
- **类型安全**: 在 `utils/env.ts` 中定义环境变量对象，提供默认值

```typescript
// ✅ 正确示例 - 环境变量配置
/**
 * 环境变量配置模块
 */
export const env = {
  // API 配置
  apiUrl: import.meta.env.BACKEND_SERVER_API_URL || 'http://localhost:6512',
  apiPrefix: import.meta.env.BACKEND_SERVER_API_PREFIX || '/api',
  apiTimeout: Number(import.meta.env.BACKEND_SERVER_API_TIMEOUT) || 100000,
}

// 使用方式
import { env } from '@/utils/env'
const baseURL = `${env.apiUrl}${env.apiPrefix}`
```

### 环境变量列表

- `BACKEND_SERVER_API_URL`: 后端 API 服务器地址
- `BACKEND_SERVER_API_PREFIX`: API 路径前缀
- `BACKEND_SERVER_API_TIMEOUT`: API 请求超时时间（毫秒）

## 最佳实践

### 性能优化

- **代码分割**: 使用 React.lazy 和 Suspense 进行路由级别的代码分割
- **Memo 优化**: 合理使用 React.memo、useMemo、useCallback 避免不必要的重渲染
- **图片优化**: 使用适当的图片格式和尺寸，考虑使用 WebP

### 错误处理

- **错误边界**: 使用 Error Boundary 捕获组件树错误
- **API 错误**: 统一处理 API 错误，提供友好的错误提示
- **类型安全**: 使用 TypeScript 类型守卫进行运行时类型检查

### 代码组织

- **单一职责**: 每个组件、函数只负责一个功能
- **可复用性**: 提取可复用的逻辑到自定义 Hooks 或工具函数
- **命名清晰**: 使用有意义的变量名和函数名，避免缩写
- **JSDoc 注释**: 为所有公开的函数、组件、类型添加 JSDoc 注释
- **类型导入**: 使用 `import type` 导入类型，提高代码可读性

```typescript
// ✅ 正确示例 - 类型导入
import type { User, LoginRequest } from '@/types/user'
import { login } from '@/api/userApi'

// ❌ 错误示例
import { User, LoginRequest } from '@/types/user'
```

### 测试

- **单元测试**: 为工具函数和业务逻辑编写单元测试
- **组件测试**: 使用 React Testing Library 测试组件
- **类型测试**: 利用 TypeScript 的类型系统进行编译时检查

## 禁止事项

- ❌ 禁止使用 `any` 类型（除非特殊情况）
- ❌ 禁止在组件中直接使用 `console.log`（生产环境）
- ❌ 禁止使用内联样式（优先使用 Tailwind）
- ❌ 禁止使用类组件（统一使用函数式组件）
- ❌ 禁止忽略 TypeScript 类型错误（使用 `@ts-ignore`）
- ❌ 禁止提交未格式化的代码（必须通过 Prettier 格式化）
- ❌ 禁止手动创建 `src/components/ui/` 目录下的组件（必须通过 shadcn CLI 安装）
- ❌ 禁止使用其他 UI 组件库（如 Material-UI、Ant Design 等），优先使用 shadcn/ui
- ❌ 禁止在 Store 中直接使用 localStorage（应在 Store 的 action 方法中统一处理）
- ❌ 禁止在组件中直接访问 `import.meta.env`（应通过 `utils/env.ts` 模块访问）
- ❌ 禁止在 API 函数中直接使用 `axios`（应使用 `utils/http.ts` 导出的 `http` 实例）
