# Redis Package

这是一个基于 go-redis/v9 的高性能 Redis 缓存包，提供了连接管理、缓存服务、监控、健康检查和键管理功能。

## 特性

- **高性能连接池**: 优化的 Redis 连接池配置，支持单机和集群模式
- **完整缓存接口**: 支持所有 Redis 数据类型和操作
- **实时监控**: 操作指标、性能统计、时间窗口分析
- **健康检查**: 全面的 Redis 健康检查，包括连接、内存、性能等
- **键管理**: 统一的缓存键命名规范和过期策略
- **错误处理**: 完善的错误分类和处理机制
- **分布式锁**: 内置分布式锁支持
- **缓存模式**: 支持 GetOrSet、Remember 等缓存模式

## 快速开始

### 1. 配置 Redis

```go
cfg := &config.RedisConfig{
    Host:            "localhost",
    Port:            6379,
    Password:        "",
    Database:        0,
    PoolSize:        10,
    MinIdleConns:    5,
    MaxRetries:      3,
    DialTimeout:     5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
    PoolTimeout:     4 * time.Second,
    IdleTimeout:     5 * time.Minute,
    MaxConnAge:      30 * time.Minute,
}
```

### 2. 初始化 Redis 包

```go
// 初始化
logger := logrus.New()
if err := redis.Initialize(cfg, logger); err != nil {
    log.Fatal(err)
}
defer redis.Shutdown()

// 等待健康状态
if err := redis.WaitForHealthy(30 * time.Second); err != nil {
    log.Printf("Redis not healthy: %v", err)
}
```

### 3. 使用缓存服务

```go
// 获取缓存服务
cache := redis.GetCacheService()

// 基础操作
err := cache.Set(ctx, "key", "value", time.Hour)
value, err := cache.Get(ctx, "key")

// JSON 操作
user := &User{ID: 1, Name: "John"}
err = cache.SetJSON(ctx, redis.UserKey(1), user, time.Hour)

var loadedUser User
err = cache.GetJSON(ctx, redis.UserKey(1), &loadedUser)

// 缓存模式
result, err := cache.GetOrSet(ctx, "expensive_key", time.Hour, func() (interface{}, error) {
    // 执行昂贵的操作
    return computeExpensiveValue(), nil
})
```

## API 文档

### 初始化函数

#### `Initialize(cfg *config.RedisConfig, logger *logrus.Logger) error`
初始化 Redis 包，创建连接池和监控器。

#### `Shutdown() error`
关闭 Redis 连接和清理资源。

#### `TestConnection(cfg *config.RedisConfig) error`
测试 Redis 连接（不初始化全局实例）。

### 缓存服务

#### 基础操作
- `Get(ctx, key) (string, error)` - 获取字符串值
- `Set(ctx, key, value, expiration) error` - 设置字符串值
- `Delete(ctx, key) error` - 删除键
- `Exists(ctx, key) (bool, error)` - 检查键是否存在

#### 批量操作
- `MGet(ctx, keys...) ([]interface{}, error)` - 批量获取
- `MSet(ctx, pairs...) error` - 批量设置
- `MDelete(ctx, keys...) error` - 批量删除

#### JSON 操作
- `GetJSON(ctx, key, dest) error` - 获取 JSON 对象
- `SetJSON(ctx, key, value, expiration) error` - 设置 JSON 对象

#### 哈希操作
- `HGet(ctx, key, field) (string, error)` - 获取哈希字段
- `HSet(ctx, key, values...) error` - 设置哈希字段
- `HGetAll(ctx, key) (map[string]string, error)` - 获取所有哈希字段

#### 列表操作
- `LPush(ctx, key, values...) error` - 左侧推入
- `RPush(ctx, key, values...) error` - 右侧推入
- `LPop(ctx, key) (string, error)` - 左侧弹出
- `RPop(ctx, key) (string, error)` - 右侧弹出
- `LRange(ctx, key, start, stop) ([]string, error)` - 范围获取

#### 集合操作
- `SAdd(ctx, key, members...) error` - 添加成员
- `SMembers(ctx, key) ([]string, error)` - 获取所有成员
- `SIsMember(ctx, key, member) (bool, error)` - 检查成员存在

#### 有序集合操作
- `ZAdd(ctx, key, members...) error` - 添加有序成员
- `ZRange(ctx, key, start, stop) ([]string, error)` - 范围获取
- `ZRangeWithScores(ctx, key, start, stop) ([]redis.Z, error)` - 带分数范围获取

