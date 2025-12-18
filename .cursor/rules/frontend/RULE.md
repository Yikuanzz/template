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
- **代码规范**: ESLint 9 + Prettier 3

## 项目架构

### 目录结构

```shell
frontend/
├── src/
│   ├── api/          # API 接口定义和调用
│   ├── components/   # 可复用组件
│   ├── pages/        # 页面组件
│   ├── store/        # Zustand 状态管理
│   ├── assets/       # 静态资源
│   ├── App.tsx       # 根组件
│   ├── main.tsx      # 应用入口
│   └── index.css     # 全局样式
├── utils/            # 工具函数和类型定义
│   ├── modules/      # 模块化工具（http, env 等）
│   └── types/        # TypeScript 类型定义
└── public/           # 公共静态资源
```

## 代码风格规范

### TypeScript

- **严格模式**: 启用 TypeScript 严格类型检查
- **类型定义**: 所有函数、组件、变量必须明确类型
- **禁止使用 `any`**: 除非特殊情况，必须使用具体类型
- **类型文件**: 类型定义统一放在 `utils/types/` 目录

### React 组件

- **函数式组件**: 统一使用函数式组件，不使用类组件
- **Hooks 优先**: 使用 React Hooks 进行状态管理和副作用处理
- **组件命名**: 使用 PascalCase，文件名与组件名保持一致
- **Props 类型**: 使用 TypeScript interface 定义组件 Props

```typescript
// ✅ 正确示例
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

```tsx
// ✅ 正确示例
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow-md">
  <h2 className="text-xl font-bold text-gray-800">标题</h2>
</div>
```

## 状态管理

### Zustand Store

- **Store 位置**: 统一放在 `src/store/` 目录
- **命名规范**: 使用 camelCase，文件名以 `store.ts` 结尾
- **类型定义**: Store 状态和操作必须定义明确的类型

```typescript
// ✅ 正确示例
import { create } from 'zustand'

interface UserState {
  user: User | null
  setUser: (user: User) => void
  clearUser: () => void
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  clearUser: () => set({ user: null }),
}))
```

## API 调用

### HTTP 客户端

- **使用 Axios**: 统一使用 Axios 进行 HTTP 请求
- **配置位置**: HTTP 客户端配置放在 `utils/modules/http.ts`
- **环境变量**: API 配置通过环境变量管理（`.env` 文件）
- **类型安全**: API 请求和响应必须定义 TypeScript 类型

### API 接口组织

- **目录结构**: API 接口按功能模块组织在 `src/api/` 目录
- **命名规范**: 使用 camelCase，文件名以 `api.ts` 结尾
- **错误处理**: 统一处理 API 错误，使用 try-catch 或拦截器

```typescript
// ✅ 正确示例
import axios from 'axios'
import type { User } from '@/utils/types'

export const getUserById = async (id: string): Promise<User> => {
  const response = await axios.get(`/api/users/${id}`)
  return response.data
}
```

## 路由管理

### React Router

- **路由配置**: 使用 React Router 7 的最新 API
- **路由定义**: 路由配置统一管理，使用类型安全的路由定义
- **懒加载**: 页面组件使用 React.lazy 进行代码分割

```typescript
// ✅ 正确示例
import { lazy } from 'react'
import { createBrowserRouter } from 'react-router'

const HomePage = lazy(() => import('@/pages/Home'))
const AboutPage = lazy(() => import('@/pages/About'))

export const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/about', element: <AboutPage /> },
])
```

## 环境变量

### 配置管理

- **环境变量文件**: 使用 `.env` 文件管理环境变量
- **变量命名**: 使用大写字母和下划线，如 `BACKEND_SERVER_API_URL`
- **类型定义**: 环境变量类型定义在 `utils/types/env.ts`
- **访问方式**: 通过 `utils/modules/env.ts` 模块访问环境变量

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
