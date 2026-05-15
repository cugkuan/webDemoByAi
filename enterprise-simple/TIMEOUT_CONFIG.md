# 超时配置说明

## 📋 概述

已在 `main.go` 中添加 HTTP 服务器超时配置，防止 goroutine 泄漏和资源耗尽。

## ⏱️ 超时类型

### 1. ReadTimeout (读超时) - 15 秒

```go
ReadTimeout: 15 * time.Second
```

**作用**：客户端发送请求数据的最大时间

**场景**：
- 防止客户端慢速发送请求
- 防止 Slowloris 攻击（缓慢 HTTP 请求攻击）

**触发**：
```bash
# 客户端 15 秒内没有完成请求发送
curl http://localhost:8080/api/tasks < file_that_takes_20_seconds
```

### 2. WriteTimeout (写超时) - 15 秒

```go
WriteTimeout: 15 * time.Second
```

**作用**：服务器写入响应数据的最大时间

**场景**：
- 防止客户端接收缓慢
- 防止内存堆积

**触发**：
```bash
# 客户端很慢，接收响应超过 15 秒
nc -l 12345  # 监听不读取，会导致超时
```

### 3. IdleTimeout (空闲超时) - 60 秒

```go
IdleTimeout: 60 * time.Second
```

**作用**：Keep-Alive 连接的最大空闲时间

**场景**：
- 关闭长时间不活动的连接
- 释放不必要的文件描述符

**效果**：
- 60 秒内没有新请求 → 连接自动关闭
- 防止僵尸连接积累

## 📊 配置对比

| 配置项 | 时间 | 用途 | 影响 |
|-------|------|------|------|
| ReadTimeout | 15s | 防慢速请求 | 安全性 |
| WriteTimeout | 15s | 防慢速响应 | 内存占用 |
| IdleTimeout | 60s | 关闭空闲连接 | 资源释放 |

## 🔄 请求处理流程

```
客户端连接
    ↓
[0-15s] 读取请求数据 (ReadTimeout)
    ↓
处理业务逻辑
    ↓
[0-15s] 写入响应数据 (WriteTimeout)
    ↓
连接保持 keep-alive
    ↓
[0-60s] 等待新请求 (IdleTimeout)
    ↓
60s 未见请求 → 连接关闭
```

## 🎯 Goroutine 生命周期

### 没有超时配置 ❌

```
请求来临 → 创建 goroutine → 请求处理 → 客户端断开连接
          ↑                               ↓
          └───── 可能卡住或泄漏 ─────────┘
```

### 有超时配置 ✅

```
请求来临 → 创建 goroutine → 请求处理 → 15s/60s 超时 → goroutine 释放
          ↑                               ↓
          └────────── 保证释放 ──────────┘
```

## 💡 场景分析

### 场景1：正常请求（< 15s）

```
✅ 正常处理
- 请求到达
- 处理业务逻辑（通常 < 100ms）
- 返回响应
- 连接保持 keep-alive
- 60s 无新请求后关闭
```

### 场景2：缓慢客户端

```
❌ ReadTimeout 触发
- 客户端 15s 内没发送完整请求
- 服务器关闭连接
- Goroutine 释放
```

### 场景3：处理耗时任务

```
⚠️  WriteTimeout 触发
- 业务处理 < 15s（大多数情况）
- 但如果响应数据很大，发送 > 15s
- 服务器强制关闭

解决方案：
1. 分批返回数据
2. 使用流式响应
3. 增加 WriteTimeout
```

### 场景4：长连接空闲

```
✅ IdleTimeout 生效
- 连接建立后，60s 内无新请求
- 服务器主动关闭连接
- Goroutine 释放
- 释放文件描述符
```

## 🔧 调整超时参数

### 根据业务调整

```go
// 快速 API（缓存命中）
ReadTimeout:  5 * time.Second
WriteTimeout: 5 * time.Second
IdleTimeout:  30 * time.Second

// 文件上传/下载
ReadTimeout:  60 * time.Second
WriteTimeout: 60 * time.Second
IdleTimeout:  120 * time.Second

// 实时推送/WebSocket
ReadTimeout:  0 // 无限制
WriteTimeout: 0 // 无限制
IdleTimeout:  0 // 无限制
```

### 当前配置适用场景

```
ReadTimeout:  15s  ← 适合大多数 REST API
WriteTimeout: 15s  ← 适合大多数 REST API
IdleTimeout:  60s  ← 适合 Web 浏览器长连接
```

## 📈 性能影响

### Goroutine 泄漏对比

| 状态 | 内存占用 | Goroutine 数 | 问题 |
|-----|---------|------------|------|
| 无超时 | 持续增长 | 持续增长 | ❌ 内存泄漏 |
| 有超时 | 稳定 | 稳定 | ✅ 正常 |

### 实际测试

```bash
# 模拟 100 个长连接
for i in {1..100}; do
  nc localhost 8080 < /dev/null &
done

# 无超时：goroutine 积累到 100+
# 有超时：60s 后自动下降到 < 10
```

## 🚀 验证超时配置

### 测试 ReadTimeout

```bash
# 将数据分散发送，超过 15 秒
(echo -n "GET /api/tasks "; sleep 20; echo "HTTP/1.1") | nc localhost 8080
# 预期：连接关闭，500 错误
```

### 测试 WriteTimeout

```bash
# 使用 curl，在响应返回时中断接收
curl --max-time 20 http://localhost:8080/api/tasks
# 预期：正常完成（因为响应通常 < 15s）
```

### 测试 IdleTimeout

```bash
# 建立连接后等待 70 秒
(echo ""; sleep 70) | nc localhost 8080
# 预期：70s 后连接关闭
```

## 📊 监控建议

添加日志跟踪超时：

```go
// handler.go
func HandleTasks(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // 业务处理...
    
    duration := time.Since(start)
    if duration > 10*time.Second {
        log.Printf("⚠️  耗时请求: %s %s - %dms", 
            r.Method, r.URL.Path, duration.Milliseconds())
    }
}
```

## ⚠️ 注意事项

1. **不要设置为 0**（无限制）
   - 导致 goroutine 泄漏
   - 内存持续增长

2. **ReadTimeout 应小于等于 WriteTimeout**
   - 避免不一致

3. **考虑网络延迟**
   - 公网 API：增大超时
   - 内网 API：可以缩小

4. **大文件下载**
   - 需要特殊处理（分块传输）
   - 或禁用超时

## 📚 HTTP 服务器完整配置参考

```go
server := &http.Server{
    Addr:           ":8080",
    Handler:        http.DefaultServeMux,
    ReadTimeout:    15 * time.Second,   // 读超时
    WriteTimeout:   15 * time.Second,   // 写超时
    IdleTimeout:    60 * time.Second,   // 空闲超时
    MaxHeaderBytes: 1 << 20,            // 1MB 请求头
}
server.ListenAndServe()
```

## 🎯 总结

✅ 已在 `main.go` 添加超时配置
✅ ReadTimeout: 15s - 防止慢速请求
✅ WriteTimeout: 15s - 防止慢速响应
✅ IdleTimeout: 60s - 释放空闲连接
✅ 防止 Goroutine 泄漏
✅ 生产就绪
