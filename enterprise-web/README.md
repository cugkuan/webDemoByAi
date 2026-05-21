# enterprise-web

企业级任务管理系统前端，基于 React + Vite 构建。

## 项目结构

```
src/
├── api/              # API 层 - HTTP 请求封装
│   ├── client.js     # Axios 实例（拦截器、可取消请求）
│   ├── auth.js       # 认证 API（登录/注册）
│   ├── tasks.js      # 任务 API（CRUD + 分页）
│   └── system.js     # 系统 API（信息/状态/健康检查）
├── components/       # 通用组件
│   ├── Navbar.jsx    # 导航栏
│   ├── ProtectedRoute.jsx  # 路由守卫
│   ├── Pagination.jsx      # 分页组件
│   ├── ErrorBoundary.jsx   # 错误边界
│   ├── Skeleton.jsx        # 骨架屏
│   └── TaskItem.jsx        # 任务项（memo 优化）
├── context/          # 全局状态
│   ├── AuthContext.jsx     # 认证上下文
│   └── useAuth.js          # 认证 Hook
├── pages/            # 页面组件
│   ├── HomePage.jsx        # 首页
│   ├── LoginPage.jsx       # 登录
│   ├── RegisterPage.jsx    # 注册
│   ├── TasksPage.jsx       # 任务管理
│   └── ProfilePage.jsx     # 个人信息 + 系统状态
├── config.js         # 全局常量配置
├── App.jsx           # 路由配置（含代码分割）
├── main.jsx          # 入口
└── index.css         # 全局样式
```

## 技术栈

| 技术 | 用途 |
|------|------|
| React 19 | UI 框架 |
| React Router 7 | 路由管理 |
| Axios | HTTP 客户端 |
| Vite 8 | 构建工具 |
| ESLint | 代码规范 |
| CSS Modules | 样式隔离 |

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:5173）
npm run dev

# 代码检查
npm run lint

# 生产构建
npm run build
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_API_BASE` | `http://localhost:8080/api` | API 地址 |

## 组件说明

### 通用组件

| 组件 | 说明 |
|------|------|
| `Navbar` | 顶部导航栏，根据登录状态显示不同菜单 |
| `ProtectedRoute` | 路由守卫，未登录跳转 /login |
| `Pagination` | 分页组件，支持省略号、ARIA 无障碍 |
| `ErrorBoundary` | 错误边界，捕获渲染异常并显示重试按钮 |
| `Skeleton` | 骨架屏，含 TaskSkeleton 和 ProfileSkeleton |
| `TaskItem` | 任务项，支持编辑/完成/删除，React.memo 优化 |

### 页面组件

| 页面 | 路由 | 说明 |
|------|------|------|
| HomePage | `/` | 首页，展示功能特性 |
| LoginPage | `/login` | 登录，防重复提交 |
| RegisterPage | `/register` | 注册，含前端校验 |
| TasksPage | `/tasks` | 任务管理 CRUD + 分页 |
| ProfilePage | `/profile` | 用户信息 + 系统状态 |

## 最佳实践

- **代码分割**: 使用 `React.lazy` + `Suspense` 实现路由级懒加载
- **样式隔离**: 使用 CSS Modules，每个组件独立样式文件
- **性能优化**: TaskItem 使用 `React.memo`，避免不必要的重渲染
- **可取消请求**: 支持 AbortController，组件卸载时取消未完成的请求
- **防重复提交**: 使用 `useRef` 标记提交状态
- **统一错误处理**: Axios 响应拦截器 + ErrorBoundary 双重保障
- **无障碍**: 表单标签关联、ARIA 属性、语义化 HTML
