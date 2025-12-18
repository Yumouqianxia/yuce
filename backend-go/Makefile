# Makefile for backend-go

.PHONY: help build test clean dev docker-build docker-up docker-down lint fmt deps docs

# 默认目标
.DEFAULT_GOAL := help

# 项目变量
PROJECT_NAME := backend-go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 构建标志
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)
BUILD_FLAGS := -ldflags "$(LDFLAGS) -w -s"

# 二进制文件输出目录
BIN_DIR := bin

help: ## 显示帮助信息
	@echo "可用的命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## 安装依赖
	@echo "安装 Go 依赖..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "安装开发工具..."
	$(GOCMD) install github.com/cosmtrek/air@latest
	$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@echo "安装 golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
	fi

fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOCMD) fmt ./...
	goimports -w .

lint: ## 运行代码检查
	@echo "运行静态检查..."
	staticcheck ./...
	@echo "运行 golangci-lint..."
	golangci-lint run --config .golangci.yml

test: ## 运行测试
	@echo "运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-short: ## 运行短测试
	@echo "运行短测试..."
	$(GOTEST) -short ./...

bench: ## 运行基准测试
	@echo "运行基准测试..."
	$(GOTEST) -bench=. -benchmem ./...

build: clean ## 构建所有服务
	@echo "构建项目..."
	@mkdir -p $(BIN_DIR)
	@echo "构建 API 服务..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/api cmd/api/main.go
	@echo "构建 WebSocket 服务..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/websocket cmd/websocket/main.go
	@echo "构建 Worker 服务..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/worker cmd/worker/main.go
	@echo "构建完成！"

build-api: ## 构建 API 服务
	@echo "构建 API 服务..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/api cmd/api/main.go

build-websocket: ## 构建 WebSocket 服务
	@echo "构建 WebSocket 服务..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/websocket cmd/websocket/main.go

build-worker: ## 构建 Worker 服务
	@echo "构建 Worker 服务..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/worker cmd/worker/main.go

clean: ## 清理构建文件
	@echo "清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

docs: ## 生成 API 文档
	@echo "生成 Swagger 文档..."
	swag init -g cmd/api/main.go -o docs

dev: ## 启动开发环境
	@echo "启动开发环境..."
	./scripts/dev.sh start

dev-setup: ## 设置开发环境
	@echo "设置开发环境..."
	./scripts/dev.sh setup

dev-stop: ## 停止开发环境
	@echo "停止开发环境..."
	./scripts/dev.sh stop

dev-restart: ## 重启开发环境
	@echo "重启开发环境..."
	./scripts/dev.sh restart

dev-logs: ## 查看开发环境日志
	@echo "查看开发环境日志..."
	./scripts/dev.sh logs

dev-clean: ## 清理开发环境
	@echo "清理开发环境..."
	./scripts/dev.sh clean

run-api: ## 运行 API 服务
	@echo "运行 API 服务..."
	$(GOCMD) run cmd/api/main.go

run-websocket: ## 运行 WebSocket 服务
	@echo "运行 WebSocket 服务..."
	$(GOCMD) run cmd/websocket/main.go

run-worker: ## 运行 Worker 服务
	@echo "运行 Worker 服务..."
	$(GOCMD) run cmd/worker/main.go

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(PROJECT_NAME):$(VERSION) .
	docker build -t $(PROJECT_NAME):latest .

docker-up: ## 启动 Docker 服务
	@echo "启动 Docker 服务..."
	docker-compose up -d

docker-up-dev: ## 启动开发 Docker 服务
	@echo "启动开发 Docker 服务..."
	docker-compose -f docker-compose.dev.yml up -d

docker-down: ## 停止 Docker 服务
	@echo "停止 Docker 服务..."
	docker-compose down

docker-down-dev: ## 停止开发 Docker 服务
	@echo "停止开发 Docker 服务..."
	docker-compose -f docker-compose.dev.yml down

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

docker-restart: ## 重启 Docker 服务
	@echo "重启 Docker 服务..."
	docker-compose restart

install: build ## 安装二进制文件到 GOPATH/bin
	@echo "安装二进制文件..."
	cp $(BIN_DIR)/api $(GOPATH)/bin/
	cp $(BIN_DIR)/websocket $(GOPATH)/bin/
	cp $(BIN_DIR)/worker $(GOPATH)/bin/

check: fmt lint test ## 运行所有检查

pre-commit: ## 运行 pre-commit 检查
	@echo "运行 pre-commit 检查..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "pre-commit 未安装，跳过检查"; \
	fi

fix: ## 自动修复代码格式问题
	@echo "自动修复代码格式..."
	$(GOCMD) fmt ./...
	goimports -w .
	golangci-lint run --fix --config .golangci.yml

quality: fmt lint ## 运行代码质量检查

ci: deps check build ## CI 流水线

release: clean check build docs ## 发布版本

.PHONY: all
all: deps fmt lint test build docs ## 执行所有任务