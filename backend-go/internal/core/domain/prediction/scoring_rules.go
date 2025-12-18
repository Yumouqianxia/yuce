package prediction

import (
	"context"
	"time"
)

// ScoringRule 预测积分规则
type ScoringRule struct {
	ID                      uint      `json:"id" gorm:"primaryKey"`
	Name                    string    `json:"name" gorm:"size:100;not null"`
	Description             string    `json:"description" gorm:"size:500"`
	CorrectTeamCorrectScore int       `json:"correct_team_correct_score" gorm:"default:0"` // 预测正确队伍和比分
	CorrectTeamWrongScore   int       `json:"correct_team_wrong_score" gorm:"default:0"`   // 预测正确队伍错误比分
	WrongTeamCorrectScore   int       `json:"wrong_team_correct_score" gorm:"default:0"`   // 预测错误队伍正确比分
	WrongTeamWrongScore     int       `json:"wrong_team_wrong_score" gorm:"default:0"`     // 预测错误队伍错误比分
	IsActive                bool      `json:"is_active" gorm:"default:true"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ScoringRule) TableName() string {
	return "scoring_rules"
}

// CalculatePoints 根据规则计算积分
func (sr *ScoringRule) CalculatePoints(prediction *Prediction) int {
	if prediction.Match == nil || !prediction.Match.IsFinished() {
		return 0
	}

	// 检查队伍预测是否正确
	teamCorrect := prediction.PredictedWinner == prediction.Match.Winner

	// 检查比分预测是否正确
	scoreCorrect := prediction.PredictedScoreA == prediction.Match.ScoreA &&
		prediction.PredictedScoreB == prediction.Match.ScoreB

	// 根据四种情况计算积分
	switch {
	case teamCorrect && scoreCorrect:
		return sr.CorrectTeamCorrectScore
	case teamCorrect && !scoreCorrect:
		return sr.CorrectTeamWrongScore
	case !teamCorrect && scoreCorrect:
		return sr.WrongTeamCorrectScore
	default: // !teamCorrect && !scoreCorrect
		return sr.WrongTeamWrongScore
	}
}

// CreateScoringRuleRequest 创建积分规则请求
type CreateScoringRuleRequest struct {
	Name                    string `json:"name" validate:"required,max=100"`
	Description             string `json:"description" validate:"max=500"`
	CorrectTeamCorrectScore int    `json:"correct_team_correct_score" validate:"min=0"`
	CorrectTeamWrongScore   int    `json:"correct_team_wrong_score" validate:"min=0"`
	WrongTeamCorrectScore   int    `json:"wrong_team_correct_score" validate:"min=0"`
	WrongTeamWrongScore     int    `json:"wrong_team_wrong_score" validate:"min=0"`
}

// UpdateScoringRuleRequest 更新积分规则请求
type UpdateScoringRuleRequest struct {
	Name                    string `json:"name" validate:"max=100"`
	Description             string `json:"description" validate:"max=500"`
	CorrectTeamCorrectScore *int   `json:"correct_team_correct_score" validate:"omitempty,min=0"`
	CorrectTeamWrongScore   *int   `json:"correct_team_wrong_score" validate:"omitempty,min=0"`
	WrongTeamCorrectScore   *int   `json:"wrong_team_correct_score" validate:"omitempty,min=0"`
	WrongTeamWrongScore     *int   `json:"wrong_team_wrong_score" validate:"omitempty,min=0"`
	IsActive                *bool  `json:"is_active"`
}

// ScoringRuleRepository 积分规则仓储接口
type ScoringRuleRepository interface {
	// CreateScoringRule 创建积分规则
	CreateScoringRule(ctx context.Context, rule *ScoringRule) error

	// GetScoringRuleByID 根据ID获取积分规则
	GetScoringRuleByID(ctx context.Context, id uint) (*ScoringRule, error)

	// GetActiveScoringRule 获取当前激活的积分规则
	GetActiveScoringRule(ctx context.Context) (*ScoringRule, error)

	// UpdateScoringRule 更新积分规则
	UpdateScoringRule(ctx context.Context, rule *ScoringRule) error

	// ListScoringRules 获取所有积分规则
	ListScoringRules(ctx context.Context) ([]ScoringRule, error)

	// SetActiveRule 设置激活的规则
	SetActiveRule(ctx context.Context, ruleID uint) error

	// DeleteScoringRule 删除积分规则
	DeleteScoringRule(ctx context.Context, id uint) error
}

// ScoringRuleService 积分规则服务接口
type ScoringRuleService interface {
	// CreateScoringRule 创建积分规则
	CreateScoringRule(ctx context.Context, req *CreateScoringRuleRequest) (*ScoringRule, error)

	// GetScoringRule 获取积分规则
	GetScoringRule(ctx context.Context, id uint) (*ScoringRule, error)

	// GetActiveScoringRule 获取当前激活的积分规则
	GetActiveScoringRule(ctx context.Context) (*ScoringRule, error)

	// UpdateScoringRule 更新积分规则
	UpdateScoringRule(ctx context.Context, id uint, req *UpdateScoringRuleRequest) (*ScoringRule, error)

	// ListScoringRules 获取所有积分规则
	ListScoringRules(ctx context.Context) ([]ScoringRule, error)

	// SetActiveRule 设置激活的规则
	SetActiveRule(ctx context.Context, ruleID uint) error

	// DeleteScoringRule 删除积分规则
	DeleteScoringRule(ctx context.Context, id uint) error

	// CalculatePointsWithRule 使用指定规则计算积分
	CalculatePointsWithRule(ctx context.Context, prediction *Prediction, ruleID *uint) (int, error)
}
