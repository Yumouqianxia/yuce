package mysql

import (
	"context"

	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/match"
	"gorm.io/gorm"
)

// MatchRepository MySQL 比赛仓储实现
type MatchRepository struct {
	db *gorm.DB
}

// NewMatchRepository 创建比赛仓储实例
func NewMatchRepository(db *gorm.DB) match.Repository {
	return &MatchRepository{
		db: db,
	}
}

// Create 创建比赛
func (r *MatchRepository) Create(ctx context.Context, m *match.Match) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// GetByID 根据 ID 获取比赛
func (r *MatchRepository) GetByID(ctx context.Context, id uint) (*match.Match, error) {
	var m match.Match
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrMatchNotFound
		}
		return nil, err
	}
	return &m, nil
}

// List 获取比赛列表
func (r *MatchRepository) List(ctx context.Context, filter match.ListFilter) ([]match.Match, error) {
	var matches []match.Match

	query := r.db.WithContext(ctx)

	// 优化查询：使用复合索引的顺序构建WHERE条件
	// 按照索引 idx_matches_status_tournament_start_time 的顺序
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Tournament != "" {
		query = query.Where("tournament = ?", filter.Tournament)
	}

	if filter.StartDate != nil {
		query = query.Where("start_time >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("start_time <= ?", *filter.EndDate)
	}

	// 按开始时间排序 (利用索引)
	query = query.Order("start_time ASC")

	// 应用分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&matches).Error
	return matches, err
}

// Update 更新比赛信息
func (r *MatchRepository) Update(ctx context.Context, m *match.Match) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// UpdateStatus 更新比赛状态
func (r *MatchRepository) UpdateStatus(ctx context.Context, id uint, status match.MatchStatus) error {
	return r.db.WithContext(ctx).Model(&match.Match{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// SetResult 设置比赛结果
func (r *MatchRepository) SetResult(ctx context.Context, id uint, scoreA, scoreB int, winner string) error {
	updates := map[string]interface{}{
		"score_a": scoreA,
		"score_b": scoreB,
		"winner":  winner,
		"status":  match.MatchStatusFinished,
	}

	return r.db.WithContext(ctx).Model(&match.Match{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateScore 更新比赛比分（用于直播比赛）
func (r *MatchRepository) UpdateScore(ctx context.Context, id uint, scoreA, scoreB int) error {
	updates := map[string]interface{}{
		"score_a": scoreA,
		"score_b": scoreB,
	}

	return r.db.WithContext(ctx).Model(&match.Match{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetUpcoming 获取即将开始的比赛
func (r *MatchRepository) GetUpcoming(ctx context.Context, limit int) ([]match.Match, error) {
	var matches []match.Match

	// 使用索引 idx_matches_status_start_time 优化查询
	query := r.db.WithContext(ctx).
		Where("status = ?", match.MatchStatusUpcoming).
		Order("start_time ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&matches).Error
	return matches, err
}

// GetLive 获取正在进行的比赛
func (r *MatchRepository) GetLive(ctx context.Context) ([]match.Match, error) {
	var matches []match.Match

	// 使用索引 idx_matches_status_start_time 优化查询
	err := r.db.WithContext(ctx).
		Where("status = ?", match.MatchStatusLive).
		Order("start_time ASC").
		Find(&matches).Error

	return matches, err
}

// GetFinished 获取已结束的比赛
func (r *MatchRepository) GetFinished(ctx context.Context, limit int) ([]match.Match, error) {
	var matches []match.Match

	// 使用索引 idx_matches_status_start_time 优化查询
	query := r.db.WithContext(ctx).
		Where("status = ?", match.MatchStatusFinished).
		Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&matches).Error
	return matches, err
}

// GetFinishedMatches 获取所有已结束的比赛（用于积分计算）
func (r *MatchRepository) GetFinishedMatches(ctx context.Context) ([]match.Match, error) {
	var matches []match.Match

	err := r.db.WithContext(ctx).
		Where("status = ?", match.MatchStatusFinished).
		Order("start_time ASC").
		Find(&matches).Error

	return matches, err
}

// Delete 删除比赛
func (r *MatchRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&match.Match{}, id).Error
}
