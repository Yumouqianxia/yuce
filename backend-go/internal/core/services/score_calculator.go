package services

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/types"
	"github.com/sirupsen/logrus"
)

// ScoreCalculator 积分计算器接口
type ScoreCalculator interface {
	CalculateScore(ctx context.Context, prediction *PredictionInfo, match *MatchInfo, rule *sport.ScoringRule) (*types.ScoreBreakdown, error)
	PreviewScore(ctx context.Context, req *types.PreviewScoreRequest) (*types.ScoreBreakdown, error)
}

// DefaultScoreCalculator 默认积分计算器实现
type DefaultScoreCalculator struct {
	logger *logrus.Logger
}

// NewDefaultScoreCalculator 创建默认积分计算器
func NewDefaultScoreCalculator(logger *logrus.Logger) *DefaultScoreCalculator {
	return &DefaultScoreCalculator{
		logger: logger,
	}
}

// PredictionInfo 预测信息
type PredictionInfo struct {
	ID                  uint      `json:"id"`
	UserID              uint      `json:"user_id"`
	MatchID             uint      `json:"match_id"`
	PredictedWinner     string    `json:"predicted_winner"`
	PredictedScoreA     int       `json:"predicted_score_a"`
	PredictedScoreB     int       `json:"predicted_score_b"`
	IsCorrect           bool      `json:"is_correct"`
	ModificationCount   int       `json:"modification_count"`
	VoteCount           int       `json:"vote_count"`
	CreatedAt           time.Time `json:"created_at"`
	LastModifiedAt      time.Time `json:"last_modified_at"`
}

