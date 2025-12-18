# 预测系统后端 - Go 版本

基于 Go + MySQL + Redis 技术栈的高性能预测系统后端。

## 技术栈

- **后端框架**: Go + Gin
- **数据库**: MySQL 8.0 + GORM
- **缓存**: Redis 6.0
- **实时通信**: WebSocket (Gorilla)
- **认证**: JWT + bcrypt
- **配置管理**: Viper
- **日志**: Logrus

## 项目结构

```
backend-go/
├── cmd/                     # 应用程序入口
│   ├── api/                 # API 服务
│   ├── websocket/           # WebSocket 服务
│   └── worker/              # 后台任务服务
├── internal/                # 私有应用代码
│   ├── core/                # 核心业务逻辑
│   │   ├── domain/          # 领域模型
│   │   └── ports/           # 端口（接口）
│   ├── adapters/            # 适配器实现
│   │   ├── http/            # HTTP 适配器
│   │   ├── persistence/     # 持久化适配器
│   │   └── external/        # 外部服务适配器
│   ├── config/              # 配置管理
│   └── shared/              # 共享组件
├── pkg/                     # 公共库代码
├── api/                     # API 定义
├── migrations/              # 数据库迁移
├── scripts/                 # 脚本文件
├── deployments/             # 部署配置
├── tests/                   # 测试文件
├── docs/                    # 文档
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## 开发环境要求

### 1. 安装 Go

请先安装 Go 1.21 或更高版本：

**Windows:**

1. 访问 https://golang.org/dl/
2. 下载 Windows 安装包
3. 运行安装程序
4. 重启命令行工具

**验证安装:**

```bash
go version
```

### 2. 安装依赖工具

```bash
# 安装代码格式化和静态分析工具
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装热重载工具
go install github.com/cosmtrek/air@latest

# 安装 Swagger 文档生成工具
go install github.com/swaggo/swag/cmd/swag@latest
```

### 3. 初始化项目

```bash
# 进入项目目录
cd backend-go

# 初始化 Go 模块
go mod init backend-go

# 安装依赖
go mod tidy
```

## 开发工具配置

### VS Code 扩展

- Go (官方 Go 扩展)
- Go Test Explorer
- REST Client

### 代码质量工具

- `gofmt`: 代码格式化
- `staticcheck`: 静态分析 (替代已弃用的 golint)
- `golangci-lint`: 综合代码检查
- `go vet`: 代码检查

## 性能目标

- API 响应时间提升 150%+
- 支持 2000+ 并发连接
- 内存占用 < 500MB
- 启动时间 < 5 秒

## 开发指南

1. 遵循 Go 代码规范和最佳实践
2. 使用六边形架构模式
3. 编写单元测试和集成测试
4. 使用结构化日志
5. 实现优雅关闭

## 部署

支持 Docker 容器化部署，详见 `deployments/` 目录。
