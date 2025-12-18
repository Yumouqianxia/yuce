package services

import (
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/scoring"
)

// scoringCalculator 积分计算器实现
type scoringCalculator struct{}

// NewScoringCalculator 创建积分计算器
func NewScoringCalculator() scoring.Calculator {
	return &scoringCalculator{}
}

// Calculate 计算预测积分
func (c *scoringCalculator) Calculate(pred *prediction.Prediction, rule *prediction.ScoringRule) *scoring.PointsCalculationResult {
	if pred == nil || pred.Match == nil || !pred.Match.IsFinished() {
		return &scoring.PointsCalculationResult{
			PredictionID: pred.ID,
			UserID:       pred.UserID,
			MatchID:      pred.MatchID,
			Points:       0,
			IsCorrect:    false,
			Reason:       "比赛未结束",
		}
	}

	// 获取预测准确性
	accuracy := c.GetAccuracy(pred)

	return c.CalculateWithAccuracy(pred, accuracy, rule)
}

// CalculateWithAccuracy 根据准确性计算积分
func (c *scoringCalculator) CalculateWithAccuracy(pred *prediction.Prediction, accuracy scoring.PredictionAccuracy, rule *prediction.ScoringRule) *scoring.PointsCalculationResult {
	// 计算基础积分
	basePoints := accuracy.CalculateBasePoints(rule)

	// 计算热门奖励
	popularityBonus := scoring.CalculatePopularityBonus(pred.VoteCount)

	// 总积分
	totalPoints := basePoints + popularityBonus.Bonus

	// 判断是否正确
	isCorrect := accuracy == scoring.AccuracyPerfect || accuracy == scoring.AccuracyTeamOnly

	// 构建原因
	reason := scoring.BuildPointsReason(accuracy, basePoints, popularityBonus)

	return &scoring.PointsCalculationResult{
		PredictionID: pred.ID,
		UserID:       pred.UserID,
		MatchID:      pred.MatchID,
		Points:       totalPoints,
		IsCorrect:    isCorrect,
		Reason:       reason,
	}
}

// GetAccuracy 获取预测准确性
func (c *scoringCalculator) GetAccuracy(pred *prediction.Prediction) scoring.PredictionAccuracy {
	return scoring.GetPredictionAccuracy(pred, pred.Match)
}
