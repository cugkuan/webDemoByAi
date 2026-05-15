# 🚀 快速参考

## 核心代码位置

| 文件 | 功能 | 关键函数 |
|-----|------|--------|
| `cache.go` | 2级缓存实现 | `InitCache()`, `getL1/L2()`, `setL1/L2()` |
| `database.go` | 数据库操作 + 缓存 | `GetAllTasks()`, `CreateTask()`, `UpdateTask()` |
| `handler.go` | HTTP 路由 | `HandleTasks()` |
| `main.go` | 应用入口 | 初始化缓存和数据库 |

## 缓存流程

### 查询流程（读操作）

```
GetAllTasks()
  ├─ getL1(key) ────────────────────> 命中 → 返回
  │
  ├─ getL2(key) ────────────────────> 命中 → setL1() → 返回
  │
  └─ DB.Find() ──────────────────────> 查询 → setL1() → setL2() → 返回
```

### 修改流程（写操作）

```
UpdateTask(id, ...)
  ├─ DB.Save()  ─────────────────> 更新数据库
  │
  └─ invalidateTaskCache()
     ├─ deleteL1()  ─────────────> 清除本地缓存
     └─ deleteL2()  ─────────────> 清除 Redis 缓存
```

## 性能指标

| 操作 | 无缓存 | L1命中 | L2命中 |
|-----|-------|--------|--------|
| 时间 | 10-50ms | 0.1ms | 1-5ms |
| 加速 | - | 100-500x | 5-10x |

## 启动命令

```bash
# 1. 编译
go build -o app .

# 2. 启动 Redis（可选，支持自动降级）
redis-server &

# 3. 启动应用
./app

# 4. 测试
bash demo.sh
```

## 环境变量

```bash
# MySQL 连接字符串
export MYSQL_DSN="root@tcp(127.0.0.1:3306)/task_db?charset=utf8mb4&parseTime=True&loc=Local"

# 应用会自动连接 localhost:6379 的 Redis
```

## 常见调整

### 改变缓存时间

编辑 `cache.go`：

```go
const (
    CacheL1TTL = 30 * time.Second  // L1 过期时间
    CacheL2TTL = 5 * time.Minute   // L2 过期时间
)
```

### 添加新的缓存项

```go
// 在 cache.go 中定义新的 key
const NewCacheKey = "new:cache:key"

// 在数据库操作中使用
setL1(NewCacheKey, value)
setL2(ctx, NewCacheKey, value)
deleteL1(NewCacheKey)
deleteL2(ctx, NewCacheKey)
```

## 监控点

- Redis 连接状态（启动日志）
- 缓存命中次数（可添加计数器）
- 本地缓存大小（可添加统计）
- 数据库查询次数（可添加日志）

## 文件大小参考

```
app          ~12MB  (编译后的二进制)
cache.go     ~3.6KB (缓存模块)
database.go  ~3.2KB (数据库层)
CACHE_README ~6.4KB (完整文档)
```

---

**更多详情** → 查看 `CACHE_README.md`