#### 高级操作
- `Increment(ctx, key) (int64, error)` - 递增
- `IncrementBy(ctx, key, value) (int64, error)` - 按值递增
- `Lock(ctx, key, expiration) (bool, error)` - 获取分布式锁
- `Unlock(ctx, key) error` - 释放分布式锁

#### 缓存模式
- `GetOrSet(ctx, key, expiration, fn) (interface{}, error)` - 获取或设置
- `Remember(ctx, key, expiration, fn) (interface{}, error)` - 记忆缓存

### 键管理

#### 用户相关键
```go
redis.UserKey(userID)                    // user:123
redis.UserSessionKey(sessionID)          // user:session:abc123
redis.UserProfileKey(userID)             // user:profile:123
redis.UserStatsKey(userID, tournament)   // user:stats:123:SPRING
```

#### 比赛相关键
```go
redis.MatchKey(matchID)                  // match:456
redis.MatchListKey(tournament, status)   // match:list:SPRING:UPCOMING
redis.MatchStatsKey(matchID)             // match:stats:456
```

#### 预测相关键
```go
redis.PredictionKey(predictionID)        // prediction:789
redis.PredictionListKey(matchID, sortBy) // prediction:list:456:votes
redis.UserPredictionKey(userID, matchID) // prediction:user:123:456
```

#### 排行榜相关键
```go
redis.LeaderboardKey(tournament)         // leaderboard:SPRING
redis.LeaderboardTopKey(tournament, 10)  // leaderboard:top:SPRING:10
```

### 监控和健康检查

#### 获取指标
```go
client := redis.GetClient()
metrics := client.GetMetrics()
summary := metrics.GetSummary()

// 获取当前 QPS
qps := metrics.GetCurrentQPS()

// 获取健康评分
score := metrics.GetHealthScore()
```

#### 健康检查
```go
checker := redis.GetHealthChecker()

// 快速检查
result := checker.QuickCheck(ctx)

// 完整检查
result = checker.Check(ctx)

// 检查是否健康
healthy := checker.IsHealthy()
```

## 配置选项

### Redis 配置

```go
type RedisConfig struct {
    Host            string        // Redis 主机
    Port            int           // Redis 端口
    Password        string        // 密码
    Database        int           // 数据库编号
    PoolSize        int           // 连接池大小
    MinIdleConns    int           // 最小空闲连接数
    MaxRetries      int           // 最大重试次数
    DialTimeout     time.Duration // 连接超时
    ReadTimeout     time.Duration // 读取超时
    WriteTimeout    time.Duration // 写入超时
    PoolTimeout     time.Duration // 连接池超时
    IdleTimeout     time.Duration // 空闲超时
    MaxConnAge      time.Duration // 连接最大生存时间
    Cluster         ClusterConfig // 集群配置
}
```

### 推荐配置

对于 2C4G 服务器的推荐配置：

```go
cfg := &config.RedisConfig{
    Host:            "localhost",
    Port:            6379,
    PoolSize:        10,  // 连接池大小
    MinIdleConns:    5,   // 最小空闲连接
    MaxRetries:      3,   // 重试次数
    DialTimeout:     5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
    PoolTimeout:     4 * time.Second,
    IdleTimeout:     5 * time.Minute,
    MaxConnAge:      30 * time.Minute,
}
```

## 缓存策略

### 预定义过期时间
```go
redis.ExpirationShort      // 5分钟
redis.ExpirationMedium     // 30分钟
redis.ExpirationLong       // 2小时
redis.ExpirationDaily      // 24小时
redis.ExpirationWeekly     // 7天
redis.ExpirationLeaderboard // 5分钟（排行榜）
redis.ExpirationMatchData   // 1分钟（比赛数据）
```

### 缓存策略配置
```go
// 排行榜缓存（5分钟过期）
strategy := redis.LeaderboardCacheStrategy

// 比赛数据缓存（1分钟过期）
strategy = redis.MatchDataCacheStrategy

// 用户资料缓存（30分钟过期）
strategy = redis.UserProfileCacheStrategy
```

## 监控指标

### 操作指标
- 总操作数
- 成功/失败操作数
- 操作延迟统计
- 操作类型分布

### 缓存指标
- 缓存命中数/未命中数
- 缓存命中率
- 缓存操作延迟

### 连接池指标
- 总连接数/空闲连接数
- 连接池使用率
- 连接超时次数

### 健康评分
系统会根据错误率、缓存命中率、延迟等计算健康评分（0-100）。

## 错误处理

