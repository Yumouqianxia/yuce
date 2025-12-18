package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"backend-go/internal/shared/logger"

	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Duration   time.Duration              `json:"duration"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
	System     map[string]interface{}     `json:"system"`
	Metrics    map[string]interface{}     `json:"metrics,omitempty"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// DatabaseHealthChecker 数据库健康检查器
type DatabaseHealthChecker struct {
	db      *gorm.DB
	timeout time.Duration
}

// NewDatabaseHealthChecker 创建数据库健康检查器
func NewDatabaseHealthChecker(db *gorm.DB, timeout time.Duration) *DatabaseHealthChecker {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &DatabaseHealthChecker{
		db:      db,
		timeout: timeout,
	}
}

// Name 返回检查器名称
func (h *DatabaseHealthChecker) Name() string {
	return "database"
}

// Check 执行健康检查
func (h *DatabaseHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Failed to get database instance: %v", err),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 检查连接
	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database ping failed: %v", err),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 获取连接统计
	stats := sqlDB.Stats()
	details := map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}

	// 判断状态
	status := HealthStatusHealthy
	message := "Database is healthy"

	if stats.OpenConnections == 0 {
		status = HealthStatusUnhealthy
		message = "No database connections available"
	} else if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
		status = HealthStatusDegraded
		message = "Database connection usage is high"
	}

	return ComponentHealth{
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// RedisHealthChecker Redis健康检查器
type RedisHealthChecker struct {
	client  redis.UniversalClient
	timeout time.Duration
}

// NewRedisHealthChecker 创建Redis健康检查器
func NewRedisHealthChecker(client redis.UniversalClient, timeout time.Duration) *RedisHealthChecker {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &RedisHealthChecker{
		client:  client,
		timeout: timeout,
	}
}

// Name 返回检查器名称
func (h *RedisHealthChecker) Name() string {
	return "redis"
}

// Check 执行健康检查
func (h *RedisHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// 执行PING命令
	pong, err := h.client.Ping(ctx).Result()
	if err != nil {
		return ComponentHealth{
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Redis ping failed: %v", err),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	// 获取Redis信息
	_, err = h.client.Info(ctx, "server", "memory", "stats").Result()
	details := map[string]interface{}{
		"ping_response": pong,
	}

	if err == nil {
		// 解析info信息（简化版）
		details["info_available"] = true
	} else {
		details["info_error"] = err.Error()
	}

	// 获取连接池统计（UniversalClient 在单机/集群下接口不同，做断言）
	totalConns := -1
	if poolStatsProvider, ok := h.client.(interface{ PoolStats() *redis.PoolStats }); ok {
		if ps := poolStatsProvider.PoolStats(); ps != nil {
			details["pool_stats"] = map[string]interface{}{
				"hits":        ps.Hits,
				"misses":      ps.Misses,
				"timeouts":    ps.Timeouts,
				"total_conns": ps.TotalConns,
				"idle_conns":  ps.IdleConns,
				"stale_conns": ps.StaleConns,
			}
			totalConns = int(ps.TotalConns)
		}
	}

	status := HealthStatusHealthy
	message := "Redis is healthy"

	// 检查连接池状态（当能获取到统计信息时）
	if totalConns == 0 {
		status = HealthStatusDegraded
		message = "No Redis connections in pool"
	}

	return ComponentHealth{
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// MemoryHealthChecker 内存健康检查器
type MemoryHealthChecker struct {
	maxMemoryMB uint64
}

// NewMemoryHealthChecker 创建内存健康检查器
func NewMemoryHealthChecker(maxMemoryMB uint64) *MemoryHealthChecker {
	if maxMemoryMB == 0 {
		maxMemoryMB = 1024 // 默认1GB
	}
	return &MemoryHealthChecker{
		maxMemoryMB: maxMemoryMB,
	}
}

// Name 返回检查器名称
func (h *MemoryHealthChecker) Name() string {
	return "memory"
}

// Check 执行健康检查
func (h *MemoryHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMB := m.Alloc / 1024 / 1024

	details := map[string]interface{}{
		"alloc_mb":       currentMB,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"max_memory_mb":  h.maxMemoryMB,
	}

	status := HealthStatusHealthy
	message := "Memory usage is normal"

	usagePercent := float64(currentMB) / float64(h.maxMemoryMB) * 100

	if usagePercent > 90 {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Memory usage is critical: %.1f%%", usagePercent)
	} else if usagePercent > 80 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Memory usage is high: %.1f%%", usagePercent)
	}

	details["usage_percent"] = usagePercent

	return ComponentHealth{
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// HealthService 健康检查服务
type HealthService struct {
	checkers []HealthChecker
	version  string
	timeout  time.Duration
	mu       sync.RWMutex
}

// NewHealthService 创建健康检查服务
func NewHealthService(version string, timeout time.Duration) *HealthService {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &HealthService{
		checkers: make([]HealthChecker, 0),
		version:  version,
		timeout:  timeout,
	}
}

// AddChecker 添加健康检查器
func (s *HealthService) AddChecker(checker HealthChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers = append(s.checkers, checker)
}

// Check 执行所有健康检查
func (s *HealthService) Check(ctx context.Context) HealthResponse {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.mu.RLock()
	checkers := make([]HealthChecker, len(s.checkers))
	copy(checkers, s.checkers)
	s.mu.RUnlock()

	components := make(map[string]ComponentHealth)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并发执行所有检查
	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			health := c.Check(ctx)

			mu.Lock()
			components[c.Name()] = health
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	// 确定整体状态
	overallStatus := HealthStatusHealthy
	for _, health := range components {
		if health.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if health.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	// 获取系统信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	system := map[string]interface{}{
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
		"gc_count":   m.NumGC,
		"uptime":     time.Since(start).String(),
	}

	return HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Duration:   time.Since(start),
		Version:    s.version,
		Components: components,
		System:     system,
	}
}

// HealthMiddleware 健康检查中间件
func HealthMiddleware(service *HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单健康检查
		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
			})
			return
		}

		// 详细健康检查
		if c.Request.URL.Path == "/health/detailed" {
			ctx := c.Request.Context()
			health := service.Check(ctx)

			statusCode := http.StatusOK
			if health.Status == HealthStatusUnhealthy {
				statusCode = http.StatusServiceUnavailable
			} else if health.Status == HealthStatusDegraded {
				statusCode = http.StatusPartialContent
			}

			c.JSON(statusCode, health)
			return
		}

		c.Next()
	}
}

// ReadinessMiddleware 就绪检查中间件
func ReadinessMiddleware(service *HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ready" {
			ctx := c.Request.Context()
			health := service.Check(ctx)

			// 就绪检查更严格，任何组件不健康都返回未就绪
			if health.Status != HealthStatusHealthy {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":  "not_ready",
					"message": "Service is not ready",
					"details": health.Components,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    "ready",
				"timestamp": time.Now(),
			})
			return
		}

		c.Next()
	}
}

