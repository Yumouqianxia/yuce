package ports

import (
	"context"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/types"
)

// ScoringRuleService 积分规则服务接口
type ScoringRuleService interface {
	// 基础CRUD操作
	CreateScoringRule(ctx context.Context, req *CreateScoringRuleRequest) (*sport.ScoringRule, error)
	GetScoringRule(ctx context.Context, id uint) (*sport.ScoringRule, error)
	UpdateScoringRule(ctx context.Context, id uint, req *UpdateScoringRuleRequest) (*sport.ScoringRule, error)
	DeleteScoringRule(ctx context.Context, id uint) error
	ListScoringRules(ctx context.Context, req *ListScoringRulesRequest) (*ListScoringRulesResponse, error)

	// 规则管理
	GetActiveScoringRule(ctx context.Context, sportTypeID uint) (*sport.ScoringRule, error)
	SetActiveScoringRule(ctx context.Context, id uint) error
	GetScoringRulesBySportType(ctx context.Context, sportTypeID uint) ([]*sport.ScoringRule, error)

	// 积分计算
	CalculateScore(ctx context.Context, predictionID uint) (*types.ScoreBreakdown, error)
	PreviewScore(ctx context.Context, req *types.PreviewScoreRequest) (*types.ScoreBreakdown, error)
	
	// 批量重算
	RecalculateScores(ctx context.Context, sportTypeID uint, ruleID uint) (*RecalculateResult, error)
}

// ScoringRuleRepository 积分规则仓储接口
type ScoringRuleRepository interface {
	Create(ctx context.Context, rule *sport.ScoringRule) error
	GetByID(ctx context.Context, id uint) (*sport.ScoringRule, error)
	GetBySportTypeID(ctx context.Context, sportTypeID uint) ([]*sport.ScoringRule, error)
	GetActiveBySportTypeID(ctx context.Context, sportTypeID uint) (*sport.ScoringRule, error)
	Update(ctx context.Context, rule *sport.ScoringRule) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, options *ListScoringRulesOptions) ([]*sport.ScoringRule, error)
	Count(ctx context.Context, options *ListScoringRulesOptions) (int64, error)
	SetActive(ctx context.Context, id uint) error
}

// 请求和响应结构体

// CreateScoringRuleRequest 创建积分规则请求
type CreateScoringRuleRequest struct {
	SportTypeID  uint   `json:"sport_type_id" validate:"required"`
	Name         string `json:"name" validate:"required,max=100"`
	Description  string `json:"description" validate:"max=1000"`
	IsActive     bool   `json:"is_active"`

	// 基础积分设置
	BasePoints           int     `json:"base_points" validate:"required,min=1,max=1000"`
	EnableDifficulty     bool    `json:"enable_difficulty"`
	DifficultyMultiplier float64 `json:"difficulty_multiplier" validate:"min=0.1,max=10.0"`

	// 奖励组件开关
	EnableVoteReward bool `json:"enable_vote_reward"`
	VoteRewardPoints int  `json:"vote_reward_points" validate:"min=0,max=100"`
	MaxVoteReward    int  `json:"max_vote_reward" validate:"min=0,max=1000"`

	EnableTimeReward bool `json:"enable_time_reward"`
	TimeRewardPoints int  `json:"time_reward_points" validate:"min=0,max=100"`
	TimeRewardHours  int  `json:"time_reward_hours" validate:"min=1,max=168"` // 最大7天

	// 惩罚组件开关
	EnableModifyPenalty bool `json:"enable_modify_penalty"`
	ModifyPenaltyPoints int  `json:"modify_penalty_points" validate:"min=0,max=100"`
	MaxModifyPenalty    int  `json:"max_modify_penalty" validate:"min=0,max=1000"`
}

// UpdateScoringRuleRequest 更新积分规则请求
type UpdateScoringRuleRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`

	// 基础积分设置
	BasePoints           *int     `json:"base_points" validate:"omitempty,min=1,max=1000"`
	EnableDifficulty     *bool    `json:"enable_difficulty"`
	DifficultyMultiplier *float64 `json:"difficulty_multiplier" validate:"omitempty,min=0.1,max=10.0"`

	// 奖励组件开关
	EnableVoteReward *bool `json:"enable_vote_reward"`
	VoteRewardPoints *int  `json:"vote_reward_points" validate:"omitempty,min=0,max=100"`
	MaxVoteReward    *int  `json:"max_vote_reward" validate:"omitempty,min=0,max=1000"`

	EnableTimeReward *bool `json:"enable_time_reward"`
	TimeRewardPoints *int  `json:"time_reward_points" validate:"omitempty,min=0,max=100"`
	TimeRewardHours  *int  `json:"time_reward_hours" validate:"omitempty,min=1,max=168"`

	// 惩罚组件开关
	EnableModifyPenalty *bool `json:"enable_modify_penalty"`
	ModifyPenaltyPoints *int  `json:"modify_penalty_points" validate:"omitempty,min=0,max=100"`
	MaxModifyPenalty    *int  `json:"max_modify_penalty" validate:"omitempty,min=0,max=1000"`
}

// ListScoringRulesRequest 积分规则列表请求
type ListScoringRulesRequest struct {
	SportTypeID *uint  `json:"sport_type_id"`
	IsActive    *bool  `json:"is_active"`
	OrderBy     string `json:"order_by" validate:"omitempty,oneof=name created_at"`
	Page        int    `json:"page" validate:"min=1"`
	PageSize    int    `json:"page_size" validate:"min=1,max=100"`
}

// ListScoringRulesResponse 积分规则列表响应
type ListScoringRulesResponse struct {
	ScoringRules []*sport.ScoringRule `json:"scoring_rules"`
	Total        int64                `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"page_size"`
	TotalPages   int                  `json:"total_pages"`
}

// RecalculateResult 重算结果
type RecalculateResult struct {
	TotalPredictions    int `json:"total_predictions"`    // 总预测数
	UpdatedPredictions  int `json:"updated_predictions"`  // 更新的预测数
	FailedPredictions   int `json:"failed_predictions"`   // 失败的预测数
	TotalPointsChanged  int `json:"total_points_changed"` // 总积分变化
}

// ListScoringRulesOptions 积分规则列表查询选项（仓储层使用）
type ListScoringRulesOptions struct {
	SportTypeID *uint
	IsActive    *bool
	OrderBy     string
	Limit       int
	Offset      int
}