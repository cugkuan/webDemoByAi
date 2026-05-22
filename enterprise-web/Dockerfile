# ============================================================
# Stage 1: Build
# ============================================================
FROM node:22-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY package.json package-lock.json ./
RUN npm ci

# 复制源码并构建
COPY . .
RUN npm run build

# ============================================================
# Stage 2: Nginx 运行
# ============================================================
FROM nginx:1.27-alpine

# 设置时区
ENV TZ=Asia/Shanghai

# 复制构建产物到 Nginx 静态目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 健康检查（使用 curl 替代 wget，避免 Alpine 中 localhost 解析为 IPv6 的问题）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -sf http://127.0.0.1:80/ > /dev/null || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
