// Package database provides MySQL database connection management and utilities.
//
// This package implements a robust database layer using GORM ORM with MySQL driver,
// featuring connection pooling, health checks, transaction support, and comprehensive
// logging capabilities. It follows Go best practices for database management and
// provides a clean API for database operations.
//
// Key Features:
//   - Connection pooling with configurable parameters
//   - Health monitoring and metrics collection
//   - Transaction support with context
//   - Structured logging with different levels
//   - Graceful connection management
//
// Example usage:
//
//	cfg := &config.DatabaseConfig{
//		Host:     "localhost",
//		Port:     3306,
//		Database: "myapp",
//		Username: "user",
//		Password: "pass",
//	}
//
//	db, err := database.NewDB(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Use the database
//	var users []User
//	err = db.Find(&users).Error
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"backend-go/internal/config"
	applogger "backend-go/internal/shared/logger"
)

// DB represents a database connection wrapper that embeds GORM's DB instance
// and provides additional functionality for connection management, health checks,
// and metrics collection.
//
// The DB struct maintains references to both the GORM instance and the underlying
// sql.DB for direct access when needed. It also stores the configuration used
// to establish the connection for reference and debugging purposes.
type DB struct {
	*gorm.DB                        // Embedded GORM database instance
	config   *config.DatabaseConfig // Database configuration
	sqlDB    *sql.DB                // Underlying SQL database connection
}

// NewDB creates a new database connection using the provided configuration.
//
// This function establishes a connection to MySQL database using GORM ORM,
// configures connection pooling parameters, and performs initial health checks.
// It returns a DB instance that wraps the GORM database with additional utilities.
//
// Parameters:
//   - cfg: Database configuration containing connection parameters, pool settings,
//     and other database-specific options. Must not be nil.
//
// Returns:
//   - *DB: A database wrapper instance with embedded GORM DB and utilities
//   - error: Any error that occurred during connection establishment
//
// The function performs the following operations:
//  1. Validates the provided configuration
//  2. Creates GORM configuration with optimized settings
//  3. Establishes MySQL connection with proper driver configuration
//  4. Configures connection pool parameters
//  5. Performs initial connectivity test
//  6. Sets up logging and monitoring
//
// Example:
//
//	cfg := &config.DatabaseConfig{
//		Host:            "localhost",
//		Port:            3306,
//		Database:        "myapp",
//		Username:        "user",
//		Password:        "password",
//		MaxOpenConns:    20,
//		MaxIdleConns:    10,
//		ConnMaxLifetime: time.Hour,
//	}
//
//	db, err := NewDB(cfg)
//	if err != nil {
//		return fmt.Errorf("failed to connect to database: %w", err)
//	}
//	defer db.Close()
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}

	// 创建 GORM 配置
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",    // 表名前缀
			SingularTable: false, // 使用复数表名
			NoLowerCase:   false, // 使用小写
		},
		Logger:                                   createGormLogger(cfg),
		DisableForeignKeyConstraintWhenMigrating: false, // 启用外键约束
		SkipDefaultTransaction:                   false, // 启用默认事务
		PrepareStmt:                              true,  // 启用预编译语句
		CreateBatchSize:                          1000,  // 批量创建大小
	}

	// 创建 MySQL 驱动配置
	mysqlConfig := mysql.Config{
		DSN:                       cfg.GetDSN(),
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  false, // 启用 datetime 精度
		DontSupportRenameIndex:    false, // 支持重命名索引
		DontSupportRenameColumn:   false, // 支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 sql.DB 实例
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	if err := configureConnectionPool(sqlDB, cfg); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbInstance := &DB{
		DB:     db,
		config: cfg,
		sqlDB:  sqlDB,
	}

	applogger.Info("Database connection established successfully")

	return dbInstance, nil
}

