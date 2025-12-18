package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// HealthChecker Redis 健康检查器
type HealthChecker struct {
	client *Client
}

// HealthResult 健康检查结果
type HealthResult struct {
	Healthy   bool                   `json:"healthy"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details"`
	Checks    map[string]CheckResult `json:"checks"`
	Score     float64                `json:"score"`
}

// CheckResult 单项检查结果
type CheckResult struct {
	Healthy   bool          `json:"healthy"`
	Message   string        `json:"message"`
	Duration  time.Duration `json:"duration"`
	Value     interface{}   `json:"value,omitempty"`
	Threshold interface{}   `json:"threshold,omitempty"`
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(client *Client) *HealthChecker {
	return &HealthChecker{
		client: client,
	}
}

// GetHealthChecker 获取全局健康检查器
func GetHealthChecker() *HealthChecker {
	return NewHealthChecker(GetClient())
}

// Check 执行完整健康检查
func (hc *HealthChecker) Check(ctx context.Context) *HealthResult {
	start := time.Now()

	result := &HealthResult{
		Timestamp: start,
		Checks:    make(map[string]CheckResult),
		Details:   make(map[string]interface{}),
	}

	// 执行各项检查
	hc.checkConnection(ctx, result)
	hc.checkMemory(ctx, result)
	hc.checkPerformance(ctx, result)
	hc.checkConnections(ctx, result)
	hc.checkReplication(ctx, result)
	hc.checkPersistence(ctx, result)

	// 计算总体健康状态
	result.Duration = time.Since(start)
	result.Healthy = hc.calculateOverallHealth(result)
	result.Score = hc.calculateHealthScore(result)

	if result.Healthy {
		result.Status = "healthy"
		result.Message = "All checks passed"
	} else {
		result.Status = "unhealthy"
		result.Message = hc.getUnhealthyMessage(result)
	}

	return result
}

// QuickCheck 执行快速健康检查
func (hc *HealthChecker) QuickCheck(ctx context.Context) *HealthResult {
	start := time.Now()

	result := &HealthResult{
		Timestamp: start,
		Checks:    make(map[string]CheckResult),
		Details:   make(map[string]interface{}),
	}

	// 只执行关键检查
	hc.checkConnection(ctx, result)
	hc.checkPerformance(ctx, result)

	result.Duration = time.Since(start)
	result.Healthy = hc.calculateOverallHealth(result)
	result.Score = hc.calculateHealthScore(result)

	if result.Healthy {
		result.Status = "healthy"
		result.Message = "Quick checks passed"
	} else {
		result.Status = "unhealthy"
		result.Message = hc.getUnhealthyMessage(result)
	}

	return result
}

// IsHealthy 检查是否健康
func (hc *HealthChecker) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := hc.QuickCheck(ctx)
	return result.Healthy
}

// checkConnection 检查连接
func (hc *HealthChecker) checkConnection(ctx context.Context, result *HealthResult) {
	start := time.Now()

	err := hc.client.Ping(ctx)
	duration := time.Since(start)

	if err != nil {
		result.Checks["connection"] = CheckResult{
			Healthy:  false,
			Message:  fmt.Sprintf("Connection failed: %v", err),
			Duration: duration,
		}
		return
	}

	result.Checks["connection"] = CheckResult{
		Healthy:   true,
		Message:   "Connection successful",
		Duration:  duration,
		Value:     duration.Milliseconds(),
		Threshold: 1000, // 1秒阈值
	}
}

// checkMemory 检查内存使用
func (hc *HealthChecker) checkMemory(ctx context.Context, result *HealthResult) {
	start := time.Now()

	info, err := hc.client.rdb.Info(ctx, "memory").Result()
	if err != nil {
		result.Checks["memory"] = CheckResult{
			Healthy:  false,
			Message:  fmt.Sprintf("Failed to get memory info: %v", err),
			Duration: time.Since(start),
		}
		return
	}

	memoryInfo := parseRedisInfo(info)

	// 解析内存使用情况
	usedMemory, _ := strconv.ParseInt(memoryInfo["used_memory"], 10, 64)
	maxMemory, _ := strconv.ParseInt(memoryInfo["maxmemory"], 10, 64)

	var memoryUsagePercent float64
	if maxMemory > 0 {
		memoryUsagePercent = float64(usedMemory) / float64(maxMemory) * 100
	}

	// 检查内存使用率
	healthy := true
	message := "Memory usage normal"

	if maxMemory > 0 && memoryUsagePercent > 90 {
		healthy = false
		message = fmt.Sprintf("Memory usage too high: %.2f%%", memoryUsagePercent)
	}

	result.Checks["memory"] = CheckResult{
		Healthy:   healthy,
		Message:   message,
		Duration:  time.Since(start),
		Value:     memoryUsagePercent,
		Threshold: 90.0,
	}

	result.Details["memory_info"] = map[string]interface{}{
		"used_memory_mb":             float64(usedMemory) / 1024 / 1024,
		"max_memory_mb":              float64(maxMemory) / 1024 / 1024,
		"memory_usage_percent":       memoryUsagePercent,
		"memory_fragmentation_ratio": memoryInfo["mem_fragmentation_ratio"],
	}
}

