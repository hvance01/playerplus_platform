#!/bin/bash
# Railway Deploy Webhook Script
# 用于接收 Gitee webhook 并触发 Railway 部署
#
# 部署到 VPS: /opt/railway-webhook/
# 配置 systemd 服务运行

set -e

# === 配置 ===
RAILWAY_API_TOKEN="${RAILWAY_API_TOKEN:-}"  # 从环境变量读取
SERVICE_ID="4a60c1dd-be4b-4624-aecc-7c736a09b8ec"
ENVIRONMENT_ID="a7f4ef40-9acc-4aef-8821-ca5092bbaf03"
WEBHOOK_SECRET="${WEBHOOK_SECRET:-}"  # Gitee webhook 密钥（可选）
LOG_FILE="/var/log/railway-webhook.log"
PORT="${PORT:-9000}"

# === 日志函数 ===
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# === 触发 Railway 部署 ===
trigger_deploy() {
    log "Triggering Railway deployment..."

    # 使用 serviceInstanceRedeploy mutation
    RESPONSE=$(curl -s -X POST \
        -H "Authorization: Bearer $RAILWAY_API_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "query": "mutation { serviceInstanceRedeploy(serviceId: \"'"$SERVICE_ID"'\", environmentId: \"'"$ENVIRONMENT_ID"'\") }"
        }' \
        "https://backboard.railway.com/graphql/v2")

    log "Railway API Response: $RESPONSE"

    # 检查响应
    if echo "$RESPONSE" | grep -q '"errors"'; then
        log "ERROR: Deployment failed"
        return 1
    fi

    log "SUCCESS: Deployment triggered"
    return 0
}

# === 验证 Gitee 签名（可选）===
verify_signature() {
    local payload="$1"
    local signature="$2"

    if [ -z "$WEBHOOK_SECRET" ]; then
        return 0  # 未配置密钥，跳过验证
    fi

    local expected=$(echo -n "$payload" | openssl dgst -sha256 -hmac "$WEBHOOK_SECRET" | sed 's/^.* //')
    if [ "$signature" = "$expected" ]; then
        return 0
    fi
    return 1
}

# === HTTP 响应 ===
send_response() {
    local status="$1"
    local message="$2"
    local body="{\"status\":\"$status\",\"message\":\"$message\"}"

    echo -e "HTTP/1.1 $status\r"
    echo -e "Content-Type: application/json\r"
    echo -e "Content-Length: ${#body}\r"
    echo -e "Connection: close\r"
    echo -e "\r"
    echo "$body"
}

# === 处理请求 ===
handle_request() {
    local method=""
    local path=""
    local content_length=0
    local x_gitee_token=""

    # 读取请求头
    while IFS= read -r line; do
        line=$(echo "$line" | tr -d '\r')
        [ -z "$line" ] && break

        if [[ "$line" =~ ^(GET|POST|PUT|DELETE)\ (.*)\ HTTP ]]; then
            method="${BASH_REMATCH[1]}"
            path="${BASH_REMATCH[2]}"
        elif [[ "$line" =~ ^Content-Length:\ ([0-9]+) ]]; then
            content_length="${BASH_REMATCH[1]}"
        elif [[ "$line" =~ ^X-Gitee-Token:\ (.*) ]]; then
            x_gitee_token="${BASH_REMATCH[1]}"
        fi
    done

    # 读取请求体
    local body=""
    if [ "$content_length" -gt 0 ]; then
        body=$(head -c "$content_length")
    fi

    log "Received $method $path"

    # 路由处理
    case "$path" in
        /health)
            send_response "200 OK" "healthy"
            ;;
        /webhook|/webhook/)
            if [ "$method" != "POST" ]; then
                send_response "405 Method Not Allowed" "Only POST allowed"
                return
            fi

            # 验证签名（如果配置了密钥）
            if [ -n "$WEBHOOK_SECRET" ] && [ -n "$x_gitee_token" ]; then
                if [ "$x_gitee_token" != "$WEBHOOK_SECRET" ]; then
                    log "ERROR: Invalid webhook token"
                    send_response "401 Unauthorized" "Invalid token"
                    return
                fi
            fi

            # 检查是否是 push 事件
            if echo "$body" | grep -q '"hook_name"'; then
                # 这是 Gitee webhook 请求
                local ref=$(echo "$body" | grep -o '"ref":"[^"]*"' | head -1 | cut -d'"' -f4)
                log "Gitee push event: $ref"

                # 只处理 main/master 分支
                if [[ "$ref" == "refs/heads/main" ]] || [[ "$ref" == "refs/heads/master" ]]; then
                    if trigger_deploy; then
                        send_response "200 OK" "Deployment triggered"
                    else
                        send_response "500 Internal Server Error" "Deployment failed"
                    fi
                else
                    log "Skipping non-main branch: $ref"
                    send_response "200 OK" "Skipped non-main branch"
                fi
            else
                # 手动触发
                if trigger_deploy; then
                    send_response "200 OK" "Deployment triggered"
                else
                    send_response "500 Internal Server Error" "Deployment failed"
                fi
            fi
            ;;
        *)
            send_response "404 Not Found" "Not found"
            ;;
    esac
}

# === 主函数 ===
main() {
    if [ -z "$RAILWAY_API_TOKEN" ]; then
        echo "ERROR: RAILWAY_API_TOKEN not set"
        exit 1
    fi

    log "Starting Railway webhook server on port $PORT..."

    # 使用 socat 或 nc 监听
    if command -v socat &> /dev/null; then
        socat TCP-LISTEN:$PORT,reuseaddr,fork EXEC:"$0 --handle"
    elif command -v nc &> /dev/null; then
        while true; do
            nc -l -p $PORT -c "$0 --handle"
        done
    else
        echo "ERROR: Neither socat nor nc found"
        exit 1
    fi
}

# === 入口 ===
if [ "$1" = "--handle" ]; then
    handle_request
else
    main
fi
