# ==================== Build Stage ====================
# 支持自定义基础镜像（如 1Panel 镜像仓库），默认使用官方
ARG BASE_GOLANG_IMAGE=golang:1.24-alpine
# 供生产阶段使用的基础镜像参数，需在首次 FROM 之前声明以兼容部分 builder
ARG BASE_ALPINE_IMAGE=alpine:3.19
FROM ${BASE_GOLANG_IMAGE} AS builder

# 可选：在 build 时通过 --build-arg HTTP_PROXY/HTTPS_PROXY/NO_PROXY 走本地代理
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY
ENV http_proxy=${HTTP_PROXY} \
    https_proxy=${HTTPS_PROXY} \
    no_proxy=${NO_PROXY} \
    GOPROXY=https://goproxy.cn,direct

# 使用国内镜像加速 Alpine 软件源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置环境变量
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 可选：构建阶段网络代理（由 compose build args 传入）

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN go build -ldflags="-s -w" -o /app/main ./cmd/api

# ==================== Test Stage ====================
FROM builder AS test-stage

# 运行测试
RUN go test -v ./...

# ==================== Production Stage ====================
# 同样支持自定义运行时基础镜像
ARG BASE_ALPINE_IMAGE=alpine:3.19
FROM ${BASE_ALPINE_IMAGE} AS production

# 使用国内镜像加速 Alpine 软件源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main /app/main

# 复制配置文件
COPY --chown=appuser:appgroup configs/ /app/configs/
COPY --chown=appuser:appgroup migrations/ /app/migrations/

# 创建必要的目录
RUN mkdir -p /app/logs /app/uploads && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
ENTRYPOINT ["/app/main"]