### 错误类型
```go
// 基础错误
redis.ErrKeyNotFound
redis.ErrKeyExists
redis.ErrConnectionFailed
redis.ErrOperationTimeout

// 检查错误类型
if redis.IsConnectionError(err) {
    // 处理连接错误
}

if redis.IsRetryableError(err) {
    // 可重试错误
}
```

### 错误包装
```go
err := redis.WrapError("get", "user:123", originalErr)
err = redis.WrapErrorWithCode("set", "user:123", "TIMEOUT", originalErr)
```

## 性能优化

### 1. 连接池优化
```go
// 根据服务器资源调整连接数
PoolSize: 10,      // 2C4G 服务器推荐值
MinIdleConns: 5,   // 保持适量空闲连接
```

### 2. 批量操作
```go
// 使用批量操作提高性能
cache.MSet(ctx, "key1", "value1", "key2", "value2")
values, err := cache.MGet(ctx, "key1", "key2", "key3")
```

### 3. 管道操作
```go
// 使用原始客户端进行管道操作
client := redis.GetClient()
pipe := client.GetRedisClient().Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
_, err := pipe.Exec(ctx)
```

### 4. 缓存预热
```go
// 预热关键数据
cache.SetJSON(ctx, redis.LeaderboardKey("SPRING"), leaderboardData, redis.ExpirationLeaderboard)
```

## 测试

运行测试：

```bash
# 运行所有测试
go test ./pkg/redis -v

# 运行基准测试
go test ./pkg/redis -bench=. -benchmem

# 运行特定测试
go test ./pkg/redis -run TestCacheService -v
```

## 示例

### 排行榜缓存示例
```go
func GetLeaderboard(ctx context.Context, tournament string) ([]LeaderboardEntry, error) {
    cache := redis.GetCacheService()
    key := redis.LeaderboardKey(tournament)
    
    var leaderboard []LeaderboardEntry
    err := cache.GetJSON(ctx, key, &leaderboard)
    if err == nil {
        return leaderboard, nil
    }
    
    if err != redis.ErrKeyNotFound {
        return nil, err
    }
    
    // 从数据库获取数据
    leaderboard, err = fetchLeaderboardFromDB(tournament)
    if err != nil {
        return nil, err
    }
    
    // 缓存数据
    cache.SetJSON(ctx, key, leaderboard, redis.ExpirationLeaderboard)
    
    return leaderboard, nil
}
```

### 分布式锁示例
```go
func UpdateUserStats(ctx context.Context, userID uint) error {
    cache := redis.GetCacheService()
    lockKey := fmt.Sprintf("user_stats_update:%d", userID)
    
    // 获取锁
    acquired, err := cache.Lock(ctx, lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    if !acquired {
        return errors.New("failed to acquire lock")
    }
    defer cache.Unlock(ctx, lockKey)
    
    // 执行更新操作
    return updateUserStatsInDB(userID)
}
```

### 缓存模式示例
```go
func GetUserProfile(ctx context.Context, userID uint) (*UserProfile, error) {
    cache := redis.GetCacheService()
    key := redis.UserProfileKey(userID)
    
    result, err := cache.GetOrSet(ctx, key, redis.ExpirationMedium, func() (interface{}, error) {
        return fetchUserProfileFromDB(userID)
    })
    if err != nil {
        return nil, err
    }
    
    return result.(*UserProfile), nil
}
```

## 最佳实践

1. **键命名**: 使用包提供的键生成函数，保持命名一致性
2. **过期时间**: 根据数据特性选择合适的过期时间
3. **批量操作**: 尽可能使用批量操作提高性能
4. **错误处理**: 区分不同类型的错误，实现合适的重试策略
5. **监控告警**: 监控关键指标并设置告警
6. **连接管理**: 使用全局实例，避免创建多个连接
7. **资源清理**: 应用退出时调用 `Shutdown()` 清理资源

## 故障排除

### 连接问题
1. 检查 Redis 服务是否运行
2. 验证连接配置（主机、端口、密码）
3. 检查网络连接和防火墙设置
4. 查看连接池状态

### 性能问题
1. 查看操作延迟统计
2. 检查缓存命中率
3. 监控连接池使用率
4. 分析慢操作日志

### 内存问题
1. 检查 Redis 内存使用情况
2. 调整过期时间策略
3. 清理无用的键
4. 监控内存碎片率

## 依赖

- [go-redis/v9](https://github.com/redis/go-redis) - Redis 客户端
- [logrus](https://github.com/sirupsen/logrus) - 日志库
- [backend-go/internal/config](../internal/config) - 配置管理