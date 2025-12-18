package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/shared/logger"
)

// CacheMonitoringService 缓存监控服务接口
type CacheMonitoringService interface {
	// StartMonitoring 启动缓存监控
	StartMonitoring(ctx context.Context)

	// StopMonitoring 停止缓存监控
	StopMonitoring()

	// GetMetrics 获取缓存指标
	GetMetrics() CacheMetrics

	// CheckHitRateThreshold 检查命中率阈值
	CheckHitRateThreshold() bool

	// GetDetailedReport 获取详细报告
	GetDetailedReport() CacheReport
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	LeaderboardStats CacheStats   `json:"leaderboard_stats"`
	RedisStats       interface{}  `json:"redis_stats"`
	SystemHealth     SystemHealth `json:"system_health"`
	LastCheck        time.Time    `json:"last_check"`
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	RedisConnected  bool    `json:"redis_connected"`
	HitRateHealthy  bool    `json:"hit_rate_healthy"`
	ResponseTimeMs  float64 `json:"response_time_ms"`
	MemoryUsageMB   float64 `json:"memory_usage_mb"`
	ConnectionCount int     `json:"connection_count"`
}

// CacheReport 缓存报告
type CacheReport struct {
	Summary         CacheMetrics  `json:"summary"`
	HourlyStats     []HourlyStats `json:"hourly_stats"`
	Recommendations []string      `json:"recommendations"`
	Alerts          []Alert       `json:"alerts"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

// HourlyStats 每小时统计
type HourlyStats struct {
	Hour        time.Time `json:"hour"`
	Requests    int64     `json:"requests"`
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	HitRate     float64   `json:"hit_rate"`
	AvgResponse float64   `json:"avg_response_ms"`
}

// Alert 告警信息
type Alert struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// cacheMonitoringService 缓存监控服务实现
type cacheMonitoringService struct {
	leaderboardCache LeaderboardCacheService
	redisClient      interface{} // Redis客户端，用于获取系统指标

	// 监控配置
	monitorInterval  time.Duration
	hitRateThreshold float64

	// 监控状态
	monitoring    bool
	monitorTicker *time.Ticker
	stopChan      chan struct{}
	monitorMutex  sync.RWMutex

	// 历史数据
	hourlyStats []HourlyStats
	alerts      []Alert
	statsMutex  sync.RWMutex

	// 当前指标
	currentMetrics CacheMetrics
	metricsMutex   sync.RWMutex
}

// CacheMonitoringConfig 缓存监控配置
type CacheMonitoringConfig struct {
	MonitorInterval  time.Duration `mapstructure:"monitor_interval"`
	HitRateThreshold float64       `mapstructure:"hit_rate_threshold"`
}

// NewCacheMonitoringService 创建缓存监控服务
func NewCacheMonitoringService(
	leaderboardCache LeaderboardCacheService,
	redisClient interface{},
	config CacheMonitoringConfig,
) CacheMonitoringService {
	// 设置默认配置
	if config.MonitorInterval == 0 {
		config.MonitorInterval = 1 * time.Minute
	}
	if config.HitRateThreshold == 0 {
		config.HitRateThreshold = 90.0 // 90% 命中率阈值
	}

	return &cacheMonitoringService{
		leaderboardCache: leaderboardCache,
		redisClient:      redisClient,
		monitorInterval:  config.MonitorInterval,
		hitRateThreshold: config.HitRateThreshold,
		stopChan:         make(chan struct{}),
		hourlyStats:      make([]HourlyStats, 0),
		alerts:           make([]Alert, 0),
	}
}

// StartMonitoring 启动缓存监控
func (s *cacheMonitoringService) StartMonitoring(ctx context.Context) {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	if s.monitoring {
		return // 已经在监控中
	}

	s.monitoring = true
	s.monitorTicker = time.NewTicker(s.monitorInterval)

	go func() {
		logger.Infof("Cache monitoring started (interval: %v, hit rate threshold: %.1f%%)",
			s.monitorInterval, s.hitRateThreshold)

		// 立即执行一次监控
		s.performMonitoring(ctx)

		for {
			select {
			case <-s.monitorTicker.C:
				s.performMonitoring(ctx)
			case <-s.stopChan:
				logger.Info("Cache monitoring stopped")
				return
			}
		}
	}()
}

// StopMonitoring 停止缓存监控
func (s *cacheMonitoringService) StopMonitoring() {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	if !s.monitoring {
		return
	}

	s.monitoring = false
	if s.monitorTicker != nil {
		s.monitorTicker.Stop()
		s.monitorTicker = nil
	}

	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}

// GetMetrics 获取缓存指标
func (s *cacheMonitoringService) GetMetrics() CacheMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()
	return s.currentMetrics
}

// CheckHitRateThreshold 检查命中率阈值
func (s *cacheMonitoringService) CheckHitRateThreshold() bool {
	metrics := s.GetMetrics()
	return metrics.LeaderboardStats.HitRate >= s.hitRateThreshold
}

// GetDetailedReport 获取详细报告
func (s *cacheMonitoringService) GetDetailedReport() CacheReport {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	metrics := s.GetMetrics()

	report := CacheReport{
		Summary:         metrics,
		HourlyStats:     make([]HourlyStats, len(s.hourlyStats)),
		Recommendations: s.generateRecommendations(metrics),
		Alerts:          make([]Alert, len(s.alerts)),
		GeneratedAt:     time.Now(),
	}

	copy(report.HourlyStats, s.hourlyStats)
	copy(report.Alerts, s.alerts)

	return report
}

// performMonitoring 执行监控检查
func (s *cacheMonitoringService) performMonitoring(ctx context.Context) {
	start := time.Now()

	// 获取排行榜缓存统计
	leaderboardStats := s.leaderboardCache.GetCacheStats()

	// 获取Redis统计（如果可用）
	var redisStats interface{}
	// TODO: 实现Redis统计获取

	// 检查系统健康状态
	systemHealth := s.checkSystemHealth(ctx)

	// 更新当前指标
	metrics := CacheMetrics{
		LeaderboardStats: leaderboardStats,
		RedisStats:       redisStats,
		SystemHealth:     systemHealth,
		LastCheck:        time.Now(),
	}

	s.metricsMutex.Lock()
	s.currentMetrics = metrics
	s.metricsMutex.Unlock()

	// 记录每小时统计
	s.recordHourlyStats(leaderboardStats)

	// 检查告警条件
	s.checkAlerts(metrics)

	monitorDuration := time.Since(start)
	logger.Debugf("Cache monitoring completed in %v (hit rate: %.1f%%)",
		monitorDuration, leaderboardStats.HitRate)
}

// checkSystemHealth 检查系统健康状态
func (s *cacheMonitoringService) checkSystemHealth(ctx context.Context) SystemHealth {
	health := SystemHealth{
		RedisConnected: true, // TODO: 实际检查Redis连接
		HitRateHealthy: s.CheckHitRateThreshold(),
	}

	// 测试响应时间
	start := time.Now()
	_, err := s.leaderboardCache.GetLeaderboard(ctx, "GLOBAL")
	if err != nil {
		health.RedisConnected = false
		health.ResponseTimeMs = -1
	} else {
		health.ResponseTimeMs = float64(time.Since(start).Nanoseconds()) / 1e6
	}

	// TODO: 获取内存使用情况和连接数
	health.MemoryUsageMB = 0
	health.ConnectionCount = 0

	return health
}

// recordHourlyStats 记录每小时统计
func (s *cacheMonitoringService) recordHourlyStats(stats CacheStats) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	now := time.Now()
	currentHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	// 检查是否需要创建新的小时记录
	if len(s.hourlyStats) == 0 || s.hourlyStats[len(s.hourlyStats)-1].Hour.Before(currentHour) {
		hourlyStats := HourlyStats{
			Hour:        currentHour,
			Requests:    stats.TotalRequests,
			Hits:        stats.CacheHits,
			Misses:      stats.CacheMisses,
			HitRate:     stats.HitRate,
			AvgResponse: 0, // TODO: 计算平均响应时间
		}

		s.hourlyStats = append(s.hourlyStats, hourlyStats)

		// 保持最近24小时的数据
		if len(s.hourlyStats) > 24 {
			s.hourlyStats = s.hourlyStats[1:]
		}
	} else {
		// 更新当前小时的统计
		lastIndex := len(s.hourlyStats) - 1
		s.hourlyStats[lastIndex].Requests = stats.TotalRequests
		s.hourlyStats[lastIndex].Hits = stats.CacheHits
		s.hourlyStats[lastIndex].Misses = stats.CacheMisses
		s.hourlyStats[lastIndex].HitRate = stats.HitRate
	}
}

// checkAlerts 检查告警条件
func (s *cacheMonitoringService) checkAlerts(metrics CacheMetrics) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	now := time.Now()

	// 检查命中率告警
	if metrics.LeaderboardStats.HitRate < s.hitRateThreshold && metrics.LeaderboardStats.TotalRequests > 10 {
		alert := Alert{
			Level:     "warning",
			Message:   fmt.Sprintf("Cache hit rate (%.1f%%) is below threshold (%.1f%%)", metrics.LeaderboardStats.HitRate, s.hitRateThreshold),
			Timestamp: now,
			Resolved:  false,
		}
		s.alerts = append(s.alerts, alert)
		logger.Warnf("Cache hit rate alert: %s", alert.Message)
	}

	// 检查Redis连接告警
	if !metrics.SystemHealth.RedisConnected {
		alert := Alert{
			Level:     "critical",
			Message:   "Redis connection is not available",
			Timestamp: now,
			Resolved:  false,
		}
		s.alerts = append(s.alerts, alert)
		logger.Errorf("Redis connection alert: %s", alert.Message)
	}

	// 检查响应时间告警
	if metrics.SystemHealth.ResponseTimeMs > 1000 { // 1秒阈值
		alert := Alert{
			Level:     "warning",
			Message:   fmt.Sprintf("Cache response time (%.1fms) is high", metrics.SystemHealth.ResponseTimeMs),
			Timestamp: now,
			Resolved:  false,
		}
		s.alerts = append(s.alerts, alert)
		logger.Warnf("Response time alert: %s", alert.Message)
	}

	// 保持最近100个告警
	if len(s.alerts) > 100 {
		s.alerts = s.alerts[len(s.alerts)-100:]
	}
}

// generateRecommendations 生成优化建议
func (s *cacheMonitoringService) generateRecommendations(metrics CacheMetrics) []string {
	var recommendations []string

	// 基于命中率的建议
	if metrics.LeaderboardStats.HitRate < 80 {
		recommendations = append(recommendations, "Consider increasing cache expiration time to improve hit rate")
		recommendations = append(recommendations, "Implement cache prewarming for frequently accessed tournaments")
	}

	// 基于响应时间的建议
	if metrics.SystemHealth.ResponseTimeMs > 500 {
		recommendations = append(recommendations, "Consider optimizing database queries for leaderboard data")
		recommendations = append(recommendations, "Review Redis configuration and network latency")
	}

	// 基于请求量的建议
	if metrics.LeaderboardStats.TotalRequests > 10000 {
		recommendations = append(recommendations, "Consider implementing multiple cache layers")
		recommendations = append(recommendations, "Monitor Redis memory usage and consider scaling")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Cache performance is optimal")
	}

	return recommendations
}
