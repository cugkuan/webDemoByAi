# 日志系统示例输出

## 系统启动输出

```
[ENTERPRISE] 2025-01-15 14:30:45.123 main.go:45: === 系统启动 ===
[ENTERPRISE] 2025-01-15 14:30:45.124 main.go:48: 初始化缓存系统...
[ENTERPRISE] 2025-01-15 14:30:45.125 cache.go:51: 初始化 Redis 客户端...
[ENTERPRISE] 2025-01-15 14:30:45.230 cache.go:65: ✓ Redis 连接成功
[ENTERPRISE] 2025-01-15 14:30:45.231 cache.go:68: 启动缓存清理后台任务...
[ENTERPRISE] 2025-01-15 14:30:45.232 main.go:51: 初始化数据库...
[ENTERPRISE] 2025-01-15 14:30:45.350 database.go:38: MySQL 数据库连接成功
[ENTERPRISE] 2025-01-15 14:30:45.351 main.go:54: 配置路由...
[ENTERPRISE] 2025-01-15 14:30:45.352 main.go:75: HTTP 服务器启动在 :8080

╔════════════════════════════════════════════╗
║      Gin REST API + 2级缓存系统启动         ║
╚════════════════════════════════════════════╝

📍 服务地址: http://localhost:8080
```

## 请求流程示例 1: 缓存未命中 + 数据库查询

```
[ENTERPRISE] 2025-01-15 14:31:10.100 router.go:68: REQUEST: GET /api/tasks HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:10.101 router.go:116: Handler: getTasks - 获取所有任务
[ENTERPRISE] 2025-01-15 14:31:10.102 database.go:47: CACHE MISS: L1 - 获取所有任务
[ENTERPRISE] 2025-01-15 14:31:10.103 database.go:51: CACHE MISS: L2 - 获取所有任务
[ENTERPRISE] 2025-01-15 14:31:10.104 database.go:55: DB QUERY: SELECT 所有任务
[ENTERPRISE] 2025-01-15 14:31:10.145 database.go:59: DB SUCCESS: 查询到 2 条任务记录
[ENTERPRISE] 2025-01-15 14:31:10.146 cache.go:94: CACHE SET: L1 - tasks:all (TTL: 30s)
[ENTERPRISE] 2025-01-15 14:31:10.230 cache.go:127: CACHE SET: L2 - tasks:all (TTL: 5m)
[ENTERPRISE] 2025-01-15 14:31:10.231 router.go:119: Handler: getTasks - 成功返回 2 个任务
[ENTERPRISE] 2025-01-15 14:31:10.232 router.go:70: RESPONSE: GET /api/tasks | Status: 200 | Duration: 132ms
```

## 请求流程示例 2: 缓存命中

```
[ENTERPRISE] 2025-01-15 14:31:12.100 router.go:68: REQUEST: GET /api/tasks HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:12.101 router.go:116: Handler: getTasks - 获取所有任务
[ENTERPRISE] 2025-01-15 14:31:12.102 database.go:47: CACHE HIT: L1 - 获取所有任务
[ENTERPRISE] 2025-01-15 14:31:12.103 router.go:119: Handler: getTasks - 成功返回 2 个任务
[ENTERPRISE] 2025-01-15 14:31:12.104 router.go:70: RESPONSE: GET /api/tasks | Status: 200 | Duration: 4ms
```

## 请求流程示例 3: 创建任务（数据库修改 + 缓存失效）

```
[ENTERPRISE] 2025-01-15 14:31:20.100 router.go:68: REQUEST: POST /api/tasks HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:20.101 router.go:187: Handler: createTask - 创建任务 Title=学习 Docker, Done=false
[ENTERPRISE] 2025-01-15 14:31:20.102 database.go:107: DB CREATE: 创建任务 - Title=学习 Docker, Done=false
[ENTERPRISE] 2025-01-15 14:31:20.135 database.go:112: DB SUCCESS: 任务创建成功 ID=3
[ENTERPRISE] 2025-01-15 14:31:20.136 database.go:114: CACHE INVALIDATE: 清除所有任务列表缓存
[ENTERPRISE] 2025-01-15 14:31:20.137 cache.go:115: CACHE DELETE: L1 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:20.195 cache.go:160: CACHE DELETE: L2 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:20.196 router.go:216: Handler: createTask - 成功创建任务 ID=3
[ENTERPRISE] 2025-01-15 14:31:20.197 router.go:70: RESPONSE: POST /api/tasks | Status: 201 | Duration: 97ms
```

