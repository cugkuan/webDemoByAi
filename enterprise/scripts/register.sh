#!/bin/bash
# ============================================================
# 用户注册/登录脚本
# 注册新用户并获取 JWT Token，或使用已有账号登录
# 用法: bash scripts/register.sh [base_url]
# 默认: http://localhost:8080
# ============================================================

BASE_URL="${1:-http://localhost:8080}"

# ── 颜色 ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC}        用户注册 / 登录脚本               ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# ── 选择操作 ──
echo -e "  ${YELLOW}请选择操作:${NC}"
echo -e "  ${BOLD}1${NC}) 注册新用户"
echo -e "  ${BOLD}2${NC}) 登录已有账号"
echo -e "  ${BOLD}3${NC}) 注册并登录（注册后自动登录）"
echo ""
read -p "  请输入选项 (1/2/3): " ACTION

# ── 输入用户名密码 ──
echo ""
read -p "  用户名: " USERNAME
read -s -p "  密码: " PASSWORD
echo ""

# 参数校验
if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ]; then
    echo ""
    echo -e "  ${RED}✘ 用户名和密码不能为空${NC}"
    exit 1
fi

if [ ${#PASSWORD} -lt 6 ]; then
    echo ""
    echo -e "  ${RED}✘ 密码长度不能少于6位${NC}"
    exit 1
fi

# ── 注册 ──
do_register() {
    echo ""
    echo -e "  ${YELLOW}▸ 注册用户: $USERNAME${NC}"
    echo -e "  ${DIM}\$ curl -s -X POST $BASE_URL/api/auth/register \\${NC}"
    echo -e "  ${DIM}    -H \"Content-Type: application/json\" \\${NC}"
    echo -e "  ${DIM}    -d '{\"username\":\"$USERNAME\",\"password\":\"******\"}'${NC}"

    RESP=$(curl -s -X POST "$BASE_URL/api/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    echo ""
    echo "$RESP" | python3 -m json.tool 2>/dev/null | sed 's/^/  /'

    # 检查是否成功
    CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code', 0))" 2>/dev/null)
    if [ "$CODE" = "201" ] || [ "$CODE" = "200" ]; then
        TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
        echo ""
        echo -e "  ${GREEN}✔ 注册成功!${NC}"
        echo -e "  Token: ${DIM}${TOKEN:0:50}...${NC}"
        echo "$TOKEN"
        return 0
    else
        MSG=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('message', '未知错误'))" 2>/dev/null)
        echo ""
        echo -e "  ${RED}✘ 注册失败: $MSG${NC}"
        return 1
    fi
}

# ── 登录 ──
do_login() {
    echo ""
    echo -e "  ${YELLOW}▸ 登录用户: $USERNAME${NC}"
    echo -e "  ${DIM}\$ curl -s -X POST $BASE_URL/api/auth/login \\${NC}"
    echo -e "  ${DIM}    -H \"Content-Type: application/json\" \\${NC}"
    echo -e "  ${DIM}    -d '{\"username\":\"$USERNAME\",\"password\":\"******\"}'${NC}"

    RESP=$(curl -s -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    echo ""
    echo "$RESP" | python3 -m json.tool 2>/dev/null | sed 's/^/  /'

    # 检查是否成功
    CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code', 0))" 2>/dev/null)
    if [ "$CODE" = "200" ]; then
        TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
        echo ""
        echo -e "  ${GREEN}✔ 登录成功!${NC}"
        echo -e "  Token: ${DIM}${TOKEN:0:50}...${NC}"
        echo "$TOKEN"
        return 0
    else
        MSG=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('message', '未知错误'))" 2>/dev/null)
        echo ""
        echo -e "  ${RED}✘ 登录失败: $MSG${NC}"
        return 1
    fi
}

# ── 执行 ──
echo ""
case $ACTION in
    1)
        do_register
        ;;
    2)
        do_login
        ;;
    3)
        do_register && do_login
        ;;
    *)
        echo -e "  ${RED}无效选项${NC}"
        exit 1
        ;;
esac
echo ""
