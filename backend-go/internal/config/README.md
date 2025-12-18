# 配置管理系统

本配置管理系统基于 Viper 实现，提供了分层配置、环境变量支持、配置验证、热重载等功能。

## 功能特性

- ✅ **分层配置**: 支持默认值、配置文件、环境变量的优先级覆盖
- ✅ **多环境支持**: 开发、测试、生产环境的配置模板
- ✅ **配置验证**: 完整的配置结构验证和业务逻辑检查
- ✅ **热重载**: 配置文件变更时自动重新加载
- ✅ **功能开关**: 灵活的功能开关配置
- ✅ **外部服务**: 邮件、文件存储、监控等外部服务配置
- ✅ **安全检查**: 生产环境安全配置检查
- ✅ **配置工具**: 命令行工具支持配置管理操作

## 快速开始

### 基本使用

```go
package main

import (
    "log"
    "backend-go/internal/config"
)

func main() {
    // 加载默认配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 使用配置
    fmt.Printf("Server: %s\n", cfg.Server.GetServerAddr())
    fmt.Printf("Database: %s\n", cfg.Database.GetDSN())
}
```

### 环境特定配置

```go
// 为特定环境加载配置
cfg, err := config.LoadForEnvironment(config.EnvProduction)
if err != nil {
    log.Fatalf("Failed to load production config: %v", err)
}
```

### 自定义配置选项

```go
opts := &config.LoadOptions{
    ConfigPath:   "./custom/path",
    ConfigName:   "myconfig",
    ConfigType:   "yaml",
    EnvPrefix:    "MYAPP",
    SkipValidate: false,
}

cfg, err := config.Load(opts)
```

## 配置结构

### 服务器配置

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  mode: "release"  # debug, release, test
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
```

### 数据库配置

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: ""
  database: "prediction_system"
  charset: "utf8mb4"
  collation: "utf8mb4_unicode_ci"
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "30m"
  ssl:
    mode: "disable"  # disable, require, verify-ca, verify-full
  migration:
    enabled: true
    auto_create: false
    path: "./migrations"
```

### Redis 配置

```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  pool_size: 10
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  cluster:
    enabled: false
    addresses: []
```

### 认证配置

```yaml
auth:
  jwt_secret: "your-jwt-secret-key"
  jwt_expiration_hours: 24
  refresh_token_exp_days: 30
  bcrypt_cost: 12
  session_timeout: "24h"
  max_login_attempts: 5
  lockout_duration: "15m"
  password_policy:
    min_length: 8
    require_upper: true
    require_lower: true
    require_number: true
    require_special: false
```

### 功能开关

```yaml
features:
  enable_swagger: false
  enable_pprof: false
  enable_metrics: true
  enable_cors: true
  enable_rate_limit: true
  enable_health_check: true
  enable_graceful_shutdown: true
  cache_leaderboard: true
  cache_match_data: true
  rate_limit:
    requests_per_second: 100
    burst_size: 200
    window_size: "1m"
  cors:
    allowed_origins:
      - "http://localhost:3000"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allow_credentials: true
```

## 环境变量

配置支持通过环境变量覆盖，环境变量使用 `BACKEND_` 前缀：

```bash
# 服务器配置
export BACKEND_SERVER_HOST=0.0.0.0
export BACKEND_SERVER_PORT=8080

# 数据库配置
export BACKEND_DATABASE_HOST=localhost
export BACKEND_DATABASE_PASSWORD=secret

# 认证配置
export BACKEND_AUTH_JWT_SECRET=your-production-secret-key
```

## 配置验证

系统提供完整的配置验证功能：

```go
validator := config.NewConfigValidator()
if err := validator.Validate(cfg); err != nil {
    // 处理验证错误
    errors := config.FormatValidationErrors(err)
    for _, e := range errors {
        fmt.Printf("Field: %s, Error: %s\n", e.Field, e.Message)
    }
}
```

### 验证规则

- **必填字段**: 服务器地址、端口、数据库连接信息等
- **数值范围**: 端口号 1-65535、连接池大小等
- **格式验证**: 邮箱格式、URL 格式、时间格式等
- **业务逻辑**: 生产环境安全检查、依赖关系验证等

## 配置热重载

支持配置文件变更时自动重新加载：

```go
loader := config.NewConfigLoader()

// 加载配置
cfg, err := loader.LoadConfig(nil)
if err != nil {
    log.Fatal(err)
}

// 添加变更监听器
loader.AddWatcher(config.ConfigWatcherFunc(func(oldCfg, newCfg *config.Config) error {
    log.Printf("Config changed: %s -> %s", 
        oldCfg.Server.GetServerAddr(), 
        newCfg.Server.GetServerAddr())
    return nil
}))

// 启动文件监听
if err := loader.WatchConfig(); err != nil {
    log.Printf("Failed to start config watcher: %v", err)
}
```

