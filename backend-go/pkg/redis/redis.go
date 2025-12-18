// Package redis provides Redis client management and caching utilities.
//
// This package implements a comprehensive Redis client wrapper that supports both
// single-node and cluster configurations. It provides connection pooling, health
// monitoring, metrics collection, and a clean API for Redis operations.
//
// Key Features:
//   - Support for both single-node and cluster Redis deployments
//   - Connection pooling with configurable parameters
//   - Health monitoring and automatic reconnection
//   - Metrics collection for performance monitoring
//   - Thread-safe operations with proper synchronization
//   - Graceful shutdown and resource cleanup
//
// The package follows the singleton pattern for global client management while
// also supporting multiple client instances for different use cases.
//
// Example usage:
//
//	// Initialize global client
//	cfg := &config.RedisConfig{
//		Host:     "localhost",
//		Port:     6379,
//		Database: 0,
//		PoolSize: 10,
//	}
//
//	err := redis.Initialize(cfg, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer redis.Shutdown()
//
//	// Use global client
//	client := redis.GetClient()
//	err = client.Set(ctx, "key", "value", time.Hour)
package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"backend-go/internal/config"
)

// Client represents a Redis client wrapper that provides enhanced functionality
// over the standard Redis client. It supports both single-node and cluster
// configurations and includes built-in metrics collection, health monitoring,
// and connection management.
//
// The Client struct is thread-safe and can be used concurrently from multiple
// goroutines. It maintains internal state for connection pooling, metrics,
// and configuration management.
//
// Fields:
//   - rdb: The underlying Redis universal client (supports both single and cluster)
//   - config: Redis configuration used for connection parameters
//   - logger: Structured logger for Redis operations and debugging
//   - metrics: Metrics collector for performance monitoring
//   - mu: Read-write mutex for thread-safe operations
type Client struct {
	rdb     redis.UniversalClient // Universal Redis client (single/cluster)
	config  *config.RedisConfig   // Redis configuration
	logger  *logrus.Logger        // Structured logger
	metrics *Metrics              // Metrics collector
	mu      sync.RWMutex          // Thread-safety mutex
}

// clientInstance 全局客户端实例
var (
	clientInstance *Client
	clientOnce     sync.Once
)

// Initialize initializes the global Redis client instance using the singleton pattern.
//
// This function should be called once during application startup to establish
// the global Redis connection. It uses sync.Once to ensure thread-safe
// initialization and prevents multiple initialization attempts.
//
// Parameters:
//   - cfg: Redis configuration containing connection parameters, pool settings,
//     and cluster configuration. Must not be nil.
//   - logger: Structured logger for Redis operations. If nil, a default logger
//     will be created.
//
// Returns:
//   - error: Any error that occurred during client initialization
//
// The function performs the following operations:
//  1. Creates a new Redis client using the provided configuration
//  2. Tests the connection to ensure Redis is accessible
//  3. Sets up metrics collection and logging
//  4. Stores the client instance globally for later access
//
// Example:
//
//	cfg := &config.RedisConfig{
//		Host:     "localhost",
//		Port:     6379,
//		Database: 0,
//		PoolSize: 10,
//	}
//
//	logger := logrus.New()
//	err := Initialize(cfg, logger)
//	if err != nil {
//		log.Fatalf("Failed to initialize Redis: %v", err)
//	}
//
// Note: This function uses sync.Once internally, so multiple calls are safe
// but only the first call will actually initialize the client.
func Initialize(cfg *config.RedisConfig, logger *logrus.Logger) error {
	var err error
	clientOnce.Do(func() {
		clientInstance, err = NewClient(cfg, logger)
	})
	return err
}

// GetClient 获取全局 Redis 客户端实例
func GetClient() *Client {
	if clientInstance == nil {
		panic("Redis client not initialized. Call Initialize() first.")
	}
	return clientInstance
}