// checkPerformance 检查性能指标
func (hc *HealthChecker) checkPerformance(ctx context.Context, result *HealthResult) {
	start := time.Now()

	metrics := hc.client.GetMetrics()

	// 检查错误率
	errorRate := metrics.GetCurrentErrorRate()
	errorHealthy := errorRate < 5.0 // 错误率低于5%

	// 检查缓存命中率
	cacheHitRate := metrics.GetCurrentCacheHitRate()
	totalCacheOps := metrics.CacheHits + metrics.CacheMisses
	cacheHealthy := true
	if totalCacheOps > 100 { // 有足够样本时才检查
		cacheHealthy = cacheHitRate > 50.0 // 缓存命中率高于50%
	}

	// 检查平均延迟
	avgLatency := float64(metrics.TotalLatency) / float64(metrics.TotalOperations) / 1e6
	latencyHealthy := avgLatency < 100.0 // 平均延迟低于100ms

	overall := errorHealthy && cacheHealthy && latencyHealthy

	var message string
	if overall {
		message = "Performance metrics normal"
	} else {
		issues := []string{}
		if !errorHealthy {
			issues = append(issues, fmt.Sprintf("high error rate: %.2f%%", errorRate))
		}
		if !cacheHealthy {
			issues = append(issues, fmt.Sprintf("low cache hit rate: %.2f%%", cacheHitRate))
		}
		if !latencyHealthy {
			issues = append(issues, fmt.Sprintf("high latency: %.2fms", avgLatency))
		}
		message = "Performance issues: " + strings.Join(issues, ", ")
	}

	result.Checks["performance"] = CheckResult{
		Healthy:  overall,
		Message:  message,
		Duration: time.Since(start),
	}

	result.Details["performance"] = map[string]interface{}{
		"error_rate":       errorRate,
		"cache_hit_rate":   cacheHitRate,
		"avg_latency_ms":   avgLatency,
		"total_operations": metrics.TotalOperations,
		"qps":              metrics.GetCurrentQPS(),
	}
}

// checkConnections 检查连接池状态
func (hc *HealthChecker) checkConnections(ctx context.Context, result *HealthResult) {
	start := time.Now()

	poolStats := hc.client.rdb.PoolStats()
	if poolStats == nil {
		result.Checks["connections"] = CheckResult{
			Healthy:  false,
			Message:  "Unable to get pool stats",
			Duration: time.Since(start),
		}
		return
	}

	// 检查连接池使用率
	poolUsage := float64(poolStats.TotalConns-poolStats.IdleConns) / float64(hc.client.config.PoolSize) * 100

	healthy := true
	message := "Connection pool normal"

	if poolUsage > 90 {
		healthy = false
		message = fmt.Sprintf("Connection pool usage too high: %.2f%%", poolUsage)
	}

	if poolStats.Timeouts > 0 {
		healthy = false
		message = fmt.Sprintf("Connection pool timeouts detected: %d", poolStats.Timeouts)
	}

	result.Checks["connections"] = CheckResult{
		Healthy:   healthy,
		Message:   message,
		Duration:  time.Since(start),
		Value:     poolUsage,
		Threshold: 90.0,
	}

	result.Details["connection_pool"] = map[string]interface{}{
		"total_conns":   poolStats.TotalConns,
		"idle_conns":    poolStats.IdleConns,
		"stale_conns":   poolStats.StaleConns,
		"hits":          poolStats.Hits,
		"misses":        poolStats.Misses,
		"timeouts":      poolStats.Timeouts,
		"pool_usage":    poolUsage,
		"max_pool_size": hc.client.config.PoolSize,
	}
}

// checkReplication 检查复制状态
func (hc *HealthChecker) checkReplication(ctx context.Context, result *HealthResult) {
	start := time.Now()

	info, err := hc.client.rdb.Info(ctx, "replication").Result()
	if err != nil {
		result.Checks["replication"] = CheckResult{
			Healthy:  false,
			Message:  fmt.Sprintf("Failed to get replication info: %v", err),
			Duration: time.Since(start),
		}
		return
	}

	replicationInfo := parseRedisInfo(info)
	role := replicationInfo["role"]

	healthy := true
	message := fmt.Sprintf("Role: %s", role)

	if role == "master" {
		connectedSlaves, _ := strconv.Atoi(replicationInfo["connected_slaves"])
		message = fmt.Sprintf("Master with %d slaves", connectedSlaves)
	} else if role == "slave" {
		masterLinkStatus := replicationInfo["master_link_status"]
		if masterLinkStatus != "up" {
			healthy = false
			message = fmt.Sprintf("Slave disconnected from master: %s", masterLinkStatus)
		} else {
			message = "Slave connected to master"
		}
	}

	result.Checks["replication"] = CheckResult{
		Healthy:  healthy,
		Message:  message,
		Duration: time.Since(start),
	}

	result.Details["replication"] = replicationInfo
}

