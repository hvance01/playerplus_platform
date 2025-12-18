# PlayerPlus Platform 部署指南

本文档记录了 PlayerPlus Platform 在 Railway 上的部署经验和常见问题解决方案。

## 部署架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              整体架构                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────────────────┐    │
│   │  中国用户    │─CN2─>│  VPS (LA)   │─────>│  Railway 后端           │    │
│   └─────────────┘      │  Nginx 代理  │      │  (playerplus-backend)   │    │
│                        └──────┬──────┘      └─────────────────────────┘    │
│                               │                                             │
│                               v                                             │
│                        ┌─────────────┐      ┌─────────────────────────┐    │
│                        │ Cloudflare  │<─────│  VModel API             │    │
│                        │ R2 存储     │      │  (直连，绕过CDN)         │    │
│                        └─────────────┘      └─────────────────────────┘    │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                           Railway Project (profound-wisdom)                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐                        ┌─────────────────┐            │
│  │ playerplus-     │                        │    Postgres     │            │
│  │ backend         │                        │                 │            │
│  │ Go + Vue SPA    │                        │ 数据库          │            │
│  └─────────────────┘                        └─────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────┘
```

## VPS 反向代理 (Hostdare LA CN2)

| 配置 | 值 |
|------|-----|
| IP | 31.40.214.114 |
| 系统 | Ubuntu |
| 线路 | CN2 (中国优化) |
| Nginx 配置目录 | `/etc/nginx/conf.d/` |

### Nginx 配置文件

| 文件 | 用途 | 上游 |
|------|------|------|
| `platform-proxy.conf` | 平台全站反向代理 | Railway 后端 |
| `r2-proxy.conf` | 媒体 CDN 反向代理 | Cloudflare R2 |

### 安全配置

- **SSL**: Let's Encrypt 自动续期
- **HSTS**: 启用
- **TLS 验证**: 启用 (proxy_ssl_verify on)

## 服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 平台（通过VPS） | https://platform.playerplus.cn | 中国用户访问 |
| 媒体 CDN | https://cdn.playerplus.cn | 媒体文件加速 |
| Railway 后端 | https://ordumf4h.up.railway.app | 原始后端 |
| R2 直连 | https://pub-xxx.r2.dev | VModel API 访问 |

## 存储配置

### 环境变量

| 变量 | 值 | 说明 |
|------|-----|------|
| `STORAGE_PUBLIC_URL` | `https://cdn.playerplus.cn` | CDN URL (中国用户) |
| `STORAGE_DIRECT_URL` | `https://pub-xxx.r2.dev` | R2 直连 (VModel API) |
| `MINIO_PUBLIC_ENDPOINT` | `xxx.r2.cloudflarestorage.com` | R2 S3 API |
| `BUCKET_NAME` | `playerplus-media` | 存储桶名称 |

### URL 转换逻辑

VModel API 调用时，系统自动将 CDN URL 转换为 R2 直连 URL：

```
CDN URL:    https://cdn.playerplus.cn/playerplus-media/videos/xxx.mp4
Direct URL: https://pub-xxx.r2.dev/videos/xxx.mp4
```

这样可以避免 VModel API 访问 CDN 时的"冷启动"超时问题。

## 自定义域名配置

### 域名信息

| 域名 | 类型 | 指向 |
|------|------|------|
| `platform.playerplus.cn` | A 记录 | 31.40.214.114 (VPS) |
| `cdn.playerplus.cn` | A 记录 | 31.40.214.114 (VPS) |

### 配置步骤

1. **Railway 添加自定义域名**
   ```bash
   railway domain platform.playerplus.cn -s playerplus-backend
   ```

2. **阿里云 DNS 配置**
   - 记录类型: CNAME
   - 主机记录: platform
   - 记录值: ordumf4h.up.railway.app
   - TTL: 600

3. **验证配置**
   ```bash
   dig platform.playerplus.cn CNAME +short
   # 期望: ordumf4h.up.railway.app.
   ```

### 备案说明

- **ICP 备案**: 不需要（服务器在海外）
- **公安备案**: 建议完成（面向国内用户）
- **域名实名认证**: 必须（.cn 域名要求）

## 部署步骤

### 1. 创建服务

```bash
# 链接到 Railway 项目
railway link -p <project-id> -e production

# 创建后端服务
railway add -s playerplus-backend

# 链接到服务
railway link -p <project-id> -e production -s playerplus-backend
```

### 2. 配置环境变量

必须配置以下环境变量：

