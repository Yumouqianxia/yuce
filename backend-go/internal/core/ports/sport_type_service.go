package ports

import (
	"context"
	"time"

	"backend-go/internal/core/domain/sport"
)

// SportTypeService 运动类型服务接口
type SportTypeService interface {
	// 基础CRUD操作
	CreateSportType(ctx context.Context, req *CreateSportTypeRequest) (*sport.SportType, error)
	GetSportType(ctx context.Context, id uint) (*sport.SportType, error)
	GetSportTypeByCode(ctx context.Context, code string) (*sport.SportType, error)
	UpdateSportType(ctx context.Context, id uint, req *UpdateSportTypeRequest) (*sport.SportType, error)
	DeleteSportType(ctx context.Context, id uint) error
	ListSportTypes(ctx context.Context, req *ListSportTypesRequest) (*ListSportTypesResponse, error)

	// 配置管理
	GetSportConfiguration(ctx context.Context, sportTypeID uint) (*sport.SportConfiguration, error)
	UpdateSportConfiguration(ctx context.Context, sportTypeID uint, req *UpdateSportConfigurationRequest) (*sport.SportConfiguration, error)

	// 批量操作
	BatchUpdateConfiguration(ctx context.Context, req *BatchUpdateConfigRequest) error
	
	// 统计信息
	GetSportTypeStats(ctx context.Context, sportTypeID uint) (*SportTypeStats, error)
}

// SportTypeRepository 运动类型仓储接口
type SportTypeRepository interface {
	Create(ctx context.Context, sportType *sport.SportType) error
	GetByID(ctx context.Context, id uint) (*sport.SportType, error)
	GetByCode(ctx context.Context, code string) (*sport.SportType, error)
	Update(ctx context.Context, sportType *sport.SportType) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, options *ListSportTypesOptions) ([]*sport.SportType, error)
	Count(ctx context.Context, options *ListSportTypesOptions) (int64, error)
	
	// 配置相关
	CreateConfiguration(ctx context.Context, config *sport.SportConfiguration) error
	UpdateConfiguration(ctx context.Context, config *sport.SportConfiguration) error
	GetConfiguration(ctx context.Context, sportTypeID uint) (*sport.SportConfiguration, error)
}

// 请求和响应结构体

// CreateSportTypeRequest 创建运动类型请求
type CreateSportTypeRequest struct {
	Name        string                `json:"name" validate:"required,max=100"`
	Code        string                `json:"code" validate:"required,max=20,alphanum"`
	Category    sport.SportCategory   `json:"category" validate:"required,oneof=esports traditional"`
	Icon        string                `json:"icon" validate:"max=255"`
	Banner      string                `json:"banner" validate:"max=255"`
	Description string                `json:"description" validate:"max=1000"`
	IsActive    bool                  `json:"is_active"`
	SortOrder   int                   `json:"sort_order"`
}

// UpdateSportTypeRequest 更新运动类型请求
type UpdateSportTypeRequest struct {
	Name        *string               `json:"name" validate:"omitempty,max=100"`
	Code        *string               `json:"code" validate:"omitempty,max=20,alphanum"`
	Category    *sport.SportCategory  `json:"category" validate:"omitempty,oneof=esports traditional"`
	Icon        *string               `json:"icon" validate:"omitempty,max=255"`
	Banner      *string               `json:"banner" validate:"omitempty,max=255"`
	Description *string               `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool                 `json:"is_active"`
	SortOrder   *int                  `json:"sort_order"`
}

// ListSportTypesRequest 运动类型列表请求
type ListSportTypesRequest struct {
	Category string `json:"category" validate:"omitempty,oneof=esports traditional"`
	IsActive *bool  `json:"is_active"`
	OrderBy  string `json:"order_by" validate:"omitempty,oneof=name code sort_order created_at"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// ListSportTypesResponse 运动类型列表响应
type ListSportTypesResponse struct {
	SportTypes []*sport.SportType `json:"sport_types"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// UpdateSportConfigurationRequest 更新运动配置请求
type UpdateSportConfigurationRequest struct {
	// 功能开关
	EnableRealtime    *bool `json:"enable_realtime"`
	EnableChat        *bool `json:"enable_chat"`
	EnableVoting      *bool `json:"enable_voting"`
	EnablePrediction  *bool `json:"enable_prediction"`
	EnableLeaderboard *bool `json:"enable_leaderboard"`

	// 预测设置
	AllowModification    *bool `json:"allow_modification"`
	MaxModifications     *int  `json:"max_modifications" validate:"omitempty,min=0,max=10"`
	ModificationDeadline *int  `json:"modification_deadline" validate:"omitempty,min=0,max=1440"` // 最大24小时

	// 投票设置
	EnableSelfVoting *bool `json:"enable_self_voting"`
	MaxVotesPerUser  *int  `json:"max_votes_per_user" validate:"omitempty,min=1,max=100"`
	VotingDeadline   *int  `json:"voting_deadline" validate:"omitempty,min=0,max=1440"` // 最大24小时
}

// BatchUpdateConfigRequest 批量更新配置请求
type BatchUpdateConfigRequest struct {
	SportTypeIDs []uint                              `json:"sport_type_ids" validate:"required,min=1"`
	Config       *UpdateSportConfigurationRequest   `json:"config" validate:"required"`
}

// SportTypeStats 运动类型统计信息
type SportTypeStats struct {
	SportTypeID     uint      `json:"sport_type_id"`
	MatchCount      int64     `json:"match_count"`
	PredictionCount int64     `json:"prediction_count"`
	UserCount       int64     `json:"user_count"`
	LastMatchTime   *time.Time `json:"last_match_time"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListSportTypesOptions 运动类型列表查询选项（仓储层使用）
type ListSportTypesOptions struct {
	Category string
	IsActive *bool
	OrderBy  string
	Limit    int
	Offset   int
}