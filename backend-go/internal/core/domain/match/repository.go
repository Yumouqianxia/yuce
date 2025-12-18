package match

import (
	"context"
	"time"
)

// Repository 比赛仓储接口
type Repository interface {
	// Create 创建比赛
	Create(ctx context.Context, match *Match) error

	// GetByID 根据 ID 获取比赛
	GetByID(ctx context.Context, id uint) (*Match, error)

	// List 获取比赛列表
	List(ctx context.Context, filter ListFilter) ([]Match, error)

	// Update 更新比赛信息
	Update(ctx context.Context, match *Match) error

	// UpdateStatus 更新比赛状态
	UpdateStatus(ctx context.Context, id uint, status MatchStatus) error

	// SetResult 设置比赛结果
	SetResult(ctx context.Context, id uint, scoreA, scoreB int, winner string) error

	// UpdateScore 更新比赛比分（用于直播比赛）
	UpdateScore(ctx context.Context, id uint, scoreA, scoreB int) error

	// GetUpcoming 获取即将开始的比赛
	GetUpcoming(ctx context.Context, limit int) ([]Match, error)

	// GetLive 获取正在进行的比赛
	GetLive(ctx context.Context) ([]Match, error)

	// GetFinished 获取已结束的比赛
	GetFinished(ctx context.Context, limit int) ([]Match, error)

	// GetFinishedMatches 获取所有已结束的比赛（用于积分计算）
	GetFinishedMatches(ctx context.Context) ([]Match, error)

	// Delete 删除比赛
	Delete(ctx context.Context, id uint) error
}

// ListFilter 列表过滤器
type ListFilter struct {
	Tournament Tournament  `json:"tournament"`
	Status     MatchStatus `json:"status"`
	StartDate  *time.Time  `json:"start_date"`
	EndDate    *time.Time  `json:"end_date"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}
