# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PlayPlus Platform - 内部 AI 工具平台，提供视频换脸、Prompt 管理、AI 文案生成等能力。

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
- **换脸 API**: DeepSwap (申请中，当前使用 mock)
- **LLM API**: 文案生成 (具体服务商待定)

## Auth

- 邮箱验证码登录，限制 `@playerplus.cn` 域名
- 验证码有效期 10 分钟
- Token 存储在 localStorage
