package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/logger"
	"backend-go/pkg/redis"
)

// LeaderboardCacheService 排行榜缓存服务接口
type LeaderboardCacheService interface {
	// GetLeaderboard 获取排行榜（带缓存）
	GetLeaderboard(ctx context.Context, tournament string) ([]user.LeaderboardEntry, error)

	// InvalidateLeaderboard 使排行榜缓存失效
	InvalidateLeaderboard(ctx context.Context, tournament string) error

	// PrewarmCache 预热缓存
	PrewarmCache(ctx context.Context) error

	// RefreshCache 刷新缓存
	RefreshCache(ctx context.Context, tournament string) error

	// GetCacheStats 获取缓存统计信息
	GetCacheStats() CacheStats

	// StartScheduledRefresh 启动定时刷新
	StartScheduledRefresh(ctx context.Context)

	// StopScheduledRefresh 停止定时刷新
	StopScheduledRefresh()
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalRequests int64     `json:"total_requests"`
	CacheHits     int64     `json:"cache_hits"`
	CacheMisses   int64     `json:"cache_misses"`
	HitRate       float64   `json:"hit_rate"`
	LastUpdated   time.Time `json:"last_updated"`
}

// leaderboardCacheService 排行榜缓存服务实现
type leaderboardCacheService struct {
	userRepo     user.Repository
	cacheService redis.CacheService

	// 缓存配置
	cacheExpiration time.Duration
	refreshInterval time.Duration

	// 统计信息
	stats      CacheStats
	statsMutex sync.RWMutex

	// 定时刷新控制
	refreshTicker *time.Ticker
	stopChan      chan struct{}
	refreshMutex  sync.Mutex
}

