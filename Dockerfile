# ==================== Build Stage ====================
FROM node:20-alpine AS builder

# 安装pnpm
RUN npm install -g pnpm

# 设置工作目录
WORKDIR /app

# 复制package文件和pnpm相关文件
COPY package*.json pnpm-lock.yaml pnpm-workspace.yaml ./

# 安装依赖（使用缓存挂载优化）
RUN --mount=type=cache,target=/root/.local/share/pnpm \
    pnpm install --frozen-lockfile --prod=false

# 复制源代码
COPY . .

# 构建应用（跳过类型检查确保先可用）
RUN pnpm run build:prod

# ==================== Production Stage ====================
FROM nginx:1.25-alpine AS production

# 安装curl用于健康检查
# 切换国内镜像并安装 curl
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache curl

# 复制自定义nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 从构建阶段复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# nginx 基础镜像已包含 nginx 用户与组，这里仅确保目录权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d

# 创建 nginx 运行时需要的目录
RUN touch /var/run/nginx.pid && chown -R nginx:nginx /var/run/nginx.pid

# 切换到非root用户
USER nginx

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]