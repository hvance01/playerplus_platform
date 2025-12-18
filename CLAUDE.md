# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**PlayerPlus Platform** - 内部 AI 工具平台，提供视频换脸、Prompt 管理、AI 文案生成等能力。

### 模块文档

- **后端开发**: [backend/CLAUDE.md](backend/CLAUDE.md)
- **前端开发**: [frontend/CLAUDE.md](frontend/CLAUDE.md)

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22+ / Gin / PostgreSQL |
| Frontend | Vue 3 / TypeScript / Ant Design Vue |
| Storage | Cloudflare R2 (S3 兼容) |
| CDN | VPS (LA, CN2) + Nginx 反向代理 |
| Deployment | Railway (单二进制) |

## Project Structure

```
playerplus_platform/
├── backend/                 # Go 后端 → 详见 backend/CLAUDE.md
│   ├── cmd/server/          # 入口
│   └── internal/            # 业务代码
├── frontend/                # Vue 前端 → 详见 frontend/CLAUDE.md
│   └── src/                 # 源码
├── Makefile                 # 统一构建命令
├── railway.json             # Railway 部署配置
└── README.md                # 项目说明
```

## External APIs

| Service | Description | Status |
|---------|-------------|--------|
| VModel API | 视频人脸检测与换脸 | ✅ 已集成 |
| Cloudflare R2 | 对象存储 | ✅ 已集成 |
| VPS CDN | 中国访问加速 (CN2) | ✅ 已集成 |
| Resend | 邮件发送服务 | 🚧 待配置 |
| LLM API | 文案生成 | 📋 待定 |

### VModel API

- 人脸检测: `POST /api/predictions` (video-face-detect)
- 换脸执行: `POST /api/predictions` (video-multi-face-swap)
- 结果查询: `GET /api/predictions/:id`

## 部署架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           中国用户访问流程                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   中国用户 ──(CN2优化)──> VPS (LA) ──> Railway 后端                      │
│                            │                                            │
│                            └──> Cloudflare R2 (媒体文件)                 │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│                           VModel API 访问流程                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   VModel API ──(直连)──> Cloudflare R2 (绕过CDN，避免超时)               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 域名配置

| 域名 | 用途 | 指向 |
|------|------|------|
| `platform.playerplus.cn` | 平台主域名 | VPS (31.40.214.114) |
| `cdn.playerplus.cn` | 媒体 CDN | VPS (31.40.214.114) |

### VPS 配置 (Hostdare LA CN2)

- **IP**: 31.40.214.114
- **Nginx 配置**: `/etc/nginx/conf.d/`
  - `platform-proxy.conf` - 平台全站反向代理
  - `r2-proxy.conf` - 媒体 CDN 反向代理
- **SSL**: Let's Encrypt 自动续期

## Railway Services

项目部署在 Railway `profound-wisdom` 项目中：

| Service | Description | Endpoint |
|---------|-------------|----------|
| playerplus-backend | Go 后端 + Vue 前端 | `platform.playerplus.cn` |
| PostgreSQL | 主数据库 | `nozomi.proxy.rlwy.net:28246/railway` |

### 存储配置

| 配置项 | 值 | 说明 |
|--------|-----|------|
| `STORAGE_PUBLIC_URL` | `https://cdn.playerplus.cn` | CDN URL (中国用户访问) |
| `STORAGE_DIRECT_URL` | `https://pub-xxx.r2.dev` | R2 直连 URL (VModel API 访问) |
| `MINIO_PUBLIC_ENDPOINT` | `xxx.r2.cloudflarestorage.com` | R2 S3 API 端点 |

## Auth

- **当前**: 固定账号密码登录 (`test` / `test`)
- **待修复**: 邮箱验证码登录，限制 `@playerplus.cn` 域名
- Token 存储在 localStorage

## Development Progress

### ✅ 已完成

- [x] 项目基础架构（Go + Vue + Railway）
- [x] 用户认证（基础登录）
- [x] Cloudflare R2 存储集成
- [x] VModel 换脸 API 集成
  - 人脸检测 API
  - 多人脸选择功能
  - 异步换脸处理
  - 结果轮询和下载
- [x] 本地开发环境配置
- [x] 视频上传进度条显示
- [x] 视频结果转存到 R2（解决 VModel CDN 国内访问问题）
- [x] 存储缓存 TTL 机制（24小时自动清理）
- [x] 转存进度状态（transferring 状态，转存完成后才显示成功）
- [x] 自定义域名配置（platform.playerplus.cn）
- [x] VPS 全站反向代理（CN2 加速中国访问）
- [x] 媒体 CDN 配置（cdn.playerplus.cn）
- [x] VModel API 直连 R2（解决首次检测超时问题）

### 🚧 进行中

- [ ] 修复邮件验证码登录 (Resend API 配置)

### 📋 待开发

- [ ] Prompt 管理功能
- [ ] LLM 文案生成
- [ ] 一键换装 (V2)
- [ ] 批量处理 (V2)
