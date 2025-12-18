#!/bin/bash

# 开发环境启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    echo -e "${GREEN}开发环境脚本使用说明${NC}"
    echo -e "用法: $0 [选项]"
    echo -e ""
    echo -e "选项:"
    echo -e "  ${YELLOW}setup${NC}     - 安装开发工具和依赖"
    echo -e "  ${YELLOW}start${NC}     - 启动开发环境（默认）"
    echo -e "  ${YELLOW}stop${NC}      - 停止开发环境"
    echo -e "  ${YELLOW}restart${NC}   - 重启开发环境"
    echo -e "  ${YELLOW}logs${NC}      - 查看服务日志"
    echo -e "  ${YELLOW}test${NC}      - 运行测试"
    echo -e "  ${YELLOW}lint${NC}      - 运行代码检查"
    echo -e "  ${YELLOW}clean${NC}     - 清理开发环境"
    echo -e "  ${YELLOW}help${NC}      - 显示此帮助信息"
    echo -e ""
    echo -e "示例:"
    echo -e "  $0 setup    # 首次设置开发环境"
    echo -e "  $0 start    # 启动开发环境"
    echo -e "  $0 test     # 运行测试"
}

# 检查必要的工具
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}$1 未安装，请先安装${NC}"
        return 1
    fi
    return 0
}

# 检查所有必要工具
check_prerequisites() {
    echo -e "${BLUE}检查必要工具...${NC}"
    
    local missing_tools=()
    
    if ! check_tool go; then
        missing_tools+=("go")
    fi
    
    if ! check_tool docker; then
        missing_tools+=("docker")
    fi
    
    if ! check_tool docker-compose; then
        missing_tools+=("docker-compose")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo -e "${RED}缺少以下工具: ${missing_tools[*]}${NC}"
        echo -e "${YELLOW}请先安装这些工具，然后重新运行脚本${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}所有必要工具已安装${NC}"
}

# 安装开发工具
install_dev_tools() {
    echo -e "${BLUE}安装开发工具...${NC}"
    
    # 安装 Air 热重载工具
    if ! command -v air &> /dev/null; then
        echo -e "${YELLOW}安装 Air 热重载工具...${NC}"
        go install github.com/cosmtrek/air@latest
    else
        echo -e "${GREEN}Air 已安装${NC}"
    fi
    
    # 安装 staticcheck
    if ! command -v staticcheck &> /dev/null; then
        echo -e "${YELLOW}安装 staticcheck...${NC}"
        go install honnef.co/go/tools/cmd/staticcheck@latest
    else
        echo -e "${GREEN}staticcheck 已安装${NC}"
    fi
    
    # 安装 golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "${YELLOW}安装 golangci-lint...${NC}"
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    else
        echo -e "${GREEN}golangci-lint 已安装${NC}"
    fi
    
    # 安装 Swagger 工具
    if ! command -v swag &> /dev/null; then
        echo -e "${YELLOW}安装 Swagger 工具...${NC}"
        go install github.com/swaggo/swag/cmd/swag@latest
    else
        echo -e "${GREEN}swag 已安装${NC}"
    fi
    
    # 安装 goimports
    if ! command -v goimports &> /dev/null; then
        echo -e "${YELLOW}安装 goimports...${NC}"
        go install golang.org/x/tools/cmd/goimports@latest
    else
        echo -e "${GREEN}goimports 已安装${NC}"
    fi
    
    # 安装 pre-commit (可选)
    if command -v pip3 &> /dev/null; then
        if ! command -v pre-commit &> /dev/null; then
            echo -e "${YELLOW}安装 pre-commit...${NC}"
            pip3 install pre-commit
            pre-commit install
        else
            echo -e "${GREEN}pre-commit 已安装${NC}"
        fi
    fi
}

# 启动开发服务
start_services() {
    echo -e "${BLUE}启动开发服务...${NC}"
    
    # 启动 Docker 服务
    echo -e "${YELLOW}启动 MySQL 和 Redis...${NC}"
    docker-compose -f docker-compose.dev.yml up -d
    
    # 等待服务启动
    echo -e "${YELLOW}等待服务启动...${NC}"
    sleep 10
    
    # 检查 MySQL 连接
    echo -e "${YELLOW}检查 MySQL 连接...${NC}"
    local mysql_ready=false
    for i in {1..30}; do
        if docker-compose -f docker-compose.dev.yml exec -T mysql-dev mysqladmin ping -h"localhost" --silent 2>/dev/null; then
            mysql_ready=true
            break
        fi
        echo -e "${YELLOW}等待 MySQL 启动... ($i/30)${NC}"
        sleep 2
    done
    
    if [ "$mysql_ready" = true ]; then
        echo -e "${GREEN}MySQL 已启动${NC}"
    else
        echo -e "${RED}MySQL 启动超时${NC}"
        exit 1
    fi
    
    # 检查 Redis 连接
    echo -e "${YELLOW}检查 Redis 连接...${NC}"
    local redis_ready=false
    for i in {1..15}; do
        if docker-compose -f docker-compose.dev.yml exec -T redis-dev redis-cli ping 2>/dev/null | grep -q PONG; then
            redis_ready=true
            break
        fi
        echo -e "${YELLOW}等待 Redis 启动... ($i/15)${NC}"
        sleep 2
    done
    
    if [ "$redis_ready" = true ]; then
        echo -e "${GREEN}Redis 已启动${NC}"
    else
        echo -e "${RED}Redis 启动超时${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}所有服务已启动${NC}"
    echo -e "${BLUE}管理工具访问地址:${NC}"
    echo -e "  - phpMyAdmin: ${YELLOW}http://localhost:8081${NC}"
    echo -e "  - Redis Commander: ${YELLOW}http://localhost:8082${NC}"
}