// NewClient 创建新的 Redis 客户端
func NewClient(cfg *config.RedisConfig, logger *logrus.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}

	// 创建 Redis 客户端选项
	var rdb redis.UniversalClient

	if cfg.Cluster.Enabled {
		// 集群模式
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           cfg.Cluster.Addresses,
			Password:        cfg.Password,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			MaxRetries:      cfg.MaxRetries,
			DialTimeout:     cfg.DialTimeout,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			PoolTimeout:     cfg.PoolTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			ConnMaxLifetime: cfg.MaxConnAge,
		})
	} else {
		// 单机模式
		rdb = redis.NewClient(&redis.Options{
			Addr:            cfg.GetRedisAddr(),
			Password:        cfg.Password,
			DB:              cfg.Database,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			MaxRetries:      cfg.MaxRetries,
			DialTimeout:     cfg.DialTimeout,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			PoolTimeout:     cfg.PoolTimeout,
			ConnMaxIdleTime: cfg.IdleTimeout,
			ConnMaxLifetime: cfg.MaxConnAge,
		})
	}

	client := &Client{
		rdb:     rdb,
		config:  cfg,
		logger:  logger,
		metrics: NewMetrics(),
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"addr":      cfg.GetRedisAddr(),
		"database":  cfg.Database,
		"pool_size": cfg.PoolSize,
		"cluster":   cfg.Cluster.Enabled,
	}).Info("Redis client initialized successfully")

	return client, nil
}

// Ping 测试 Redis 连接
func (c *Client) Ping(ctx context.Context) error {
	start := time.Now()
	defer func() {
		c.metrics.RecordOperation("ping", time.Since(start), nil)
	}()

	result := c.rdb.Ping(ctx)
	if err := result.Err(); err != nil {
		c.metrics.RecordOperation("ping", time.Since(start), err)
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rdb != nil {
		err := c.rdb.Close()
		c.logger.Info("Redis client closed")
		return err
	}
	return nil
}

// GetRedisClient 获取原始 Redis 客户端（用于高级操作）
func (c *Client) GetRedisClient() redis.UniversalClient {
	return c.rdb
}

// Redis操作方法 - 代理到底层客户端
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.LPush(ctx, key, values...).Err()
}

func (c *Client) LTrim(ctx context.Context, key string, start, stop int64) error {
	return c.rdb.LTrim(ctx, key, start, stop).Err()
}

func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.LRange(ctx, key, start, stop).Result()
}

func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.rdb.SAdd(ctx, key, members...).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.SCard(ctx, key).Result()
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, key).Result()
}

// GetConfig 获取配置
func (c *Client) GetConfig() *config.RedisConfig {
	return c.config
}

// GetMetrics 获取指标
func (c *Client) GetMetrics() *Metrics {
	return c.metrics
}

// Shutdown 关闭全局客户端实例
func Shutdown() error {
	if clientInstance != nil {
		return clientInstance.Close()
	}
	return nil
}

// TestConnection 测试 Redis 连接（不初始化全局实例）
func TestConnection(cfg *config.RedisConfig) error {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // 减少测试时的日志输出

	client, err := NewClient(cfg, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx)
}

// WaitForHealthy 等待 Redis 变为健康状态
func WaitForHealthy(timeout time.Duration) error {
	if clientInstance == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Redis to become healthy")
		case <-ticker.C:
			if err := clientInstance.Ping(ctx); err == nil {
				return nil
			}
		}
	}
}

// GetConnectionInfo 获取连接信息
func (c *Client) GetConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := map[string]interface{}{
		"addr":           c.config.GetRedisAddr(),
		"database":       c.config.Database,
		"pool_size":      c.config.PoolSize,
		"min_idle_conns": c.config.MinIdleConns,
		"cluster_mode":   c.config.Cluster.Enabled,
	}

	// 获取连接池统计信息
	if poolStats := c.rdb.PoolStats(); poolStats != nil {
		info["pool_stats"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}
	}

	return info
}

// GetStatus 获取完整状态信息
func (c *Client) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"connection": c.GetConnectionInfo(),
		"metrics":    c.metrics.GetSummary(),
		"health":     c.IsHealthy(),
	}

	// 获取 Redis 服务器信息
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if info, err := c.rdb.Info(ctx, "server", "memory", "stats").Result(); err == nil {
		status["server_info"] = parseRedisInfo(info)
	}

	return status
}

// IsHealthy 检查 Redis 是否健康
func (c *Client) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.Ping(ctx) == nil
}

// parseRedisInfo 解析 Redis INFO 命令输出
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result
}
