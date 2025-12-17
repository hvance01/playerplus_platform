# Backend CLAUDE.md

后端开发指南，适用于 `playerplus_platform/backend` 目录。

## 架构概览

采用 Go Clean Architecture，使用 **Gin** Web 框架：

```
┌─────────────────────────────────────────────────────────┐
│                      Handler Layer                       │
│              (HTTP 请求处理、输入验证)                    │
├─────────────────────────────────────────────────────────┤
│                      Service Layer                       │
│         (业务逻辑、VModel/MinIO 集成)                    │
├─────────────────────────────────────────────────────────┤
│                    Repository Layer                      │
│              (数据库访问、PostgreSQL)                    │
├─────────────────────────────────────────────────────────┤
│                      Model Layer                         │
│                    (领域实体定义)                         │
└─────────────────────────────────────────────────────────┘
```

## 目录结构

```
backend/
├── cmd/server/
│   └── main.go              # 入口点，初始化服务
├── internal/
│   ├── config/
│   │   └── config.go        # 配置加载（单例模式）
│   ├── handler/
│   │   ├── api/
│   │   │   ├── auth.go      # 认证处理
│   │   │   ├── face.go      # 人脸检测处理
│   │   │   ├── faceswap.go  # 换脸任务处理
│   │   │   └── media.go     # 媒体上传处理
│   │   ├── router.go        # 路由配置
│   │   └── embed.go         # 静态文件嵌入
│   ├── middleware/
│   │   └── auth.go          # JWT 认证中间件
│   ├── model/               # 数据模型定义
│   ├── repository/
│   │   ├── db.go            # 数据库连接管理
│   │   ├── user.go          # 用户数据访问
│   │   └── faceswap.go      # 换脸任务数据访问
│   └── service/
│       ├── auth.go          # 认证业务逻辑
│       ├── faceswap.go      # 换脸业务逻辑
│       ├── storage.go       # MinIO 存储操作
│       ├── vmodel.go        # VModel API 客户端
│       └── vmodel_test.go   # VModel 单元测试
├── .env                     # 环境变量（不提交）
├── .env.example             # 环境变量模板
├── go.mod
└── go.sum
```

## 开发命令

在 `backend/` 目录下执行：

```bash
# 启动服务 (localhost:8080)
source .env && go run ./cmd/server

# 运行所有测试
source .env && go test ./...

# 运行特定测试
source .env && go test -v -run TestName ./path/to/package

# 代码检查
golangci-lint run

# 构建二进制
go build -o bin/server ./cmd/server
```

## API 路由

定义在 `internal/handler/router.go`：

```
/api
├── /health                     # GET  健康检查
├── /auth
│   ├── /login                  # POST 固定账号登录
│   ├── /send-code              # POST 发送验证码
│   └── /verify                 # POST 验证码验证
├── /faceswap (v1 - 遗留)
│   ├── /upload                 # POST 上传媒体
│   ├── /swap                   # POST 执行换脸
│   └── /tasks/:id              # GET  查询任务状态
└── /v2 (推荐使用)
    ├── /media
    │   ├── /upload             # POST 上传视频/图像
    │   ├── /upload/face        # POST 上传人脸图像
    │   └── /upload/frame       # POST 上传视频帧
    ├── /face
    │   ├── /detect             # POST 从 URL 检测人脸
    │   └── /detect/upload      # POST 从上传文件检测人脸
    └── /faceswap
        ├── /create             # POST 创建换脸任务
        └── /task/:id           # GET  获取任务状态
```

## 环境变量

在 `backend/.env` 中配置：

| 变量 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `PORT` | 否 | 服务端口 | `8080` |
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

## 核心服务

### VModel API (`service/vmodel.go`)

换脸核心服务，包含：
- `DetectFaces()` - 检测视频/图片中的人脸
- `CreateFaceSwap()` - 创建换脸任务
- `GetPredictionStatus()` - 查询任务状态
- `GetCredits()` - 查询 API 余额

### Storage (`service/storage.go`)

MinIO 存储服务：
- `UploadFile()` - 上传文件到 MinIO
- `GetPublicURL()` - 获取文件公网 URL
- `DeleteFile()` - 删除文件
- `TransferFromVModel()` - 异步转存 VModel 结果视频到 MinIO
- `GetTransferStatus()` - 获取转存状态
- `CleanupExpiredCache()` - 清理过期缓存（24小时 TTL）
- `StartCacheCleanupJob()` - 启动定时清理任务

## 测试指南

```bash
# 检查 VModel 余额
source .env && go test -v -run TestVModelCredits ./internal/service/

# 测试人脸检测
source .env && go test -v -run TestVModelDetectFaces ./internal/service/ -timeout 180s

# 完整流程测试
source .env && go test -v -run TestVModelFullFlow ./internal/service/ -timeout 300s
```

## 依赖

主要依赖（`go.mod`）：
- `github.com/gin-gonic/gin` - Web 框架
- `github.com/lib/pq` - PostgreSQL 驱动
- `github.com/minio/minio-go/v7` - MinIO 客户端
- `github.com/resend/resend-go/v2` - 邮件服务
- `github.com/joho/godotenv` - .env 加载