## 请求流程示例 4: 获取单个任务（第一次 + 第二次）

```
【第一次 - 缓存未命中】
[ENTERPRISE] 2025-01-15 14:31:30.100 router.go:68: REQUEST: GET /api/tasks/1 HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:30.101 router.go:143: Handler: getTask - 获取任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:30.102 database.go:75: CACHE MISS: L1 - 获取任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:30.103 database.go:79: CACHE MISS: L2 - 获取任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:30.104 database.go:82: DB QUERY: SELECT 任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:30.110 database.go:85: DB SUCCESS: 查询到任务 ID=1, Title=学习 Go
[ENTERPRISE] 2025-01-15 14:31:30.111 cache.go:94: CACHE SET: L1 - task:1 (TTL: 30s)
[ENTERPRISE] 2025-01-15 14:31:30.160 cache.go:127: CACHE SET: L2 - task:1 (TTL: 5m)
[ENTERPRISE] 2025-01-15 14:31:30.161 router.go:169: Handler: getTask - 成功返回任务 ID=1, Title=学习 Go
[ENTERPRISE] 2025-01-15 14:31:30.162 router.go:70: RESPONSE: GET /api/tasks/1 | Status: 200 | Duration: 62ms

【第二次 - L1 缓存命中】
[ENTERPRISE] 2025-01-15 14:31:31.100 router.go:68: REQUEST: GET /api/tasks/1 HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:31.101 router.go:143: Handler: getTask - 获取任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:31.102 database.go:75: CACHE HIT: L1 - 获取任务 ID=1
[ENTERPRISE] 2025-01-15 14:31:31.103 router.go:169: Handler: getTask - 成功返回任务 ID=1, Title=学习 Go
[ENTERPRISE] 2025-01-15 14:31:31.104 router.go:70: RESPONSE: GET /api/tasks/1 | Status: 200 | Duration: 4ms
```

## 请求流程示例 5: 更新任务

```
[ENTERPRISE] 2025-01-15 14:31:40.100 router.go:68: REQUEST: PUT /api/tasks/2 HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:40.101 router.go:236: Handler: updateTask - 更新任务 ID=2, Title=构建 REST API v2, Done=true
[ENTERPRISE] 2025-01-15 14:31:40.102 database.go:122: DB UPDATE: 更新任务 ID=2 - Title=构建 REST API v2, Done=true
[ENTERPRISE] 2025-01-15 14:31:40.125 database.go:133: DB SUCCESS: 任务更新成功 ID=2
[ENTERPRISE] 2025-01-15 14:31:40.126 database.go:135: CACHE INVALIDATE: 清除任务缓存 ID=2
[ENTERPRISE] 2025-01-15 14:31:40.127 cache.go:115: CACHE DELETE: L1 - task:2
[ENTERPRISE] 2025-01-15 14:31:40.180 cache.go:160: CACHE DELETE: L2 - task:2
[ENTERPRISE] 2025-01-15 14:31:40.181 cache.go:115: CACHE DELETE: L1 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:40.235 cache.go:160: CACHE DELETE: L2 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:40.236 router.go:274: Handler: updateTask - 成功更新任务 ID=2
[ENTERPRISE] 2025-01-15 14:31:40.237 router.go:70: RESPONSE: PUT /api/tasks/2 | Status: 200 | Duration: 137ms
```

## 请求流程示例 6: 删除任务

```
[ENTERPRISE] 2025-01-15 14:31:50.100 router.go:68: REQUEST: DELETE /api/tasks/3 HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:50.101 router.go:293: Handler: deleteTask - 删除任务 ID=3
[ENTERPRISE] 2025-01-15 14:31:50.102 database.go:142: DB DELETE: 删除任务 ID=3
[ENTERPRISE] 2025-01-15 14:31:50.125 database.go:148: DB SUCCESS: 任务删除成功 ID=3
[ENTERPRISE] 2025-01-15 14:31:50.126 database.go:150: CACHE INVALIDATE: 清除任务缓存 ID=3
[ENTERPRISE] 2025-01-15 14:31:50.127 cache.go:115: CACHE DELETE: L1 - task:3
[ENTERPRISE] 2025-01-15 14:31:50.180 cache.go:160: CACHE DELETE: L2 - task:3
[ENTERPRISE] 2025-01-15 14:31:50.181 cache.go:115: CACHE DELETE: L1 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:50.235 cache.go:160: CACHE DELETE: L2 - tasks:all
[ENTERPRISE] 2025-01-15 14:31:50.236 router.go:309: Handler: deleteTask - 成功删除任务 ID=3
[ENTERPRISE] 2025-01-15 14:31:50.237 router.go:70: RESPONSE: DELETE /api/tasks/3 | Status: 200 | Duration: 137ms
```

