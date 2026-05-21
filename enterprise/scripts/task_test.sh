#!/bin/bash
# ============================================================
# 任务 (Task) API 接口测试脚本
# 测试所有任务相关的 REST API 端点
# 用法: bash scripts/task_test.sh [base_url]
# 默认: http://localhost:8080
# ============================================================

BASE_URL="${1:-http://localhost:8080}"
PASS=0
FAIL=0

# ── 颜色 ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

print_pass() {
    echo -e "  ${GREEN}✔${NC} ${BOLD}PASS${NC} $1"
    PASS=$((PASS+1))
}

print_fail() {
    echo -e "  ${RED}✘${NC} ${BOLD}FAIL${NC} $1"
    echo -e "  ${DIM}    $2${NC}"
    FAIL=$((FAIL+1))
}

print_section() {
    local title="$1"
    local len=${#title}
    local total=56
    local pad=$(( (total - len) / 2 - 2 ))
    echo ""
    echo -e "${BLUE}┌$(printf '─%.0s' $(seq 1 $total))┐${NC}"
    printf "${BLUE}│${NC}%${pad}s${BOLD}%s${NC}%${pad}s${BLUE}│${NC}\n" "" "$title" ""
    echo -e "${BLUE}└$(printf '─%.0s' $(seq 1 $total))┘${NC}"
}

# 显示 curl 命令（带 | jq，方便复制直接执行）
print_curl() {
    echo -e "  ${DIM}\$ $1 | jq${NC}"
}

# 显示 curl 命令（仅 HTTP 状态码，不带管道）
print_curl_http() {
    echo -e "  ${DIM}\$ $1${NC}"
}

print_json() {
    echo "$1" | python3 -m json.tool 2>/dev/null | sed 's/^/  /'
}

print_http() {
    echo -e "  ${DIM}→ HTTP $1${NC}"
}

# ════════════════════════════════════════════════════════════
# 1. 任务 CRUD
# ════════════════════════════════════════════════════════════
print_section "1. 任务 CRUD"

# 先注册/登录获取 token
echo ""
echo -e "  ${YELLOW}▸ 获取认证 Token${NC}"
RESP=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"tasktestuser","password":"test123456"}' 2>/dev/null)
TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then
    # 尝试登录（如果已注册）
    RESP=$(curl -s -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"tasktestuser","password":"test123456"}' 2>/dev/null)
    TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")
fi

if [ -z "$TOKEN" ]; then
    echo -e "  ${RED}无法获取 Token，请确保服务已启动${NC}"
    exit 1
fi
echo -e "  ${DIM}Token: ${TOKEN:0:20}...${NC}"

echo ""
echo -e "  ${YELLOW}▸ 创建任务${NC}"
print_curl "curl -s -X POST $BASE_URL/api/tasks -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d '{\"title\":\"API测试任务\",\"done\":false}'"
RESP=$(curl -s -X POST "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"API测试任务","done":false}')
print_json "$RESP"
TASK_ID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
if [ -n "$TASK_ID" ] && [ "$TASK_ID" -gt 0 ] 2>/dev/null; then
    print_pass "创建任务 (id=$TASK_ID)"
else
    print_fail "创建任务" "期望 data.id > 0"
fi

echo ""
echo -e "  ${YELLOW}▸ 获取单个任务${NC}"
print_curl "curl -s $BASE_URL/api/tasks/$TASK_ID -H \"Authorization: Bearer \$TOKEN\""
RESP=$(curl -s "$BASE_URL/api/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
print_json "$RESP"
TASK_TITLE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['title'])" 2>/dev/null || echo "")
if [ "$TASK_TITLE" = "API测试任务" ]; then
    print_pass "获取单个任务 (title=$TASK_TITLE)"
else
    print_fail "获取单个任务" "期望 title='API测试任务'"
fi

echo ""
echo -e "  ${YELLOW}▸ 更新任务${NC}"
print_curl "curl -s -X PUT $BASE_URL/api/tasks/$TASK_ID -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d '{\"title\":\"已更新任务\",\"done\":true}'"
RESP=$(curl -s -X PUT "$BASE_URL/api/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"已更新任务","done":true}')
print_json "$RESP"
UPDATED_TITLE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['title'])" 2>/dev/null || echo "")
UPDATED_DONE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['done'])" 2>/dev/null || echo "")
if [ "$UPDATED_TITLE" = "已更新任务" ] && [ "$UPDATED_DONE" = "True" ]; then
    print_pass "更新任务 (title=$UPDATED_TITLE, done=$UPDATED_DONE)"
else
    print_fail "更新任务" "期望 title='已更新任务', done=true"
fi

echo ""
echo -e "  ${YELLOW}▸ 删除任务${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X DELETE $BASE_URL/api/tasks/$TASK_ID -H \"Authorization: Bearer \$TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    print_pass "删除任务 (status=$HTTP_CODE)"
else
    print_fail "删除任务" "期望 HTTP 200"
fi

# ════════════════════════════════════════════════════════════
# 2. 边界情况测试
# ════════════════════════════════════════════════════════════
print_section "2. 边界情况测试"

echo ""
echo -e "  ${YELLOW}▸ 获取不存在的任务${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" $BASE_URL/api/tasks/99999 -H \"Authorization: Bearer \$TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/tasks/99999" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
    print_pass "获取不存在的任务 (status=$HTTP_CODE)"
else
    print_fail "获取不存在的任务" "期望 HTTP 404"
fi

