package services

import (
	"context"
	"time"

	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/redis"
)

// CacheServiceManager 缓存服务管理器
type CacheServiceManager struct {
	LeaderboardCache    LeaderboardCacheService
	InvalidationService LeaderboardInvalidationService
	MonitoringService   CacheMonitoringService
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Leaderboard LeaderboardCacheConfig `mapstructure:"leaderboard"`
	Monitoring  CacheMonitoringConfig  `mapstructure:"monitoring"`
}

// NewCacheServiceManager 创建缓存服务管理器
func NewCacheServiceManager(
	userRepo user.Repository,
	redisClient redis.CacheService,
	config CacheConfig,
) *CacheServiceManager {
	// 创建排行榜缓存服务
	leaderboardCache := NewLeaderboardCacheService(userRepo, redisClient, config.Leaderboard)

	// 创建缓存失效服务
	invalidationService := NewLeaderboardInvalidationService(leaderboardCache)

	// 创建监控服务
	monitoringService := NewCacheMonitoringService(leaderboardCache, redisClient, config.Monitoring)

	return &CacheServiceManager{
		LeaderboardCache:    leaderboardCache,
		InvalidationService: invalidationService,
		MonitoringService:   monitoringService,
	}
}

// Start 启动缓存服务
func (m *CacheServiceManager) Start(ctx context.Context) error {
	logger.Info("Starting cache services...")

	// 启动监控服务
	m.MonitoringService.StartMonitoring(ctx)

	// 启动定时刷新
	m.LeaderboardCache.StartScheduledRefresh(ctx)

	// 预热缓存
	if err := m.LeaderboardCache.PrewarmCache(ctx); err != nil {
		logger.Errorf("Failed to prewarm cache: %v", err)
		// 不返回错误，因为预热失败不应该阻止服务启动
	}

	logger.Info("Cache services started successfully")
	return nil
}

// Stop 停止缓存服务
func (m *CacheServiceManager) Stop() error {
	logger.Info("Stopping cache services...")

	// 停止定时刷新
	m.LeaderboardCache.StopScheduledRefresh()

	// 停止监控服务
	m.MonitoringService.StopMonitoring()

	// 刷新所有计划的失效操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.InvalidationService.FlushScheduledInvalidations(ctx); err != nil {
		logger.Errorf("Failed to flush scheduled invalidations: %v", err)
	}

	logger.Info("Cache services stopped successfully")
	return nil
}

// GetStats 获取所有缓存统计信息
func (m *CacheServiceManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"leaderboard": m.LeaderboardCache.GetCacheStats(),
		"monitoring":  m.MonitoringService.GetMetrics(),
	}
}

// HealthCheck 缓存健康检查
func (m *CacheServiceManager) HealthCheck() map[string]interface{} {
	stats := m.LeaderboardCache.GetCacheStats()
	metrics := m.MonitoringService.GetMetrics()

	return map[string]interface{}{
		"status": "ok",
		"leaderboard_cache": map[string]interface{}{
			"hit_rate":       stats.HitRate,
			"total_requests": stats.TotalRequests,
			"healthy":        m.MonitoringService.CheckHitRateThreshold(),
		},
		"system_health": metrics.SystemHealth,
		"last_updated":  stats.LastUpdated,
	}
}

// InvalidateOnUserPointsChange 用户积分变化时的缓存失效
func (m *CacheServiceManager) InvalidateOnUserPointsChange(ctx context.Context, userID uint, tournament string) error {
	return m.InvalidationService.InvalidateOnPointsUpdate(ctx, userID, tournament)
}

// InvalidateOnMatchComplete 比赛完成时的缓存失效
func (m *CacheServiceManager) InvalidateOnMatchComplete(ctx context.Context, matchID uint, tournament string) error {
	return m.InvalidationService.InvalidateOnMatchComplete(ctx, matchID, tournament)
}

// ScheduleBatchInvalidation 计划批量失效（用于大量积分更新场景）
func (m *CacheServiceManager) ScheduleBatchInvalidation(tournaments []string, delay time.Duration) {
	for _, tournament := range tournaments {
		m.InvalidationService.ScheduleInvalidation(tournament, delay)
	}
}
