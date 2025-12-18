package database

import (
	"fmt"
	"sync"
	"time"
)

// Metrics 数据库指标
type Metrics struct {
	mu sync.RWMutex

	// 查询指标
	QueryCount     int64         `json:"query_count"`
	QueryErrors    int64         `json:"query_errors"`
	TotalQueryTime time.Duration `json:"total_query_time"`
	AvgQueryTime   time.Duration `json:"avg_query_time"`
	MaxQueryTime   time.Duration `json:"max_query_time"`
	MinQueryTime   time.Duration `json:"min_query_time"`
	SlowQueryCount int64         `json:"slow_query_count"`

	// 事务指标
	TransactionCount     int64         `json:"transaction_count"`
	TransactionErrors    int64         `json:"transaction_errors"`
	TotalTransactionTime time.Duration `json:"total_transaction_time"`
	AvgTransactionTime   time.Duration `json:"avg_transaction_time"`

	// 连接池指标
	ConnectionPool ConnectionPoolMetrics `json:"connection_pool"`

	// 时间戳
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// ConnectionPoolMetrics 连接池指标
type ConnectionPoolMetrics struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	Timestamp         time.Time     `json:"timestamp"`
}

// NewMetrics 创建新的指标实例
func NewMetrics() *Metrics {
	now := time.Now()
	return &Metrics{
		StartTime:    now,
		LastUpdated:  now,
		MinQueryTime: time.Duration(^uint64(0) >> 1), // 设置为最大值
	}
}

// RecordQuery 记录查询指标
func (m *Metrics) RecordQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryCount++
	m.TotalQueryTime += duration
	m.LastUpdated = time.Now()

	// 更新平均查询时间
	m.AvgQueryTime = m.TotalQueryTime / time.Duration(m.QueryCount)

	// 更新最大查询时间
	if duration > m.MaxQueryTime {
		m.MaxQueryTime = duration
	}

	// 更新最小查询时间
	if duration < m.MinQueryTime {
		m.MinQueryTime = duration
	}

	// 检查是否为慢查询（超过200ms）
	if duration > 200*time.Millisecond {
		m.SlowQueryCount++
	}
}

// RecordQueryError 记录查询错误
func (m *Metrics) RecordQueryError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryErrors++
	m.LastUpdated = time.Now()
}

// RecordTransaction 记录事务指标
func (m *Metrics) RecordTransaction(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TransactionCount++
	m.TotalTransactionTime += duration
	m.LastUpdated = time.Now()

	// 更新平均事务时间
	m.AvgTransactionTime = m.TotalTransactionTime / time.Duration(m.TransactionCount)
}

// RecordTransactionError 记录事务错误
func (m *Metrics) RecordTransactionError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TransactionErrors++
	m.LastUpdated = time.Now()
}

// UpdateConnectionPool 更新连接池指标
func (m *Metrics) UpdateConnectionPool(metrics ConnectionPoolMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectionPool = metrics
	m.LastUpdated = time.Now()
}

// GetSummary 获取指标摘要
func (m *Metrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime)

	// 计算错误率
	var queryErrorRate, transactionErrorRate float64
	if m.QueryCount > 0 {
		queryErrorRate = float64(m.QueryErrors) / float64(m.QueryCount) * 100
	}
	if m.TransactionCount > 0 {
		transactionErrorRate = float64(m.TransactionErrors) / float64(m.TransactionCount) * 100
	}

	// 计算QPS（每秒查询数）
	var qps float64
	if uptime.Seconds() > 0 {
		qps = float64(m.QueryCount) / uptime.Seconds()
	}

	return map[string]interface{}{
		"uptime": uptime,
		"queries": map[string]interface{}{
			"total":        m.QueryCount,
			"errors":       m.QueryErrors,
			"error_rate":   queryErrorRate,
			"qps":          qps,
			"avg_duration": m.AvgQueryTime,
			"max_duration": m.MaxQueryTime,
			"min_duration": m.MinQueryTime,
			"slow_queries": m.SlowQueryCount,
		},
		"transactions": map[string]interface{}{
			"total":        m.TransactionCount,
			"errors":       m.TransactionErrors,
			"error_rate":   transactionErrorRate,
			"avg_duration": m.AvgTransactionTime,
		},
		"connection_pool": map[string]interface{}{
			"open_connections":    m.ConnectionPool.OpenConnections,
			"in_use":              m.ConnectionPool.InUse,
			"idle":                m.ConnectionPool.Idle,
			"wait_count":          m.ConnectionPool.WaitCount,
			"wait_duration":       m.ConnectionPool.WaitDuration,
			"max_idle_closed":     m.ConnectionPool.MaxIdleClosed,
			"max_lifetime_closed": m.ConnectionPool.MaxLifetimeClosed,
		},
		"last_updated": m.LastUpdated,
	}
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 重置所有字段，但保留 mutex
	m.QueryCount = 0
	m.QueryErrors = 0
	m.TotalQueryTime = 0
	m.AvgQueryTime = 0
	m.MaxQueryTime = 0
	m.MinQueryTime = time.Duration(^uint64(0) >> 1)
	m.SlowQueryCount = 0
	m.TransactionCount = 0
	m.TransactionErrors = 0
	m.TotalTransactionTime = 0
	m.AvgTransactionTime = 0
	m.ConnectionPool = ConnectionPoolMetrics{}
	m.LastUpdated = now
	m.StartTime = now
}

