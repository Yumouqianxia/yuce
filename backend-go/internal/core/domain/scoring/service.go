package scoring

import (
	"context"

	"backend-go/internal/core/domain/prediction"
)

// Service 积分计算服务接口
type Service interface {
	// CalculateMatchPoints 计算比赛结束后的所有预测积分
	CalculateMatchPoints(ctx context.Context, matchID uint) (*MatchPointsCalculation, error)

	// CalculateMatchPointsWithRule 使用指定规则计算比赛积分
	CalculateMatchPointsWithRule(ctx context.Context, matchID uint, ruleID *uint) (*MatchPointsCalculation, error)

	// CalculatePredictionPoints 计算单个预测的积分
	CalculatePredictionPoints(ctx context.Context, predictionID uint, ruleID *uint) (*PointsCalculationResult, error)

	// RecalculateAllPoints 重新计算所有已结束比赛的积分
	RecalculateAllPoints(ctx context.Context, ruleID *uint) error

	// GetPointsHistory 获取用户积分历史
	GetPointsHistory(ctx context.Context, userID uint, tournament string) ([]PointsUpdateEvent, error)

	// ProcessPointsUpdate 处理积分更新（更新用户积分和排行榜）
	ProcessPointsUpdate(ctx context.Context, results []PointsCalculationResult, tournament string) error
}

// Calculator 积分计算器接口
type Calculator interface {
	// Calculate 计算预测积分
	Calculate(pred *prediction.Prediction, rule *prediction.ScoringRule) *PointsCalculationResult

	// CalculateWithAccuracy 根据准确性计算积分
	CalculateWithAccuracy(pred *prediction.Prediction, accuracy PredictionAccuracy, rule *prediction.ScoringRule) *PointsCalculationResult

	// GetAccuracy 获取预测准确性
	GetAccuracy(pred *prediction.Prediction) PredictionAccuracy
}

// Repository 积分计算仓储接口
type Repository interface {
	// SavePointsCalculation 保存积分计算结果
	SavePointsCalculation(ctx context.Context, calculation *MatchPointsCalculation) error

	// GetPointsHistory 获取积分历史
	GetPointsHistory(ctx context.Context, userID uint, tournament string) ([]PointsUpdateEvent, error)

	// SavePointsUpdateEvent 保存积分更新事件
	SavePointsUpdateEvent(ctx context.Context, event *PointsUpdateEvent) error

	// GetMatchCalculation 获取比赛的积分计算结果
	GetMatchCalculation(ctx context.Context, matchID uint) (*MatchPointsCalculation, error)

	// IsMatchProcessed 检查比赛是否已处理积分
	IsMatchProcessed(ctx context.Context, matchID uint) (bool, error)
}
