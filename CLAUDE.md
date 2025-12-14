# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PlayerPlus Platform - 内部 AI 工具平台，提供视频换脸、Prompt 管理、AI 文案生成等能力。

### Feature Priority

| Priority | Feature | Description |
|----------|---------|-------------|
| P0 | 视频换脸 | 上传视频+人脸照片，自动替换视频中人脸 |
| P0 | Prompt管理 | 创建、编辑、分类管理Prompt模板，支持变量占位符 |
| P1 | LLM文案生成 | 选择Prompt模板，填入变量，调用AI生成商品文案 |
| P2 | 一键换装 | 上传模特图+服装图，AI生成穿搭效果图 (V2) |
| P2 | 批量处理 | 批量提交视频/文案需求 (V2) |

## Tech Stack

- **Backend**: Go 1.22+ / Gin
- **Frontend**: Vue 3 + Ant Design Vue + Vite
- **Database**: PostgreSQL (Railway)
- **Deployment**: Railway (单服务，单二进制)
- **Architecture**: Go embed 前端静态文件，无跨域

## Deployment

单二进制部署到 Railway，Go 服务同时提供 API 和静态文件：

```
https://app.railway.app
  ├── /api/*     → Go API handlers
  └── /*         → Vue SPA (embedded)
```

## Development Commands

### 开发模式（热重载）

```bash
# 终端 1: 后端
cd backend && go run ./cmd/server    # localhost:8080

# 终端 2: 前端
cd frontend && pnpm dev              # localhost:5173 (proxy → 8080)
```

### 生产构建 & 本地测试

```bash
# 一键构建
make build

# 或手动
cd frontend && pnpm build
cd backend && go build -o bin/server ./cmd/server

# 本地运行生产版本
./backend/bin/server                 # localhost:8080
```

### 测试

```bash
# 后端测试
cd backend && go test ./...

# 单个测试
cd backend && go test -v -run TestName ./path/to/package

# 前端测试
cd frontend && pnpm test
```

### Lint

```bash
cd backend && golangci-lint run
cd frontend && pnpm lint
```

## Project Structure

```
playplus_platform/
├── backend/
│   ├── cmd/server/
│   │   └── main.go          # 入口，embed 前端
│   ├── internal/
│   │   ├── handler/         # HTTP handlers
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据库访问
│   │   └── model/           # 领域模型
│   ├── embed.go             # //go:embed frontend/dist
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── views/           # 页面组件
│   │   ├── components/      # 通用组件
│   │   ├── api/             # API client
│   │   ├── stores/          # Pinia stores
│   │   └── router/          # Vue Router
│   ├── dist/                # 构建产物 (git ignored)
│   ├── package.json
│   └── vite.config.ts       # dev proxy 配置
│
├── Makefile                 # 统一构建命令
├── railway.json             # Railway 部署配置
└── CLAUDE.md
```

## External APIs

- **邮件服务**: Resend (3000封/月免费额度)
- **换脸 API**: Akool API (已集成)
  - 人脸检测: `POST /api/open/v4/faceswap/highres/specifyimage`
  - 换脸执行: `POST /api/open/v4/faceswap/highres/async`
  - 结果查询: `GET /api/open/v4/faceswap/highres/info/by_ids`
- **LLM API**: 文案生成 (具体服务商待定)

## Railway Services

项目部署在 Railway `profound-wisdom` 项目中：

| Service | Description | Endpoint |
|---------|-------------|----------|
| PostgreSQL | 主数据库 | `nozomi.proxy.rlwy.net:28246/railway` |
| MinIO | S3兼容对象存储 | `bucket-production-acf6.up.railway.app` |
| MinIO Console | 管理界面 | `console-production-fa67.up.railway.app` |

## Environment Variables

本地开发需要 `backend/.env` 文件（已在 .gitignore 中）：

```bash
# Server
PORT=8080

# Database (Railway Postgres)
DATABASE_URL=postgresql://...

# Akool API (Face Swap)
AKOOL_CLIENT_ID=xxx
AKOOL_API_KEY=xxx
AKOOL_BASE_URL=https://openapi.akool.com
AKOOL_DETECT_URL=https://sg3.akool.com

# MinIO Storage (Railway)
MINIO_PUBLIC_ENDPOINT=https://bucket-production-acf6.up.railway.app
MINIO_ROOT_USER=xxx
MINIO_ROOT_PASSWORD=xxx
BUCKET_NAME=playerplus-media
STORAGE_PUBLIC_URL=https://bucket-production-acf6.up.railway.app

# Resend (Email) - Optional
RESEND_API_KEY=
```

## Auth

- **当前**: 固定账号密码登录 (`test` / `test`)
- **待修复**: 邮箱验证码登录，限制 `@playerplus.cn` 域名
- Token 存储在 localStorage

---

## Development Progress

### ✅ 已完成

- [x] 项目基础架构（Go + Vue + Railway）
- [x] 用户认证（邮箱验证码登录）
- [x] MinIO 存储服务部署和集成
- [x] Akool 换脸 API 集成
  - 人脸检测 API
  - 多人脸选择功能
  - 异步换脸处理
  - 结果轮询和下载
- [x] 本地开发环境配置（.env）

### 🚧 进行中

- [ ] 本地端到端测试（需要安装 Go: `brew install go`）
- [ ] 修复邮件验证码登录 (Resend API 配置)

### 📋 待开发

- [ ] Prompt 管理功能
- [ ] LLM 文案生成
- [ ] 一键换装 (V2)
- [ ] 批量处理 (V2)
