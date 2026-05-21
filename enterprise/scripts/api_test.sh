#!/bin/bash
# ============================================================
# API 接口测试脚本
# 测试所有 REST API 端点，显示完整 curl 命令
# 用法: bash scripts/api_test.sh [base_url]
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
# 1. 健康检查
# ════════════════════════════════════════════════════════════
print_section "1. 健康检查"

echo ""
print_curl "curl -s $BASE_URL/health"
RESP=$(curl -s "$BASE_URL/health")
print_json "$RESP"
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='ok'" 2>/dev/null; then
    print_pass "/health"
else
    print_fail "/health" "期望 status=ok"
fi

echo ""
print_curl "curl -s $BASE_URL/health/liveness"
RESP=$(curl -s "$BASE_URL/health/liveness")
print_json "$RESP"
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='alive'" 2>/dev/null; then
    print_pass "/health/liveness"
else
    print_fail "/health/liveness" "期望 status=alive"
fi

echo ""
print_curl "curl -s $BASE_URL/health/readiness"
RESP=$(curl -s "$BASE_URL/health/readiness")
print_json "$RESP"
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='ready'" 2>/dev/null; then
    print_pass "/health/readiness"
else
    print_fail "/health/readiness" "期望 status=ready"
fi

# ════════════════════════════════════════════════════════════
# 2. 认证测试
# ════════════════════════════════════════════════════════════
print_section "2. 认证测试"

echo ""
echo -e "  ${YELLOW}▸ 用户注册${NC}"
print_curl "curl -s -X POST $BASE_URL/api/auth/register -H \"Content-Type: application/json\" -d '{\"username\":\"testuser\",\"password\":\"test123456\"}'"
RESP=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123456"}')
print_json "$RESP"
TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")
USERNAME=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['user']['username'])" 2>/dev/null || echo "")
if [ -n "$TOKEN" ] && [ "$USERNAME" = "testuser" ]; then
    print_pass "用户注册 (username=$USERNAME)"
else
    print_fail "用户注册" "期望 data.token 和 data.user.username"
fi

echo ""
echo -e "  ${YELLOW}▸ 重复注册${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X POST $BASE_URL/api/auth/register -H \"Content-Type: application/json\" -d '{\"username\":\"testuser\",\"password\":\"test123456\"}'"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123456"}')
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "409" ]; then
    print_pass "重复注册 (status=$HTTP_CODE)"
else
    print_fail "重复注册" "期望 HTTP 409"
fi

echo ""
echo -e "  ${YELLOW}▸ 用户登录${NC}"
print_curl "curl -s -X POST $BASE_URL/api/auth/login -H \"Content-Type: application/json\" -d '{\"username\":\"testuser\",\"password\":\"test123456\"}'"
RESP=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"test123456"}')
print_json "$RESP"
TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")
if [ -n "$TOKEN" ]; then
    print_pass "用户登录 (token=${TOKEN:0:20}...)"
else
    print_fail "用户登录" "期望 data.token"
fi

echo ""
echo -e "  ${YELLOW}▸ 登录失败（错误密码）${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X POST $BASE_URL/api/auth/login -H \"Content-Type: application/json\" -d '{\"username\":\"testuser\",\"password\":\"wrongpassword\"}'"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"wrongpassword"}')
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "401" ]; then
    print_pass "登录失败 (status=$HTTP_CODE)"
else
    print_fail "登录失败" "期望 HTTP 401"
fi

echo ""
echo -e "  ${YELLOW}▸ 获取用户信息（需认证）${NC}"
print_curl "curl -s $BASE_URL/api/profile -H \"Authorization: Bearer $TOKEN\""
RESP=$(curl -s "$BASE_URL/api/profile" \
    -H "Authorization: Bearer $TOKEN")
print_json "$RESP"
PROFILE_USERNAME=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['username'])" 2>/dev/null || echo "")
if [ "$PROFILE_USERNAME" = "testuser" ]; then
    print_pass "获取用户信息 (username=$PROFILE_USERNAME)"
