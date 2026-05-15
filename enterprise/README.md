# Enterprise - 企业级 Go Web 项目

基于 Gin + GORM + Redis 的企业级 REST API 服务，采用分层架构和二级缓存设计。

## 技术栈

| 组件 | 技术 |
|------|------|
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis + 本地内存（二级缓存） |
| 日志 | zerolog |
| 文档 | Swagger |
| 部署 | Docker / Docker Compose |

## 项目结构

```
enterprise/
├── cmd/server/          # 入口
├── config/              # 配置管理
├── internal/
│   ├── cache/           # 二级缓存（L1 本地内存 + L2 Redis）
│   ├── database/        # 数据库连接
│   ├── handler/         # HTTP 处理器
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── router/          # 路由配置
│   └── service/         # 业务逻辑层
├── middleware/           # 中间件（日志、错误处理、限流、请求ID）
├── pkg/
│   ├── logger/          # 日志初始化
│   └── response/        # 统一响应格式
├── errors/              # 统一错误处理
├── migrations/          # 数据库迁移脚本
├── scripts/             # 辅助脚本
├── docs/                # Swagger 文档
├── Dockerfile           # 多阶段构建
├── docker-compose.yml   # 容器编排
└── config.yaml          # 配置文件
```

## 快速开始

### 使用 Docker（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 停止服务
docker-compose down
```

### 本地开发

```bash
# 确保 MySQL 和 Redis 已启动

# 运行
go run ./cmd/server/

# 测试
go test ./...
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/health/liveness` | 存活检查 |
| GET | `/health/readiness` | 就绪检查 |
| GET | `/api/tasks` | 获取所有任务 |
| GET | `/api/tasks/count` | 统计任务总数 |
| GET | `/api/tasks/:id` | 获取单个任务 |
| POST | `/api/tasks` | 创建任务 |
| PUT | `/api/tasks/:id` | 更新任务 |
| DELETE | `/api/tasks/:id` | 删除任务 |
| DELETE | `/api/cache` | 清除所有缓存 |
| DELETE | `/api/cache/:id` | 清除任务缓存 |
| GET | `/swagger/*any` | Swagger API 文档 |

## 架构特点

### 分层架构

```
Handler（HTTP） → Service（业务） → Repository（数据）
```

- **Handler**: 处理 HTTP 请求/响应，参数校验
- **Service**: 业务逻辑，缓存策略
- **Repository**: 数据库操作

### 二级缓存

```
请求 → L1（本地内存，30s）→ L2（Redis，5min）→ 数据库
```

写操作自动清除相关缓存，保证数据一致性。

### 中间件

- **RequestID**: 每个请求分配唯一 ID，支持链路追踪
- **Logger**: 结构化日志，按状态码分级
- **ErrorHandler**: 统一错误处理
- **RateLimit**: 令牌桶限流
- **CORS**: 跨域支持

## 日志

### 输出位置

日志**同时输出到控制台和文件**，方便实时查看和事后分析。

**Docker 环境**：日志自动写入容器内 `/app/logs/app.log`，通过 Docker volume 持久化。

**本地开发**：在 `config.yaml` 中取消注释即可启用文件日志：

```yaml
log:
  file_path: "logs/app.log"
```

### 查看日志

```bash
# 实时查看（Docker）
docker-compose logs -f app

# 实时监控日志文件（类似 tail -f）
docker-compose exec app tail -f /app/logs/app.log

# 查看全部日志
docker-compose exec app cat /app/logs/app.log

# 复制到宿主机分析
docker cp enterprise-app:/app/logs/app.log ./app.log
```

### 日志分析

```bash
# 查看所有错误
docker-compose exec app grep "error" /app/logs/app.log

# 查看缓存命中率
docker-compose exec app sh -c 'echo "HITS: $(grep -c cache_hit /app/logs/app.log), MISS: $(grep -c cache_miss /app/logs/app.log)"'

# 找出最慢的请求（前5个）
docker-compose exec app sh -c 'grep "request completed" /app/logs/app.log | sort -t" " -k12 -rn | head -5'

# 查看所有 4xx/5xx 错误请求
docker-compose exec app grep -E '"status":[45]' /app/logs/app.log
```

## 配置

支持 YAML 文件 + 环境变量覆盖：

```bash
# 环境变量覆盖示例
export MYSQL_DSN="user:pass@tcp(host:3306)/db?charset=utf8mb4"
export REDIS_ADDR="redis:6379"
export GIN_MODE="release"
export LOG_LEVEL="debug"
export LOG_FILE_PATH="/app/logs/app.log"
```
