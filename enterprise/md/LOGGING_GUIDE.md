# Enterprise 应用日志系统指南

## 概述

已为 enterprise 目录下的代码添加了完整的结构化日志系统，支持追踪请求流程、缓存命中/未命中、数据库操作和系统事件。

## 日志级别和前缀

所有日志统一使用 `[ENTERPRISE]` 前缀和时间戳，格式示例：
```
[ENTERPRISE] 2025-01-15 10:30:45 INFO: 任务创建成功 ID=1
```

### 日志分类

| 前缀 | 说明 | 示例 |
|-----|------|------|
| `REQUEST:` | HTTP 请求日志 | `REQUEST: GET /api/tasks HTTP/1.1 \| IP: 127.0.0.1` |
| `RESPONSE:` | HTTP 响应日志 | `RESPONSE: GET /api/tasks \| Status: 200 \| Duration: 15ms` |
| `CACHE HIT:` | 缓存命中 | `CACHE HIT: L1 - 获取任务 ID=1` |
| `CACHE MISS:` | 缓存未命中 | `CACHE MISS: L1 - 获取所有任务` |
| `CACHE SET:` | 缓存写入 | `CACHE SET: L1 - task:1 (TTL: 30s)` |
| `CACHE DELETE:` | 缓存删除 | `CACHE DELETE: L1 - task:1` |
| `CACHE INVALIDATE:` | 缓存失效 | `CACHE INVALIDATE: 清除任务缓存 ID=1` |
| `DB QUERY:` | 数据库查询 | `DB QUERY: SELECT 任务 ID=1` |
| `DB SUCCESS:` | 数据库操作成功 | `DB SUCCESS: 查询到任务 ID=1, Title=学习Go` |
| `DB CREATE:` | 数据库创建 | `DB CREATE: 创建任务 - Title=新任务, Done=false` |
| `DB UPDATE:` | 数据库更新 | `DB UPDATE: 更新任务 ID=1 - Title=新标题, Done=true` |
| `DB DELETE:` | 数据库删除 | `DB DELETE: 删除任务 ID=1` |
| `ERROR:` | 错误日志 | `ERROR: 数据库连接失败: connection refused` |
| `CLEANUP:` | 定期清理任务 | `CLEANUP: 清理过期本地缓存 5 项` |
| `Handler:` | 处理器执行 | `Handler: getTasks - 获取所有任务` |

## 日志输出位置

### main.go
- 系统启动/关闭
- 缓存系统初始化
- 数据库初始化
- HTTP 服务器启动

### database.go
- 数据库连接和迁移
- CRUD 操作（CREATE, READ, UPDATE, DELETE）
- 缓存一致性操作
- 查询记录（数量、ID、标题等）

### cache.go
- L1/L2 缓存命中/未命中
- 缓存读写操作
- 缓存删除操作
- Redis 连接和心跳检测
- 后台清理任务
- 序列化/反序列化错误

### router.go
- HTTP 请求进入和响应离开
- 请求延迟（Duration）
- 参数验证错误
- 各 handler 业务逻辑执行

## 关键业务流程日志

### 1. 获取任务列表流程
```
REQUEST: GET /api/tasks HTTP/1.1 | IP: 127.0.0.1
Handler: getTasks - 获取所有任务
CACHE MISS: L1 - 获取所有任务
CACHE MISS: L2 - 获取所有任务
DB QUERY: SELECT 所有任务
DB SUCCESS: 查询到 5 条任务记录
CACHE SET: L1 - tasks:all (TTL: 30s)
CACHE SET: L2 - tasks:all (TTL: 5m)
Handler: getTasks - 成功返回 5 个任务
RESPONSE: GET /api/tasks | Status: 200 | Duration: 45ms
```

### 2. 二级缓存命中流程（后续请求）
```
REQUEST: GET /api/tasks HTTP/1.1 | IP: 127.0.0.1
Handler: getTasks - 获取所有任务
CACHE HIT: L1 - 获取所有任务
Handler: getTasks - 成功返回 5 个任务
RESPONSE: GET /api/tasks | Status: 200 | Duration: 2ms
```

