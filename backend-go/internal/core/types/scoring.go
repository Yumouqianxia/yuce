package types

import "time"

// ScoreBreakdown 积分明细
type ScoreBreakdown struct {
	BaseScore       int     `json:"base_score"`        // 基础积分
	DifficultyBonus int     `json:"difficulty_bonus"`  // 难度奖励
	VoteReward      int     `json:"vote_reward"`       // 投票奖励
	TimeReward      int     `json:"time_reward"`       // 时间奖励
	ModifyPenalty   int     `json:"modify_penalty"`    // 修改惩罚
	TotalScore      int     `json:"total_score"`       // 总积分
	Breakdown       string  `json:"breakdown"`         // 积分计算说明
}

// PreviewScoreRequest 积分预览请求
type PreviewScoreRequest struct {
	SportTypeID       uint      `json:"sport_type_id" validate:"required"`
	PredictedWinner   string    `json:"predicted_winner" validate:"required,oneof=A B DRAW"`
	PredictedScoreA   int       `json:"predicted_score_a" validate:"min=0"`
	PredictedScoreB   int       `json:"predicted_score_b" validate:"min=0"`
	ActualWinner      string    `json:"actual_winner" validate:"oneof=A B DRAW"`
	ActualScoreA      int       `json:"actual_score_a" validate:"min=0"`
	ActualScoreB      int       `json:"actual_score_b" validate:"min=0"`
	ModificationCount int       `json:"modification_count" validate:"min=0"`
	VoteCount         int       `json:"vote_count" validate:"min=0"`
	PredictionTime    time.Time `json:"prediction_time"`
	MatchStartTime    time.Time `json:"match_start_time"`
}