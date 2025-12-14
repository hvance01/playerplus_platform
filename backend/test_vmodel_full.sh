#!/bin/bash

# API Token
API_TOKEN="QSPuI_VGr4YH79s_tAN8vSkSMunQWp_KwRSOjMUnN1qr25q8P6UKQuAefn2NTJwVL550ZT_Kc8k-jjlihigaRg=="
BASE_URL="https://api.vmodel.ai"

# Version IDs
VIDEO_FACE_DETECT_VERSION="fa9317a2ad086f7633f4f9b38f35c82495b6c5f38fa2afbe32d9d9df8620b389"
VIDEO_MULTI_FACE_SWAP_VERSION="8e960283784c5b58e5f67236757c40bb6796c85e3c733d060342bdf62f9f0c64"

# 测试数据
TEST_VIDEO="https://vmodel.ai/data/model/remaker/video-face-detect/tmp2ukv7myu.mp4"
NEW_FACE_1="https://vmodel.ai/data/model/remaker/video-multiple-face-swap/t2.jpg"
NEW_FACE_2="https://vmodel.ai/data/model/remaker/video-multiple-face-swap/t1.jpg"

echo "=========================================="
echo "VModel API 完整流程测试"
echo "=========================================="

# ========== Test 1: Video Face Detect ==========
echo ""
echo "=== Step 1: Video Face Detect ==="
echo "视频: $TEST_VIDEO"

DETECT_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/tasks/v1/create" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"version\": \"${VIDEO_FACE_DETECT_VERSION}\",
    \"input\": {
        \"source\": \"${TEST_VIDEO}\"
    }
  }")

echo "Create Task Response:"
echo "$DETECT_RESPONSE" | jq .

DETECT_TASK_ID=$(echo "$DETECT_RESPONSE" | jq -r '.result.task_id // empty')
DETECT_CODE=$(echo "$DETECT_RESPONSE" | jq -r '.code // empty')

if [ "$DETECT_CODE" != "200" ] || [ -z "$DETECT_TASK_ID" ]; then
    echo "❌ Face Detect 任务创建失败"
    exit 1
fi

echo "✅ Face Detect 任务创建成功: $DETECT_TASK_ID"

# 等待检测完成
echo ""
echo "等待检测完成..."
for i in {1..30}; do
    sleep 2
    STATUS_RESPONSE=$(curl -s "${BASE_URL}/api/tasks/v1/get/${DETECT_TASK_ID}" \
      -H "Authorization: Bearer ${API_TOKEN}")

    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.result.status // empty')
    echo "[$i] Status: $STATUS"

    if [ "$STATUS" = "succeeded" ]; then
        echo ""
        echo "✅ Face Detect 完成!"

        # 提取 detect_id 和 faces
        OUTPUT=$(echo "$STATUS_RESPONSE" | jq -r '.result.output[0]')
        DETECT_ID=$(echo "$OUTPUT" | jq -r '.id // empty')
        FACES=$(echo "$OUTPUT" | jq '.faces')
        FACE_COUNT=$(echo "$FACES" | jq 'length')

        echo ""
        echo "检测结果:"
        echo "  detect_id: $DETECT_ID"
        echo "  检测到 $FACE_COUNT 张人脸"
        echo "$FACES" | jq .
        break
    elif [ "$STATUS" = "failed" ]; then
        echo "❌ Face Detect 失败!"
        echo "$STATUS_RESPONSE" | jq .
        exit 1
    fi
done

if [ -z "$DETECT_ID" ]; then
    echo "❌ 未能获取 detect_id"
    exit 1
fi

# ========== Test 2: Video Multiple Face Swap ==========
echo ""
echo "=========================================="
echo "=== Step 2: Video Multiple Face Swap ==="

# 构建 face_map - 注意需要正确转义
# face_map 的值本身是一个 JSON 字符串
FACE_MAP='[{\"face_id\":0,\"target\":\"'${NEW_FACE_1}'\"},{\"face_id\":1,\"target\":\"'${NEW_FACE_2}'\"}]'
echo "detect_id: $DETECT_ID"
echo "face_map: $FACE_MAP"

# 使用 heredoc 来避免转义问题
SWAP_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/tasks/v1/create" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
    "version": "${VIDEO_MULTI_FACE_SWAP_VERSION}",
    "input": {
        "detect_id": "${DETECT_ID}",
        "face_map": "[{\"face_id\":0,\"target\":\"${NEW_FACE_1}\"},{\"face_id\":1,\"target\":\"${NEW_FACE_2}\"}]"
    }
}
EOF
)

echo ""
echo "Create Swap Task Response:"
echo "$SWAP_RESPONSE" | jq .

SWAP_TASK_ID=$(echo "$SWAP_RESPONSE" | jq -r '.result.task_id // empty')
SWAP_CODE=$(echo "$SWAP_RESPONSE" | jq -r '.code // empty')

if [ "$SWAP_CODE" != "200" ] || [ -z "$SWAP_TASK_ID" ]; then
    echo "❌ Face Swap 任务创建失败"
    exit 1
fi

echo "✅ Face Swap 任务创建成功: $SWAP_TASK_ID"

# 等待换脸完成
echo ""
echo "等待换脸完成..."
for i in {1..60}; do
    sleep 3
    STATUS_RESPONSE=$(curl -s "${BASE_URL}/api/tasks/v1/get/${SWAP_TASK_ID}" \
      -H "Authorization: Bearer ${API_TOKEN}")

    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.result.status // empty')
    echo "[$i] Status: $STATUS"

    if [ "$STATUS" = "succeeded" ]; then
        echo ""
        echo "✅ Face Swap 完成!"
        echo "$STATUS_RESPONSE" | jq .

        RESULT_URL=$(echo "$STATUS_RESPONSE" | jq -r '.result.output[0] // empty')
        echo ""
        echo "=========================================="
        echo "🎉 换脸结果视频: $RESULT_URL"
        echo "=========================================="
        break
    elif [ "$STATUS" = "failed" ]; then
        echo "❌ Face Swap 失败!"
        echo "$STATUS_RESPONSE" | jq .
        exit 1
    fi
done

echo ""
echo "=========================================="
echo "✅ 完整流程测试通过!"
echo "=========================================="
