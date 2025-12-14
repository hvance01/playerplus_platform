#!/bin/bash

# API Token (新的)
API_TOKEN="QSPuI_VGr4YH79s_tAN8vSkSMunQWp_KwRSOjMUnN1qr25q8P6UKQuAefn2NTJwVL550ZT_Kc8k-jjlihigaRg=="
BASE_URL="https://api.vmodel.ai"

echo "=========================================="
echo "VModel API 单元测试 - 使用官方示例数据"
echo "=========================================="

# Test 1: Create Task - 使用VModel官方示例数据
echo ""
echo "=== Test 1: Create Task (Video Face Swap) ==="
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/tasks/v1/create" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "85e248d268bcc04f5302cf9645663c2c12acd03c953ec1a4bbfdc252a65bddc0",
    "input": {
        "source": "https://data.vmodel.ai/data/model-example/vmodel/video-face-swap/source.jpg",
        "target": "https://data.vmodel.ai/data/model-example/vmodel/video-face-swap/target.mp4",
        "keep_fps": false
    }
  }')

echo "Response:"
echo "$RESPONSE" | jq .

# 检查是否成功
CODE=$(echo "$RESPONSE" | jq -r '.code // empty')
TASK_ID=$(echo "$RESPONSE" | jq -r '.result.task_id // empty')

if [ "$CODE" = "200" ] && [ -n "$TASK_ID" ]; then
    echo "✅ Create Task 成功! Task ID: $TASK_ID"

    # Test 2: Get Task
    echo ""
    echo "=== Test 2: Get Task Status ==="
    for i in {1..15}; do
        sleep 3
        STATUS_RESPONSE=$(curl -s "${BASE_URL}/api/tasks/v1/get/${TASK_ID}" \
          -H "Authorization: Bearer ${API_TOKEN}")

        STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.result.status // empty')
        echo "[$i] Status: $STATUS"

        if [ "$STATUS" = "succeeded" ]; then
            echo ""
            echo "✅ 任务完成!"
            echo "$STATUS_RESPONSE" | jq .
            break
        elif [ "$STATUS" = "failed" ]; then
            echo ""
            echo "❌ 任务失败!"
            echo "$STATUS_RESPONSE" | jq .
            break
        fi
    done
else
    echo "❌ Create Task 失败"
    echo "Code: $CODE"
    echo "Message: $(echo "$RESPONSE" | jq -r '.message // empty')"
fi
