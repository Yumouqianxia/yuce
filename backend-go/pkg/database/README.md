# Database Package

这是一个基于 GORM 和 MySQL 的高性能数据库包，提供了连接管理、监控、健康检查和指标收集功能。

## 特性

- **高性能连接池**: 优化的 MySQL 连接池配置，支持最大 20 个连接
- **实时监控**: 连接池状态、查询性能、错误率等指标监控
- **健康检查**: 全面的数据库健康检查，包括连接、性能、磁盘空间等
- **指标收集**: 查询时间、事务统计、慢查询检测等
- **自动重连**: 连接断开时自动重连机制
- **性能优化**: 预编译语句缓存，查询日志记录
- **错误处理**: 统一的错误处理和日志记录

## 快速开始

### 1. 配置数据库

```go
cfg := &config.DatabaseConfig{
    Host:            "localhost",
    Port:            3306,
    Username:        "root",
    Password:        "password",
    Database:        "myapp",
    Charset:         "utf8mb4",
    Collation:       "utf8mb4_unicode_ci",
    MaxOpenConns:    20,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 30 * time.Minute,
}
```

### 2. 初始化数据库包

```go
// 初始化
if err := database.Initialize(cfg); err != nil {
    log.Fatal(err)
}
defer database.Shutdown()

// 等待健康状态
if err := database.WaitForHealthy(30 * time.Second); err != nil {
    log.Printf("Database not healthy: %v", err)
}
```

### 3. 使用数据库

```go
// 获取数据库实例
db := database.GetDB()

// 执行查询（带指标记录）
err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
    var users []User
    return db.WithContext(ctx).Find(&users).Error
})

// 执行事务
err = database.WithTransaction(ctx, func(db *database.DB) error {
    // 事务操作
    return db.WithContext(ctx).Create(&user).Error
})
```

## API 文档

### 初始化函数

#### `Initialize(cfg *config.DatabaseConfig) error`
初始化数据库包，创建连接池和监控器。

#### `Shutdown() error`
关闭数据库连接和清理资源。

#### `TestConnection(cfg *config.DatabaseConfig) error`
测试数据库连接（不初始化全局实例）。

### 数据库操作

#### `GetDB() *DB`
获取数据库实例。

#### `ExecuteWithMetrics(ctx context.Context, fn func(*DB) error) error`
执行查询并自动记录性能指标。

#### `WithTransaction(ctx context.Context, fn func(*DB) error) error`
执行数据库事务。

### 监控和健康检查

#### `HealthCheck(ctx context.Context) error`
执行数据库健康检查。

#### `GetMetrics() *Metrics`
获取数据库性能指标。

#### `GetConnectionInfo() map[string]interface{}`
获取连接池信息。

#### `GetStatus() map[string]interface{}`
获取完整的数据库状态信息。

### 健康检查器

```go
checker := database.GetHealthChecker()

// 快速检查
result := checker.QuickCheck(ctx)

// 完整检查
result = checker.Check(ctx)

// 检查是否健康
healthy := checker.IsHealthy()
```

## 配置选项

### 数据库配置

```go
type DatabaseConfig struct {
    Host            string        // 数据库主机
    Port            int           // 数据库端口
    Username        string        // 用户名
    Password        string        // 密码
    Database        string        // 数据库名
    Charset         string        // 字符集
    Collation       string        // 排序规则
    MaxOpenConns    int           // 最大打开连接数
    MaxIdleConns    int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大生存时间
    ConnMaxIdleTime time.Duration // 连接最大空闲时间
    SSL             SSLConfig     // SSL 配置
    Migration       MigrationConfig // 迁移配置
}
```

### 推荐配置

对于 2C4G 服务器的推荐配置：

```go
cfg := &config.DatabaseConfig{
    MaxOpenConns:    20,  // 最大连接数
    MaxIdleConns:    10,  // 空闲连接数
    ConnMaxLifetime: time.Hour,        // 连接生存时间
    ConnMaxIdleTime: 30 * time.Minute, // 空闲超时
}
```

## 监控指标

### 查询指标
- 总查询数
- 查询错误数
- 平均查询时间
- 最大/最小查询时间
- 慢查询数量

### 事务指标
- 总事务数
- 事务错误数
- 平均事务时间

### 连接池指标
- 打开连接数
- 使用中连接数
- 空闲连接数
- 等待时间和次数

### 健康评分
系统会根据错误率、慢查询率、连接池状态等计算健康评分（0-100）。

## 性能优化

### 1. 连接池优化
```go
// 根据服务器资源调整连接数
MaxOpenConns: 20,  // 2C4G 服务器推荐值
MaxIdleConns: 10,  // 保持适量空闲连接
```

### 2. 查询优化
```go
// 启用预编译语句缓存
PrepareStmt: true,

// 设置慢查询阈值
db.SetSlowQueryThreshold(200 * time.Millisecond)
```

### 3. 监控告警
```go
metrics := database.GetMetrics()
alerts := metrics.CheckAlerts()
for _, alert := range alerts {
    log.Printf("Alert: %s - %s", alert.Type, alert.Message)
}
```

## 错误处理

包提供了统一的错误处理机制：

```go
// 自动记录错误指标
err := database.ExecuteWithMetrics(ctx, func(db *database.DB) error {
    return db.WithContext(ctx).Create(&user).Error
})
if err != nil {
    // 错误已自动记录到指标中
    log.Printf("Database operation failed: %v", err)
}
```

## 测试

运行测试：

```bash
# 运行所有测试
go test ./pkg/database -v

# 运行基准测试
go test ./pkg/database -bench=. -benchmem

# 运行特定测试
go test ./pkg/database -run TestMetrics -v
```

## 示例

查看 `examples/database_example.go` 获取完整的使用示例。

## 最佳实践

1. **连接管理**: 使用包提供的全局实例，避免创建多个连接
2. **事务处理**: 使用 `WithTransaction` 确保事务正确处理
3. **错误处理**: 使用 `ExecuteWithMetrics` 自动记录性能指标
4. **健康检查**: 定期检查数据库健康状态
5. **监控告警**: 监控关键指标并设置告警
6. **资源清理**: 应用退出时调用 `Shutdown()` 清理资源

## 故障排除

### 连接问题
1. 检查数据库服务是否运行
2. 验证连接配置（主机、端口、用户名、密码）
3. 检查网络连接和防火墙设置

### 性能问题
1. 查看慢查询日志
2. 检查连接池使用率
3. 监控查询执行时间
4. 优化数据库索引

### 内存问题
1. 调整连接池大小
2. 检查连接泄漏
3. 监控内存使用情况

## 依赖

- [GORM](https://gorm.io/) - ORM 框架
- [MySQL Driver](https://github.com/go-sql-driver/mysql) - MySQL 驱动
- [Logrus](https://github.com/sirupsen/logrus) - 日志库