# PlayerPlus Platform

PlayerPlus 内部 AI 工具平台，提供视频换脸、Prompt 管理、AI 文案生成等能力。

## 功能特性

| 优先级 | 功能 | 状态 |
|--------|------|------|
| P0 | 视频换脸 | ✅ 已完成 |
| P0 | 上传进度条 | ✅ 已完成 |
| P0 | 视频结果转存 | ✅ 已完成 |
| P0 | Prompt管理 | 🚧 待开发 |
| P1 | LLM文案生成 | 🚧 待开发 |
| P2 | 一键换装 / 批量处理 | 📋 计划中 |

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+ & pnpm
- PostgreSQL (或使用 Railway 连接)

### 一键启动（推荐）

```bash
git clone <repository-url>
cd playerplus_platform
./scripts/dev-start.sh
```

### 手动启动

```bash
# 1. 配置环境变量
cp backend/.env.example backend/.env

# 2. 安装依赖
make deps

# 3. 启动服务
make dev
```

**访问地址：**
- 前端: http://localhost:5173
- 后端 API: http://localhost:8080/api
- 默认登录: `test` / `test`

## 常用命令

```bash
make dev      # 启动开发环境
make build    # 构建生产版本
make test     # 运行测试
make lint     # 代码检查
```

## API 文档

### 认证

```bash
POST /api/auth/login
{"username": "test", "password": "test"}
# → {"token": "xxx", "user": "test"}
```

后续请求携带 Header: `Authorization: Bearer <token>`

### 视频换脸

```bash
# 1. 检测人脸
POST /api/v2/face/detect
{"image_url": "https://example.com/video.mp4"}

# 2. 创建换脸任务
POST /api/v2/faceswap/create
{
  "target_video_url": "https://...",
  "detect_id": "XdkJepD5XDO",
  "face_swaps": [{"face_id": 0, "source_image_url": "https://..."}],
  "face_enhance": true
}

# 3. 查询任务状态
GET /api/v2/faceswap/task/:task_id
```

### 媒体上传

```bash
POST /api/v2/media/upload
Content-Type: multipart/form-data
file=@/path/to/video.mp4
```

## 部署

项目采用**单二进制部署**模式，部署到 Railway：

```bash
# 构建
make build

# 本地运行生产版本
./backend/bin/server
```

Railway 自动检测 `railway.json` 配置并构建部署。

## 环境变量

| 变量 | 必需 | 说明 |
|------|------|------|
| `DATABASE_URL` | 是 | PostgreSQL 连接串 |
| `VMODEL_API_TOKEN` | 是* | VModel API Token |
| `MINIO_PUBLIC_ENDPOINT` | 是* | MinIO 公网地址 |
| `MINIO_ROOT_USER` | 是* | MinIO 访问密钥 |
| `MINIO_ROOT_PASSWORD` | 是* | MinIO 密钥 |

> *未配置时进入 Mock 模式

完整环境变量说明见 [backend/CLAUDE.md](backend/CLAUDE.md)。

## 常见问题

**Q: 后端启动报错 "VModel API not configured"**

确保 `backend/.env` 中配置了 `VMODEL_API_TOKEN`。

**Q: 前端访问 API 返回 404**

1. 确保后端在 8080 端口运行
2. 使用 `pnpm dev` 启动前端（Vite 代理生效）

**Q: 人脸检测超时**

- 使用较短的视频 (< 30 秒)
- 确保视频 URL 可公网访问

**Q: 换脸结果下载慢或无法访问**

结果视频会自动从 VModel CDN 转存到 MinIO，前端会等待转存完成后再显示下载链接。

**Q: 如何检查 VModel 余额**

```bash
curl -s -X POST 'https://api.vmodel.ai/api/users/v1/account/credits/left' \
  -H "Authorization: Bearer $VMODEL_API_TOKEN" \
  -H "Content-Type: application/json" -d '{}'
```

## License

Internal use only.