## 错误场景日志

### 任务不存在
```
[ENTERPRISE] 2025-01-15 14:32:00.100 router.go:68: REQUEST: GET /api/tasks/999 HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:32:00.101 router.go:143: Handler: getTask - 获取任务 ID=999
[ENTERPRISE] 2025-01-15 14:32:00.102 database.go:75: CACHE MISS: L1 - 获取任务 ID=999
[ENTERPRISE] 2025-01-15 14:32:00.103 database.go:79: CACHE MISS: L2 - 获取任务 ID=999
[ENTERPRISE] 2025-01-15 14:32:00.104 database.go:82: DB QUERY: SELECT 任务 ID=999
[ENTERPRISE] 2025-01-15 14:32:00.110 database.go:86: ERROR: 数据库查询失败 ID=999 - record not found
[ENTERPRISE] 2025-01-15 14:32:00.111 router.go:160: ERROR: getTask - 任务不存在 ID=999
[ENTERPRISE] 2025-01-15 14:32:00.112 router.go:70: RESPONSE: GET /api/tasks/999 | Status: 404 | Duration: 12ms
```

### 无效的任务 ID
```
[ENTERPRISE] 2025-01-15 14:32:10.100 router.go:68: REQUEST: GET /api/tasks/abc HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:32:10.101 router.go:140: ERROR: getTask - 无效的 ID 参数: abc
[ENTERPRISE] 2025-01-15 14:32:10.102 router.go:70: RESPONSE: GET /api/tasks/abc | Status: 400 | Duration: 2ms
```

### 无效的请求体
```
[ENTERPRISE] 2025-01-15 14:32:20.100 router.go:68: REQUEST: POST /api/tasks HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:32:20.101 router.go:193: ERROR: createTask - 无效的请求体: EOF
[ENTERPRISE] 2025-01-15 14:32:20.102 router.go:70: RESPONSE: POST /api/tasks | Status: 400 | Duration: 3ms
```

## 后台清理任务日志

```
[ENTERPRISE] 2025-01-15 14:35:00.100 cache.go:82: CLEANUP: 清理过期本地缓存 3 项
[ENTERPRISE] 2025-01-15 14:36:00.100 cache.go:82: CLEANUP: 清理过期本地缓存 5 项
[ENTERPRISE] 2025-01-15 14:37:00.100 cache.go:82: CLEANUP: 清理过期本地缓存 2 项
```

## 系统健康检查

```
[ENTERPRISE] 2025-01-15 14:31:05.100 router.go:68: REQUEST: GET /health HTTP/1.1 | IP: 127.0.0.1
[ENTERPRISE] 2025-01-15 14:31:05.101 router.go:100: Handler: healthCheck - 执行健康检查
[ENTERPRISE] 2025-01-15 14:31:05.102 router.go:101: Handler: healthCheck - DB: ok, Redis: ok
[ENTERPRISE] 2025-01-15 14:31:05.103 router.go:70: RESPONSE: GET /health | Status: 200 | Duration: 3ms
```

## 日志分析示例

### 提取特定用户的所有操作
```bash
grep "IP: 192.168.1.100" app.log
```

### 找出最慢的请求
```bash
grep "RESPONSE:" app.log | sort -t: -k8 -rn | head -10
```

### 统计缓存命中率
```bash
HITS=$(grep "CACHE HIT" app.log | wc -l)
MISS=$(grep "CACHE MISS" app.log | wc -l)
echo "缓存命中率: $((HITS * 100 / (HITS + MISS)))%"
```

### 找出所有失败的请求
```bash
grep "Status: [45]" app.log
```

### 监控 Redis 连接问题
```bash
grep -i "redis\|error" app.log
```
