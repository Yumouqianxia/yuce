package prediction

import (
	"context"
)

// Repository 预测仓储接口
type Repository interface {
	// CreatePrediction 创建预测
	CreatePrediction(ctx context.Context, prediction *Prediction) error

	// GetPredictionByID 根据 ID 获取预测
	GetPredictionByID(ctx context.Context, id uint) (*Prediction, error)

	// GetPredictionByUserAndMatch 根据用户和比赛获取预测
	GetPredictionByUserAndMatch(ctx context.Context, userID, matchID uint) (*Prediction, error)

	// UpdatePrediction 更新预测
	UpdatePrediction(ctx context.Context, prediction *Prediction) error

	// GetPredictionsByMatch 获取比赛的所有预测
	GetPredictionsByMatch(ctx context.Context, matchID uint, userID *uint) ([]PredictionWithVotes, error)

	// GetPredictionsByUser 获取用户的所有预测
	GetPredictionsByUser(ctx context.Context, userID uint) ([]Prediction, error)

	// UpdatePredictionPoints 更新预测积分
	UpdatePredictionPoints(ctx context.Context, predictionID uint, points int, isCorrect bool) error

	// GetFeaturedPredictions 获取精选预测
	GetFeaturedPredictions(ctx context.Context, limit int) ([]PredictionWithVotes, error)

	// SetFeatured 设置精选状态
	SetFeatured(ctx context.Context, predictionID uint, featured bool) error

	// DeletePrediction 删除预测
	DeletePrediction(ctx context.Context, id uint) error
}

// VoteRepository 投票仓储接口
type VoteRepository interface {
	// CreateVote 创建投票
	CreateVote(ctx context.Context, vote *Vote) error

	// GetVote 获取投票
	GetVote(ctx context.Context, userID, predictionID uint) (*Vote, error)

	// DeleteVote 删除投票
	DeleteVote(ctx context.Context, userID, predictionID uint) error

	// GetVotesByPrediction 获取预测的所有投票
	GetVotesByPrediction(ctx context.Context, predictionID uint) ([]Vote, error)

	// GetVotesByUser 获取用户的所有投票
	GetVotesByUser(ctx context.Context, userID uint) ([]Vote, error)

	// ExistsVote 检查投票是否存在
	ExistsVote(ctx context.Context, userID, predictionID uint) (bool, error)

	// GetVoteCount 获取预测的投票数
	GetVoteCount(ctx context.Context, predictionID uint) (int, error)

	// CreateVoteWithCount 创建投票并更新计数（事务性操作）
	CreateVoteWithCount(ctx context.Context, vote *Vote) error

	// DeleteVoteWithCount 删除投票并更新计数（事务性操作）
	DeleteVoteWithCount(ctx context.Context, userID, predictionID uint) error

	// GetVoteStats 获取投票统计
	GetVoteStats(ctx context.Context, predictionIDs []uint) ([]VoteStats, error)

	// GetTopVotedPredictions 获取投票数最高的预测
	GetTopVotedPredictions(ctx context.Context, limit int) ([]VoteStats, error)
}
