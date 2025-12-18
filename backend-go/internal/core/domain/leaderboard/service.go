package leaderboard

import (
	"context"
)

// Service 排行榜服务接口
type Service interface {
	// GetLeaderboard 获取排行榜
	GetLeaderboard(ctx context.Context, tournament string, limit int) ([]LeaderboardEntry, error)

	// GetUserRank 获取用户排名信息
	GetUserRank(ctx context.Context, userID uint, tournament string) (*UserRankInfo, error)

	// GetLeaderboardStats 获取排行榜统计信息
	GetLeaderboardStats(ctx context.Context, tournament string) (*LeaderboardStats, error)

	// RefreshLeaderboard 刷新排行榜缓存
	RefreshLeaderboard(ctx context.Context, tournament string) error

	// UpdateUserPoints 更新用户积分并刷新排行榜
	UpdateUserPoints(ctx context.Context, userID uint, points int, tournament string) error

	// GetTopUsers 获取前N名用户
	GetTopUsers(ctx context.Context, tournament string, limit int) ([]LeaderboardEntry, error)

	// GetUsersAroundRank 获取指定排名周围的用户
	GetUsersAroundRank(ctx context.Context, tournament string, rank int, radius int) ([]LeaderboardEntry, error)
}

// CacheService 排行榜缓存服务接口
type CacheService interface {
	// GetLeaderboard 从缓存获取排行榜
	GetLeaderboard(ctx context.Context, tournament string) ([]LeaderboardEntry, error)

	// SetLeaderboard 设置排行榜缓存
	SetLeaderboard(ctx context.Context, tournament string, entries []LeaderboardEntry) error

	// GetUserRank 从缓存获取用户排名
	GetUserRank(ctx context.Context, userID uint, tournament string) (*UserRankInfo, error)

	// SetUserRank 设置用户排名缓存
	SetUserRank(ctx context.Context, userID uint, tournament string, rankInfo *UserRankInfo) error

	// InvalidateLeaderboard 使排行榜缓存失效
	InvalidateLeaderboard(ctx context.Context, tournament string) error

	// InvalidateUserRank 使用户排名缓存失效
	InvalidateUserRank(ctx context.Context, userID uint, tournament string) error

	// GetLeaderboardStats 从缓存获取排行榜统计
	GetLeaderboardStats(ctx context.Context, tournament string) (*LeaderboardStats, error)

	// SetLeaderboardStats 设置排行榜统计缓存
	SetLeaderboardStats(ctx context.Context, tournament string, stats *LeaderboardStats) error
}

// Repository 排行榜仓储接口
type Repository interface {
	// GetLeaderboard 从数据库获取排行榜
	GetLeaderboard(ctx context.Context, tournament string, limit int) ([]LeaderboardEntry, error)

	// GetUserRank 从数据库获取用户排名
	GetUserRank(ctx context.Context, userID uint, tournament string) (*UserRankInfo, error)

	// GetLeaderboardStats 从数据库获取排行榜统计
	GetLeaderboardStats(ctx context.Context, tournament string) (*LeaderboardStats, error)

	// GetTopUsers 获取前N名用户
	GetTopUsers(ctx context.Context, tournament string, limit int) ([]LeaderboardEntry, error)

	// GetUsersAroundRank 获取指定排名周围的用户
	GetUsersAroundRank(ctx context.Context, tournament string, rank int, radius int) ([]LeaderboardEntry, error)
}