// LeaderboardCacheConfig 排行榜缓存配置
type LeaderboardCacheConfig struct {
	CacheExpiration time.Duration `mapstructure:"cache_expiration"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// NewLeaderboardCacheService 创建排行榜缓存服务
func NewLeaderboardCacheService(
	userRepo user.Repository,
	cacheService redis.CacheService,
	config LeaderboardCacheConfig,
) LeaderboardCacheService {
	// 设置默认配置
	if config.CacheExpiration == 0 {
		config.CacheExpiration = 5 * time.Minute // 5分钟过期
	}
	if config.RefreshInterval == 0 {
		config.RefreshInterval = 2 * time.Minute // 2分钟刷新一次
	}

	return &leaderboardCacheService{
		userRepo:        userRepo,
		cacheService:    cacheService,
		cacheExpiration: config.CacheExpiration,
		refreshInterval: config.RefreshInterval,
		stats: CacheStats{
			LastUpdated: time.Now(),
		},
		stopChan: make(chan struct{}),
	}
}

// GetLeaderboard 获取排行榜（带缓存）
func (s *leaderboardCacheService) GetLeaderboard(ctx context.Context, tournament string) ([]user.LeaderboardEntry, error) {
	s.incrementTotalRequests()

	cacheKey := s.getLeaderboardCacheKey(tournament)

	// 尝试从缓存获取
	var entries []user.LeaderboardEntry
	err := s.cacheService.GetJSON(ctx, cacheKey, &entries)
	if err == nil {
		s.incrementCacheHits()
		logger.Debugf("Leaderboard cache hit for tournament: %s", tournament)
		return entries, nil
	}

	// 缓存未命中，从数据库获取
	if err != redis.ErrKeyNotFound {
		logger.Warnf("Failed to get leaderboard from cache: %v", err)
	}

	s.incrementCacheMisses()
	logger.Debugf("Leaderboard cache miss for tournament: %s", tournament)

	// 从数据库获取数据
	entries, err = s.userRepo.GetLeaderboard(ctx, tournament, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard from database: %w", err)
	}

	// 设置缓存（异步进行，不阻塞响应）
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.cacheService.SetJSON(cacheCtx, cacheKey, entries, s.cacheExpiration); err != nil {
			logger.Errorf("Failed to set leaderboard cache for tournament %s: %v", tournament, err)
		} else {
			logger.Debugf("Leaderboard cached for tournament: %s", tournament)
		}
	}()

	return entries, nil
}

// InvalidateLeaderboard 使排行榜缓存失效
func (s *leaderboardCacheService) InvalidateLeaderboard(ctx context.Context, tournament string) error {
	cacheKey := s.getLeaderboardCacheKey(tournament)

	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		logger.Errorf("Failed to invalidate leaderboard cache for tournament %s: %v", tournament, err)
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	logger.Infof("Leaderboard cache invalidated for tournament: %s", tournament)
	return nil
}

// PrewarmCache 预热缓存
func (s *leaderboardCacheService) PrewarmCache(ctx context.Context) error {
	tournaments := []string{"GLOBAL", "SPRING", "SUMMER", "AUTUMN", "WINTER"}

	logger.Info("Starting leaderboard cache prewarming...")

	for _, tournament := range tournaments {
		if err := s.RefreshCache(ctx, tournament); err != nil {
			logger.Errorf("Failed to prewarm cache for tournament %s: %v", tournament, err)
			continue
		}
		logger.Debugf("Cache prewarmed for tournament: %s", tournament)
	}

	logger.Info("Leaderboard cache prewarming completed")
	return nil
}

// RefreshCache 刷新缓存
func (s *leaderboardCacheService) RefreshCache(ctx context.Context, tournament string) error {
	// 从数据库获取最新数据
	entries, err := s.userRepo.GetLeaderboard(ctx, tournament, 50)
	if err != nil {
		return fmt.Errorf("failed to get leaderboard from database: %w", err)
	}

	// 更新缓存
	cacheKey := s.getLeaderboardCacheKey(tournament)
	if err := s.cacheService.SetJSON(ctx, cacheKey, entries, s.cacheExpiration); err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	logger.Debugf("Leaderboard cache refreshed for tournament: %s", tournament)
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *leaderboardCacheService) GetCacheStats() CacheStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	stats := s.stats
	if stats.TotalRequests > 0 {
		stats.HitRate = float64(stats.CacheHits) / float64(stats.TotalRequests) * 100
	}

	return stats
}

// StartScheduledRefresh 启动定时刷新
func (s *leaderboardCacheService) StartScheduledRefresh(ctx context.Context) {
	s.refreshMutex.Lock()
	defer s.refreshMutex.Unlock()

	if s.refreshTicker != nil {
		return // 已经启动
	}

	s.refreshTicker = time.NewTicker(s.refreshInterval)

	go func() {
		logger.Infof("Leaderboard cache scheduled refresh started (interval: %v)", s.refreshInterval)

		for {
			select {
			case <-s.refreshTicker.C:
				s.performScheduledRefresh(ctx)
			case <-s.stopChan:
				logger.Info("Leaderboard cache scheduled refresh stopped")
				return
			}
		}
	}()
}

// StopScheduledRefresh 停止定时刷新
func (s *leaderboardCacheService) StopScheduledRefresh() {
	s.refreshMutex.Lock()
	defer s.refreshMutex.Unlock()

	if s.refreshTicker != nil {
		s.refreshTicker.Stop()
		s.refreshTicker = nil
	}

	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}

// performScheduledRefresh 执行定时刷新
func (s *leaderboardCacheService) performScheduledRefresh(ctx context.Context) {
	tournaments := []string{"GLOBAL", "SPRING", "SUMMER", "AUTUMN", "WINTER"}

	logger.Debug("Performing scheduled leaderboard cache refresh...")

	for _, tournament := range tournaments {
		refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		if err := s.RefreshCache(refreshCtx, tournament); err != nil {
			logger.Errorf("Failed to refresh cache for tournament %s during scheduled refresh: %v", tournament, err)
		}

		cancel()
	}

	// 记录统计信息更新时间
	s.statsMutex.Lock()
	s.stats.LastUpdated = time.Now()
	s.statsMutex.Unlock()

	logger.Debug("Scheduled leaderboard cache refresh completed")
}

// getLeaderboardCacheKey 获取排行榜缓存键
func (s *leaderboardCacheService) getLeaderboardCacheKey(tournament string) string {
	return fmt.Sprintf("leaderboard:%s", tournament)
}

// incrementTotalRequests 增加总请求数
func (s *leaderboardCacheService) incrementTotalRequests() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.TotalRequests++
}

// incrementCacheHits 增加缓存命中数
func (s *leaderboardCacheService) incrementCacheHits() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.CacheHits++
}

// incrementCacheMisses 增加缓存未命中数
func (s *leaderboardCacheService) incrementCacheMisses() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	s.stats.CacheMisses++
}
