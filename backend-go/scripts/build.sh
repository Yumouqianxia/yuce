#!/bin/bash

# 构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="backend-go"
VERSION=${VERSION:-"latest"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}开始构建 ${PROJECT_NAME}...${NC}"

# 检查 Go 版本
echo -e "${YELLOW}检查 Go 版本...${NC}"
go version

# 检查依赖
echo -e "${YELLOW}检查依赖...${NC}"
go mod tidy
go mod verify

# 代码格式化
echo -e "${YELLOW}格式化代码...${NC}"
go fmt ./...

# 静态检查
echo -e "${YELLOW}运行静态检查...${NC}"
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
else
    echo -e "${RED}staticcheck 未安装，跳过静态检查${NC}"
fi

# 运行 golangci-lint
echo -e "${YELLOW}运行 golangci-lint...${NC}"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
else
    echo -e "${RED}golangci-lint 未安装，跳过 lint 检查${NC}"
fi

# 运行测试
echo -e "${YELLOW}运行测试...${NC}"
go test -v -race -coverprofile=coverage.out ./...

# 生成测试覆盖率报告
if [ -f coverage.out ]; then
    echo -e "${YELLOW}生成测试覆盖率报告...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}测试覆盖率报告已生成: coverage.html${NC}"
fi

# 构建二进制文件
echo -e "${YELLOW}构建二进制文件...${NC}"

# 设置构建标志
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# 构建 API 服务
echo -e "${YELLOW}构建 API 服务...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="${LDFLAGS} -w -s" \
    -o bin/api \
    cmd/api/main.go

# 构建 WebSocket 服务
echo -e "${YELLOW}构建 WebSocket 服务...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="${LDFLAGS} -w -s" \
    -o bin/websocket \
    cmd/websocket/main.go

# 构建 Worker 服务
echo -e "${YELLOW}构建 Worker 服务...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="${LDFLAGS} -w -s" \
    -o bin/worker \
    cmd/worker/main.go

# 生成 Swagger 文档
echo -e "${YELLOW}生成 Swagger 文档...${NC}"
if command -v swag &> /dev/null; then
    swag init -g cmd/api/main.go -o docs
    echo -e "${GREEN}Swagger 文档已生成${NC}"
else
    echo -e "${RED}swag 未安装，跳过文档生成${NC}"
fi

echo -e "${GREEN}构建完成！${NC}"
echo -e "${GREEN}二进制文件位置:${NC}"
echo -e "  - API 服务: bin/api"
echo -e "  - WebSocket 服务: bin/websocket"
echo -e "  - Worker 服务: bin/worker"

# 显示文件大小
if [ -f bin/api ]; then
    echo -e "${GREEN}文件大小:${NC}"
    ls -lh bin/
fi