// MatchInfo 比赛信息
type MatchInfo struct {
	ID          uint      `json:"id"`
	SportTypeID uint      `json:"sport_type_id"`
	TeamA       string    `json:"team_a"`
	TeamB       string    `json:"team_b"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	Winner      string    `json:"winner"`
	ScoreA      int       `json:"score_a"`
	ScoreB      int       `json:"score_b"`
	CreatedAt   time.Time `json:"created_at"`
}



// CalculateScore 计算积分
func (c *DefaultScoreCalculator) CalculateScore(ctx context.Context, prediction *PredictionInfo, match *MatchInfo, rule *sport.ScoringRule) (*types.ScoreBreakdown, error) {
	c.logger.WithFields(logrus.Fields{
		"prediction_id": prediction.ID,
		"match_id":      match.ID,
		"rule_id":       rule.ID,
	}).Debug("Calculating score")

	breakdown := &types.ScoreBreakdown{}
	
	// 如果预测错误，直接返回0分
	if !prediction.IsCorrect {
		breakdown.TotalScore = 0
		breakdown.Breakdown = "预测错误，无积分"
		return breakdown, nil
	}

	// 1. 基础积分
	breakdown.BaseScore = rule.BasePoints

	// 2. 难度系数奖励
	if rule.EnableDifficulty {
		difficultyBonus := int(float64(rule.BasePoints) * (rule.DifficultyMultiplier - 1.0))
		breakdown.DifficultyBonus = difficultyBonus
	}

	// 3. 投票奖励
	if rule.EnableVoteReward && prediction.VoteCount > 0 {
		voteReward := prediction.VoteCount * rule.VoteRewardPoints
		if voteReward > rule.MaxVoteReward {
			voteReward = rule.MaxVoteReward
		}
		breakdown.VoteReward = voteReward
	}

	// 4. 时间奖励
	if rule.EnableTimeReward {
		timeReward := c.calculateTimeReward(prediction.CreatedAt, match.StartTime, rule)
		breakdown.TimeReward = timeReward
	}

	// 5. 修改惩罚
	if rule.EnableModifyPenalty && prediction.ModificationCount > 0 {
		modifyPenalty := prediction.ModificationCount * rule.ModifyPenaltyPoints
		if modifyPenalty > rule.MaxModifyPenalty {
			modifyPenalty = rule.MaxModifyPenalty
		}
		breakdown.ModifyPenalty = modifyPenalty
	}

	// 计算总积分
	breakdown.TotalScore = breakdown.BaseScore + breakdown.DifficultyBonus + 
		breakdown.VoteReward + breakdown.TimeReward - breakdown.ModifyPenalty

	// 确保积分不为负数
	if breakdown.TotalScore < 0 {
		breakdown.TotalScore = 0
	}

	// 生成积分说明
	breakdown.Breakdown = c.generateBreakdownText(breakdown, rule)

	c.logger.WithFields(logrus.Fields{
		"prediction_id": prediction.ID,
		"total_score":   breakdown.TotalScore,
		"breakdown":     breakdown.Breakdown,
	}).Info("Score calculated")

	return breakdown, nil
}

// PreviewScore 预览积分计算
func (c *DefaultScoreCalculator) PreviewScore(ctx context.Context, req *types.PreviewScoreRequest) (*types.ScoreBreakdown, error) {
	// 这里需要获取积分规则，暂时使用默认规则进行演示
	// 在实际实现中，应该从数据库获取对应运动类型的激活规则
	rule := &sport.ScoringRule{
		BasePoints:           10,
		EnableDifficulty:     true,
		DifficultyMultiplier: 1.5,
		EnableVoteReward:     true,
		VoteRewardPoints:     1,
		MaxVoteReward:        10,
		EnableTimeReward:     true,
		TimeRewardPoints:     5,
		TimeRewardHours:      24,
		EnableModifyPenalty:  true,
		ModifyPenaltyPoints:  2,
		MaxModifyPenalty:     6,
	}

	// 构造预测信息
	prediction := &PredictionInfo{
		PredictedWinner:   req.PredictedWinner,
		PredictedScoreA:   req.PredictedScoreA,
		PredictedScoreB:   req.PredictedScoreB,
		IsCorrect:         c.isPredictionCorrect(req),
		ModificationCount: req.ModificationCount,
		VoteCount:         req.VoteCount,
		CreatedAt:         req.PredictionTime,
	}

	// 构造比赛信息
	match := &MatchInfo{
		SportTypeID: req.SportTypeID,
		StartTime:   req.MatchStartTime,
		Winner:      req.ActualWinner,
		ScoreA:      req.ActualScoreA,
		ScoreB:      req.ActualScoreB,
	}

	return c.CalculateScore(ctx, prediction, match, rule)
}

// calculateTimeReward 计算时间奖励
func (c *DefaultScoreCalculator) calculateTimeReward(predictionTime, matchStartTime time.Time, rule *sport.ScoringRule) int {
	// 计算预测时间与比赛开始时间的间隔
	duration := matchStartTime.Sub(predictionTime)
	hours := int(duration.Hours())

	// 如果预测时间在奖励时间范围内，给予时间奖励
	if hours >= rule.TimeRewardHours {
		return rule.TimeRewardPoints
	}

	return 0
}

// isPredictionCorrect 判断预测是否正确
func (c *DefaultScoreCalculator) isPredictionCorrect(req *types.PreviewScoreRequest) bool {
	// 简单的预测正确性判断
	if req.ActualWinner == "" {
		return false // 比赛未结束
	}

	// 判断胜负预测是否正确
	return req.PredictedWinner == req.ActualWinner
}

// generateBreakdownText 生成积分说明文本
func (c *DefaultScoreCalculator) generateBreakdownText(breakdown *types.ScoreBreakdown, rule *sport.ScoringRule) string {
	text := fmt.Sprintf("基础积分: %d", breakdown.BaseScore)

	if breakdown.DifficultyBonus > 0 {
		text += fmt.Sprintf(" + 难度奖励: %d", breakdown.DifficultyBonus)
	}

	if breakdown.VoteReward > 0 {
		text += fmt.Sprintf(" + 投票奖励: %d", breakdown.VoteReward)
	}

	if breakdown.TimeReward > 0 {
		text += fmt.Sprintf(" + 时间奖励: %d", breakdown.TimeReward)
	}

	if breakdown.ModifyPenalty > 0 {
		text += fmt.Sprintf(" - 修改惩罚: %d", breakdown.ModifyPenalty)
	}

	text += fmt.Sprintf(" = 总积分: %d", breakdown.TotalScore)

	return text
}