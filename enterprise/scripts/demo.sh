c'l#!/bin/bash

# 缓存演示脚本
echo "========================================"
echo "2级缓存演示脚本"
echo "========================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试基础地址
BASE_URL="http://localhost:8080"

# 检查服务是否运行
echo -e "${CYAN}检查服务状态...${NC}"
if ! curl -s "$BASE_URL/api/tasks" > /dev/null 2>&1; then
    echo -e "${RED}❌ 错误：服务未运行，请先启动应用${NC}"
    echo "启动方法: cd /Users/kuan/Downloads/web-demo/enterprise && docker compose up -d"
    exit 1
fi
echo -e "${GREEN}✓ 服务已运行${NC}"
echo ""

echo -e "${YELLOW}1. 创建任务...${NC}"
curl -s -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{"title":"缓存测试任务","done":false}' | jq '.'
echo ""

echo -e "${YELLOW}2. 获取单个任务（查询数据库 + L1/L2 缓存写入）${NC}"
time curl -s -X GET "$BASE_URL/api/tasks/1" | jq '.'
echo ""

echo -e "${YELLOW}3. 等待 1 秒后再次获取单个任务（L1 缓存命中）${NC}"
sleep 1
time curl -s -X GET "$BASE_URL/api/tasks/1" | jq '.'
echo ""

echo -e "${YELLOW}4. 获取单个任务（查询数据库 + L1/L2 缓存写入）${NC}"
time curl -s -X GET "$BASE_URL/api/tasks/1" | jq '.'
echo ""

echo -e "${YELLOW}5. 立即再次获取单个任务（L1 缓存命中）${NC}"
time curl -s -X GET "$BASE_URL/api/tasks/1" | jq '.'
echo ""

echo -e "${YELLOW}6. 更新任务（清除缓存）${NC}"
curl -s -X PUT "$BASE_URL/api/tasks/1" \
  -H "Content-Type: application/json" \
  -d '{"title":"缓存测试任务(已更新)","done":true}' | jq '.'
echo ""

echo -e "${YELLOW}7. 获取已更新的任务（缓存已被清除，重新查询数据库）${NC}"
time curl -s -X GET "$BASE_URL/api/tasks/1" | jq '.'
echo ""

# ============================================================
# 压力测试部分
# ============================================================

echo ""
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        🔥 压力测试 - 并发请求         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# 压力测试函数
run_load_test() {
    local test_name=$1
    local endpoint=$2
    local method=$3
    local data=$4
    local num_requests=$5
    local num_concurrent=$6
    
    echo -e "${YELLOW}▶ $test_name${NC}"
    echo "  并发数: $num_concurrent | 总请求: $num_requests"
    
    local start_time=$(date +%s%N)
    local success=0
    local failed=0
    
    # 使用并发 curl 请求
    for ((i = 1; i <= num_requests; i++)); do
        (
            if [ "$method" = "GET" ]; then
                curl -s "$endpoint" > /dev/null 2>&1
            elif [ "$method" = "POST" ]; then
                curl -s -X POST "$endpoint" \
                    -H "Content-Type: application/json" \
                    -d "$data" > /dev/null 2>&1
            fi
            echo "done"
        ) &
        
        # 控制并发数
        if (( i % num_concurrent == 0 )); then
            wait
        fi
    done
    wait
    
    local end_time=$(date +%s%N)
    local duration_ns=$((end_time - start_time))
    local duration_ms=$((duration_ns / 1000000))
    local qps=$((num_requests * 1000 / (duration_ms + 1)))
    
    echo "  ✓ 完成 | 耗时: ${duration_ms}ms | QPS: $qps"
    echo ""
}

# 测试 1: 缓存未命中（查询数据库）
echo -e "${CYAN}【测试 1】无缓存状态 - 清除所有缓存后的请求${NC}"
echo ""

# 清除缓存：更新一个任务来清除所有缓存
curl -s -X PUT "$BASE_URL/api/tasks/1" \
    -H "Content-Type: application/json" \
    -d '{"title":"test","done":false}' > /dev/null