// GetQueryStats 获取查询统计信息
func (m *Metrics) GetQueryStats() QueryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return QueryStats{
		TotalQueries:  m.QueryCount,
		TotalErrors:   m.QueryErrors,
		TotalDuration: m.TotalQueryTime,
		AvgDuration:   m.AvgQueryTime,
		MaxDuration:   m.MaxQueryTime,
		MinDuration:   m.MinQueryTime,
		SlowQueries:   m.SlowQueryCount,
	}
}

// GetTransactionStats 获取事务统计信息
func (m *Metrics) GetTransactionStats() TransactionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return TransactionStats{
		TotalTransactions: m.TransactionCount,
		TotalErrors:       m.TransactionErrors,
		TotalDuration:     m.TotalTransactionTime,
		AvgDuration:       m.AvgTransactionTime,
	}
}

// QueryStats 查询统计信息
type QueryStats struct {
	TotalQueries  int64         `json:"total_queries"`
	TotalErrors   int64         `json:"total_errors"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	SlowQueries   int64         `json:"slow_queries"`
}

// TransactionStats 事务统计信息
type TransactionStats struct {
	TotalTransactions int64         `json:"total_transactions"`
	TotalErrors       int64         `json:"total_errors"`
	TotalDuration     time.Duration `json:"total_duration"`
	AvgDuration       time.Duration `json:"avg_duration"`
}

// Monitor 数据库监控器
type Monitor struct {
	stopCh chan struct{}
	mu     sync.RWMutex
}

// NewMonitor 创建新的监控器
func NewMonitor() *Monitor {
	return &Monitor{
		stopCh: make(chan struct{}),
	}
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.stopCh:
		// 已经停止
	default:
		close(m.stopCh)
	}
}

// IsRunning 检查监控是否正在运行
func (m *Monitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-m.stopCh:
		return false
	default:
		return true
	}
}

// PerformanceAlert 性能警报
type PerformanceAlert struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// AlertType 警报类型
const (
	AlertTypeSlowQuery      = "slow_query"
	AlertTypeHighErrorRate  = "high_error_rate"
	AlertTypeConnectionPool = "connection_pool"
	AlertTypeDeadlock       = "deadlock"
)

// AlertSeverity 警报严重程度
const (
	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityCritical = "critical"
)

// CheckAlerts 检查性能警报
func (m *Metrics) CheckAlerts() []PerformanceAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var alerts []PerformanceAlert
	now := time.Now()

	// 检查慢查询率
	if m.QueryCount > 0 {
		slowQueryRate := float64(m.SlowQueryCount) / float64(m.QueryCount) * 100
		if slowQueryRate > 10 { // 超过10%的查询是慢查询
			alerts = append(alerts, PerformanceAlert{
				Type:      AlertTypeSlowQuery,
				Message:   fmt.Sprintf("High slow query rate: %.2f%%", slowQueryRate),
				Severity:  AlertSeverityWarning,
				Timestamp: now,
				Metrics: map[string]interface{}{
					"slow_query_rate": slowQueryRate,
					"slow_queries":    m.SlowQueryCount,
					"total_queries":   m.QueryCount,
				},
			})
		}
	}

	// 检查错误率
	if m.QueryCount > 0 {
		errorRate := float64(m.QueryErrors) / float64(m.QueryCount) * 100
		if errorRate > 5 { // 超过5%的查询出错
			severity := AlertSeverityWarning
			if errorRate > 15 {
				severity = AlertSeverityCritical
			}

			alerts = append(alerts, PerformanceAlert{
				Type:      AlertTypeHighErrorRate,
				Message:   fmt.Sprintf("High query error rate: %.2f%%", errorRate),
				Severity:  severity,
				Timestamp: now,
				Metrics: map[string]interface{}{
					"error_rate":    errorRate,
					"query_errors":  m.QueryErrors,
					"total_queries": m.QueryCount,
				},
			})
		}
	}

	// 检查连接池状态
	if m.ConnectionPool.WaitCount > 0 && m.ConnectionPool.WaitDuration > 5*time.Second {
		alerts = append(alerts, PerformanceAlert{
			Type:      AlertTypeConnectionPool,
			Message:   fmt.Sprintf("High connection pool wait time: %v", m.ConnectionPool.WaitDuration),
			Severity:  AlertSeverityWarning,
			Timestamp: now,
			Metrics: map[string]interface{}{
				"wait_count":       m.ConnectionPool.WaitCount,
				"wait_duration":    m.ConnectionPool.WaitDuration,
				"open_connections": m.ConnectionPool.OpenConnections,
				"in_use":           m.ConnectionPool.InUse,
			},
		})
	}

	return alerts
}

// GetHealthScore 获取数据库健康评分（0-100）
func (m *Metrics) GetHealthScore() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	score := 100.0

	// 基于错误率扣分
	if m.QueryCount > 0 {
		errorRate := float64(m.QueryErrors) / float64(m.QueryCount) * 100
		score -= errorRate * 2 // 每1%错误率扣2分
	}

	// 基于慢查询率扣分
	if m.QueryCount > 0 {
		slowQueryRate := float64(m.SlowQueryCount) / float64(m.QueryCount) * 100
		score -= slowQueryRate // 每1%慢查询率扣1分
	}

	// 基于连接池等待时间扣分
	if m.ConnectionPool.WaitDuration > time.Second {
		waitSeconds := m.ConnectionPool.WaitDuration.Seconds()
		score -= waitSeconds * 5 // 每秒等待时间扣5分
	}

	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
