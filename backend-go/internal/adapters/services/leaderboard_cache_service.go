package services

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/core/domain/leaderboard"
	"backend-go/pkg/redis"
)

// leaderboardCacheService 排行榜缓存服务实现
type leaderboardCacheService struct {
	cache redis.CacheService
}

// NewLeaderboardCacheService 创建排行榜缓存服务
func NewLeaderboardCacheService(cache redis.CacheService) leaderboard.CacheService {
	return &leaderboardCacheService{
		cache: cache,
	}
}

// 缓存键常量
const (
	leaderboardKeyPrefix = "leaderboard"
	userRankKeyPrefix    = "user_rank"
	statsKeyPrefix       = "leaderboard_stats"
	cacheExpiration      = 5 * time.Minute // 5分钟缓存过期时间
)

// buildLeaderboardKey 构建排行榜缓存键
func (s *leaderboardCacheService) buildLeaderboardKey(tournament string) string {
	return fmt.Sprintf("%s:%s", leaderboardKeyPrefix, tournament)
}

// buildUserRankKey 构建用户排名缓存键
func (s *leaderboardCacheService) buildUserRankKey(userID uint, tournament string) string {
	return fmt.Sprintf("%s:%d:%s", userRankKeyPrefix, userID, tournament)
}

// buildStatsKey 构建统计缓存键
func (s *leaderboardCacheService) buildStatsKey(tournament string) string {
	return fmt.Sprintf("%s:%s", statsKeyPrefix, tournament)
}

// GetLeaderboard 从缓存获取排行榜
func (s *leaderboardCacheService) GetLeaderboard(ctx context.Context, tournament string) ([]leaderboard.LeaderboardEntry, error) {
	key := s.buildLeaderboardKey(tournament)

	var entries []leaderboard.LeaderboardEntry
	err := s.cache.GetJSON(ctx, key, &entries)
	if err != nil {
		if err == redis.ErrKeyNotFound {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("获取排行榜缓存失败: %w", err)
	}

	return entries, nil
}

// SetLeaderboard 设置排行榜缓存
func (s *leaderboardCacheService) SetLeaderboard(ctx context.Context, tournament string, entries []leaderboard.LeaderboardEntry) error {
	key := s.buildLeaderboardKey(tournament)

	err := s.cache.SetJSON(ctx, key, entries, cacheExpiration)
	if err != nil {
		return fmt.Errorf("设置排行榜缓存失败: %w", err)
	}

	return nil
}

// GetUserRank 从缓存获取用户排名
func (s *leaderboardCacheService) GetUserRank(ctx context.Context, userID uint, tournament string) (*leaderboard.UserRankInfo, error) {
	key := s.buildUserRankKey(userID, tournament)

	var rankInfo leaderboard.UserRankInfo
	err := s.cache.GetJSON(ctx, key, &rankInfo)
	if err != nil {
		if err == redis.ErrKeyNotFound {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("获取用户排名缓存失败: %w", err)
	}

	return &rankInfo, nil
}

// SetUserRank 设置用户排名缓存
func (s *leaderboardCacheService) SetUserRank(ctx context.Context, userID uint, tournament string, rankInfo *leaderboard.UserRankInfo) error {
	key := s.buildUserRankKey(userID, tournament)

	err := s.cache.SetJSON(ctx, key, rankInfo, cacheExpiration)
	if err != nil {
		return fmt.Errorf("设置用户排名缓存失败: %w", err)
	}

	return nil
}

// InvalidateLeaderboard 使排行榜缓存失效
func (s *leaderboardCacheService) InvalidateLeaderboard(ctx context.Context, tournament string) error {
	key := s.buildLeaderboardKey(tournament)

	err := s.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("使排行榜缓存失效失败: %w", err)
	}

	return nil
}

// InvalidateUserRank 使用户排名缓存失效
func (s *leaderboardCacheService) InvalidateUserRank(ctx context.Context, userID uint, tournament string) error {
	key := s.buildUserRankKey(userID, tournament)

	err := s.cache.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("使用户排名缓存失效失败: %w", err)
	}

	return nil
}

// GetLeaderboardStats 从缓存获取排行榜统计
func (s *leaderboardCacheService) GetLeaderboardStats(ctx context.Context, tournament string) (*leaderboard.LeaderboardStats, error) {
	key := s.buildStatsKey(tournament)

	var stats leaderboard.LeaderboardStats
	err := s.cache.GetJSON(ctx, key, &stats)
	if err != nil {
		if err == redis.ErrKeyNotFound {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("获取排行榜统计缓存失败: %w", err)
	}

	return &stats, nil
}

// SetLeaderboardStats 设置排行榜统计缓存
func (s *leaderboardCacheService) SetLeaderboardStats(ctx context.Context, tournament string, stats *leaderboard.LeaderboardStats) error {
	key := s.buildStatsKey(tournament)

	err := s.cache.SetJSON(ctx, key, stats, cacheExpiration)
	if err != nil {
		return fmt.Errorf("设置排行榜统计缓存失败: %w", err)
	}

	return nil
}
