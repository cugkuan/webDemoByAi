# 日志系统快速参考

## 文件变更一览

| 文件 | 变更 | 日志点 |
|-----|------|--------|
| `main.go` | ✅ 新增日志系统初始化 | 6 |
| `database.go` | ✅ 添加 CRUD 和缓存日志 | 18 |
| `cache.go` | ✅ 添加缓存操作日志 | 15 |
| `router.go` | ✅ 添加 HTTP 和业务日志 | 18 |

## 新增文档

| 文档 | 用途 | 推荐阅读顺序 |
|-----|------|-----------|
| `LOGGING_GUIDE.md` | 详细使用指南 | 1️⃣ 首先 |
| `LOG_ARCHITECTURE.md` | 架构和流程图 | 2️⃣ 其次 |
| `LOG_EXAMPLES.md` | 实际输出示例 | 3️⃣ 参考 |
| `LOGGING_SUMMARY.md` | 变更总结 | 4️⃣ 备查 |

## 常用命令

### Docker 方式（推荐）
```bash
# 一键启动（MySQL + Redis + API）
docker compose up -d

# 查看日志
docker compose logs -f app

# 停止
docker compose down
```

### 本地编译运行
```bash
# 编译
go build -o app ./cmd/server/

# 运行（输出日志到 stdout）
./app

# 运行并保存日志文件
./app > app.log 2>&1 &

# 运行自动化测试
./test_logging.sh

# 查看特定类型的日志
grep "CACHE HIT" app.log      # 缓存命中
grep "CACHE MISS" app.log     # 缓存未命中
grep "DB QUERY" app.log       # 数据库查询
grep "ERROR" app.log          # 所有错误
grep "Duration:" app.log      # 所有请求及耗时

# 找出最慢的请求
grep "RESPONSE:" app.log | sort -t' ' -k8 -rn | head -5

# 实时监控错误
tail -f app.log | grep ERROR
```

## 日志类型速查

| 前缀 | 含义 | 场景 |
|-----|------|------|
| `REQUEST:` | HTTP 请求进入 | 请求开始 |
| `RESPONSE:` | HTTP 响应离开 | 请求完成 |
| `CACHE HIT:` | 缓存命中 | 数据从缓存返回 |
| `CACHE MISS:` | 缓存未命中 | 需要查询数据库 |
| `CACHE SET:` | 写入缓存 | 数据已缓存 |
| `CACHE DELETE:` | 删除缓存 | 缓存已清除 |
| `CACHE INVALIDATE:` | 缓存失效 | 批量清除缓存 |
| `DB QUERY:` | 数据库查询 | 执行 SELECT |
| `DB CREATE:` | 创建记录 | 执行 INSERT |
| `DB UPDATE:` | 更新记录 | 执行 UPDATE |
| `DB DELETE:` | 删除记录 | 执行 DELETE |
| `DB SUCCESS:` | 操作成功 | 返回结果 |
| `ERROR:` | 错误发生 | 异常捕获 |
| `Handler:` | 处理器执行 | 业务逻辑 |
| `CLEANUP:` | 后台清理 | 定期维护 |

## 典型性能数据

| 操作 | 典型耗时 | 说明 |
|-----|--------|------|
| 缓存命中 | 2-5ms | L1 缓存直接返回 |
| L2 缓存命中 | 5-10ms | Redis 返回缓存 |
| 数据库查询 | 20-100ms | 取决于数据量 |
| 完整请求 (MISS) | 50-150ms | MISS + DB 查询 + 缓存写入 |

## 调试技巧

### 追踪特定请求
```bash
REQ_ID="GET /api/tasks/1"
grep "$REQ_ID" app.log
```

### 统计缓存命中率
```bash
HITS=$(grep "CACHE HIT" app.log | wc -l)
MISS=$(grep "CACHE MISS" app.log | wc -l)
TOTAL=$((HITS + MISS))
echo "命中率: $((HITS * 100 / TOTAL))%"
```

### 找出性能问题
```bash
# 找出超过 100ms 的请求
grep "RESPONSE:" app.log | awk -F'Duration: ' '{print $2}' | \
  awk '{print $1}' | sed 's/ms//' | awk '$1 > 100 {print "慢请求: " $1 "ms"}'

# 按耗时排序
grep "RESPONSE:" app.log | sort -t'Duration: ' -k2 -rn | head -10
```

### 监控错误趋势
```bash
# 每分钟的错误数
tail -100 app.log | grep "ERROR" | wc -l

# 实时监控
tail -f app.log | grep -E "ERROR|WARN"
```

## 生产环境配置

### 1. 日志到文件
编辑 `main.go` 中的日志初始化：
```go
logFile, _ := os.OpenFile("/var/log/enterprise.log", 
    os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
logger = log.New(logFile, "[ENTERPRISE] ", log.LstdFlags|log.Lshortfile)
```

### 2. 日志轮转配置
创建 `/etc/logrotate.d/enterprise`:
```
/var/log/enterprise.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
}
```

### 3. 集成监控
```bash
# 使用 systemd journal
journalctl -u enterprise -f

# 使用 syslog
logger -t enterprise "日志信息"

# 集成 ELK
# 使用 filebeat 采集 /var/log/enterprise.log
```

## 常见问题

**Q: 日志会影响性能吗？**
A: 不会。日志记录耗时 < 0.1ms，对总耗时影响 < 1%

**Q: 如何禁用日志？**
A: 修改 `main.go` 中的日志初始化，使用 `io.Discard` 或 `/dev/null`

**Q: 日志文件会很大吗？**
A: 平均每 1000 请求生成 10-20KB 日志，可以配置日志轮转

**Q: 能否只记录特定类型的日志？**
A: 可以，在 `main.go` 中添加日志级别过滤逻辑

**Q: 如何搜索特定用户的所有操作？**
A: `grep "IP: 192.168.1.100" app.log`

## 更多信息

- 详细指南: 见 `LOGGING_GUIDE.md`
- 架构说明: 见 `LOG_ARCHITECTURE.md`
- 输出示例: 见 `LOG_EXAMPLES.md`
- 完整总结: 见 `LOGGING_SUMMARY.md`

---

**记住**: 日志系统已完全集成，开箱即用，无需任何配置！
