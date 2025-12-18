package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/config"
	applogger "backend-go/internal/shared/logger"
)

// Package-level variables
var (
	defaultManager *Manager
	defaultChecker *HealthChecker
)

// Initialize 初始化数据库包
func Initialize(cfg *config.DatabaseConfig) error {
	if cfg == nil {
		return fmt.Errorf("database config is required")
	}

	// 创建数据库管理器
	manager, err := NewManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create database manager: %w", err)
	}

	// 创建健康检查器
	checker := NewHealthChecker(manager)

	// 设置全局实例
	defaultManager = manager
	defaultChecker = checker

	applogger.Info("Database package initialized successfully")
	return nil
}

// Shutdown 关闭数据库连接
func Shutdown() error {
	if defaultManager != nil {
		return defaultManager.Close()
	}
	return nil
}

// GetManager 获取默认数据库管理器
func GetManager() *Manager {
	return defaultManager
}

// GetDB 获取默认数据库实例
func GetDB() *DB {
	if defaultManager != nil {
		return defaultManager.GetRawDB()
	}
	return nil
}

// GetHealthChecker 获取默认健康检查器
func GetHealthChecker() *HealthChecker {
	return defaultChecker
}

// HealthCheck 执行健康检查
func HealthCheck(ctx context.Context) error {
	if defaultManager == nil {
		return fmt.Errorf("database not initialized")
	}
	return defaultManager.HealthCheck(ctx)
}

// QuickHealthCheck 快速健康检查
func QuickHealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return HealthCheck(ctx)
}

// GetMetrics 获取数据库指标
func GetMetrics() *Metrics {
	if defaultManager != nil {
		return defaultManager.GetMetrics()
	}
	return nil
}

// GetConnectionInfo 获取连接信息
func GetConnectionInfo() map[string]interface{} {
	if defaultManager != nil {
		return defaultManager.GetConnectionInfo()
	}
	return map[string]interface{}{"status": "not_initialized"}
}

// IsInitialized 检查是否已初始化
func IsInitialized() bool {
	return defaultManager != nil
}

// IsHealthy 检查是否健康
func IsHealthy() bool {
	if defaultChecker != nil {
		return defaultChecker.IsHealthy()
	}
	return false
}

// WaitForHealthy 等待数据库变为健康状态
func WaitForHealthy(timeout time.Duration) error {
	if defaultChecker == nil {
		return fmt.Errorf("health checker not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return defaultChecker.WaitForHealthy(ctx, time.Second)
}

// ExecuteWithMetrics 执行查询并记录指标
func ExecuteWithMetrics(ctx context.Context, fn func(*DB) error) error {
	if defaultManager == nil {
		return fmt.Errorf("database not initialized")
	}

	db := defaultManager.GetRawDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	start := time.Now()
	err := fn(db)
	duration := time.Since(start)

	// 记录指标
	metrics := defaultManager.GetMetrics()
	if err != nil {
		metrics.RecordQueryError()
	} else {
		metrics.RecordQuery(duration)
	}

	return err
}

// WithTransaction 执行事务
func WithTransaction(ctx context.Context, fn func(*DB) error) error {
	if defaultManager == nil {
		return fmt.Errorf("database not initialized")
	}

	db := defaultManager.GetRawDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	start := time.Now()
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建一个临时的 DB 实例用于事务
		txDB := &DB{
			DB:     tx,
			config: db.config,
			sqlDB:  db.sqlDB,
		}
		return fn(txDB)
	})
	duration := time.Since(start)

	// 记录事务指标
	metrics := defaultManager.GetMetrics()
	if err != nil {
		metrics.RecordTransactionError()
	} else {
		metrics.RecordTransaction(duration)
	}

	return err
}

// Reconnect 重新连接数据库
func Reconnect() error {
	if defaultManager == nil {
		return fmt.Errorf("database not initialized")
	}
	return defaultManager.Reconnect()
}

// GetStatus 获取数据库状态摘要
func GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"initialized": IsInitialized(),
		"healthy":     IsHealthy(),
		"timestamp":   time.Now(),
	}

	if defaultManager != nil {
		status["connection"] = defaultManager.GetConnectionInfo()

		if metrics := defaultManager.GetMetrics(); metrics != nil {
			status["metrics"] = metrics.GetSummary()
		}
	}

	if defaultChecker != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		healthResult := defaultChecker.QuickCheck(ctx)
		status["health_check"] = map[string]interface{}{
			"status":   healthResult.Status,
			"message":  healthResult.Message,
			"duration": healthResult.Duration,
		}
	}

	return status
}

// ValidateConfig 验证数据库配置
func ValidateConfig(cfg *config.DatabaseConfig) error {
	if cfg == nil {
		return fmt.Errorf("database config is nil")
	}

	if cfg.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", cfg.Port)
	}

	if cfg.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if cfg.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive")
	}

	if cfg.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive")
	}

	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
	}

	if cfg.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be positive")
	}

	return nil
}

// TestConnection 测试数据库连接（不初始化全局实例）
func TestConnection(cfg *config.DatabaseConfig) error {
	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 创建临时连接进行测试
	db, err := NewDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	defer db.Close()

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	applogger.Info("Database connection test successful")
	return nil
}

// MustInitialize 初始化数据库包，失败时 panic
func MustInitialize(cfg *config.DatabaseConfig) {
	if err := Initialize(cfg); err != nil {
		applogger.Fatal("Failed to initialize database package")
		panic(err)
	}
}

// MustGetDB 获取数据库实例，未初始化时 panic
func MustGetDB() *DB {
	db := GetDB()
	if db == nil {
		panic("database not initialized")
	}
	return db
}

// MustGetManager 获取数据库管理器，未初始化时 panic
func MustGetManager() *Manager {
	manager := GetManager()
	if manager == nil {
		panic("database manager not initialized")
	}
	return manager
}
