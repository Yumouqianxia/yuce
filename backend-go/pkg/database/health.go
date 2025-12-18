package database

import (
	"context"
	"fmt"
	"time"

	applogger "backend-go/internal/shared/logger"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck 健康检查结果
type HealthCheckResult struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details"`
	Checks    []IndividualCheck      `json:"checks"`
	Score     float64                `json:"score"`
}

// IndividualCheck 单项检查结果
type IndividualCheck struct {
	Name     string        `json:"name"`
	Status   HealthStatus  `json:"status"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	manager *Manager
	timeout time.Duration
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(manager *Manager) *HealthChecker {
	return &HealthChecker{
		manager: manager,
		timeout: 10 * time.Second,
	}
}

// SetTimeout 设置检查超时时间
func (hc *HealthChecker) SetTimeout(timeout time.Duration) {
	hc.timeout = timeout
}

// Check 执行完整的健康检查
func (hc *HealthChecker) Check(ctx context.Context) *HealthCheckResult {
	start := time.Now()

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	result := &HealthCheckResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
		Checks:    make([]IndividualCheck, 0),
	}

	// 执行各项检查
	checks := []func(context.Context) IndividualCheck{
		hc.checkConnection,
		hc.checkConnectionPool,
		hc.checkQueryPerformance,
		hc.checkDiskSpace,
		hc.checkReplication,
	}

	healthyCount := 0
	degradedCount := 0

	for _, checkFunc := range checks {
		check := checkFunc(checkCtx)
		result.Checks = append(result.Checks, check)

		switch check.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}

	// 计算总体状态
	totalChecks := len(checks)
	if healthyCount == totalChecks {
		result.Status = HealthStatusHealthy
		result.Message = "All database checks passed"
	} else if healthyCount+degradedCount == totalChecks {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("%d checks degraded", degradedCount)
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("%d checks failed", totalChecks-healthyCount-degradedCount)
	}

	// 添加详细信息
	result.Duration = time.Since(start)
	result.Details = hc.getHealthDetails()
	result.Score = hc.manager.GetMetrics().GetHealthScore()

	return result
}

// checkConnection 检查数据库连接
func (hc *HealthChecker) checkConnection(ctx context.Context) IndividualCheck {
	start := time.Now()
	check := IndividualCheck{
		Name: "database_connection",
	}

	err := hc.manager.HealthCheck(ctx)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database connection failed"
		check.Error = err.Error()
		applogger.Error("Database connection health check failed")
	} else {
		check.Status = HealthStatusHealthy
		check.Message = "Database connection is healthy"
	}

	return check
}

// checkConnectionPool 检查连接池状态
func (hc *HealthChecker) checkConnectionPool(ctx context.Context) IndividualCheck {
	start := time.Now()
	check := IndividualCheck{
		Name: "connection_pool",
	}

	db := hc.manager.GetRawDB()
	if db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database instance not available"
		check.Duration = time.Since(start)
		return check
	}

	stats := db.Stats()
	config := db.GetConfig()

	check.Duration = time.Since(start)

	// 检查连接池使用率
	if config.MaxOpenConns > 0 {
		usage := float64(stats.OpenConnections) / float64(config.MaxOpenConns)

		if usage > 0.9 {
			check.Status = HealthStatusDegraded
			check.Message = fmt.Sprintf("High connection pool usage: %.1f%%", usage*100)
		} else if usage > 0.95 {
			check.Status = HealthStatusUnhealthy
			check.Message = fmt.Sprintf("Critical connection pool usage: %.1f%%", usage*100)
		} else {
			check.Status = HealthStatusHealthy
			check.Message = fmt.Sprintf("Connection pool usage: %.1f%%", usage*100)
		}
	} else {
		check.Status = HealthStatusHealthy
		check.Message = "Connection pool is healthy"
	}

	// 检查等待时间
	if stats.WaitDuration > 5*time.Second {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("High connection wait time: %v", stats.WaitDuration)
	}

	return check
}

// checkQueryPerformance 检查查询性能
func (hc *HealthChecker) checkQueryPerformance(ctx context.Context) IndividualCheck {
	start := time.Now()
	check := IndividualCheck{
		Name: "query_performance",
	}

	db := hc.manager.GetDB()
	if db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database instance not available"
		check.Duration = time.Since(start)
		return check
	}

	// 执行简单的性能测试查询
	queryStart := time.Now()
	var result int
	err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	queryDuration := time.Since(queryStart)

	check.Duration = time.Since(start)

	if err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Query performance test failed"
		check.Error = err.Error()
	} else if queryDuration > 500*time.Millisecond {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("Slow query performance: %v", queryDuration)
	} else if queryDuration > 1*time.Second {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Very slow query performance: %v", queryDuration)
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("Query performance is good: %v", queryDuration)
	}

	return check
}

// checkDiskSpace 检查磁盘空间（通过查询 MySQL 系统表）
func (hc *HealthChecker) checkDiskSpace(ctx context.Context) IndividualCheck {
	start := time.Now()
	check := IndividualCheck{
		Name: "disk_space",
	}

	db := hc.manager.GetDB()
	if db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database instance not available"
		check.Duration = time.Since(start)
		return check
	}

	// 查询数据库大小
	var dbSize float64
	err := db.WithContext(ctx).Raw(`
		SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as db_size_mb
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&dbSize).Error

	check.Duration = time.Since(start)

	if err != nil {
		check.Status = HealthStatusDegraded
		check.Message = "Could not check disk space"
		check.Error = err.Error()
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("Database size: %.2f MB", dbSize)
	}

	return check
}

