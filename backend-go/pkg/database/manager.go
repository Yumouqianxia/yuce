package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/config"
	applogger "backend-go/internal/shared/logger"
)

// Manager 数据库管理器
type Manager struct {
	db     *DB
	config *config.DatabaseConfig
	mu     sync.RWMutex

	// 监控相关
	metrics *Metrics
	monitor *Monitor
}

// NewManager 创建数据库管理器
func NewManager(cfg *config.DatabaseConfig) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}

	// 如果启用了自动创建数据库
	if cfg.Migration.AutoCreate {
		if err := CreateDatabase(cfg); err != nil {
			applogger.Warn("Failed to create database, continuing anyway")
		}
	}

	// 创建数据库连接
	db, err := NewDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	manager := &Manager{
		db:      db,
		config:  cfg,
		metrics: NewMetrics(),
		monitor: NewMonitor(),
	}

	// 启动监控
	manager.startMonitoring()

	applogger.Info("Database manager initialized successfully")
	return manager, nil
}

// GetDB 获取数据库实例
func (m *Manager) GetDB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db.DB
}

// GetRawDB 获取原始数据库实例
func (m *Manager) GetRawDB() *DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 停止监控
	m.monitor.Stop()

	// 关闭数据库连接
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	return m.db.HealthCheck(ctx)
}

// GetMetrics 获取数据库指标
func (m *Manager) GetMetrics() *Metrics {
	return m.metrics
}

// GetConnectionInfo 获取连接信息
func (m *Manager) GetConnectionInfo() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return map[string]interface{}{"status": "disconnected"}
	}

	info := m.db.GetConnectionInfo()
	info["status"] = "connected"
	info["metrics"] = m.metrics.GetSummary()

	return info
}

// Reconnect 重新连接数据库
func (m *Manager) Reconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	applogger.Info("Attempting to reconnect to database")

	// 关闭现有连接
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			applogger.Warn("Failed to close existing database connection")
		}
	}

	// 创建新连接
	db, err := NewDB(m.config)
	if err != nil {
		return fmt.Errorf("failed to reconnect to database: %w", err)
	}

	m.db = db
	applogger.Info("Database reconnected successfully")
	return nil
}

// startMonitoring 启动数据库监控
func (m *Manager) startMonitoring() {
	// 启动连接池监控
	go m.monitorConnectionPool()

	// 启动性能监控
	go m.monitorPerformance()
}

// monitorConnectionPool 监控连接池
func (m *Manager) monitorConnectionPool() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectConnectionPoolMetrics()
		case <-m.monitor.stopCh:
			return
		}
	}
}

// monitorPerformance 监控性能
func (m *Manager) monitorPerformance() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectPerformanceMetrics()
		case <-m.monitor.stopCh:
			return
		}
	}
}

// collectConnectionPoolMetrics 收集连接池指标
func (m *Manager) collectConnectionPoolMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return
	}

	stats := m.db.Stats()

	m.metrics.UpdateConnectionPool(ConnectionPoolMetrics{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		Timestamp:         time.Now(),
	})

	// 检查连接池健康状态
	if stats.OpenConnections == 0 {
		applogger.Warn("No open database connections detected")
	}

	// 检查等待时间
	if stats.WaitDuration > 5*time.Second {
		applogger.Warn("High database connection wait time detected")
	}

	// 检查连接使用率
	if m.config.MaxOpenConns > 0 {
		usage := float64(stats.OpenConnections) / float64(m.config.MaxOpenConns)
		if usage > 0.8 {
			applogger.Warn("High database connection usage detected")
		}
	}
}

// collectPerformanceMetrics 收集性能指标
func (m *Manager) collectPerformanceMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试查询性能
	start := time.Now()
	var result int
	err := m.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	duration := time.Since(start)

	if err != nil {
		m.metrics.RecordQueryError()
		applogger.Error("Database performance test query failed")
		return
	}

	m.metrics.RecordQuery(duration)

	// 如果查询时间过长，记录警告
	if duration > 100*time.Millisecond {
		applogger.Warn("Slow database performance test query detected")
	}
}

// WithTransaction 执行事务
func (m *Manager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	start := time.Now()
	err := m.db.WithContext(ctx).Transaction(fn)
	duration := time.Since(start)

	// 记录事务指标
	if err != nil {
		m.metrics.RecordTransactionError()
	} else {
		m.metrics.RecordTransaction(duration)
	}

	return err
}

// ExecuteWithMetrics 执行查询并记录指标
func (m *Manager) ExecuteWithMetrics(ctx context.Context, fn func(*gorm.DB) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	start := time.Now()
	err := fn(m.db.WithContext(ctx))
	duration := time.Since(start)

	// 记录查询指标
	if err != nil {
		m.metrics.RecordQueryError()
	} else {
		m.metrics.RecordQuery(duration)
	}

	return err
}

// GetSlowQueries 获取慢查询信息（如果支持）
func (m *Manager) GetSlowQueries(ctx context.Context, limit int) ([]SlowQuery, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var queries []SlowQuery

	// 查询 MySQL 慢查询日志表（需要启用慢查询日志）
	sql := `
		SELECT 
			start_time,
			user_host,
			query_time,
			lock_time,
			rows_sent,
			rows_examined,
			sql_text
		FROM mysql.slow_log 
		ORDER BY start_time DESC 
		LIMIT ?
	`

	err := m.db.WithContext(ctx).Raw(sql, limit).Scan(&queries).Error
	if err != nil {
		// 如果慢查询日志表不可用，返回空结果而不是错误
		applogger.Debug("Slow query log table not available")
		return []SlowQuery{}, nil
	}

	return queries, nil
}

// OptimizeConnectionPool 优化连接池配置
func (m *Manager) OptimizeConnectionPool() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	stats := m.db.Stats()
	sqlDB := m.db.GetSQLDB()

	// 基于当前使用情况调整连接池
	if stats.WaitCount > 0 && stats.OpenConnections < m.config.MaxOpenConns {
		// 如果有等待且未达到最大连接数，可以考虑增加连接
		newMaxOpen := min(m.config.MaxOpenConns, stats.OpenConnections+5)
		sqlDB.SetMaxOpenConns(newMaxOpen)

		applogger.Info("Optimized database connection pool")
	}

	return nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SlowQuery 慢查询信息
type SlowQuery struct {
	StartTime    time.Time `gorm:"column:start_time"`
	UserHost     string    `gorm:"column:user_host"`
	QueryTime    float64   `gorm:"column:query_time"`
	LockTime     float64   `gorm:"column:lock_time"`
	RowsSent     int       `gorm:"column:rows_sent"`
	RowsExamined int       `gorm:"column:rows_examined"`
	SQLText      string    `gorm:"column:sql_text"`
}
