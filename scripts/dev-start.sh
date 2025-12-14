#!/bin/bash
#
# PlayerPlus Platform - 一键启动开发环境
# Usage: ./scripts/dev-start.sh
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   PlayerPlus Platform - 开发环境启动    ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ===================
# 1. Check Prerequisites
# ===================
echo -e "${YELLOW}[1/5] 检查前置依赖...${NC}"

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 未安装${NC}"
        echo -e "  安装方式: $2"
        return 1
    else
        local version=$($3 2>&1 | head -n1)
        echo -e "${GREEN}✓ $1${NC} - $version"
        return 0
    fi
}

MISSING_DEPS=0

check_command "go" "brew install go" "go version" || MISSING_DEPS=1
check_command "node" "brew install node" "node --version" || MISSING_DEPS=1
check_command "pnpm" "npm install -g pnpm" "pnpm --version" || MISSING_DEPS=1

if [ $MISSING_DEPS -eq 1 ]; then
    echo ""
    echo -e "${RED}请先安装缺失的依赖后重试${NC}"
    exit 1
fi

echo ""

# ===================
# 2. Check Environment File
# ===================
echo -e "${YELLOW}[2/5] 检查环境配置...${NC}"

ENV_FILE="$PROJECT_ROOT/backend/.env"
ENV_EXAMPLE="$PROJECT_ROOT/backend/.env.example"

if [ ! -f "$ENV_FILE" ]; then
    if [ -f "$ENV_EXAMPLE" ]; then
        echo -e "${YELLOW}⚠ .env 文件不存在，从模板创建...${NC}"
        cp "$ENV_EXAMPLE" "$ENV_FILE"
        echo -e "${YELLOW}⚠ 请编辑 backend/.env 填写实际配置后重新运行${NC}"
        echo ""
        echo "必填配置项:"
        echo "  - DATABASE_URL: PostgreSQL 连接字符串"
        echo "  - VMODEL_API_TOKEN: VModel API Token (从 vmodel.ai 获取)"
        echo "  - MINIO_* 相关配置: MinIO 存储服务"
        echo ""
        echo -e "编辑命令: ${BLUE}vim $ENV_FILE${NC}"
        exit 1
    else
        echo -e "${RED}✗ .env.example 模板文件不存在${NC}"
        exit 1
    fi
fi

# Validate required env vars
source "$ENV_FILE"

MISSING_ENV=0
check_env() {
    local var_name=$1
    local var_value="${!var_name}"
    if [ -z "$var_value" ] || [[ "$var_value" == *"your_"* ]] || [[ "$var_value" == *"password"* && ${#var_value} -lt 10 ]]; then
        echo -e "${RED}✗ $var_name 未配置或使用默认值${NC}"
        MISSING_ENV=1
    else
        echo -e "${GREEN}✓ $var_name${NC}"
    fi
}

check_env "DATABASE_URL"
check_env "VMODEL_API_TOKEN"
check_env "MINIO_PUBLIC_ENDPOINT"

if [ $MISSING_ENV -eq 1 ]; then
    echo ""
    echo -e "${YELLOW}⚠ 部分环境变量未配置，服务可能进入 Mock 模式${NC}"
    echo -e "继续启动? (y/N) "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "已取消"
        exit 1
    fi
fi

echo ""

# ===================
# 3. Install Dependencies
# ===================
echo -e "${YELLOW}[3/5] 安装依赖...${NC}"

# Backend
echo -e "${BLUE}  → Go modules...${NC}"
cd "$PROJECT_ROOT/backend"
go mod download

# Frontend
echo -e "${BLUE}  → Node packages...${NC}"
cd "$PROJECT_ROOT/frontend"
if [ ! -d "node_modules" ]; then
    pnpm install
else
    echo "    (node_modules 已存在，跳过)"
fi

echo -e "${GREEN}✓ 依赖安装完成${NC}"
echo ""

# ===================
# 4. Start Services
# ===================
echo -e "${YELLOW}[4/5] 启动服务...${NC}"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}正在停止服务...${NC}"
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    echo -e "${GREEN}服务已停止${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start backend
echo -e "${BLUE}  → 启动后端 (localhost:8080)...${NC}"
cd "$PROJECT_ROOT/backend"
source .env
go run ./cmd/server &
BACKEND_PID=$!

# Wait for backend to start
sleep 2

# Check if backend is running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}✗ 后端启动失败${NC}"
    exit 1
fi

# Start frontend
echo -e "${BLUE}  → 启动前端 (localhost:5173)...${NC}"
cd "$PROJECT_ROOT/frontend"
pnpm dev &
FRONTEND_PID=$!

# Wait for frontend to start
sleep 3

echo ""

# ===================
# 5. Ready
# ===================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   ✓ 开发环境已就绪!                    ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "  前端: ${BLUE}http://localhost:5173${NC}"
echo -e "  后端: ${BLUE}http://localhost:8080${NC}"
echo -e "  API:  ${BLUE}http://localhost:8080/api/health${NC}"
echo ""
echo -e "  登录凭证: ${YELLOW}test / test${NC}"
echo ""
echo -e "  按 ${RED}Ctrl+C${NC} 停止所有服务"
echo ""

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
