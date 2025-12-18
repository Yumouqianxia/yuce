package mysql

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/core/domain/leaderboard"
	"backend-go/internal/core/domain/user"
	"gorm.io/gorm"
)

// LeaderboardRepository 排行榜仓储 MySQL 实现
type LeaderboardRepository struct {
	db *gorm.DB
}

// NewLeaderboardRepository 创建排行榜仓储
func NewLeaderboardRepository(db *gorm.DB) leaderboard.Repository {
	return &LeaderboardRepository{
		db: db,
	}
}

// GetLeaderboard 从数据库获取排行榜
func (r *LeaderboardRepository) GetLeaderboard(ctx context.Context, tournament string, limit int) ([]leaderboard.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var users []user.User
	query := r.db.WithContext(ctx).
		Model(&user.User{}).
		Order("points DESC, createdAt ASC").
		Limit(limit)

	// 如果需要按锦标赛过滤，这里可以添加相应的逻辑
	// 目前先返回全局排行榜

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("获取排行榜失败: %w", err)
	}

	// 转换为排行榜条目
	entries := make([]leaderboard.LeaderboardEntry, len(users))
	for i, u := range users {
		entries[i] = leaderboard.LeaderboardEntry{
			UserID:     u.ID,
			Username:   u.Username,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Points:     u.Points,
			Rank:       i + 1,
			Tournament: tournament,
		}
	}

	return entries, nil
}

// GetUserRank 从数据库获取用户排名
func (r *LeaderboardRepository) GetUserRank(ctx context.Context, userID uint, tournament string) (*leaderboard.UserRankInfo, error) {
	// 获取用户信息
	var u user.User
	if err := r.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 计算用户排名
	var rank int64
	err := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("points > ? OR (points = ? AND createdAt < ?)", u.Points, u.Points, u.CreatedAt).
		Count(&rank).Error

	if err != nil {
		return nil, fmt.Errorf("计算用户排名失败: %w", err)
	}

	rank++ // 排名从1开始

	return &leaderboard.UserRankInfo{
		UserID:       u.ID,
		Username:     u.Username,
		Nickname:     u.Nickname,
		Points:       u.Points,
		Rank:         int(rank),
		Tournament:   tournament,
		RankChange:   0, // 排名变化需要额外的逻辑来计算
		PointsChange: 0, // 积分变化需要额外的逻辑来计算
	}, nil
}

// GetLeaderboardStats 从数据库获取排行榜统计
func (r *LeaderboardRepository) GetLeaderboardStats(ctx context.Context, tournament string) (*leaderboard.LeaderboardStats, error) {
	var stats struct {
		TotalUsers   int64   `json:"total_users"`
		TopScore     int     `json:"top_score"`
		AverageScore float64 `json:"average_score"`
	}

	// 使用原生SQL查询确保正确性
	query := `
		SELECT 
			COUNT(*) as total_users,
			COALESCE(MAX(points), 0) as top_score,
			COALESCE(AVG(points), 0) as average_score
		FROM users
	`
	
	err := r.db.WithContext(ctx).Raw(query).Row().Scan(&stats.TotalUsers, &stats.TopScore, &stats.AverageScore)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜统计失败: %w", err)
	}

	return &leaderboard.LeaderboardStats{
		TotalUsers:   int(stats.TotalUsers),
		TopScore:     stats.TopScore,
		AverageScore: stats.AverageScore,
		LastUpdated:  time.Now(),
		Tournament:   tournament,
	}, nil
}

// GetTopUsers 获取前N名用户
func (r *LeaderboardRepository) GetTopUsers(ctx context.Context, tournament string, limit int) ([]leaderboard.LeaderboardEntry, error) {
	return r.GetLeaderboard(ctx, tournament, limit)
}

// GetUsersAroundRank 获取指定排名周围的用户
func (r *LeaderboardRepository) GetUsersAroundRank(ctx context.Context, tournament string, rank int, radius int) ([]leaderboard.LeaderboardEntry, error) {
	if rank <= 0 {
		rank = 1
	}
	if radius <= 0 {
		radius = 5
	}

	// 计算偏移量和限制
	offset := rank - radius - 1
	if offset < 0 {
		offset = 0
	}
	limit := radius*2 + 1

	var users []user.User
	err := r.db.WithContext(ctx).
		Model(&user.User{}).
		Order("points DESC, createdAt ASC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("获取排名周围用户失败: %w", err)
	}

	// 转换为排行榜条目
	entries := make([]leaderboard.LeaderboardEntry, len(users))
	for i, u := range users {
		entries[i] = leaderboard.LeaderboardEntry{
			UserID:     u.ID,
			Username:   u.Username,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Points:     u.Points,
			Rank:       offset + i + 1,
			Tournament: tournament,
		}
	}

	return entries, nil
}