// LivenessMiddleware 存活检查中间件
func LivenessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/live" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "alive",
				"timestamp": time.Now(),
			})
			return
		}

		c.Next()
	}
}

// MetricsHealthMiddleware 指标健康检查中间件
func MetricsHealthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health/metrics" {
			// 获取基本指标
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			metrics := map[string]interface{}{
				"memory": map[string]interface{}{
					"alloc_mb":       m.Alloc / 1024 / 1024,
					"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
					"sys_mb":         m.Sys / 1024 / 1024,
					"num_gc":         m.NumGC,
				},
				"runtime": map[string]interface{}{
					"goroutines": runtime.NumGoroutine(),
					"go_version": runtime.Version(),
					"num_cpu":    runtime.NumCPU(),
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
				"metrics":   metrics,
			})
			return
		}

		c.Next()
	}
}

// StartupProbe 启动探针
func StartupProbe(service *HealthService, maxRetries int, interval time.Duration) error {
	logger.Info("Starting startup probe...")

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		health := service.Check(ctx)
		cancel()

		if health.Status == HealthStatusHealthy {
			logger.Info("Startup probe successful")
			return nil
		}

		logger.Warnf("Startup probe failed (attempt %d/%d): %s", i+1, maxRetries, health.Status)

		if i < maxRetries-1 {
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("startup probe failed after %d attempts", maxRetries)
}