# 停止开发服务
stop_services() {
    echo -e "${BLUE}停止开发服务...${NC}"
    docker-compose -f docker-compose.dev.yml down
    echo -e "${GREEN}服务已停止${NC}"
}

# 重启开发服务
restart_services() {
    echo -e "${BLUE}重启开发服务...${NC}"
    stop_services
    start_services
}

# 查看服务日志
show_logs() {
    echo -e "${BLUE}查看服务日志...${NC}"
    docker-compose -f docker-compose.dev.yml logs -f
}

# 运行测试
run_tests() {
    echo -e "${BLUE}运行测试...${NC}"
    
    # 设置测试环境变量
    export BACKEND_DATABASE_HOST=localhost
    export BACKEND_DATABASE_PORT=3306
    export BACKEND_DATABASE_USERNAME=root
    export BACKEND_DATABASE_PASSWORD=123456
    export BACKEND_DATABASE_DATABASE=prediction_system_test
    export BACKEND_REDIS_HOST=localhost
    export BACKEND_REDIS_PORT=6379
    export BACKEND_REDIS_DATABASE=1
    export BACKEND_AUTH_JWT_SECRET=test-secret-key
    export BACKEND_LOG_LEVEL=debug
    
    # 运行测试
    go test -v -race -coverprofile=coverage.out ./...
    
    # 生成覆盖率报告
    if [ -f coverage.out ]; then
        go tool cover -html=coverage.out -o coverage.html
        echo -e "${GREEN}测试覆盖率报告已生成: coverage.html${NC}"
    fi
}

# 运行代码检查
run_lint() {
    echo -e "${BLUE}运行代码检查...${NC}"
    
    # 格式化代码
    echo -e "${YELLOW}格式化代码...${NC}"
    go fmt ./...
    goimports -w .
    
    # 运行 staticcheck
    echo -e "${YELLOW}运行 staticcheck...${NC}"
    staticcheck ./...
    
    # 运行 golangci-lint
    echo -e "${YELLOW}运行 golangci-lint...${NC}"
    golangci-lint run --config .golangci.yml
    
    echo -e "${GREEN}代码检查完成${NC}"
}

# 清理开发环境
clean_env() {
    echo -e "${BLUE}清理开发环境...${NC}"
    
    # 停止服务
    stop_services
    
    # 清理 Docker 资源
    echo -e "${YELLOW}清理 Docker 资源...${NC}"
    docker-compose -f docker-compose.dev.yml down -v --remove-orphans
    
    # 清理构建文件
    echo -e "${YELLOW}清理构建文件...${NC}"
    rm -rf bin/
    rm -f coverage.out coverage.html
    rm -rf tmp/
    
    echo -e "${GREEN}清理完成${NC}"
}

# 设置开发环境
setup_env() {
    echo -e "${GREEN}设置开发环境...${NC}"
    
    check_prerequisites
    
    # 安装 Go 依赖
    echo -e "${YELLOW}安装 Go 依赖...${NC}"
    go mod tidy
    go mod verify
    
    install_dev_tools
    
    # 生成 Swagger 文档
    echo -e "${YELLOW}生成 Swagger 文档...${NC}"
    swag init -g cmd/api/main.go -o docs
    
    echo -e "${GREEN}开发环境设置完成！${NC}"
}

# 启动开发环境
start_dev() {
    echo -e "${GREEN}启动开发环境...${NC}"
    
    check_prerequisites
    start_services
    
    # 设置环境变量
    export BACKEND_DATABASE_HOST=localhost
    export BACKEND_DATABASE_PORT=3306
    export BACKEND_DATABASE_USERNAME=root
    export BACKEND_DATABASE_PASSWORD=123456
    export BACKEND_DATABASE_DATABASE=prediction_system_dev
    export BACKEND_REDIS_HOST=localhost
    export BACKEND_REDIS_PORT=6379
    export BACKEND_LOG_LEVEL=debug
    
    echo -e "${GREEN}开发环境已准备就绪！${NC}"
    echo -e "${GREEN}可用的命令:${NC}"
    echo -e "  - ${YELLOW}air${NC}: 启动 API 服务（热重载）"
    echo -e "  - ${YELLOW}go run cmd/api/main.go${NC}: 直接运行 API 服务"
    echo -e "  - ${YELLOW}go run cmd/websocket/main.go${NC}: 运行 WebSocket 服务"
    echo -e "  - ${YELLOW}go run cmd/worker/main.go${NC}: 运行 Worker 服务"
    echo -e "  - ${YELLOW}make test${NC}: 运行测试"
    echo -e "  - ${YELLOW}make lint${NC}: 运行代码检查"
    echo -e ""
    echo -e "${BLUE}管理工具:${NC}"
    echo -e "  - phpMyAdmin: ${YELLOW}http://localhost:8081${NC}"
    echo -e "  - Redis Commander: ${YELLOW}http://localhost:8082${NC}"
    
    # 询问是否启动 Air 热重载
    echo -e ""
    read -p "是否启动 Air 热重载服务 (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}启动 API 服务（热重载）...${NC}"
        air -c .air.toml
    fi
}

# 主函数
main() {
    case "${1:-start}" in
        setup)
            setup_env
            ;;
        start)
            start_dev
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        logs)
            show_logs
            ;;
        test)
            run_tests
            ;;
        lint)
            run_lint
            ;;
        clean)
            clean_env
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"