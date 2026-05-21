#!/bin/bash

# 日志测试脚本 - 演示所有日志类型（需要 JWT 认证）
# 用法: chmod +x test_logging.sh && ./test_logging.sh

set -e

BASE_URL="${1:-http://localhost:8080}"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ── 获取认证 Token ──
get_token() {
    local username="logtestuser_$(date +%s)"
    local password="logtest123"

    local resp=$(curl -s -X POST "$BASE_URL/api/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    local token=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

    if [ -z "$token" ]; then
        resp=$(curl -s -X POST "$BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"username\":\"$username\",\"password\":\"$password\"}")
        token=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
    fi

    if [ -z "$token" ]; then
        resp=$(curl -s -X POST "$BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"testuser","password":"test123456"}')
        token=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
    fi

    echo "$token"
}

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Enterprise 日志系统测试脚本               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo "⚠️  请确保服务已启动: go run main.go"
echo ""

# 获取 Token
echo -e "${YELLOW}获取认证 Token...${NC}"
TOKEN=$(get_token)
if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ 无法获取 Token${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Token 获取成功${NC}"
echo ""

# 1. 健康检查 - 测试健康检查日志
echo -e "${YELLOW}1️⃣  测试健康检查${NC}"
echo "发送请求: GET /health"
curl -s "$BASE_URL/health" | jq .
echo ""
sleep 1

# 2. 获取单个任务 - 测试缓存 MISS + DB QUERY
echo -e "${YELLOW}2️⃣  获取单个任务（首次 - 缓存 MISS）${NC}"
echo "发送请求: GET /api/tasks/1"
curl -s "$BASE_URL/api/tasks/1" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 3. 再次获取单个任务 - 测试缓存 HIT
echo -e "${YELLOW}3️⃣  获取单个任务（第二次 - 缓存 HIT）${NC}"
echo "发送请求: GET /api/tasks/1"
curl -s "$BASE_URL/api/tasks/1" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 4. 获取单个任务 - 测试 L1+L2 缓存
echo -e "${YELLOW}4️⃣  获取单个任务 ID=1（首次 - 缓存 MISS）${NC}"
echo "发送请求: GET /api/tasks/1"
curl -s "$BASE_URL/api/tasks/1" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 5. 再次获取单个任务 - 测试缓存 HIT
echo -e "${YELLOW}5️⃣  获取单个任务 ID=1（第二次 - 缓存 HIT）${NC}"
echo "发送请求: GET /api/tasks/1"
curl -s "$BASE_URL/api/tasks/1" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 6. 创建任务 - 测试 DB CREATE + CACHE INVALIDATE
echo -e "${YELLOW}6️⃣  创建新任务（测试 DB CREATE + 缓存失效）${NC}"
echo "发送请求: POST /api/tasks"
TASK_ID=$(curl -s -X POST "$BASE_URL/api/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试日志任务","done":false}' | jq -r '.data.id')
echo "创建的任务 ID: $TASK_ID"
echo ""
sleep 1

# 7. 获取任务列表 - 测试新列表缓存
echo -e "${YELLOW}7️⃣  获取任务列表（创建后 - 新缓存）${NC}"
echo "发送请求: GET /api/tasks"
curl -s "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 8. 更新任务 - 测试 DB UPDATE + CACHE INVALIDATE
echo -e "${YELLOW}8️⃣  更新任务 ID=$TASK_ID（测试 DB UPDATE + 缓存失效）${NC}"
echo "发送请求: PUT /api/tasks/$TASK_ID"
curl -s -X PUT "$BASE_URL/api/tasks/$TASK_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"更新后的日志测试任务","done":true}' | jq .
echo ""
sleep 1

# 9. 清除指定任务缓存 - 测试 CACHE INVALIDATE
echo -e "${YELLOW}9️⃣  清除任务缓存 ID=$TASK_ID${NC}"
echo "发送请求: DELETE /api/cache/$TASK_ID"
curl -s -X DELETE "$BASE_URL/api/cache/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 10. 清除所有缓存 - 测试 CACHE INVALIDATE
echo -e "${YELLOW}🔟 清除所有缓存${NC}"
echo "发送请求: DELETE /api/cache"
curl -s -X DELETE "$BASE_URL/api/cache" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 11. 删除任务 - 测试 DB DELETE + CACHE INVALIDATE
echo -e "${YELLOW}1️⃣1️⃣  删除任务 ID=$TASK_ID（测试 DB DELETE + 缓存失效）${NC}"
echo "发送请求: DELETE /api/tasks/$TASK_ID"
curl -s -X DELETE "$BASE_URL/api/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 12. 测试错误场景 - 获取不存在的任务
echo -e "${YELLOW}1️⃣2️⃣  错误测试：获取不存在的任务 ID=999${NC}"
echo "发送请求: GET /api/tasks/999"
curl -s "$BASE_URL/api/tasks/999" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 13. 测试错误场景 - 无效的任务 ID
echo -e "${YELLOW}1️⃣3️⃣  错误测试：无效的任务 ID${NC}"
echo "发送请求: GET /api/tasks/abc"
curl -s "$BASE_URL/api/tasks/abc" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo ""
sleep 1

# 14. 系统信息 - 测试系统端点
echo -e "${YELLOW}1️⃣4️⃣  系统信息${NC}"
echo "发送请求: GET /sys/info"
curl -s "$BASE_URL/sys/info" | jq .
echo ""

# 15. 系统统计
echo -e "${YELLOW}1️⃣5️⃣  系统统计${NC}"
echo "发送请求: GET /sys/stats"
curl -s "$BASE_URL/sys/stats" | jq .
echo ""

echo -e "${GREEN}✅ 测试完成！${NC}"
echo ""
echo "查看服务日志输出，应该能看到以下日志类型:"
echo "  - REQUEST: 请求进入日志"
echo "  - Handler: 处理器执行日志"
echo "  - CACHE HIT/MISS: 缓存命中/未命中"
echo "  - CACHE SET/DELETE: 缓存写入/删除"
echo "  - DB QUERY/SUCCESS: 数据库查询成功"
echo "  - DB CREATE/UPDATE/DELETE: 数据库修改操作"
echo "  - ERROR: 错误日志"
echo "  - RESPONSE: 响应完成日志（含执行时长）"