// checkPersistence 检查持久化状态
func (hc *HealthChecker) checkPersistence(ctx context.Context, result *HealthResult) {
	start := time.Now()

	info, err := hc.client.rdb.Info(ctx, "persistence").Result()
	if err != nil {
		result.Checks["persistence"] = CheckResult{
			Healthy:  false,
			Message:  fmt.Sprintf("Failed to get persistence info: %v", err),
			Duration: time.Since(start),
		}
		return
	}

	persistenceInfo := parseRedisInfo(info)

	healthy := true
	message := "Persistence normal"

	// 检查 RDB 状态
	rdbLastBgsaveStatus := persistenceInfo["rdb_last_bgsave_status"]
	if rdbLastBgsaveStatus == "err" {
		healthy = false
		message = "RDB background save failed"
	}

	// 检查 AOF 状态
	aofEnabled := persistenceInfo["aof_enabled"] == "1"
	if aofEnabled {
		aofLastBgrewriteStatus := persistenceInfo["aof_last_bgrewrite_status"]
		if aofLastBgrewriteStatus == "err" {
			healthy = false
			message = "AOF background rewrite failed"
		}
	}

	result.Checks["persistence"] = CheckResult{
		Healthy:  healthy,
		Message:  message,
		Duration: time.Since(start),
	}

	result.Details["persistence"] = persistenceInfo
}

// calculateOverallHealth 计算总体健康状态
func (hc *HealthChecker) calculateOverallHealth(result *HealthResult) bool {
	// 关键检查项必须通过
	criticalChecks := []string{"connection", "performance"}

	for _, check := range criticalChecks {
		if checkResult, exists := result.Checks[check]; exists && !checkResult.Healthy {
			return false
		}
	}

	// 计算通过的检查项比例
	totalChecks := len(result.Checks)
	passedChecks := 0

	for _, checkResult := range result.Checks {
		if checkResult.Healthy {
			passedChecks++
		}
	}

	// 至少80%的检查项通过
	return float64(passedChecks)/float64(totalChecks) >= 0.8
}

// calculateHealthScore 计算健康评分
func (hc *HealthChecker) calculateHealthScore(result *HealthResult) float64 {
	if len(result.Checks) == 0 {
		return 0
	}

	totalScore := 0.0
	weights := map[string]float64{
		"connection":  30.0, // 连接权重最高
		"performance": 25.0, // 性能权重次之
		"memory":      20.0, // 内存权重
		"connections": 15.0, // 连接池权重
		"replication": 5.0,  // 复制权重
		"persistence": 5.0,  // 持久化权重
	}

	totalWeight := 0.0

	for checkName, checkResult := range result.Checks {
		weight := weights[checkName]
		if weight == 0 {
			weight = 10.0 // 默认权重
		}

		totalWeight += weight

		if checkResult.Healthy {
			totalScore += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return (totalScore / totalWeight) * 100
}

// getUnhealthyMessage 获取不健康消息
func (hc *HealthChecker) getUnhealthyMessage(result *HealthResult) string {
	var issues []string

	for checkName, checkResult := range result.Checks {
		if !checkResult.Healthy {
			issues = append(issues, fmt.Sprintf("%s: %s", checkName, checkResult.Message))
		}
	}

	if len(issues) == 0 {
		return "Health check failed"
	}

	return strings.Join(issues, "; ")
}

// WaitForHealthy 等待 Redis 变为健康状态
func (hc *HealthChecker) WaitForHealthy(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Redis to become healthy")
		case <-ticker.C:
			result := hc.QuickCheck(ctx)
			if result.Healthy {
				return nil
			}
		}
	}
}

// GetHealthStatus 获取健康状态摘要
func (hc *HealthChecker) GetHealthStatus() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := hc.Check(ctx)

	return map[string]interface{}{
		"healthy":   result.Healthy,
		"status":    result.Status,
		"message":   result.Message,
		"score":     result.Score,
		"timestamp": result.Timestamp,
		"duration":  result.Duration.Milliseconds(),
		"checks":    len(result.Checks),
		"passed":    hc.countPassedChecks(result),
	}
}

// countPassedChecks 计算通过的检查项数量
func (hc *HealthChecker) countPassedChecks(result *HealthResult) int {
	passed := 0
	for _, checkResult := range result.Checks {
		if checkResult.Healthy {
			passed++
		}
	}
	return passed
}
