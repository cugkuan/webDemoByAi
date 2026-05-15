# 极简 Go Web 服务

只需要 **2 个文件**，**11 行代码**！

## 文件结构

```
web-demo/
├── go.mod      # Go 模块
└── main.go     # 程序 (11 行)
```

## 代码解析

```go
http.HandleFunc("/", hello)      // 注册路由
```
- 当访问 `/` 时，调用 `hello` 函数

```go
func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}
```
- `w` 是响应写入器
- `r` 是请求对象
- 发送 "Hello, World!" 给客户端

```go
http.ListenAndServe(":8080", nil)  // 启动服务
```
- 监听 8080 端口
- `nil` 表示使用默认路由

## 运行

### 1. 构建
```bash
cd /Users/kuan/Downloads/web-demo
go build -o web-demo
```

### 2. 运行
```bash
./web-demo
```

看到这个说明就成功了：
```
服务启动于 http://localhost:8080
```

### 3. 测试（新终端）
```bash
curl http://localhost:8080
```

返回：
```
Hello, World!
```

## 就这么简单！👍
