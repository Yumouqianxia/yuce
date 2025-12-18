package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/shared/logger"
)

// LeaderboardInvalidationService 排行榜缓存失效服务接口
type LeaderboardInvalidationService interface {
	// InvalidateOnPointsUpdate 当用户积分更新时使缓存失效
	InvalidateOnPointsUpdate(ctx context.Context, userID uint, tournament string) error

	// InvalidateOnMatchComplete 当比赛完成时使缓存失效
	InvalidateOnMatchComplete(ctx context.Context, matchID uint, tournament string) error

	// BatchInvalidate 批量使缓存失效
	BatchInvalidate(ctx context.Context, tournaments []string) error

	// ScheduleInvalidation 计划延迟失效（用于批量更新场景）
	ScheduleInvalidation(tournament string, delay time.Duration)

	// FlushScheduledInvalidations 立即执行所有计划的失效操作
	FlushScheduledInvalidations(ctx context.Context) error
}

// leaderboardInvalidationService 排行榜缓存失效服务实现
type leaderboardInvalidationService struct {
	cacheService LeaderboardCacheService

	// 延迟失效管理
	scheduledInvalidations map[string]*time.Timer
	invalidationMutex      sync.RWMutex
}

// NewLeaderboardInvalidationService 创建排行榜缓存失效服务
func NewLeaderboardInvalidationService(cacheService LeaderboardCacheService) LeaderboardInvalidationService {
	return &leaderboardInvalidationService{
		cacheService:           cacheService,
		scheduledInvalidations: make(map[string]*time.Timer),
	}
}

// InvalidateOnPointsUpdate 当用户积分更新时使缓存失效
func (s *leaderboardInvalidationService) InvalidateOnPointsUpdate(ctx context.Context, userID uint, tournament string) error {
	logger.Debugf("Invalidating leaderboard cache due to points update for user %d in tournament %s", userID, tournament)

	// 使指定锦标赛的缓存失效
	if err := s.cacheService.InvalidateLeaderboard(ctx, tournament); err != nil {
		return fmt.Errorf("failed to invalidate tournament leaderboard: %w", err)
	}

	// 如果不是全局锦标赛，也要使全局排行榜失效
	if tournament != "GLOBAL" {
		if err := s.cacheService.InvalidateLeaderboard(ctx, "GLOBAL"); err != nil {
			logger.Errorf("Failed to invalidate global leaderboard: %v", err)
			// 不返回错误，因为主要的锦标赛缓存已经失效
		}
	}

	return nil
}

// InvalidateOnMatchComplete 当比赛完成时使缓存失效
func (s *leaderboardInvalidationService) InvalidateOnMatchComplete(ctx context.Context, matchID uint, tournament string) error {
	logger.Infof("Invalidating leaderboard cache due to match completion: match %d in tournament %s", matchID, tournament)

	// 比赛完成时，可能影响多个用户的积分，所以需要失效相关的排行榜
	tournaments := []string{tournament}
	if tournament != "GLOBAL" {
		tournaments = append(tournaments, "GLOBAL")
	}

	return s.BatchInvalidate(ctx, tournaments)
}

// BatchInvalidate 批量使缓存失效
func (s *leaderboardInvalidationService) BatchInvalidate(ctx context.Context, tournaments []string) error {
	var errors []error

	for _, tournament := range tournaments {
		if err := s.cacheService.InvalidateLeaderboard(ctx, tournament); err != nil {
			errors = append(errors, fmt.Errorf("failed to invalidate %s: %w", tournament, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch invalidation failed: %v", errors)
	}

	logger.Debugf("Batch invalidated leaderboard cache for tournaments: %v", tournaments)
	return nil
}

// ScheduleInvalidation 计划延迟失效（用于批量更新场景）
func (s *leaderboardInvalidationService) ScheduleInvalidation(tournament string, delay time.Duration) {
	s.invalidationMutex.Lock()
	defer s.invalidationMutex.Unlock()

	// 如果已经有计划的失效操作，取消它
	if existingTimer, exists := s.scheduledInvalidations[tournament]; exists {
		existingTimer.Stop()
	}

	// 创建新的定时器
	timer := time.AfterFunc(delay, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.cacheService.InvalidateLeaderboard(ctx, tournament); err != nil {
			logger.Errorf("Failed to execute scheduled invalidation for tournament %s: %v", tournament, err)
		} else {
			logger.Debugf("Executed scheduled invalidation for tournament: %s", tournament)
		}

		// 清理已完成的定时器
		s.invalidationMutex.Lock()
		delete(s.scheduledInvalidations, tournament)
		s.invalidationMutex.Unlock()
	})

	s.scheduledInvalidations[tournament] = timer
	logger.Debugf("Scheduled invalidation for tournament %s in %v", tournament, delay)
}

// FlushScheduledInvalidations 立即执行所有计划的失效操作
func (s *leaderboardInvalidationService) FlushScheduledInvalidations(ctx context.Context) error {
	s.invalidationMutex.Lock()
	defer s.invalidationMutex.Unlock()

	var tournaments []string

	// 收集所有计划的锦标赛并取消定时器
	for tournament, timer := range s.scheduledInvalidations {
		timer.Stop()
		tournaments = append(tournaments, tournament)
	}

	// 清空计划的失效操作
	s.scheduledInvalidations = make(map[string]*time.Timer)

	if len(tournaments) == 0 {
		return nil
	}

	logger.Infof("Flushing scheduled invalidations for tournaments: %v", tournaments)
	return s.BatchInvalidate(ctx, tournaments)
}