### 3. 创建任务流程
```
REQUEST: POST /api/tasks HTTP/1.1 | IP: 127.0.0.1
Handler: createTask - 创建任务 Title=新任务, Done=false
DB CREATE: 创建任务 - Title=新任务, Done=false
DB SUCCESS: 任务创建成功 ID=6
CACHE INVALIDATE: 清除所有任务列表缓存
CACHE DELETE: L1 - tasks:all
CACHE DELETE: L2 - tasks:all
Handler: createTask - 成功创建任务 ID=6
RESPONSE: POST /api/tasks | Status: 201 | Duration: 12ms
```

### 4. 更新任务流程
```
REQUEST: PUT /api/tasks/1 HTTP/1.1 | IP: 127.0.0.1
Handler: updateTask - 更新任务 ID=1, Title=更新后, Done=true
DB UPDATE: 更新任务 ID=1 - Title=更新后, Done=true
DB SUCCESS: 任务更新成功 ID=1
CACHE INVALIDATE: 清除任务缓存 ID=1
CACHE DELETE: L1 - task:1
CACHE DELETE: L2 - task:1
CACHE DELETE: L1 - tasks:all
CACHE DELETE: L2 - tasks:all
Handler: updateTask - 成功更新任务 ID=1
RESPONSE: PUT /api/tasks/1 | Status: 200 | Duration: 18ms
```

### 5. 错误流程示例
```
REQUEST: GET /api/tasks/999 HTTP/1.1 | IP: 127.0.0.1
Handler: getTask - 获取任务 ID=999
CACHE MISS: L1 - 获取任务 ID=999
CACHE MISS: L2 - 获取任务 ID=999
DB QUERY: SELECT 任务 ID=999
ERROR: 数据库查询失败 ID=999 - record not found
Handler: getTask - 任务不存在 ID=999
RESPONSE: GET /api/tasks/999 | Status: 404 | Duration: 8ms
```

## 日志分析技巧

### 1. 监控缓存效率
统计 `CACHE HIT` vs `CACHE MISS` 比率，L1 缓存命中率应该很高

### 2. 识别性能瓶颈
- 检查 Duration 时间，超过 100ms 的请求需要优化
- 对比有缓存和无缓存的响应时间

### 3. 调试数据一致性问题
- 追踪 `CACHE INVALIDATE` 和数据库操作的关联
- 验证缓存更新是否及时

### 4. 监控错误
搜索 `ERROR:` 前缀识别所有异常情况

### 5. 追踪缓存工作
通过 `CACHE SET/DELETE` 日志检查缓存失效策略

## 启用和禁用日志

### 临时禁用日志（生产环境可选）
编辑 `main.go` 中的日志初始化：
```go
// 方案1: 重定向到文件
logFile, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
logger = log.New(logFile, "[ENTERPRISE] ", log.LstdFlags|log.Lshortfile)

// 方案2: 禁用日志（不推荐）
logger = log.New(io.Discard, "", 0)
```

## 日志文件持久化

创建 `docker-compose.override.yml` 用于日志卷挂载：
```yaml
version: '3.8'
services:
  app:
    volumes:
      - ./logs:/app/logs
    environment:
      LOG_FILE: /app/logs/enterprise.log
```

## 最佳实践

1. **生产环境**：结合 ELK（Elasticsearch + Logstash + Kibana）进行集中日志管理
2. **开发环境**：直接输出到 stdout，通过 Docker/systemd 日志采集
3. **日志轮转**：使用 `logrotate` 管理日志文件大小
4. **敏感信息**：不记录密码、令牌等敏感数据（当前代码已避免）
5. **性能**：日志操作已优化，不会显著影响性能

## 相关文件

- `main.go` - 日志系统初始化
- `database.go` - 数据库操作日志
- `cache.go` - 缓存操作日志
- `router.go` - HTTP 请求/响应日志和 handler 日志
- `handler.go` - 原始处理器（未使用 Gin，保留供参考）
