#!/bin/bash
# ============================================================
# 用户注册/登录/退出脚本
# 注册新用户并获取 JWT Token，或使用已有账号登录
# 登录后 Token 自动保存到 /tmp/enterprise_token.txt
# 退出登录时自动读取该文件，无需手动输入 Token
# 用法: bash scripts/register.sh [base_url]
# 默认: http://localhost:8080
# ============================================================

BASE_URL="${1:-http://localhost:8080}"
TOKEN_FILE="/tmp/enterprise_token.txt"

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
echo -e "${BLUE}║${NC}        用户注册 / 登录 / 退出脚本        ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# ── 检查是否有已保存的 Token ──
SAVED_TOKEN=""
if [ -f "$TOKEN_FILE" ]; then
    SAVED_TOKEN=$(cat "$TOKEN_FILE")
    SAVED_USER=$(echo "$SAVED_TOKEN" | cut -d'|' -f1)
    SAVED_TOKEN_VAL=$(echo "$SAVED_TOKEN" | cut -d'|' -f2)
    echo -e "  ${DIM}检测到已保存的 Token（用户: $SAVED_USER）${NC}"
    echo ""
fi

# ── 选择操作 ──
echo -e "  ${YELLOW}请选择操作:${NC}"
echo -e "  ${BOLD}1${NC}) 注册新用户"
echo -e "  ${BOLD}2${NC}) 登录已有账号"
echo -e "  ${BOLD}3${NC}) 注册并登录（注册后自动登录）"
echo -e "  ${BOLD}4${NC}) 退出登录（使当前 Token 失效）"
if [ -n "$SAVED_TOKEN_VAL" ]; then
    echo -e "  ${BOLD}5${NC}) 查看已保存的 Token"
fi
echo ""
read -p "  请输入选项 (1/2/3/4${SAVED_TOKEN_VAL:+5}): " ACTION

# ── 保存 Token 到文件 ──
save_token() {
    local user="$1"
    local token="$2"
    echo "${user}|${token}" > "$TOKEN_FILE"
    chmod 600 "$TOKEN_FILE"
    echo -e "  ${DIM}Token 已保存到 $TOKEN_FILE${NC}"
}

# ── 退出登录 ──
if [ "$ACTION" = "4" ]; then
    if [ -z "$SAVED_TOKEN_VAL" ]; then
        echo ""
        echo -e "  ${GREEN}✔ 当前未登录，无需退出${NC}"
        echo ""
        exit 0
    fi

    TOKEN="$SAVED_TOKEN_VAL"
    echo -e "  ${DIM}使用已保存的 Token（用户: $SAVED_USER）${NC}"

    echo ""
    echo -e "  ${YELLOW}▸ 退出登录${NC}"
    echo -e "  ${DIM}\$ curl -s -X POST $BASE_URL/api/auth/logout \\${NC}"
    echo -e "  ${DIM}    -H \"Authorization: Bearer \$TOKEN\"${NC}"

    RESP=$(curl -s -X POST "$BASE_URL/api/auth/logout" \
        -H "Authorization: Bearer $TOKEN")

    echo ""
    echo "$RESP" | python3 -m json.tool 2>/dev/null | sed 's/^/  /'

    CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code', 0))" 2>/dev/null)
    if [ "$CODE" = "200" ]; then
        echo ""
        echo -e "  ${GREEN}✔ 退出登录成功! Token 已失效${NC}"

        # 验证 token 是否真的失效
        echo ""
        echo -e "  ${YELLOW}▸ 验证 Token 是否已失效...${NC}"
        CHECK_RESP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/profile" \
            -H "Authorization: Bearer $TOKEN")
        if [ "$CHECK_RESP" = "401" ]; then
            echo -e "  ${GREEN}✔ Token 已被拒绝 (HTTP 401)，退出登录生效${NC}"
        else
            echo -e "  ${YELLOW}⚠ Token 仍有效 (HTTP $CHECK_RESP)${NC}"
        fi

        # 清除已保存的 Token
        rm -f "$TOKEN_FILE"
        echo -e "  ${DIM}已清除保存的 Token 文件${NC}"
    else
        MSG=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('message', '未知错误'))" 2>/dev/null)
        echo ""
        echo -e "  ${RED}✘ 退出登录失败: $MSG${NC}"
    fi

    echo ""
    exit 0
fi

# ── 查看已保存的 Token ──
if [ "$ACTION" = "5" ] && [ -n "$SAVED_TOKEN_VAL" ]; then
    echo ""
    echo -e "  ${YELLOW}已保存的 Token 信息:${NC}"
    echo -e "  用户: ${BOLD}$SAVED_USER${NC}"
    echo -e "  Token: ${DIM}${SAVED_TOKEN_VAL:0:50}...${NC}"
    echo ""
    echo -e "  ${DIM}Token 文件路径: $TOKEN_FILE${NC}"
    echo ""
    exit 0
fi

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
        save_token "$USERNAME" "$TOKEN"
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
        save_token "$USERNAME" "$TOKEN"
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
