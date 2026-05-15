#!/bin/bash
# ============================================================
# 批量写入测试数据脚本
# 直接通过 MySQL 批量插入大量任务数据
# 用法: bash scripts/seed_data.sh [count]
# 默认: count=1000000
# ============================================================

COUNT="${1:-1000000}"
BATCH_SIZE=1000
TOTAL_INSERTED=0
START_TIME=$(date +%s)

# MySQL 连接参数（Docker 环境，映射到宿主机 3307 端口）
MYSQL_USER="root"
MYSQL_PASSWORD="root123"
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3307"
MYSQL_DB="task_db"
MYSQL_ARGS="-u $MYSQL_USER -p$MYSQL_PASSWORD -h $MYSQL_HOST -P $MYSQL_PORT"

# ── 颜色 ──
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

echo -e "${BLUE}┌────────────────────────────────────────────────────────┐${NC}"
echo -e "${BLUE}│${NC}  ${BOLD}批量写入测试数据${NC}"
echo -e "${BLUE}│${NC}  ────────────────"
echo -e "${BLUE}│${NC}  目标: ${BOLD}$COUNT${NC} 条任务"
echo -e "${BLUE}│${NC}  批次: $BATCH_SIZE 条/INSERT"
echo -e "${BLUE}│${NC}  MySQL: ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"
echo -e "${BLUE}└────────────────────────────────────────────────────────┘${NC}"
echo ""

# 检查 MySQL 是否可连接
if ! mysql $MYSQL_ARGS -e "SELECT 1" $MYSQL_DB > /dev/null 2>&1; then
    echo -e "${RED}✖ 无法连接 MySQL${NC}"
    exit 1
fi

# 检查表是否存在
TABLE_EXISTS=$(mysql $MYSQL_ARGS -N -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$MYSQL_DB' AND table_name='tasks'" 2>/dev/null)
if [ "$TABLE_EXISTS" = "0" ]; then
    echo -e "${YELLOW}⚠ tasks 表不存在，请先启动服务自动建表${NC}"
    exit 1
fi

# 进度显示
show_progress() {
    local current=$1
    local total=$2
    local elapsed=$(($(date +%s) - START_TIME))
    local pct=$((current * 100 / total))
    local bar_len=40
    local filled=$((pct * bar_len / 100))
    local empty=$((bar_len - filled))
    local rate=0
    [ $elapsed -gt 0 ] && rate=$((current / elapsed))
    local eta=0
    [ $rate -gt 0 ] && eta=$(((total - current) / rate))

    printf "\r  ${DIM}[${NC}"
    printf "${GREEN}%${filled}s${NC}" | tr ' ' '█'
    printf "${DIM}%${empty}s${NC}" | tr ' ' '░'
    printf "${DIM}]${NC} %3d%%  %d/%d  %d条/秒  ETA:%ds  " "$pct" "$current" "$total" "$rate" "$eta"
}

echo -e "  ${YELLOW}开始写入数据...${NC}"
echo ""

# 批量插入
for ((i=1; i<=COUNT; i+=BATCH_SIZE)); do
    # 构建批量 INSERT SQL
    SQL="INSERT INTO tasks (title, done, created_at, updated_at) VALUES "
    VALUES=()
    for ((j=0; j<BATCH_SIZE && i+j<=COUNT; j++)); do
        TITLE="测试任务_$((i+j))"
        DONE=$((RANDOM % 2))
        NOW=$(date '+%Y-%m-%d %H:%M:%S')
        VALUES+=("('$TITLE', $DONE, '$NOW', '$NOW')")
    done

    # 用逗号连接所有 VALUES
    SQL+=$(IFS=,; echo "${VALUES[*]}")

    # 执行插入
    if mysql $MYSQL_ARGS -e "$SQL" $MYSQL_DB 2>/dev/null; then
        TOTAL_INSERTED=$((TOTAL_INSERTED + BATCH_SIZE))
    else
        # 如果批量失败，回退到逐条插入
        for ((j=0; j<BATCH_SIZE && i+j<=COUNT; j++)); do
            TITLE="测试任务_$((i+j))"
            DONE=$((RANDOM % 2))
            NOW=$(date '+%Y-%m-%d %H:%M:%S')
            mysql $MYSQL_ARGS -e "INSERT INTO tasks (title, done, created_at, updated_at) VALUES ('$TITLE', $DONE, '$NOW', '$NOW')" $MYSQL_DB 2>/dev/null
            TOTAL_INSERTED=$((TOTAL_INSERTED + 1))
        done
    fi

    show_progress "$TOTAL_INSERTED" "$COUNT"
done

echo ""
echo ""

# 统计
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
RATE=0
[ $ELAPSED -gt 0 ] && RATE=$((TOTAL_INSERTED / ELAPSED))

# 验证总数
ACTUAL_COUNT=$(mysql $MYSQL_ARGS -N -e "SELECT COUNT(*) FROM tasks" $MYSQL_DB 2>/dev/null)

echo -e "${GREEN}┌────────────────────────────────────────────────────────┐${NC}"
echo -e "${GREEN}│${NC}  ${BOLD}写入完成${NC}"
echo -e "${GREEN}│${NC}  ────────"
echo -e "${GREEN}│${NC}  写入: ${BOLD}$TOTAL_INSERTED${NC} 条"
echo -e "${GREEN}│${NC}  库中: ${BOLD}$ACTUAL_COUNT${NC} 条"
echo -e "${GREEN}│${NC}  耗时: ${BOLD}${ELAPSED}s${NC}"
echo -e "${GREEN}│${NC}  速率: ${BOLD}${RATE}条/秒${NC}"
echo -e "${GREEN}└────────────────────────────────────────────────────────┘${NC}"
echo ""
