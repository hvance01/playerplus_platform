#!/bin/bash
# 一键部署 Railway Webhook 到 VPS
# 使用方法: ./deploy-webhook-to-vps.sh

set -e

# === 配置 ===
VPS_HOST="31.40.214.114"
VPS_USER="root"
REMOTE_DIR="/opt/railway-webhook"

# Railway 配置
RAILWAY_API_TOKEN="87570706-eaaa-4c82-92b5-9ef7735f63b3"
RAILWAY_SERVICE_ID="4a60c1dd-be4b-4624-aecc-7c736a09b8ec"
RAILWAY_ENVIRONMENT_ID="a7f4ef40-9acc-4aef-8821-ca5092bbaf03"

# Gitee Webhook 密钥（建议设置一个随机字符串）
WEBHOOK_SECRET="playerplus-deploy-$(openssl rand -hex 8)"

echo "=== Railway Webhook 部署脚本 ==="
echo "VPS: $VPS_USER@$VPS_HOST"
echo "Webhook Secret: $WEBHOOK_SECRET"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "1. 创建远程目录..."
ssh $VPS_USER@$VPS_HOST "mkdir -p $REMOTE_DIR"

echo "2. 上传 webhook 脚本..."
scp "$SCRIPT_DIR/railway_webhook.py" $VPS_USER@$VPS_HOST:$REMOTE_DIR/

echo "3. 安装 Python 依赖..."
ssh $VPS_USER@$VPS_HOST "pip3 install flask requests 2>/dev/null || pip install flask requests"

echo "4. 创建 systemd 服务..."
ssh $VPS_USER@$VPS_HOST "cat > /etc/systemd/system/railway-webhook.service << 'EOF'
[Unit]
Description=Railway Deploy Webhook Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$REMOTE_DIR
ExecStart=/usr/bin/python3 $REMOTE_DIR/railway_webhook.py
Restart=always
RestartSec=5

Environment=RAILWAY_API_TOKEN=$RAILWAY_API_TOKEN
Environment=RAILWAY_SERVICE_ID=$RAILWAY_SERVICE_ID
Environment=RAILWAY_ENVIRONMENT_ID=$RAILWAY_ENVIRONMENT_ID
Environment=WEBHOOK_SECRET=$WEBHOOK_SECRET
Environment=PORT=9000

[Install]
WantedBy=multi-user.target
EOF"

echo "5. 启动服务..."
ssh $VPS_USER@$VPS_HOST "systemctl daemon-reload && systemctl enable railway-webhook && systemctl restart railway-webhook"

echo "6. 检查服务状态..."
ssh $VPS_USER@$VPS_HOST "systemctl status railway-webhook --no-pager" || true

echo "7. 配置 Nginx..."
ssh $VPS_USER@$VPS_HOST "cat >> /etc/nginx/conf.d/platform-proxy.conf << 'EOF'

# Railway Webhook (added by deploy script)
location /deploy-webhook {
    proxy_pass http://127.0.0.1:9000/webhook;
    proxy_http_version 1.1;
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Gitee-Token \$http_x_gitee_token;
    proxy_connect_timeout 30s;
    proxy_read_timeout 60s;
}

location /deploy-webhook/health {
    proxy_pass http://127.0.0.1:9000/health;
    proxy_http_version 1.1;
}
EOF"

echo "8. 重载 Nginx..."
ssh $VPS_USER@$VPS_HOST "nginx -t && systemctl reload nginx"

echo ""
echo "=== 部署完成 ==="
echo ""
echo "Webhook URL: https://platform.playerplus.cn/deploy-webhook"
echo "Health Check: https://platform.playerplus.cn/deploy-webhook/health"
echo "Webhook Secret: $WEBHOOK_SECRET"
echo ""
echo "=== Gitee Webhook 配置 ==="
echo "1. 访问 https://gitee.com/playerplus/playerplus_platform/hooks"
echo "2. 点击 '添加 WebHook'"
echo "3. URL: https://platform.playerplus.cn/deploy-webhook"
echo "4. 密码: $WEBHOOK_SECRET"
echo "5. 勾选 'Push' 事件"
echo "6. 点击 '添加'"
echo ""
echo "测试命令:"
echo "curl -X POST https://platform.playerplus.cn/deploy-webhook/health"