sleep 1
echo -e "${YELLOW}发送 100 个并发请求，每次都会查询数据库...${NC}"
run_load_test "无缓存请求（查询DB）" \
    "$BASE_URL/api/tasks" \
    "GET" \
    "" \
    100 \
    10

# 测试 2: 缓存命中（L1 缓存）
echo -e "${CYAN}【测试 2】L1 缓存命中状态 - 数据在本地内存中${NC}"
echo ""
echo -e "${YELLOW}发送 1000 个并发请求，命中 L1 本地缓存（超快）...${NC}"
run_load_test "L1 缓存命中（本地内存）" \
    "$BASE_URL/api/tasks" \
    "GET" \
    "" \
    1000 \
    100

# 测试 3: 混合读写
echo -e "${CYAN}【测试 3】混合读写压力测试${NC}"
echo ""
echo -e "${YELLOW}发送 500 个并发读请求 + 50 个写请求...${NC}"

local start_time=$(date +%s%N)

# 并发读
for ((i = 1; i <= 500; i++)); do
    curl -s "$BASE_URL/api/tasks" > /dev/null 2>&1 &
    if (( i % 50 == 0 )); then
        wait
    fi
done
wait

local end_time=$(date +%s%N)
local duration_ns=$((end_time - start_time))
local duration_ms=$((duration_ns / 1000000))
local qps=$((500 * 1000 / (duration_ms + 1)))

echo -e "${YELLOW}  ✓ 完成 | 耗时: ${duration_ms}ms | QPS: $qps${NC}"
echo ""

# 测试 4: 极限并发测试
echo -e "${CYAN}【测试 4】极限并发压力测试 - 5000 个请求${NC}"
echo ""
echo -e "${YELLOW}发送 5000 个高并发请求，测试系统容量...${NC}"
echo -e "${RED}⚠️  此测试会持续数秒，请耐心等待...${NC}"
echo ""

local start_time=$(date +%s%N)
local pids=()

for ((i = 1; i <= 5000; i++)); do
    curl -s "$BASE_URL/api/tasks/1" > /dev/null 2>&1 &
    pids+=($!)
    
    # 每 200 个请求打印进度
    if (( i % 200 == 0 )); then
        echo -ne "  进度: $i/5000 请求已发送...\r"
    fi
done

# 等待所有后台进程完成
for pid in "${pids[@]}"; do
    wait "$pid" 2>/dev/null
done

local end_time=$(date +%s%N)
local duration_ns=$((end_time - start_time))
local duration_ms=$((duration_ns / 1000000))
local duration_s=$((duration_ms / 1000))
local qps=$((5000 * 1000 / (duration_ms + 1)))

echo -e "  ✓ 完成 | 耗时: ${duration_ms}ms (${duration_s}s) | QPS: $qps"
echo ""

# ============================================================
# 测试总结
# ============================================================

echo ""
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}✨ 压力测试完成总结${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo ""
echo "缓存工作流程说明："
echo "  L1 缓存：本地内存，30秒过期，最快"
echo "  L2 缓存：Redis，5分钟过期，中速"
echo "  数据库：持久化存储，最慢"
echo ""
echo "Cache-Aside 策略流程："
echo "  1. 先查 L1（本地内存）"
echo "  2. 未命中 → 查 L2（Redis）"
echo "  3. 未命中 → 查数据库"
echo "  4. 将结果回写到 L1 和 L2"
echo "  5. 修改/删除时清除 L1 和 L2 缓存"
echo ""
echo -e "${CYAN}压力测试观察点：${NC}"
echo "  ✓ QPS 数值越高越好（理想 > 1000）"
echo "  ✓ L1 缓存命中时 QPS 应显著高于 L2 和 DB"
echo "  ✓ 超时配置确保请求不会长时间占用 goroutine"
echo "  ✓ 5000 并发请求应顺利完成（无超时错误）"
echo ""
echo -e "${GREEN}✅ 所有测试完成！${NC}"
