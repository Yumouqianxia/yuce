package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"backend-go/internal/core/domain/leaderboard"
)

// leaderboardService 排行榜服务实现
type leaderboardService struct {
	repo         leaderboard.Repository
	cacheService leaderboard.CacheService
	logger       *logrus.Logger
}

// NewLeaderboardService 创建排行榜服务
func NewLeaderboardService(
	repo leaderboard.Repository,
	cacheService leaderboard.CacheService,
	logger *logrus.Logger,
) leaderboard.Service {
	return &leaderboardService{
		repo:         repo,
		cacheService: cacheService,
		logger:       logger,
	}
}

// GetLeaderboard 获取排行榜
func (s *leaderboardService) GetLeaderboard(ctx context.Context, tournament string, limit int) ([]leaderboard.LeaderboardEntry, error) {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	// 尝试从缓存获取
	entries, err := s.cacheService.GetLeaderboard(ctx, tournament)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Warn("获取排行榜缓存失败")
	}

	if entries != nil && len(entries) > 0 {
		s.logger.WithFields(logrus.Fields{
			"tournament": tournament,
			"entries":    len(entries),
			"source":     "cache",
		}).Debug("从缓存获取排行榜成功")

		// 如果需要限制数量，截取结果
		if limit > 0 && len(entries) > limit {
			entries = entries[:limit]
		}

		return entries, nil
	}

	// 缓存未命中，从数据库获取
	entries, err = s.repo.GetLeaderboard(ctx, tournament, limit)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜失败: %w", err)
	}

	// 设置缓存
	if len(entries) > 0 {
		err = s.cacheService.SetLeaderboard(ctx, tournament, entries)
		if err != nil {
			s.logger.WithError(err).WithField("tournament", tournament).Warn("设置排行榜缓存失败")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"tournament": tournament,
		"entries":    len(entries),
		"source":     "database",
	}).Debug("从数据库获取排行榜成功")

	return entries, nil
}

// GetUserRank 获取用户排名信息
func (s *leaderboardService) GetUserRank(ctx context.Context, userID uint, tournament string) (*leaderboard.UserRankInfo, error) {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	// 尝试从缓存获取
	rankInfo, err := s.cacheService.GetUserRank(ctx, userID, tournament)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"tournament": tournament,
		}).Warn("获取用户排名缓存失败")
	}

	if rankInfo != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"tournament": tournament,
			"rank":       rankInfo.Rank,
			"source":     "cache",
		}).Debug("从缓存获取用户排名成功")

		return rankInfo, nil
	}

	// 缓存未命中，从数据库获取
	rankInfo, err = s.repo.GetUserRank(ctx, userID, tournament)
	if err != nil {
		return nil, fmt.Errorf("获取用户排名失败: %w", err)
	}

	// 设置缓存
	if rankInfo != nil {
		err = s.cacheService.SetUserRank(ctx, userID, tournament, rankInfo)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":    userID,
				"tournament": tournament,
			}).Warn("设置用户排名缓存失败")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"tournament": tournament,
		"rank":       rankInfo.Rank,
		"source":     "database",
	}).Debug("从数据库获取用户排名成功")

	return rankInfo, nil
}

// GetLeaderboardStats 获取排行榜统计信息
func (s *leaderboardService) GetLeaderboardStats(ctx context.Context, tournament string) (*leaderboard.LeaderboardStats, error) {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	// 尝试从缓存获取
	stats, err := s.cacheService.GetLeaderboardStats(ctx, tournament)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Warn("获取排行榜统计缓存失败")
	}

	if stats != nil {
		s.logger.WithFields(logrus.Fields{
			"tournament": tournament,
			"source":     "cache",
		}).Debug("从缓存获取排行榜统计成功")

		return stats, nil
	}

	// 缓存未命中，从数据库获取
	stats, err = s.repo.GetLeaderboardStats(ctx, tournament)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜统计失败: %w", err)
	}

	// 设置缓存
	if stats != nil {
		err = s.cacheService.SetLeaderboardStats(ctx, tournament, stats)
		if err != nil {
			s.logger.WithError(err).WithField("tournament", tournament).Warn("设置排行榜统计缓存失败")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"tournament": tournament,
		"source":     "database",
	}).Debug("从数据库获取排行榜统计成功")

	return stats, nil
}

// RefreshLeaderboard 刷新排行榜缓存
func (s *leaderboardService) RefreshLeaderboard(ctx context.Context, tournament string) error {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	s.logger.WithField("tournament", tournament).Info("开始刷新排行榜缓存")

	// 使缓存失效
	err := s.cacheService.InvalidateLeaderboard(ctx, tournament)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Error("使排行榜缓存失效失败")
		return fmt.Errorf("使排行榜缓存失效失败: %w", err)
	}

	// 从数据库重新获取并设置缓存
	entries, err := s.repo.GetLeaderboard(ctx, tournament, 100) // 获取前100名
	if err != nil {
		return fmt.Errorf("刷新排行榜失败: %w", err)
	}

	// 设置新缓存
	err = s.cacheService.SetLeaderboard(ctx, tournament, entries)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Error("设置排行榜缓存失败")
		return fmt.Errorf("设置排行榜缓存失败: %w", err)
	}

	// 刷新统计信息缓存
	stats, err := s.repo.GetLeaderboardStats(ctx, tournament)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Warn("获取排行榜统计失败")
	} else {
		err = s.cacheService.SetLeaderboardStats(ctx, tournament, stats)
		if err != nil {
			s.logger.WithError(err).WithField("tournament", tournament).Warn("设置排行榜统计缓存失败")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"tournament": tournament,
		"entries":    len(entries),
	}).Info("排行榜缓存刷新完成")

	return nil
}

// UpdateUserPoints 更新用户积分并刷新排行榜
func (s *leaderboardService) UpdateUserPoints(ctx context.Context, userID uint, points int, tournament string) error {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"points":     points,
		"tournament": tournament,
	}).Info("更新用户积分并刷新排行榜")

	// 使用户排名缓存失效
	err := s.cacheService.InvalidateUserRank(ctx, userID, tournament)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"tournament": tournament,
		}).Warn("使用户排名缓存失效失败")
	}

	// 刷新排行榜缓存
	err = s.RefreshLeaderboard(ctx, tournament)
	if err != nil {
		s.logger.WithError(err).WithField("tournament", tournament).Error("刷新排行榜失败")
		return fmt.Errorf("刷新排行榜失败: %w", err)
	}

	return nil
}

// GetTopUsers 获取前N名用户
func (s *leaderboardService) GetTopUsers(ctx context.Context, tournament string, limit int) ([]leaderboard.LeaderboardEntry, error) {
	return s.GetLeaderboard(ctx, tournament, limit)
}

// GetUsersAroundRank 获取指定排名周围的用户
func (s *leaderboardService) GetUsersAroundRank(ctx context.Context, tournament string, rank int, radius int) ([]leaderboard.LeaderboardEntry, error) {
	// 验证锦标赛类型
	if !leaderboard.IsValidTournament(tournament) {
		tournament = string(leaderboard.TournamentGlobal)
	}

	return s.repo.GetUsersAroundRank(ctx, tournament, rank, radius)
}
