# 企业级 2 级缓存系统实现

## 📋 概述

本项目实现了生产级别的 **2 级缓存系统**，采用 **Cache-Aside** 策略：

```
┌─────────────┐
│   应用请求   │
└──────┬──────┘
       │
       ▼
┌──────────────────────┐
│  L1 缓存查询        │  (本地内存，30秒)
│  (超快，毫秒级)      │
└──────┬───────────────┘
       │ Miss
       ▼
┌──────────────────────┐
│  L2 缓存查询        │  (Redis，5分钟)
│  (快速，微秒级)      │
└──────┬───────────────┘
       │ Miss
       ▼
┌──────────────────────┐
│  数据库查询         │  (持久化)
│  (相对慢，毫秒级)    │
└──────┬───────────────┘
       │
       ▼
    写入 L1 + L2 缓存
```

## 🏗️ 架构设计

### 缓存层次

| 级别 | 存储 | 过期时间 | 特点 | 场景 |
|-----|------|---------|------|------|
| **L1** | 本地内存 | 30秒 | 超快，单机 | 热数据、频繁访问 |
| **L2** | Redis | 5分钟 | 快速，分布式 | 跨实例共享 |
| **DB** | MySQL | ∞ | 持久化 | 最终数据源 |

### 关键特性

1. **Cache-Aside 策略**
   - 应用负责缓存逻辑
   - 三层递进查询
   - 自动降级（Redis 不可用时只用本地缓存）

2. **自动缓存失效**
   - 修改操作自动清除相关缓存
   - 防止数据不一致

3. **后台清理**
   - 定时清理过期的本地缓存
   - 防止内存泄漏

4. **容错能力**
   - Redis 不可用时自动降级
   - 仅用本地缓存继续工作

## 📝 API 文件说明

### `cache.go` - 核心缓存模块

```go
// 缓存配置
const (
    CacheL1TTL = 30 * time.Second  // 本地缓存 30 秒
    CacheL2TTL = 5 * time.Minute   // Redis 缓存 5 分钟
)

// 初始化
InitCache()

// L1 操作
getL1(key)      // 获取本地缓存
setL1(key, val) // 设置本地缓存
deleteL1(key)   // 删除本地缓存

// L2 操作
getL2(ctx, key, dest)     // 获取 Redis 缓存
setL2(ctx, key, val)      // 设置 Redis 缓存
deleteL2(ctx, key)        // 删除 Redis 缓存

// 缓存失效
invalidateTaskCache(ctx, id)      // 清除特定任务缓存
invalidateAllTasksCache(ctx)      // 清除所有任务缓存
```

### `database.go` - 集成缓存的数据库操作

```go
// 读操作 - 自动使用缓存
GetAllTasks()       // 查询流程: L1 → L2 → DB
GetTaskByID(id)     // 查询流程: L1 → L2 → DB

// 写操作 - 自动清除缓存
CreateTask()        // 清除所有任务列表缓存
UpdateTask()        // 清除特定任务和列表缓存
DeleteTask()        // 清除特定任务和列表缓存
```

## 🚀 快速开始

### 1. 构建

```bash
cd /Users/kuan/Downloads/web-demo/enterprise-simple
go mod tidy
go build -o app .
```

### 2. 启动 Redis（可选）

```bash
# 如果有 Redis，启动它以获得完整的 2 级缓存体验
redis-server

# 如果没有 Redis，程序会自动降级到只用 L1 缓存
```

### 3. 启动应用

```bash
./app
```

### 4. 测试缓存

```bash
bash demo.sh
```

## 📊 性能对比

### 缓存命中场景

```
没有缓存：
GET /api/tasks/1 → MySQL 查询 ≈ 10-50ms

有 L1 缓存（命中）：
GET /api/tasks/1 → 本地内存 ≈ 0.1ms （快 100-500 倍）

有 L2 缓存（命中）：
GET /api/tasks/1 → Redis 查询 ≈ 1-5ms （快 5-10 倍）
```

### 场景分析

| 场景 | 流程 | 耗时 |
|-----|------|------|
| 首次请求 | DB 查询 + 缓存写入 | 10-50ms |
| 30秒内重复请求 | L1 缓存命中 | 0.1ms |
| 30秒后、5分钟内 | L2 缓存命中 | 1-5ms |
| 5分钟后 | DB 查询 + 缓存写入 | 10-50ms |

## 🔧 配置调优

### 调整过期时间

在 `cache.go` 中修改：

```go
const (
    CacheL1TTL = 30 * time.Second  // 改为 1 * time.Minute
    CacheL2TTL = 5 * time.Minute   // 改为 30 * time.Minute
)
```

**建议值**：
- 热数据：L1=30s, L2=5m
- 温数据：L1=5m, L2=30m
- 冷数据：L1=1m, L2=24h

### 监控缓存命中率

可以在 `cache.go` 中添加统计：

```go
type CacheStats struct {
    L1Hits    int64
    L1Misses  int64
    L2Hits    int64
    L2Misses  int64
}

func GetCacheHitRate() float64 {
    total := L1Hits + L1Misses
    if total == 0 {
        return 0
    }
    return float64(L1Hits) / float64(total) * 100
}
```

## 🛡️ 生产部署

### 高可用架构

```
多个应用实例
    ├── 本地缓存 (L1) × 3
    └── 共享 Redis 集群 (L2)
        └── MySQL 主从复制
```

### 部署清单

- [ ] 部署 Redis 集群或单机
- [ ] 配置 Redis 持久化 (AOF/RDB)
- [ ] 设置合理的过期策略
- [ ] 监控缓存命中率
- [ ] 配置告警（Redis 连接失败）
- [ ] 定期审查缓存大小

## 📈 优化建议

1. **针对不同数据类型的缓存策略**
   ```go
   // 用户数据：缓存更久
   UserCacheL1TTL = 5 * time.Minute
   
   // 实时数据：缓存较短
   RealtimeCacheL1TTL = 10 * time.Second
   ```

2. **分布式锁防止缓存击穿**
   ```go
   // 数据库不存在时的缓存穿透
   // 可以缓存一个特殊值表示不存在
   ```

3. **缓存预热**
   ```go
   func WarmupCache() {
       tasks, _ := GetAllTasks()
       // 应用启动时预加载热数据
   }
   ```

4. **缓存统计和监控**
   ```go
   // 导出 Prometheus 指标
   // 展示缓存命中率、大小等
   ```

## 🐛 故障排查

### Redis 不可用

```
日志输出：
⚠️  Redis 连接失败，仅使用本地缓存

解决：
1. 检查 Redis 是否在运行
2. 检查防火墙规则
3. 验证连接字符串
```

### 缓存不一致

```
症状：修改数据后，旧数据仍返回

原因：缓存未正确清除

解决：
1. 检查是否使用了 UpdateTask/DeleteTask 等 API
2. 检查缓存键是否正确
3. 手动清除：redisClient.FlushDB()
```

### 内存泄漏

```
症状：本地缓存持续增长

原因：清理任务未正常运行

解决：
1. 检查 goroutine 是否正常
2. 减少 L1TTL
3. 增加清理频率
```

## 📚 扩展阅读

- [Cache Patterns](https://martinfowler.com/bliki/CacheAsidePattern.html)
- [Redis 最佳实践](https://redis.io/topics/data-types)
- [Go 并发缓存](https://golang.org/doc/effective_go#sync)

## 许可

MIT
