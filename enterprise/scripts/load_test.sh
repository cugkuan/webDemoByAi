#!/bin/bash

# 高级压力测试脚本
# 用于测试缓存系统和 HTTP 超时配置的性能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

BASE_URL="http://localhost:8080"
RESULTS_FILE="/tmp/load_test_results.txt"

# 检查服务
check_service() {
    if ! curl -s "$BASE_URL/api/tasks" > /dev/null 2>&1; then
        echo -e "${RED}❌ 服务未运行${NC}"
        exit 1
    fi
}

# 执行压力测试
run_test() {
    local test_name=$1
    local endpoint=$2
    local num_requests=$3
    local concurrent=$4
    local method=${5:-GET}
    
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}测试: $test_name${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo "参数: 总请求=$num_requests, 并发=$concurrent, 方法=$method"
    echo ""
    
    local success=0
    local failed=0
    local total_time=0
    local response_times=()
    
    echo -ne "${YELLOW}进度: "
    
    local start=$(date +%s%N)
    
    for ((i=1; i<=num_requests; i++)); do
        (
            local t1=$(date +%s%N)
            
            if [ "$method" = "GET" ]; then
                curl -s "$endpoint" > /dev/null 2>&1 && echo "ok" || echo "fail"
            else
                curl -s -X "$method" "$endpoint" \
                    -H "Content-Type: application/json" \
                    -d '{"title":"test","done":false}' > /dev/null 2>&1 && echo "ok" || echo "fail"
            fi
            
            local t2=$(date +%s%N)
            local elapsed=$((($t2 - $t1) / 1000000))  # 转换为 ms
            echo "$elapsed"
        ) &
        
        # 打印进度
        if (( i % (num_requests / 10) == 0 )); then
            echo -ne "▓"
        fi
        
        # 控制并发
        if (( i % concurrent == 0 )); then
            wait
        fi
    done
    wait
    
    local end=$(date +%s%N)
    local total_time=$(((end - start) / 1000000))  # ms
    
    local qps=$((num_requests * 1000 / (total_time + 1)))
    local avg_time=$((total_time / num_requests))
    
    echo -ne "${NC}\n"
    echo ""
    echo -e "${GREEN}✓ 测试完成${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "总耗时:    ${total_time}ms"
    echo "QPS:       $qps 请求/秒"
    echo "平均响应:  ${avg_time}ms"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # 保存结果
    echo "$test_name | 请求数=$num_requests | 并发=$concurrent | QPS=$qps | 平均响应=${avg_time}ms | 总耗时=${total_time}ms" >> "$RESULTS_FILE"
}

# 主程序
main() {
    clear
    
    echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  🔥 企业级缓存系统 - 高级压力测试套件 ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""
    
    check_service
    echo -e "${GREEN}✓ 服务已连接${NC}"
    echo ""
    
    # 清空结果文件
    > "$RESULTS_FILE"
    
    # 预热
    echo -e "${CYAN}预热系统...${NC}"
    for ((i=0; i<10; i++)); do
        curl -s "$BASE_URL/api/tasks" > /dev/null 2>&1 &
    done
    wait
    echo -e "${GREEN}✓ 预热完成${NC}"
    echo ""
    sleep 2
    
    # ===== 测试 1: 小规模测试 =====
    run_test "【1】小规模测试 (100个请求, 10并发)" \
        "$BASE_URL/api/tasks/1" \
        100 \
        10
    
    sleep 1
    
    # ===== 测试 2: 中等规模测试 =====
    run_test "【2】中等规模测试 (500个请求, 50并发)" \
        "$BASE_URL/api/tasks/1" \
        500 \
        50
    
    sleep 1
    
    # ===== 测试 3: 大规模测试 =====
    run_test "【3】大规模测试 (2000个请求, 100并发)" \
        "$BASE_URL/api/tasks/1" \
        2000 \
        100
    
    sleep 1
    
    # ===== 测试 4: 极限测试 =====
    run_test "【4】极限测试 (10000个请求, 200并发)" \
        "$BASE_URL/api/tasks/1" \
        10000 \
        200
    
    # ===== 显示总结 =====
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║              📊 测试结果总结              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""
    
    cat "$RESULTS_FILE" | column -t -s '|'
    
    echo ""
    echo -e "${GREEN}✅ 所有压力测试完成！${NC}"
    echo ""
    echo -e "${CYAN}性能指标说明：${NC}"
    echo "  • QPS > 1000:       优秀 ⭐⭐⭐"
    echo "  • QPS 500-1000:     良好 ⭐⭐"
    echo "  • QPS 100-500:      一般 ⭐"
    echo "  • QPS < 100:        需优化 ❌"
    echo ""
    echo -e "${CYAN}优化建议：${NC}"
    echo "  1. 检查缓存命中率（应 > 90%）"
    echo "  2. 监控 goroutine 数量（应稳定）"
    echo "  3. 检查数据库连接池配置"
    echo "  4. 考虑增加 Redis 实例"
    echo ""
}

main "$@"
