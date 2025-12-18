# Frontend CLAUDE.md

前端开发指南，适用于 `playerplus_platform/frontend` 目录。

## 技术栈

- **框架**: Vue 3 + TypeScript
- **UI 组件**: Ant Design Vue 4
- **状态管理**: Pinia
- **构建工具**: Vite 5
- **包管理**: pnpm

## 目录结构

```
frontend/
├── src/
│   ├── api/
│   │   └── index.ts         # Axios 配置和 API 调用
│   ├── assets/              # 静态资源（CSS、图片）
│   ├── components/          # 通用组件
│   │   └── MarkdownViewer.vue  # Markdown 渲染组件
│   ├── docs/                # 功能使用说明文档
│   │   ├── faceswap.md      # 换脸功能使用说明
│   │   └── prompts.md       # Prompt 管理使用说明
│   ├── router/
│   │   └── index.ts         # Vue Router 配置
│   ├── stores/
│   │   └── auth.ts          # Pinia 认证状态
│   ├── views/
│   │   ├── LoginView.vue    # 登录页面
│   │   ├── LayoutView.vue   # 主布局容器
│   │   ├── FaceSwapView.vue # 换脸功能页面
│   │   └── PromptsView.vue  # Prompt 管理页面
│   ├── App.vue              # 根组件
│   └── main.ts              # 入口文件
├── dist/                    # 构建产物（git ignored）
├── vite.config.ts           # Vite 配置（含代理）
├── tsconfig.json            # TypeScript 配置
└── package.json
```

## 开发命令

在 `frontend/` 目录下执行：

```bash
# 安装依赖
pnpm install

# 启动开发服务器 (localhost:5173)
pnpm dev

# 构建生产版本
pnpm build

# 代码检查
pnpm lint

# 运行测试
pnpm test
```

## 路由配置

定义在 `src/router/index.ts`：

```
/login                    # 登录页面（无需认证）
/                         # 主布局（需要认证）
├── /faceswap             # 换脸功能页面
└── /prompts              # Prompt 管理页面
```

**路由守卫**：
- 检查 `requiresAuth` 元数据
- 未认证用户重定向到 `/login`
- 已认证用户访问 `/login` 重定向到首页

## 状态管理 (Pinia)

### Auth Store (`stores/auth.ts`)

```typescript
// 状态
state: {
  token: string | null,
  user: string | null
}

// Actions
login(username, password)  // 登录
logout()                   // 登出
checkAuth()                // 检查认证状态
```

Token 存储在 `localStorage`。

## API 客户端 (`api/index.ts`)

基于 Axios 封装：

```typescript
// 请求拦截器：自动添加 Authorization header
// 响应拦截器：处理 401 错误，自动登出

// API 方法
api.login(username, password)
api.detectFaces(imageUrl)
api.createFaceSwap(params)
api.getFaceSwapTask(taskId)
api.uploadMedia(file, onProgress)  // 支持上传进度回调
```

## Vite 代理配置

`vite.config.ts` 中配置开发代理：

```typescript
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

开发时前端请求 `/api/*` 会自动代理到后端 `localhost:8080`。

## 通用组件

### MarkdownViewer.vue

Markdown 渲染组件，用于显示功能使用说明：
- 使用 `marked` 库解析 Markdown
- 使用 `DOMPurify` 防止 XSS 攻击
- 支持完整的 Markdown 语法（标题、列表、代码块、表格等）

使用方式：
```vue
<MarkdownViewer :content="markdownContent" />
```

## 页面组件

### FaceSwapView.vue

换脸功能主页面，包含两个 Tab：「视频换脸」和「使用说明」。

**视频换脸流程**：
1. 上传视频/图片（带圆形进度条显示上传进度）
2. 检测人脸（显示检测到的人脸列表）
3. 选择要替换的人脸
4. 上传替换用的人脸图片
5. 创建换脸任务
6. 轮询任务状态（包含视频转存状态）
7. 显示/下载结果（从 MinIO 下载，解决 VModel CDN 国内访问问题）

**关键功能**：
- 上传进度条：使用 `a-progress` 组件显示上传百分比
- 视频转存：后端自动将 VModel 结果转存到 MinIO
- 转存进度状态：`transferring` 状态时显示"正在转存视频到服务器..."
- 轮询优化：`transferring` 状态时使用 2000ms 间隔，其他状态 3000ms
- 转存失败处理：`transfer_status=failed` 时禁用下载按钮并显示警告
- 类型复用：使用 `TaskStatusResponse['data']` 类型确保类型安全

### PromptsView.vue

Prompt 管理页面，包含两个 Tab：「Prompt 管理」和「使用说明」。

**计划功能**（待开发）：
- Prompt 模板列表
- 创建/编辑 Prompt
- 变量占位符支持
- 分类管理

使用说明从 `docs/prompts.md` 加载。

## UI 组件库

使用 Ant Design Vue 4，已全局注册。

常用组件：
- `a-button`, `a-input`, `a-form`
- `a-table`, `a-modal`, `a-message`
- `a-upload`, `a-spin`, `a-progress`

## 依赖

主要依赖（`package.json`）：
- `vue@^3.4.21` - Vue 3
- `vue-router@^4.3.0` - 路由
- `pinia@^2.1.7` - 状态管理
- `ant-design-vue@^4.1.2` - UI 组件库
- `axios@^1.6.7` - HTTP 客户端
- `marked@^15.0.0` - Markdown 解析
- `dompurify@^3.3.1` - XSS 防护

## 开发注意事项

1. **TypeScript**: 所有新代码使用 TypeScript
2. **组件命名**: 使用 PascalCase，如 `FaceSwapView.vue`
3. **API 调用**: 统一使用 `src/api/index.ts` 中的方法
4. **状态管理**: 全局状态使用 Pinia，组件内状态使用 `ref`/`reactive`
5. **样式**: 优先使用 Ant Design Vue 组件，自定义样式使用 scoped CSS
