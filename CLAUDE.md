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
| Storage | MinIO (S3 兼容) |
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
| MinIO | 对象存储 (Railway) | ✅ 已集成 |
| Resend | 邮件发送服务 | 🚧 待配置 |
| LLM API | 文案生成 | 📋 待定 |

### VModel API

- 人脸检测: `POST /api/predictions` (video-face-detect)
- 换脸执行: `POST /api/predictions` (video-multi-face-swap)
- 结果查询: `GET /api/predictions/:id`

## Railway Services

项目部署在 Railway `profound-wisdom` 项目中：

| Service | Description | Endpoint |
|---------|-------------|----------|
| PostgreSQL | 主数据库 | `nozomi.proxy.rlwy.net:28246/railway` |
| MinIO | S3兼容对象存储 | `bucket-production-acf6.up.railway.app` |
| MinIO Console | 管理界面 | `console-production-fa67.up.railway.app` |

## Auth

- **当前**: 固定账号密码登录 (`test` / `test`)
- **待修复**: 邮箱验证码登录，限制 `@playerplus.cn` 域名
- Token 存储在 localStorage

## Development Progress

### ✅ 已完成

- [x] 项目基础架构（Go + Vue + Railway）
- [x] 用户认证（基础登录）
- [x] MinIO 存储服务部署和集成
- [x] VModel 换脸 API 集成
  - 人脸检测 API
  - 多人脸选择功能
  - 异步换脸处理
  - 结果轮询和下载
- [x] 本地开发环境配置

### 🚧 进行中

- [ ] 修复邮件验证码登录 (Resend API 配置)

### 📋 待开发

- [ ] Prompt 管理功能
- [ ] LLM 文案生成
- [ ] 一键换装 (V2)
- [ ] 批量处理 (V2)