// createGormLogger 创建 GORM 日志记录器
func createGormLogger(cfg *config.DatabaseConfig) gormlogger.Interface {
	// 根据环境配置日志级别
	env := config.GetEnvironment()

	var logLevel gormlogger.LogLevel
	switch {
	case env.IsDevelopment():
		logLevel = gormlogger.Info
	case env.IsTesting():
		logLevel = gormlogger.Silent
	default:
		logLevel = gormlogger.Warn
	}

	return gormlogger.New(
		applogger.GetLogger(), // 使用应用的标准日志记录器
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,                // 忽略 ErrRecordNotFound 错误
			ParameterizedQueries:      false,               // 在日志中包含参数
			Colorful:                  env.IsDevelopment(), // 开发环境启用颜色
		},
	)
}

// configureConnectionPool 配置数据库连接池
func configureConnectionPool(sqlDB *sql.DB, cfg *config.DatabaseConfig) error {
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// 设置连接最大生存时间
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 设置连接最大空闲时间
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	applogger.Info("Database connection pool configured")

	return nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.sqlDB != nil {
		return db.sqlDB.Close()
	}
	return nil
}

// Ping 测试数据库连接
func (db *DB) Ping(ctx context.Context) error {
	if db.sqlDB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return db.sqlDB.PingContext(ctx)
}

// Stats 获取数据库连接统计信息
func (db *DB) Stats() sql.DBStats {
	if db.sqlDB == nil {
		return sql.DBStats{}
	}
	return db.sqlDB.Stats()
}

// GetSQLDB 获取底层 sql.DB 实例
func (db *DB) GetSQLDB() *sql.DB {
	return db.sqlDB
}

// GetConfig 获取数据库配置
func (db *DB) GetConfig() *config.DatabaseConfig {
	return db.config
}

// WithContext 创建带上下文的数据库会话
func (db *DB) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

// Transaction 执行事务
func (db *DB) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return db.DB.Transaction(fn, opts...)
}

// BeginTx 开始事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) *gorm.DB {
	return db.DB.Begin(opts)
}

// HealthCheck 健康检查
func (db *DB) HealthCheck(ctx context.Context) error {
	// 检查连接是否可用
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查连接池状态
	stats := db.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}

	// 执行简单查询测试
	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query test returned unexpected result: %d", result)
	}

	return nil
}

// GetConnectionInfo 获取连接信息
func (db *DB) GetConnectionInfo() map[string]interface{} {
	stats := db.Stats()

	return map[string]interface{}{
		"host":                db.config.Host,
		"port":                db.config.Port,
		"database":            db.config.Database,
		"max_open_conns":      db.config.MaxOpenConns,
		"max_idle_conns":      db.config.MaxIdleConns,
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// EnableQueryLogging 启用查询日志记录
func (db *DB) EnableQueryLogging() *gorm.DB {
	return db.DB.Session(&gorm.Session{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
}

// DisableQueryLogging 禁用查询日志记录
func (db *DB) DisableQueryLogging() *gorm.DB {
	return db.DB.Session(&gorm.Session{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
}

// SetSlowQueryThreshold 设置慢查询阈值
func (db *DB) SetSlowQueryThreshold(threshold time.Duration) *gorm.DB {
	newLogger := gormlogger.New(
		applogger.GetLogger(),
		gormlogger.Config{
			SlowThreshold:             threshold,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  config.GetEnvironment().IsDevelopment(),
		},
	)

	return db.DB.Session(&gorm.Session{
		Logger: newLogger,
	})
}

// CreateDatabase 创建数据库（如果不存在）
func CreateDatabase(cfg *config.DatabaseConfig) error {
	// 连接到 MySQL 服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Charset)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	// 创建数据库
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s COLLATE %s",
		cfg.Database, cfg.Charset, cfg.Collation)

	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	applogger.Info("Database created or already exists")

	return nil
}

// DropDatabase 删除数据库（谨慎使用）
func DropDatabase(cfg *config.DatabaseConfig) error {
	// 连接到 MySQL 服务器（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Charset)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer db.Close()

	// 删除数据库
	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", cfg.Database)

	if _, err := db.Exec(dropSQL); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	applogger.Warn("Database dropped")
	return nil
}
