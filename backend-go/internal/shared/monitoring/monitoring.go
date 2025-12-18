package monitoring

import (
	"net/http"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// MonitoringService 监控服务
type MonitoringService struct {
	config          *config.Config
	healthService   *middleware.HealthService
	businessMetrics *middleware.BusinessMetrics
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(cfg *config.Config) *MonitoringService {
	return &MonitoringService{
		config:          cfg,
		healthService:   middleware.NewHealthService("1.0.0", cfg.External.Monitoring.HealthCheck.Timeout),
		businessMetrics: middleware.GetBusinessMetrics(),
	}
}

// Initialize 初始化监控服务
func (s *MonitoringService) Initialize(db *gorm.DB, redisClient redis.UniversalClient) error {
	logger.Info("Initializing monitoring service...")

	// 添加健康检查器
	if db != nil {
		dbChecker := middleware.NewDatabaseHealthChecker(db, 5*time.Second)
		s.healthService.AddChecker(dbChecker)
		logger.Info("Added database health checker")
	}

	if redisClient != nil {
		redisChecker := middleware.NewRedisHealthChecker(redisClient, 5*time.Second)
		s.healthService.AddChecker(redisChecker)
		logger.Info("Added Redis health checker")
	}

	// 添加内存检查器
	memoryChecker := middleware.NewMemoryHealthChecker(s.config.External.Monitoring.HealthCheck.MaxMemoryMB)
	s.healthService.AddChecker(memoryChecker)
	logger.Info("Added memory health checker")

	logger.Info("Monitoring service initialized successfully")
	return nil
}

// SetupMiddleware 设置监控中间件
func (s *MonitoringService) SetupMiddleware(router *gin.Engine) {
	logger.Info("Setting up monitoring middleware...")

	// 请求ID中间件
	router.Use(middleware.DefaultRequestIDMiddleware())

	// 结构化日志中间件
	router.Use(middleware.StructuredLoggingMiddleware())

	// 日志中间件
	loggingConfig := &middleware.LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/health/detailed",
			"/ready",
			"/live",
			"/metrics",
			"/favicon.ico",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024 * 1024, // 1MB
		SlowThreshold:   s.config.Log.SlowThreshold,
	}
	router.Use(middleware.LoggingMiddleware(loggingConfig))

	// 错误日志中间件
	router.Use(middleware.ErrorLoggingMiddleware())

	// 恢复中间件
	router.Use(middleware.RecoveryLoggingMiddleware())

	// Prometheus指标中间件
	if s.config.External.Monitoring.Prometheus.Enabled {
		metricsConfig := &middleware.MetricsConfig{
			SkipPaths:     s.config.External.Monitoring.Prometheus.SkipPaths,
			NormalizePath: s.config.External.Monitoring.Prometheus.NormalizePath,
			MaxPathLabels: s.config.External.Monitoring.Prometheus.MaxPathLabels,
		}
		router.Use(middleware.MetricsMiddleware(metricsConfig))
		logger.Info("Enabled Prometheus metrics middleware")
	}

	// 审计日志中间件
	router.Use(middleware.AuditMiddleware())

	// 安全日志中间件
	router.Use(middleware.SecurityLoggingMiddleware())

	// 健康检查中间件
	if s.config.External.Monitoring.HealthCheck.Enabled {
		router.Use(middleware.HealthMiddleware(s.healthService))
		router.Use(middleware.ReadinessMiddleware(s.healthService))
		router.Use(middleware.LivenessMiddleware())
		router.Use(middleware.MetricsHealthMiddleware())
		logger.Info("Enabled health check middleware")
	}

	logger.Info("Monitoring middleware setup completed")
}

// SetupRoutes 设置监控路由
func (s *MonitoringService) SetupRoutes(router *gin.Engine) {
	logger.Info("Setting up monitoring routes...")

	// Prometheus指标端点
	if s.config.External.Monitoring.Prometheus.Enabled {
		router.GET(s.config.External.Monitoring.Prometheus.Path, gin.WrapH(promhttp.Handler()))
		logger.Infof("Prometheus metrics available at %s", s.config.External.Monitoring.Prometheus.Path)
	}

	// 健康检查端点已通过中间件处�?
	logger.Info("Health check endpoints:")
	logger.Info("  - /health - Simple health check")
	logger.Info("  - /health/detailed - Detailed health check")
	logger.Info("  - /health/metrics - Health metrics")
	logger.Info("  - /ready - Readiness probe")
	logger.Info("  - /live - Liveness probe")

	logger.Info("Monitoring routes setup completed")
}

// StartupProbe 执行启动探针
func (s *MonitoringService) StartupProbe() error {
	if !s.config.External.Monitoring.HealthCheck.Enabled {
		logger.Info("Health check disabled, skipping startup probe")
		return nil
	}

	return middleware.StartupProbe(
		s.healthService,
		s.config.External.Monitoring.HealthCheck.StartupRetries,
		s.config.External.Monitoring.HealthCheck.StartupInterval,
	)
}

// GetHealthService 获取健康检查服务
func (s *MonitoringService) GetHealthService() *middleware.HealthService {
	return s.healthService
}

// GetBusinessMetrics 获取业务指标
func (s *MonitoringService) GetBusinessMetrics() *middleware.BusinessMetrics {
	return s.businessMetrics
}

// RecordUserRegistration 记录用户注册指标
func (s *MonitoringService) RecordUserRegistration(source string) {
	s.businessMetrics.RecordUserRegistration(source)
	logger.WithFields(map[string]interface{}{
		"metric": "user_registration",
		"source": source,
	}).Info("User registration recorded")
}

// RecordUserLogin 记录用户登录指标
func (s *MonitoringService) RecordUserLogin(success bool) {
	s.businessMetrics.RecordUserLogin(success)
	logger.WithFields(map[string]interface{}{
		"metric":  "user_login",
		"success": success,
	}).Info("User login recorded")
}

// RecordPredictionCreated 记录预测创建指标
func (s *MonitoringService) RecordPredictionCreated(matchType string) {
	s.businessMetrics.RecordPredictionCreated(matchType)
	logger.WithFields(map[string]interface{}{
		"metric":     "prediction_created",
		"match_type": matchType,
	}).Info("Prediction creation recorded")
}

// RecordVoteCreated 记录投票创建指标
func (s *MonitoringService) RecordVoteCreated(predictionType string) {
	s.businessMetrics.RecordVoteCreated(predictionType)
	logger.WithFields(map[string]interface{}{
		"metric":          "vote_created",
		"prediction_type": predictionType,
	}).Info("Vote creation recorded")
}

// RecordCacheOperation 记录缓存操作指标
func (s *MonitoringService) RecordCacheOperation(cacheType string, hit bool) {
	if hit {
		s.businessMetrics.RecordCacheHit(cacheType)
	} else {
		s.businessMetrics.RecordCacheMiss(cacheType)
	}

	logger.WithFields(map[string]interface{}{
		"metric":     "cache_operation",
		"cache_type": cacheType,
		"hit":        hit,
	}).Debug("Cache operation recorded")
}

// RecordDatabaseStats 记录数据库统�?
func (s *MonitoringService) RecordDatabaseStats(open, idle, inUse int) {
	s.businessMetrics.RecordDatabaseConnections(open, idle, inUse)
}

// RecordRedisStats 记录Redis统计
func (s *MonitoringService) RecordRedisStats(active, idle int) {
	s.businessMetrics.RecordRedisConnections(active, idle)
}

// RecordWebSocketStats 记录WebSocket统计
func (s *MonitoringService) RecordWebSocketStats(connections int) {
	// TODO: 实现WebSocket连接数统计
	logger.WithFields(map[string]interface{}{
		"metric":      "websocket_connections",
		"connections": connections,
	}).Debug("WebSocket connections recorded")
}

// PerformanceTimer 性能计时�?
type PerformanceTimer struct {
	operation string
	startTime time.Time
	service   *MonitoringService
}

// StartTimer 开始性能计时
func (s *MonitoringService) StartTimer(operation string) *PerformanceTimer {
	return &PerformanceTimer{
		operation: operation,
		startTime: time.Now(),
		service:   s,
	}
}

// Stop 停止计时并记�?
func (t *PerformanceTimer) Stop() {
	t.StopWithSuccess(true, "")
}

// StopWithError 停止计时并记录错�?
func (t *PerformanceTimer) StopWithError(err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	t.StopWithSuccess(false, errorMsg)
}

// StopWithSuccess 停止计时并记录结�?
func (t *PerformanceTimer) StopWithSuccess(success bool, errorMsg string) {
	duration := time.Since(t.startTime)

	logger.LogPerformance(logger.Performance{
		Operation: t.operation,
		Duration:  duration,
		StartTime: t.startTime,
		EndTime:   time.Now(),
		Success:   success,
		Error:     errorMsg,
	})
}

// MonitoringHandler 监控处理�?
type MonitoringHandler struct {
	service *MonitoringService
}

// NewMonitoringHandler 创建监控处理�?
func NewMonitoringHandler(service *MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		service: service,
	}
}

// GetHealth 获取健康状�?
func (h *MonitoringHandler) GetHealth(c *gin.Context) {
	ctx := c.Request.Context()
	health := h.service.healthService.Check(ctx)

	statusCode := http.StatusOK
	if health.Status == middleware.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == middleware.HealthStatusDegraded {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, health)
}

// GetMetrics 获取指标信息
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	// 这个端点由Prometheus处理器处�?
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

// GetStatus 获取服务状�?
func (h *MonitoringHandler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "prediction-system",
		"version": "1.0.0",
		"status":  "running",
		"uptime":  time.Since(time.Now()).String(), // 这里应该是实际的启动时间
	})
}
