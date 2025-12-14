# PlayerPlus Platform

PlayerPlus 内部 AI 工具平台，提供视频换脸、Prompt 管理、AI 文案生成等能力。

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [开发指南](#开发指南)
- [API 文档](#api-文档)
- [测试指南](#测试指南)
- [部署](#部署)
- [架构说明](#架构说明)
- [常见问题](#常见问题)

## 功能特性

| 优先级 | 功能 | 描述 | 状态 |
|--------|------|------|------|
| P0 | 视频换脸 | 上传视频+人脸照片，自动替换视频中人脸 | ✅ 已完成 |
| P0 | Prompt管理 | 创建、编辑、分类管理Prompt模板 | 🚧 待开发 |
| P1 | LLM文案生成 | 选择Prompt模板，填入变量，调用AI生成商品文案 | 🚧 待开发 |
| P2 | 一键换装 | 上传模特图+服装图，AI生成穿搭效果图 | 📋 计划中 |
| P2 | 批量处理 | 批量提交视频/文案需求 | 📋 计划中 |

## 技术栈

### 前端
- **框架**: Vue 3 + TypeScript
- **UI 组件**: Ant Design Vue 4
- **状态管理**: Pinia
- **构建工具**: Vite 5
- **包管理**: pnpm

### 后端
- **语言**: Go 1.22+
- **Web 框架**: Gin
- **数据库**: PostgreSQL
- **对象存储**: MinIO (S3 兼容)
- **换脸 API**: VModel API

### 部署
- **平台**: Railway
- **架构**: 单二进制部署，Go embed 前端静态文件

## 项目结构

```
playplus_platform/
├── backend/                    # Go 后端
│   ├── cmd/server/            # 服务入口
│   │   └── main.go
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── handler/           # HTTP handlers
│   │   │   ├── api/           # API 处理器
│   │   │   ├── router.go      # 路由配置
│   │   │   └── embed.go       # 静态文件嵌入
│   │   ├── middleware/        # 中间件
│   │   ├── model/             # 数据模型
│   │   ├── repository/        # 数据库访问层
│   │   └── service/           # 业务逻辑层
│   │       ├── vmodel.go      # VModel API 客户端
│   │       └── storage.go     # 存储服务
│   ├── .env                   # 环境变量 (不提交)
│   ├── .env.example           # 环境变量模板
│   └── go.mod
│
├── frontend/                   # Vue 前端
│   ├── src/
│   │   ├── views/             # 页面组件
│   │   ├── components/        # 通用组件
│   │   ├── api/               # API 客户端
│   │   ├── stores/            # Pinia stores
│   │   └── router/            # Vue Router
│   ├── dist/                  # 构建产物 (git ignored)
│   ├── package.json
│   └── vite.config.ts         # Vite 配置 (含代理)
│
├── Makefile                    # 统一构建命令
├── docker-compose.yml          # 本地数据库
├── railway.json                # Railway 部署配置
└── README.md                   # 本文件
```

## 快速开始

### 一键启动（推荐）

```bash
# 克隆代码
git clone <repository-url>
cd playerplus_platform

# 配置环境变量
cp backend/.env.example backend/.env
vim backend/.env  # 填写实际配置

# 一键启动
./scripts/dev-start.sh
```

脚本会自动：
1. ✅ 检查 Go、Node.js、pnpm 是否安装
2. ✅ 验证环境变量配置
3. ✅ 安装项目依赖
4. ✅ 同时启动前后端服务
5. ✅ 按 `Ctrl+C` 一键停止所有服务

---

### 手动启动

#### 前置要求

- **Go** 1.22+: `brew install go`
- **Node.js** 18+: `brew install node`
- **pnpm**: `npm install -g pnpm`
- **jq** (可选，用于测试脚本): `brew install jq`

#### 步骤 1: 克隆代码

```bash
git clone <repository-url>
cd playerplus_platform
```

#### 步骤 2: 配置环境变量

```bash
# 复制模板
cp backend/.env.example backend/.env

# 编辑配置 (必填项见下方)
vim backend/.env
```

**必填环境变量：**

```bash
# 数据库
DATABASE_URL=postgresql://user:pass@host:port/db

# VModel API (视频换脸)
VMODEL_API_TOKEN=your_token_here

# MinIO 存储
MINIO_PUBLIC_ENDPOINT=https://your-minio.railway.app
MINIO_ROOT_USER=your_access_key
MINIO_ROOT_PASSWORD=your_secret_key
```

> 💡 **获取 VModel API Token**: 访问 [vmodel.ai](https://vmodel.ai)，注册并在 API 设置页面获取。

#### 步骤 3: 安装依赖

```bash
make deps
# 或手动执行：
# cd backend && go mod tidy
# cd frontend && pnpm install
```

#### 步骤 4: 启动开发服务

**方式一：同时启动前后端（推荐）**

```bash
make dev
```

**方式二：分别启动**

```bash
# 终端 1: 后端 (localhost:8080)
cd backend && source .env && go run ./cmd/server

# 终端 2: 前端 (localhost:5173)
cd frontend && pnpm dev
```

#### 步骤 5: 访问应用

- **前端开发**: http://localhost:5173
- **后端 API**: http://localhost:8080/api
- **健康检查**: http://localhost:8080/api/health

**默认登录凭证：**
- 用户名: `test`
- 密码: `test`

## 开发指南

### 常用命令

```bash
# 开发模式 (热重载)
make dev

# 构建生产版本
make build

# 本地运行生产版本
make run

# 运行测试
make test

# 代码检查
make lint

# 清理构建产物
make clean
```

### 后端开发

```bash
cd backend

# 运行服务
source .env && go run ./cmd/server

# 运行测试
source .env && go test ./...

# 运行特定测试
source .env && go test -v -run TestVModelDetectFaces ./internal/service/ -timeout 180s
```

### 前端开发

```bash
cd frontend

# 开发模式
pnpm dev

# 构建
pnpm build

# 代码检查
pnpm lint

# 测试
pnpm test
```

### 代理配置

前端开发服务器 (Vite) 自动将 `/api/*` 请求代理到后端：

```typescript
// frontend/vite.config.ts
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

## API 文档

### 认证

```bash
# 登录
POST /api/auth/login
Content-Type: application/json

{"username": "test", "password": "test"}

# 返回
{"token": "xxx", "user": "test"}
```

后续请求需要在 Header 中携带 Token：
```
Authorization: Bearer <token>
```

### 视频换脸 API

#### 1. 检测人脸

```bash
POST /api/v2/face/detect
Authorization: Bearer <token>
Content-Type: application/json

{
  "image_url": "https://example.com/video.mp4"
}

# 返回
{
  "code": 0,
  "data": {
    "detect_id": "XdkJepD5XDO",
    "faces": [
      {"face_id": 0, "thumbnail": "https://..."},
      {"face_id": 1, "thumbnail": "https://..."}
    ],
    "frame_image": "https://example.com/video.mp4"
  }
}
```

#### 2. 创建换脸任务

```bash
POST /api/v2/faceswap/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "target_video_url": "https://example.com/video.mp4",
  "detect_id": "XdkJepD5XDO",
  "face_swaps": [
    {
      "face_id": 0,
      "source_image_url": "https://example.com/new_face.jpg"
    }
  ],
  "face_enhance": true
}

# 返回
{
  "code": 0,
  "data": {
    "task_id": "dey23xdo5rc2flz0re",
    "status": "queuing"
  }
}
```

#### 3. 查询任务状态

```bash
GET /api/v2/faceswap/task/:task_id
Authorization: Bearer <token>

# 返回 (完成时)
{
  "code": 0,
  "data": {
    "task_id": "dey23xdo5rc2flz0re",
    "status": "completed",
    "result_url": "https://cdn.vmimgs.com/.../result.mp4"
  }
}
```

### 媒体上传 API

```bash
POST /api/v2/media/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file=@/path/to/video.mp4

# 返回
{
  "url": "https://storage.example.com/video.mp4",
  "key": "videos/abc123.mp4"
}
```

## 测试指南

### 单元测试

```bash
# 后端测试
cd backend
source .env && export VMODEL_API_TOKEN
go test -v ./...

# 前端测试
cd frontend
pnpm test
```

### 端到端测试

#### 完整视频换脸流程测试

```bash
# 1. 确保后端运行中
cd backend && source .env && go run ./cmd/server &

# 2. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test"}' | jq -r '.token')

# 3. 检测人脸 (使用 VModel 示例视频)
DETECT_RESULT=$(curl -s -X POST http://localhost:8080/api/v2/face/detect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"image_url": "https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4"}')

echo "检测结果: $DETECT_RESULT"

DETECT_ID=$(echo $DETECT_RESULT | jq -r '.data.detect_id')
FACE_URL=$(echo $DETECT_RESULT | jq -r '.data.faces[1].thumbnail')

# 4. 创建换脸任务
SWAP_RESULT=$(curl -s -X POST http://localhost:8080/api/v2/faceswap/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"target_video_url\": \"https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4\",
    \"detect_id\": \"$DETECT_ID\",
    \"face_swaps\": [{\"face_id\": 0, \"source_image_url\": \"$FACE_URL\"}],
    \"face_enhance\": true
  }")

echo "换脸任务: $SWAP_RESULT"

TASK_ID=$(echo $SWAP_RESULT | jq -r '.data.task_id')

# 5. 轮询任务状态 (每 10 秒查询一次)
while true; do
  STATUS=$(curl -s -X GET "http://localhost:8080/api/v2/faceswap/task/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN")

  echo "状态: $STATUS"

  TASK_STATUS=$(echo $STATUS | jq -r '.data.status')
  if [ "$TASK_STATUS" = "completed" ]; then
    echo "✅ 换脸完成!"
    echo "视频 URL: $(echo $STATUS | jq -r '.data.result_url')"
    break
  elif [ "$TASK_STATUS" = "failed" ]; then
    echo "❌ 换脸失败"
    break
  fi

  sleep 10
done
```

### VModel API 测试

```bash
cd backend
source .env && export VMODEL_API_TOKEN

# 检查余额
go test -v -run TestVModelCredits ./internal/service/

# 测试人脸检测
go test -v -run TestVModelDetectFaces ./internal/service/ -timeout 180s

# 完整流程测试
go test -v -run TestVModelFullFlow ./internal/service/ -timeout 300s
```

## 部署

### Railway 部署

项目配置为 Railway 单服务部署，Go 服务同时提供 API 和静态文件：

```
https://app.railway.app
├── /api/*     → Go API handlers
└── /*         → Vue SPA (embedded)
```

**部署步骤：**

1. 连接 GitHub 仓库到 Railway
2. 配置环境变量 (见 `.env.example`)
3. Railway 会自动检测并构建

**构建流程：**
1. 构建前端: `cd frontend && pnpm build`
2. 复制到后端: `cp -r frontend/dist/* backend/internal/handler/dist/`
3. 构建后端: `cd backend && go build -o bin/server ./cmd/server`

## 架构说明

### 视频换脸流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                         PlayerPlus Platform                            │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    ▼                              ▼
         ┌─────────────────┐            ┌─────────────────┐
         │    Frontend     │            │     Backend     │
         │   (Vue 3)       │───────────▶│     (Go/Gin)    │
         │   Port: 5173    │    API     │   Port: 8080    │
         └─────────────────┘            └─────────────────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    ▼                          ▼                          ▼
         ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
         │   PostgreSQL    │       │     MinIO       │       │   VModel API    │
         │   (Railway)     │       │   (Railway)     │       │  (External)     │
         │   用户/任务数据  │       │   视频/图片存储  │       │   AI 换脸处理    │
         └─────────────────┘       └─────────────────┘       └─────────────────┘
```

### VModel API 版本

| API | Version ID | 用途 |
|-----|------------|------|
| video-face-detect | `fa9317a2ad08...` | 检测视频中所有人脸 |
| video-multi-face-swap | `8e960283784c...` | 多人脸换脸处理 |

## 常见问题

### Q: 启动后端报错 "VModel API not configured"

确保 `backend/.env` 中配置了 `VMODEL_API_TOKEN`。未配置时会进入 Mock 模式。

### Q: 前端访问 API 返回 404

检查：
1. 后端是否在 8080 端口运行
2. Vite 代理配置是否正确
3. 使用 `pnpm dev` 而不是直接打开 `index.html`

### Q: 人脸检测超时

检测超时设置为 120 秒。建议：
1. 使用较短的视频 (< 30 秒)
2. 确保视频 URL 可公网访问

### Q: 换脸结果是图片而不是视频

确保：
1. 使用 V2 API (`/api/v2/faceswap/create`)
2. `detect_id` 来自人脸检测结果
3. `face_id` 与检测结果中的人脸 ID 匹配

### Q: 如何检查 VModel 余额

```bash
curl -s -X POST 'https://api.vmodel.ai/api/users/v1/account/credits/left' \
  -H "Authorization: Bearer $VMODEL_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Q: 如何只启动后端进行 API 测试

```bash
cd backend
source .env
go run ./cmd/server

# 然后用 curl 或 Postman 测试 http://localhost:8080/api/*
```

## 环境变量参考

| 变量 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `PORT` | 否 | 后端服务端口 | `8080` |
| `DATABASE_URL` | 是 | PostgreSQL 连接串 | - |
| `VMODEL_API_TOKEN` | 是* | VModel API Token | - |
| `VMODEL_BASE_URL` | 否 | VModel API 地址 | `https://api.vmodel.ai` |
| `MINIO_PUBLIC_ENDPOINT` | 是* | MinIO 公网地址 | - |
| `MINIO_ROOT_USER` | 是* | MinIO 访问密钥 | - |
| `MINIO_ROOT_PASSWORD` | 是* | MinIO 密钥 | - |
| `BUCKET_NAME` | 否 | 存储桶名称 | `playerplus-media` |
| `STORAGE_PUBLIC_URL` | 否 | 存储公网 URL | - |
| `RESEND_API_KEY` | 否 | Resend 邮件密钥 | - |

> *注：标记为"是*"的变量，未配置时服务会进入 Mock 模式。

## 已知问题

| 级别 | 问题 | 说明 |
|------|------|------|
| 🔴 高 | 硬编码测试凭证 | `test/test` 仅用于开发环境 |
| 🟡 中 | API 响应格式不一致 | V2 API 使用新格式，旧 API 待统一 |
| 🟡 中 | 缺少依赖注入 | 影响单元测试 |

## License

Internal use only.