```bash
railway variables set \
  DATABASE_URL="postgresql://..." \
  VMODEL_API_TOKEN="..." \
  VMODEL_BASE_URL="https://api.vmodel.ai" \
  MINIO_PUBLIC_ENDPOINT="https://bucket-production-acf6.up.railway.app" \
  MINIO_ROOT_USER="..." \
  MINIO_ROOT_PASSWORD="..." \
  BUCKET_NAME="playerplus-media" \
  STORAGE_PUBLIC_URL="https://bucket-production-acf6.up.railway.app"
```

### 3. 部署

```bash
railway up
```

## 遇到的问题和解决方案

### 问题 1: Nixpacks 包名错误

**错误信息**:
```
error: undefined variable 'go_1_22'
error: undefined variable 'pnpm'
```

**原因**: Nixpacks 的包名系统与标准 Nix 包名不同，版本化的包名（如 `go_1_22`、`nodejs_20`）不被支持。

**解决方案**: 改用 Dockerfile 进行构建，而不是 Nixpacks。

```dockerfile
# Multi-stage build
FROM node:20-alpine AS frontend-builder
# ... build frontend

FROM golang:1.22-alpine AS backend-builder
# ... build backend

FROM alpine:latest
# ... final image
```

### 问题 2: 前端静态文件 301 重定向循环

**错误信息**: 访问 `/` 返回 301 重定向，导致无限循环。

**原因**: `c.FileFromFS("index.html", ...)` 在某些情况下会返回 301 重定向。

**解决方案**: 使用 `fs.ReadFile` 直接读取文件内容并返回：

```go
r.NoRoute(func(c *gin.Context) {
    content, err := fs.ReadFile(staticFS, "index.html")
    if err != nil {
        c.String(http.StatusNotFound, "Not Found")
        return
    }
    c.Data(http.StatusOK, "text/html; charset=utf-8", content)
})
```

### 问题 3: 静态资源路径映射错误

**错误信息**: 访问 `/assets/xxx.js` 返回 index.html 而不是 JS 文件。

**原因**: `r.StaticFS("/assets", http.FS(staticFS))` 会将 `/assets/*` 映射到 `staticFS` 的根目录，而不是 `assets` 子目录。

**解决方案**: 使用 `fs.Sub` 获取 assets 子目录：

```go
assetsFS, err := fs.Sub(staticFS, "assets")
if err == nil {
    r.StaticFS("/assets", http.FS(assetsFS))
}
```

### 问题 4: 文件上传 Content-Type 验证失败

**错误信息**: `Invalid file type. Only images and videos are allowed`

**原因**: curl 上传文件时没有正确设置 Content-Type。

**解决方案**: 显式指定 Content-Type：

```bash
curl -F "file=@video.mp4;type=video/mp4" ...
```

### 问题 5: 环境变量未配置导致本地存储

**现象**: 上传文件返回 `/uploads/...` 本地路径而不是 MinIO URL。

**原因**: Railway 服务没有配置 MinIO 环境变量，导致使用本地存储回退。

**解决方案**: 在 Railway 中配置所有必需的环境变量。

## Dockerfile 最佳实践

```dockerfile
# Multi-stage build for PlayerPlus Platform
# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/ ./backend/
COPY --from=frontend-builder /app/frontend/dist ./backend/internal/handler/dist/
WORKDIR /app/backend
RUN go build -o server ./cmd/server

# Stage 3: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/backend/server ./server
EXPOSE 8080
CMD ["./server"]
```

## railway.json 配置

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "healthcheckPath": "/api/health",
    "restartPolicyType": "ON_FAILURE"
  }
}
```

## 验证部署

### 健康检查

```bash
curl https://platform.playerplus.cn/api/health
# 期望: {"status":"ok"}
```

### 登录测试

```bash
curl -X POST https://platform.playerplus.cn/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
# 期望: {"token":"...","user":"test"}
```

### 完整流程测试

1. 登录获取 token
2. 上传视频 (`POST /api/v2/media/upload`)
3. 检测人脸 (`POST /api/v2/face/detect`)
4. 上传替换人脸 (`POST /api/v2/media/upload/face`)
5. 创建换脸任务 (`POST /api/v2/faceswap/create`)
6. 轮询任务状态 (`GET /api/v2/faceswap/task/:id`)

## 监控和日志

### 查看构建日志

```bash
railway logs --build
```

### 查看部署日志

```bash
railway logs --deploy
```

### 查看服务列表

```bash
railway service status
```

## 常见命令

```bash
# 链接项目
railway link -p <project-id> -e production -s playerplus-backend

# 部署
railway up

# 查看变量
railway variables

# 设置变量
railway variables set KEY=VALUE

# 查看日志
railway logs

# 生成域名
railway domain
```

## 更新记录

- **2025-12-18**: 初始部署，解决 Nixpacks、静态文件、环境变量等问题
- **2025-12-18**: 完成完整流程测试验证