else
    print_fail "获取用户信息" "期望 data.username='testuser'"
fi

echo ""
echo -e "  ${YELLOW}▸ 未认证访问${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" $BASE_URL/api/profile"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/profile")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "401" ]; then
    print_pass "未认证访问 (status=$HTTP_CODE)"
else
    print_fail "未认证访问" "期望 HTTP 401"
fi

echo ""
echo -e "  ${YELLOW}▸ 无效 Token${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" $BASE_URL/api/profile -H \"Authorization: Bearer invalid_token_here\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/profile" \
    -H "Authorization: Bearer invalid_token_here")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "401" ]; then
    print_pass "无效 Token (status=$HTTP_CODE)"
else
    print_fail "无效 Token" "期望 HTTP 401"
fi

# ════════════════════════════════════════════════════════════
# 3. 缓存管理
# ════════════════════════════════════════════════════════════
print_section "3. 缓存管理"

echo ""
echo -e "  ${YELLOW}▸ 清除所有缓存${NC}"

print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X DELETE $BASE_URL/api/cache -H \"Authorization: Bearer $TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/cache" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    print_pass "清除所有缓存 (status=$HTTP_CODE)"
else
    print_fail "清除所有缓存" "期望 HTTP 200"
fi

echo ""
echo -e "  ${YELLOW}▸ 清除任务缓存${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X DELETE $BASE_URL/api/cache/1 -H \"Authorization: Bearer $TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/cache/1" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    print_pass "清除任务缓存 (status=$HTTP_CODE)"
else
    print_fail "清除任务缓存" "期望 HTTP 200"
fi

echo ""
echo -e "  ${YELLOW}▸ 清除缓存无效 ID${NC}"
print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" -X DELETE $BASE_URL/api/cache/abc -H \"Authorization: Bearer $TOKEN\""
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/cache/abc" \
    -H "Authorization: Bearer $TOKEN")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "400" ]; then
    print_pass "清除缓存无效ID (status=$HTTP_CODE)"
else
    print_fail "清除缓存无效ID" "期望 HTTP 400"
fi

# ════════════════════════════════════════════════════════════
# 4. 系统信息
# ════════════════════════════════════════════════════════════
print_section "4. 系统信息"

echo ""
echo -e "  ${YELLOW}▸ 系统信息${NC}"

print_curl "curl -s $BASE_URL/sys/info"
RESP=$(curl -s "$BASE_URL/sys/info")
print_json "$RESP"
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['code']==200" 2>/dev/null; then
    print_pass "/sys/info"
else
    print_fail "/sys/info" "期望 code=200"
fi

echo ""
echo -e "  ${YELLOW}▸ 系统状态${NC}"
print_curl "curl -s $BASE_URL/sys/stats"
RESP=$(curl -s "$BASE_URL/sys/stats")
print_json "$RESP"
if echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['code']==200" 2>/dev/null; then
    print_pass "/sys/stats"
else
    print_fail "/sys/stats" "期望 code=200"
fi

# ════════════════════════════════════════════════════════════
# 5. Swagger 文档
# ════════════════════════════════════════════════════════════
print_section "5. Swagger 文档"

echo ""
echo -e "  ${YELLOW}▸ Swagger UI${NC}"

print_curl_http "curl -s -o /dev/null -w \"%{http_code}\" $BASE_URL/swagger/index.html"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger/index.html")
print_http "$HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    print_pass "Swagger UI (status=$HTTP_CODE)"
else
    print_fail "Swagger UI" "期望 HTTP 200"
fi

# ════════════════════════════════════════════════════════════
# 总结
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${GREEN}┌$(printf '─%.0s' $(seq 1 56))┐${NC}"
echo -e "${GREEN}│${NC}  ${BOLD}测试完成${NC}  "
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
