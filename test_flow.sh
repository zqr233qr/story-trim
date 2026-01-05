#!/bin/bash

# --- StoryTrim 3.0 接口测试流程 (Enhanced) ---

SERVER_URL="http://localhost:8080/api"

echo "--- 1. 上传测试书籍 ---"
if [ ! -f "dpcq.txt" ]; then
    echo "错误: 当前目录下找不到 dpcq.txt"
    exit 1
fi

UPLOAD_RES=$(curl -s -F "file=@dpcq.txt" $SERVER_URL/upload)
echo "$UPLOAD_RES" | jq .

BOOK_ID=$(echo "$UPLOAD_RES" | jq -r '.data.book_id // empty')
CHAPTER_ID=$(echo "$UPLOAD_RES" | jq -r '.data.chapters[0].id // empty')

if [ -z "$BOOK_ID" ] || [ "$BOOK_ID" == "null" ]; then
    echo "错误: 无法从响应中获取 Book ID"
    exit 1
fi

if [ -z "$CHAPTER_ID" ] || [ "$CHAPTER_ID" == "null" ]; then
    echo "错误: 无法从响应中获取 Chapter ID。请检查后端是否正确返回了章节列表。"
    exit 1
fi

echo -e "\nBook ID: $BOOK_ID"
echo "First Chapter ID: $CHAPTER_ID"

echo -e "\n--- 2. 获取章节详情 (验证内容展开) ---"
curl -s "$SERVER_URL/chapters/$CHAPTER_ID" | jq .

echo -e "\n--- 3. 执行流式精简 (使用标准模式 ID: 2) ---"
echo "等待 AI 响应..."
curl -N -X POST "$SERVER_URL/trim/stream" \
     -H "Content-Type: application/json" \
     -d "{
       \"chapter_id\": $CHAPTER_ID,
       \"prompt_id\": 2,
       \"prompt_version\": \"v1.0\"
     }"

echo -e "\n\n--- 4. 再次执行流式精简 (验证缓存命中) ---"
echo "这一次应该瞬间返回..."
time curl -N -X POST "$SERVER_URL/trim/stream" \
     -H "Content-Type: application/json" \
     -d "{
       \"chapter_id\": $CHAPTER_ID,
       \"prompt_id\": 2,
       \"prompt_version\": \"v1.0\"
     }"

echo -e "\n--- 5. 获取章节详情 (验证摘要已生成) ---"
sleep 2
curl -s "$SERVER_URL/chapters/$CHAPTER_ID" | jq .data.summary