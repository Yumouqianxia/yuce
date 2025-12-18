package scoring

import (
	"time"

	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
)

// PointsCalculationResult 积分计算结果
type PointsCalculationResult struct {
	PredictionID uint   `json:"predictionId"`
	UserID       uint   `json:"userId"`
	MatchID      uint   `json:"matchId"`
	Points       int    `json:"points"`
	IsCorrect    bool   `json:"isCorrect"`
	Reason       string `json:"reason"` // 积分获得原因
}

// MatchPointsCalculation 比赛积分计算结果
type MatchPointsCalculation struct {
	MatchID     uint                      `json:"matchId"`
	Results     []PointsCalculationResult `json:"results"`
	TotalPoints int                       `json:"totalPoints"`
	ProcessedAt time.Time                 `json:"processedAt"`
}

// PointsUpdateEvent 积分更新事件
type PointsUpdateEvent struct {
	UserID       uint      `json:"userId"`
	MatchID      uint      `json:"matchId"`
	PredictionID uint      `json:"predictionId"`
	OldPoints    int       `json:"oldPoints"`
	NewPoints    int       `json:"newPoints"`
	PointsChange int       `json:"pointsChange"`
	Tournament   string    `json:"tournament"`
	Timestamp    time.Time `json:"createdAt"`
}

// PredictionAccuracy 预测准确性分类
type PredictionAccuracy string

const (
	AccuracyPerfect   PredictionAccuracy = "PERFECT"    // 完全正确（队伍+比分）
	AccuracyTeamOnly  PredictionAccuracy = "TEAM_ONLY"  // 仅队伍正确
	AccuracyScoreOnly PredictionAccuracy = "SCORE_ONLY" // 仅比分正确
	AccuracyWrong     PredictionAccuracy = "WRONG"      // 完全错误
)

// GetPredictionAccuracy 获取预测准确性
func GetPredictionAccuracy(pred *prediction.Prediction, match *match.Match) PredictionAccuracy {
	if pred == nil || match == nil || !match.IsFinished() {
		return AccuracyWrong
	}

	teamCorrect := pred.PredictedWinner == match.Winner
	scoreCorrect := pred.PredictedScoreA == match.ScoreA && pred.PredictedScoreB == match.ScoreB

	switch {
	case teamCorrect && scoreCorrect:
		return AccuracyPerfect
	case teamCorrect && !scoreCorrect:
		return AccuracyTeamOnly
	case !teamCorrect && scoreCorrect:
		return AccuracyScoreOnly
	default:
		return AccuracyWrong
	}
}

// GetAccuracyDescription 获取准确性描述
func (a PredictionAccuracy) GetDescription() string {
	switch a {
	case AccuracyPerfect:
		return "预测完全正确"
	case AccuracyTeamOnly:
		return "预测队伍正确"
	case AccuracyScoreOnly:
		return "预测比分正确"
	case AccuracyWrong:
		return "预测错误"
	default:
		return "未知"
	}
}

// CalculateBasePoints 根据准确性计算基础积分
func (a PredictionAccuracy) CalculateBasePoints(rule *prediction.ScoringRule) int {
	if rule == nil {
		// 使用默认积分规则
		switch a {
		case AccuracyPerfect:
			return 30
		case AccuracyTeamOnly:
			return 10
		case AccuracyScoreOnly:
			return 0
		case AccuracyWrong:
			return 0
		default:
			return 0
		}
	}

	// 使用自定义积分规则
	switch a {
	case AccuracyPerfect:
		return rule.CorrectTeamCorrectScore
	case AccuracyTeamOnly:
		return rule.CorrectTeamWrongScore
	case AccuracyScoreOnly:
		return rule.WrongTeamCorrectScore
	case AccuracyWrong:
		return rule.WrongTeamWrongScore
	default:
		return 0
	}
}

// PopularityBonus 热门奖励计算
type PopularityBonus struct {
	VoteCount int `json:"voteCount"`
	Bonus     int `json:"bonus"`
}

// CalculatePopularityBonus 计算热门奖励
func CalculatePopularityBonus(voteCount int) PopularityBonus {
	bonus := 0

	// 根据投票数给予不同的奖励
	switch {
	case voteCount >= 20:
		bonus = 10 // 超高人气
	case voteCount >= 10:
		bonus = 5 // 高人气
	case voteCount >= 5:
		bonus = 2 // 中等人气
	default:
		bonus = 0 // 无奖励
	}

	return PopularityBonus{
		VoteCount: voteCount,
		Bonus:     bonus,
	}
}

// BuildPointsReason 构建积分获得原因
func BuildPointsReason(accuracy PredictionAccuracy, basePoints int, popularityBonus PopularityBonus) string {
	reason := accuracy.GetDescription()

	if basePoints > 0 {
		reason += "，获得基础积分"
	}

	if popularityBonus.Bonus > 0 {
		reason += "，获得热门奖励"
	}

	return reason
}
