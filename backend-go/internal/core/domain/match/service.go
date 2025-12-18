package match

import (
	"context"
)

// Service 比赛服务接口
type Service interface {
	// CreateMatch 创建比赛
	CreateMatch(ctx context.Context, req *CreateMatchRequest) (*Match, error)

	// GetMatch 获取比赛详情
	GetMatch(ctx context.Context, id uint) (*Match, error)

	// ListMatches 获取比赛列表
	ListMatches(ctx context.Context, filter ListFilter) ([]Match, error)

	// UpdateMatch 更新比赛信息
	UpdateMatch(ctx context.Context, id uint, req *UpdateMatchRequest) (*Match, error)

	// StartMatch 开始比赛
	StartMatch(ctx context.Context, id uint) error

	// SetResult 设置比赛结果
	SetResult(ctx context.Context, id uint, req *SetResultRequest) error

	// CancelMatch 取消比赛
	CancelMatch(ctx context.Context, id uint) error

	// GetUpcomingMatches 获取即将开始的比赛
	GetUpcomingMatches(ctx context.Context) ([]Match, error)

	// GetLiveMatches 获取正在进行的比赛
	GetLiveMatches(ctx context.Context) ([]Match, error)

	// GetFinishedMatches 获取已结束的比赛
	GetFinishedMatches(ctx context.Context, limit int) ([]Match, error)
}