## 配置管理工具

提供命令行工具进行配置管理：

```bash
# 验证配置
go run cmd/config/main.go validate configs/config.yaml

# 生成环境配置
go run cmd/config/main.go generate production

# 比较配置差异
go run cmd/config/main.go diff config.yaml config.prod.yaml

# 导出配置
go run cmd/config/main.go export json config.json

# 应用配置模板
go run cmd/config/main.go template production

# 检查配置健康状态
go run cmd/config/main.go health

# 性能分析
go run cmd/config/main.go profile
```

## 配置模板

系统提供预定义的环境配置模板：

### 开发环境模板

```yaml
server:
  mode: debug
  host: localhost
log:
  level: debug
  format: text
features:
  enable_swagger: true
  enable_pprof: true
  enable_rate_limit: false
```

### 生产环境模板

```yaml
server:
  mode: release
  host: 0.0.0.0
log:
  level: info
  format: json
features:
  enable_swagger: false
  enable_pprof: false
  enable_rate_limit: true
```

## 高级功能

### 配置比较

```go
diffs := config.CompareConfigs(oldConfig, newConfig)
for _, diff := range diffs {
    fmt.Printf("Field: %s, Old: %v, New: %v\n", 
        diff.Field, diff.OldValue, diff.NewValue)
}
```

### 配置导出/导入

```go
// 导出配置
exporter := config.NewConfigExporter(cfg)
data, err := exporter.ExportToJSON()

// 导入配置
importer := config.NewConfigImporter()
cfg, err := importer.ImportFromFile("config.json")
```

### 配置合并

```go
merger := config.NewConfigMerger()
mergedConfig, err := merger.MergeConfigs(baseConfig, overrideConfig)
```

### 性能分析

```go
profiler := config.NewConfigProfiler()
cfg, err := profiler.ProfileLoad(func() (*config.Config, error) {
    return config.Load()
})

stats := profiler.GetStats()
fmt.Printf("Load time: %v\n", stats["avg_load_time"])
```

## 安全最佳实践

### 生产环境

1. **JWT 密钥**: 至少 32 个字符的强密钥
2. **TLS 启用**: 生产环境必须启用 HTTPS
3. **调试功能**: 关闭 Swagger、pprof 等调试功能
4. **密码策略**: 强制复杂密码策略
5. **连接限制**: 合理设置数据库连接池大小

### 敏感信息

- 使用环境变量存储敏感信息
- 不要在配置文件中硬编码密码
- 定期轮换密钥和密码
- 使用配置加密（如需要）

## 故障排除

### 常见问题

1. **配置文件未找到**
   ```
   Error: config file not found
   ```
   解决：检查配置文件路径和名称

2. **环境变量格式错误**
   ```
   Error: invalid duration format
   ```
   解决：检查时间格式，如 "30s", "5m", "1h"

3. **验证失败**
   ```
   Error: validation failed
   ```
   解决：检查必填字段和数值范围

4. **权限问题**
   ```
   Error: permission denied
   ```
   解决：检查配置文件读取权限

### 调试技巧

1. 使用 `config validate` 命令检查配置
2. 查看配置摘要了解当前设置
3. 使用 `config health` 检查连接性
4. 启用调试日志查看详细信息

## 扩展开发

### 添加新配置项

1. 在 `config.go` 中添加结构体字段
2. 在 `setDefaults` 中设置默认值
3. 在验证器中添加验证规则
4. 更新配置模板
5. 添加相应的测试

### 自定义验证器

```go
func registerCustomValidators(validate *validator.Validate) error {
    return validate.RegisterValidation("custom", func(fl validator.FieldLevel) bool {
        // 自定义验证逻辑
        return true
    })
}
```

### 配置插件

```go
type ConfigPlugin interface {
    Name() string
    Load(config *Config) error
    Validate(config *Config) error
}
```

## 测试

运行配置相关测试：

```bash
# 运行所有配置测试
go test ./internal/config/...

# 运行特定测试
go test -run TestLoad ./internal/config/

# 运行基准测试
go test -bench=. ./internal/config/
```

## 性能考虑

- 配置加载时间通常在 1-5ms
- 内存占用约 1-2KB
- 支持并发安全的配置访问
- 文件监听对性能影响很小

## 版本兼容性

- Go 1.21+
- Viper v1.17+
- 向后兼容的配置格式
- 平滑的配置迁移支持