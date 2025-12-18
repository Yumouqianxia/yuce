package prediction

import (
	"context"

	"backend-go/internal/core/domain/match"
)

// CreatePredictionRequest 创建预测请求
type CreatePredictionRequest struct {
	MatchID         uint         `json:"matchId" validate:"required"`
	PredictedWinner match.Winner `json:"predictedWinner" validate:"required,oneof=A B DRAW"`
	PredictedScoreA int          `json:"predictedScoreA" validate:"min=0"`
	PredictedScoreB int          `json:"predictedScoreB" validate:"min=0"`
}

// UpdatePredictionRequest 更新预测请求
type UpdatePredictionRequest struct {
	PredictedWinner match.Winner `json:"predictedWinner" validate:"required,oneof=A B DRAW"`
	PredictedScoreA int          `json:"predictedScoreA" validate:"min=0"`
	PredictedScoreB int          `json:"predictedScoreB" validate:"min=0"`
}

// Service 预测服务接口
type Service interface {
	// CreatePrediction 创建预测
	CreatePrediction(ctx context.Context, userID uint, req *CreatePredictionRequest) (*Prediction, error)

	// UpdatePrediction 更新预测
	UpdatePrediction(ctx context.Context, userID uint, predictionID uint, req *UpdatePredictionRequest) (*Prediction, error)

	// GetPrediction 获取预测详情
	GetPrediction(ctx context.Context, id uint) (*Prediction, error)

	// GetPredictionsByMatch 获取比赛的所有预测
	GetPredictionsByMatch(ctx context.Context, matchID uint, userID *uint) ([]PredictionWithVotes, error)

	// GetUserPredictions 获取用户的所有预测
	GetUserPredictions(ctx context.Context, userID uint) ([]Prediction, error)

	// VotePrediction 投票支持预测
	VotePrediction(ctx context.Context, userID uint, predictionID uint) error

	// UnvotePrediction 取消投票
	UnvotePrediction(ctx context.Context, userID uint, predictionID uint) error

	// CalculatePoints 计算比赛结束后的积分
	CalculatePoints(ctx context.Context, matchID uint) error

	// GetFeaturedPredictions 获取精选预测
	GetFeaturedPredictions(ctx context.Context) ([]PredictionWithVotes, error)

	// UpdateFeaturedPredictions 更新精选预测
	UpdateFeaturedPredictions(ctx context.Context) error

	// CalculatePointsWithCustomRule 使用自定义规则计算积分
	CalculatePointsWithCustomRule(ctx context.Context, matchID uint, ruleID *uint) error
}