echo ""
echo -e "  ${YELLOW}▸ 更新不存在的任务${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X PUT $BASE_URL/api/tasks/99999 -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d '{\"title\":\"test\",\"done\":false}'"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL/api/tasks/99999" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"test","done":false}')
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
    print_pass "更新不存在的任务 (status=$HTTP_CODE)"
else
    print_fail "更新不存在的任务" "期望 HTTP 404"
fi

echo ""
echo -e "  ${YELLOW}▸ 删除不存在的任务${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X DELETE $BASE_URL/api/tasks/99999 -H \"Authorization: Bearer \$TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/tasks/99999" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
    print_pass "删除不存在的任务 (status=$HTTP_CODE)"
else
    print_fail "删除不存在的任务" "期望 HTTP 404"
fi

echo ""
echo -e "  ${YELLOW}▸ 无效 ID（非数字）${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" $BASE_URL/api/tasks/abc -H \"Authorization: Bearer \$TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/tasks/abc" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "400" ]; then
    print_pass "无效 ID (status=$HTTP_CODE)"
else
    print_fail "无效 ID" "期望 HTTP 400"
fi

echo ""
echo -e "  ${YELLOW}▸ 空标题${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X POST $BASE_URL/api/tasks -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d '{\"title\":\"\",\"done\":false}'"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"","done":false}')
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "400" ]; then
    print_pass "空标题 (status=$HTTP_CODE)"
else
    print_fail "空标题" "期望 HTTP 400"
fi

echo ""
echo -e "  ${YELLOW}▸ 无效 JSON${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X POST $BASE_URL/api/tasks -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d 'not json'"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d 'not json')
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "400" ]; then
    print_pass "无效 JSON (status=$HTTP_CODE)"
else
    print_fail "无效 JSON" "期望 HTTP 400"
fi

# ════════════════════════════════════════════════════════════
# 3. 缓存穿透测试
# ════════════════════════════════════════════════════════════
print_section "3. 缓存穿透测试"

echo ""
echo -e "  ${YELLOW}▸ 创建测试任务${NC}"
RESP=$(curl -s -X POST "$BASE_URL/api/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"缓存测试","done":false}')
CACHE_TASK_ID=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
echo -e "  ${DIM}任务 ID: $CACHE_TASK_ID${NC}"

echo ""
echo -e "  ${YELLOW}▸ 第一次请求（查数据库）${NC}"
print_curl "curl -s $BASE_URL/api/tasks/$CACHE_TASK_ID -H \"Authorization: Bearer \$TOKEN\""
START=$(python3 -c "import time; print(int(time.time()*1000))")
RESP=$(curl -s "$BASE_URL/api/tasks/$CACHE_TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
END=$(python3 -c "import time; print(int(time.time()*1000))")
print_json "$RESP"
echo -e "  ${DIM}⏱ 耗时: $((END-START))ms${NC}"

echo ""
echo -e "  ${YELLOW}▸ 第二次请求（L1 缓存命中）${NC}"
print_curl "curl -s $BASE_URL/api/tasks/$CACHE_TASK_ID -H \"Authorization: Bearer \$TOKEN\""
START=$(python3 -c "import time; print(int(time.time()*1000))")
RESP=$(curl -s "$BASE_URL/api/tasks/$CACHE_TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
END=$(python3 -c "import time; print(int(time.time()*1000))")
print_json "$RESP"
echo -e "  ${DIM}⏱ 耗时: $((END-START))ms${NC}"

echo ""
echo -e "  ${YELLOW}▸ 更新任务（清除缓存）${NC}"
print_curl "curl -s -X PUT $BASE_URL/api/tasks/$CACHE_TASK_ID -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" -d '{\"title\":\"缓存已清除\",\"done\":true}'"
RESP=$(curl -s -X PUT "$BASE_URL/api/tasks/$CACHE_TASK_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"缓存已清除","done":true}')
print_json "$RESP"

echo ""
echo -e "  ${YELLOW}▸ 更新后请求（缓存已清除，重新查数据库）${NC}"
print_curl "curl -s $BASE_URL/api/tasks/$CACHE_TASK_ID -H \"Authorization: Bearer \$TOKEN\""
START=$(python3 -c "import time; print(int(time.time()*1000))")
RESP=$(curl -s "$BASE_URL/api/tasks/$CACHE_TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
END=$(python3 -c "import time; print(int(time.time()*1000))")
print_json "$RESP"
echo -e "  ${DIM}⏱ 耗时: $((END-START))ms${NC}"

print_pass "缓存穿透测试完成"

# ════════════════════════════════════════════════════════════
# 总结
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${GREEN}┌$(printf '─%.0s' $(seq 1 56))┐${NC}"
echo -e "${GREEN}│${NC}  ${BOLD}任务测试完成${NC}  "
echo -e "${GREEN}│${NC}  ────────"
echo -e "${GREEN}│${NC}  总测试: $((PASS+FAIL))"
echo -e "${GREEN}│${NC}  通过:   ${BOLD}${GREEN}$PASS${NC}"
if [ "$FAIL" -gt 0 ]; then
    echo -e "${GREEN}│${NC}  失败:   ${BOLD}${RED}$FAIL${NC}"
    echo -e "${GREEN}└$(printf '─%.0s' $(seq 1 56))┘${NC}"
    exit 1
else
    echo -e "${GREEN}│${NC}  ${BOLD}全部通过! 🎉${NC}"
    echo -e "${GREEN}└$(printf '─%.0s' $(seq 1 56))┘${NC}"
fi
echo ""
