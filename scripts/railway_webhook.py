#!/usr/bin/env python3
"""
Railway Deploy Webhook Server
接收 Gitee webhook 并触发 Railway 部署

部署到 VPS:
1. 复制到 /opt/railway-webhook/
2. 安装依赖: pip3 install flask requests
3. 配置 systemd 服务
4. 配置 Nginx 反向代理
"""

import os
import json
import logging
import requests
from flask import Flask, request, jsonify
from functools import wraps

# === 配置 ===
RAILWAY_API_TOKEN = os.environ.get('RAILWAY_API_TOKEN', '')
SERVICE_ID = os.environ.get('RAILWAY_SERVICE_ID', '4a60c1dd-be4b-4624-aecc-7c736a09b8ec')
ENVIRONMENT_ID = os.environ.get('RAILWAY_ENVIRONMENT_ID', 'a7f4ef40-9acc-4aef-8821-ca5092bbaf03')
WEBHOOK_SECRET = os.environ.get('WEBHOOK_SECRET', '')  # Gitee webhook 密钥
PORT = int(os.environ.get('PORT', 9000))
ALLOWED_BRANCHES = ['refs/heads/main', 'refs/heads/master']

# === 日志配置 ===
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)

def verify_webhook_token(f):
    """验证 Gitee webhook token"""
    @wraps(f)
    def decorated(*args, **kwargs):
        if not WEBHOOK_SECRET:
            return f(*args, **kwargs)

        token = request.headers.get('X-Gitee-Token', '')
        if token != WEBHOOK_SECRET:
            logger.warning(f"Invalid webhook token: {token[:10]}...")
            return jsonify({'error': 'Invalid token'}), 401
        return f(*args, **kwargs)
    return decorated

def trigger_railway_deploy():
    """触发 Railway 部署"""
    if not RAILWAY_API_TOKEN:
        logger.error("RAILWAY_API_TOKEN not configured")
        return False, "API token not configured"

    query = """
    mutation {
        serviceInstanceRedeploy(
            serviceId: "%s",
            environmentId: "%s"
        )
    }
    """ % (SERVICE_ID, ENVIRONMENT_ID)

    try:
        response = requests.post(
            'https://backboard.railway.com/graphql/v2',
            headers={
                'Authorization': f'Bearer {RAILWAY_API_TOKEN}',
                'Content-Type': 'application/json'
            },
            json={'query': query},
            timeout=30
        )

        data = response.json()
        logger.info(f"Railway API response: {json.dumps(data)}")

        if 'errors' in data:
            error_msg = data['errors'][0].get('message', 'Unknown error')
            logger.error(f"Railway API error: {error_msg}")
            return False, error_msg

        logger.info("Deployment triggered successfully")
        return True, "Deployment triggered"

    except requests.exceptions.RequestException as e:
        logger.error(f"Request failed: {e}")
        return False, str(e)

@app.route('/health', methods=['GET'])
def health():
    """健康检查"""
    return jsonify({
        'status': 'healthy',
        'service_id': SERVICE_ID[:8] + '...',
        'token_configured': bool(RAILWAY_API_TOKEN)
    })

@app.route('/webhook', methods=['POST'])
@verify_webhook_token
def webhook():
    """处理 Gitee webhook"""
    try:
        payload = request.get_json(force=True, silent=True) or {}
    except:
        payload = {}

    # 检查是否是 Gitee push 事件
    hook_name = payload.get('hook_name', '')
    ref = payload.get('ref', '')

    if hook_name:
        # Gitee webhook
        logger.info(f"Received Gitee webhook: hook_name={hook_name}, ref={ref}")

        # 只处理 push 事件
        if hook_name != 'push_hooks':
            logger.info(f"Skipping non-push event: {hook_name}")
            return jsonify({'status': 'skipped', 'reason': 'Not a push event'})

        # 只处理 main/master 分支
        if ref not in ALLOWED_BRANCHES:
            logger.info(f"Skipping non-main branch: {ref}")
            return jsonify({'status': 'skipped', 'reason': f'Branch {ref} not in allowed list'})

        # 获取提交信息
        commits = payload.get('commits', [])
        if commits:
            latest_commit = commits[-1]
            logger.info(f"Latest commit: {latest_commit.get('id', '')[:8]} - {latest_commit.get('message', '')[:50]}")
    else:
        # 手动触发
        logger.info("Manual trigger received")

    # 触发部署
    success, message = trigger_railway_deploy()

    if success:
        return jsonify({'status': 'success', 'message': message})
    else:
        return jsonify({'status': 'error', 'message': message}), 500

@app.route('/trigger', methods=['POST', 'GET'])
def manual_trigger():
    """手动触发部署（用于测试）"""
    # 简单的密钥验证
    secret = request.args.get('secret', '') or request.headers.get('X-Secret', '')
    if WEBHOOK_SECRET and secret != WEBHOOK_SECRET:
        return jsonify({'error': 'Invalid secret'}), 401

    success, message = trigger_railway_deploy()

    if success:
        return jsonify({'status': 'success', 'message': message})
    else:
        return jsonify({'status': 'error', 'message': message}), 500

if __name__ == '__main__':
    if not RAILWAY_API_TOKEN:
        logger.warning("WARNING: RAILWAY_API_TOKEN not set!")

    logger.info(f"Starting Railway webhook server on port {PORT}")
    logger.info(f"Service ID: {SERVICE_ID}")
    logger.info(f"Environment ID: {ENVIRONMENT_ID}")
    logger.info(f"Webhook secret configured: {bool(WEBHOOK_SECRET)}")

    app.run(host='0.0.0.0', port=PORT, debug=False)