// checkReplication 检查主从复制状态（如果配置了复制）
func (hc *HealthChecker) checkReplication(ctx context.Context) IndividualCheck {
	start := time.Now()
	check := IndividualCheck{
		Name: "replication",
	}

	db := hc.manager.GetDB()
	if db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "Database instance not available"
		check.Duration = time.Since(start)
		return check
	}

	// 检查是否为从库
	var slaveStatus []map[string]interface{}
	err := db.WithContext(ctx).Raw("SHOW SLAVE STATUS").Find(&slaveStatus).Error

	check.Duration = time.Since(start)

	if err != nil {
		// 如果查询失败，可能是权限问题或者不是复制环境
		check.Status = HealthStatusHealthy
		check.Message = "Replication not configured or not accessible"
	} else if len(slaveStatus) == 0 {
		// 不是从库
		check.Status = HealthStatusHealthy
		check.Message = "Not a slave server"
	} else {
		// 是从库，检查复制状态
		status := slaveStatus[0]
		ioRunning := status["Slave_IO_Running"]
		sqlRunning := status["Slave_SQL_Running"]

		if ioRunning == "Yes" && sqlRunning == "Yes" {
			check.Status = HealthStatusHealthy
			check.Message = "Replication is running normally"
		} else {
			check.Status = HealthStatusUnhealthy
			check.Message = fmt.Sprintf("Replication issues: IO=%v, SQL=%v", ioRunning, sqlRunning)
		}
	}

	return check
}

// getHealthDetails 获取健康检查详细信息
func (hc *HealthChecker) getHealthDetails() map[string]interface{} {
	details := make(map[string]interface{})

	// 添加连接信息
	if connInfo := hc.manager.GetConnectionInfo(); connInfo != nil {
		details["connection"] = connInfo
	}

	// 添加指标信息
	if metrics := hc.manager.GetMetrics(); metrics != nil {
		details["metrics"] = metrics.GetSummary()
	}

	// 添加配置信息
	if db := hc.manager.GetRawDB(); db != nil {
		if config := db.GetConfig(); config != nil {
			details["config"] = map[string]interface{}{
				"host":               config.Host,
				"port":               config.Port,
				"database":           config.Database,
				"max_open_conns":     config.MaxOpenConns,
				"max_idle_conns":     config.MaxIdleConns,
				"conn_max_lifetime":  config.ConnMaxLifetime,
				"conn_max_idle_time": config.ConnMaxIdleTime,
			}
		}
	}

	return details
}

// QuickCheck 快速健康检查（仅检查连接）
func (hc *HealthChecker) QuickCheck(ctx context.Context) *HealthCheckResult {
	start := time.Now()

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := &HealthCheckResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
		Checks:    make([]IndividualCheck, 0),
	}

	// 只执行连接检查
	check := hc.checkConnection(checkCtx)
	result.Checks = append(result.Checks, check)

	result.Status = check.Status
	result.Message = check.Message
	result.Duration = time.Since(start)

	return result
}

// GetStatus 获取当前状态（不执行检查）
func (hc *HealthChecker) GetStatus() HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := hc.manager.HealthCheck(ctx)
	if err != nil {
		return HealthStatusUnhealthy
	}

	// 检查指标
	metrics := hc.manager.GetMetrics()
	score := metrics.GetHealthScore()

	if score >= 90 {
		return HealthStatusHealthy
	} else if score >= 70 {
		return HealthStatusDegraded
	} else {
		return HealthStatusUnhealthy
	}
}

// IsHealthy 检查是否健康
func (hc *HealthChecker) IsHealthy() bool {
	return hc.GetStatus() == HealthStatusHealthy
}

// WaitForHealthy 等待数据库变为健康状态
func (hc *HealthChecker) WaitForHealthy(ctx context.Context, checkInterval time.Duration) error {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if hc.IsHealthy() {
				return nil
			}
			applogger.Debug("Waiting for database to become healthy")
		}
	}
}
