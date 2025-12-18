package sport

import (
	"time"
)

// SportCategory 运动类别枚举
type SportCategory string

const (
	SportCategoryEsports     SportCategory = "esports"     // 电子竞技
	SportCategoryTraditional SportCategory = "traditional" // 传统体育
)

// SportType 运动类型实体
type SportType struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"uniqueIndex;size:100;not null"` // LOL、王者荣耀、足球等
	Code        string        `json:"code" gorm:"uniqueIndex;size:20;not null"`  // lol、wzry、football等
	Category    SportCategory `json:"category" gorm:"size:20;not null"`          // esports 或 traditional
	Icon        string        `json:"icon" gorm:"size:255"`                      // 运动图标URL
	Banner      string        `json:"banner" gorm:"size:255"`                    // 运动横幅图URL
	Description string        `json:"description" gorm:"type:text"`              // 运动描述
	IsActive    bool          `json:"is_active" gorm:"default:true"`             // 是否启用
	SortOrder   int           `json:"sort_order" gorm:"default:0"`               // 首页显示顺序
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// 关联配置 (一对一关系)
	Configuration *SportConfiguration `json:"configuration,omitempty" gorm:"foreignKey:SportTypeID"`
}

// SportConfiguration 运动配置实体
type SportConfiguration struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	SportTypeID uint `json:"sport_type_id" gorm:"uniqueIndex;not null"`

	// 功能开关
	EnableRealtime    bool `json:"enable_realtime" gorm:"default:true"`     // 启用实时通信
	EnableChat        bool `json:"enable_chat" gorm:"default:false"`        // 启用聊天功能
	EnableVoting      bool `json:"enable_voting" gorm:"default:true"`       // 启用投票功能
	EnablePrediction  bool `json:"enable_prediction" gorm:"default:true"`   // 启用预测功能
	EnableLeaderboard bool `json:"enable_leaderboard" gorm:"default:true"`  // 启用排行榜

	// 预测设置
	AllowModification    bool `json:"allow_modification" gorm:"default:true"`    // 允许修改预测
	MaxModifications     int  `json:"max_modifications" gorm:"default:3"`        // 最大修改次数
	ModificationDeadline int  `json:"modification_deadline" gorm:"default:30"`   // 比赛开始前N分钟禁止修改

	// 投票设置
	EnableSelfVoting bool `json:"enable_self_voting" gorm:"default:false"`  // 允许给自己投票
	MaxVotesPerUser  int  `json:"max_votes_per_user" gorm:"default:10"`     // 每用户最大投票数
	VotingDeadline   int  `json:"voting_deadline" gorm:"default:0"`         // 比赛开始前N分钟禁止投票，0表示无限制

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联运动类型
	SportType *SportType `json:"sport_type,omitempty" gorm:"foreignKey:SportTypeID"`
}

// ScoringRule 积分规则实体
type ScoringRule struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	SportTypeID uint   `json:"sport_type_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"type:text"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	// 基础积分设置
	BasePoints           int     `json:"base_points" gorm:"default:10"`           // 基础积分
	EnableDifficulty     bool    `json:"enable_difficulty" gorm:"default:false"` // 启用难度系数
	DifficultyMultiplier float64 `json:"difficulty_multiplier" gorm:"default:1.0;type:decimal(3,2)"` // 难度系数

	// 奖励组件开关
	EnableVoteReward bool `json:"enable_vote_reward" gorm:"default:false"` // 启用投票奖励
	VoteRewardPoints int  `json:"vote_reward_points" gorm:"default:1"`     // 每票奖励积分
	MaxVoteReward    int  `json:"max_vote_reward" gorm:"default:10"`       // 最大投票奖励

	EnableTimeReward bool `json:"enable_time_reward" gorm:"default:false"` // 启用时间奖励
	TimeRewardPoints int  `json:"time_reward_points" gorm:"default:5"`     // 时间奖励积分
	TimeRewardHours  int  `json:"time_reward_hours" gorm:"default:24"`     // 时间奖励小时数

	// 惩罚组件开关
	EnableModifyPenalty bool `json:"enable_modify_penalty" gorm:"default:false"` // 启用修改惩罚
	ModifyPenaltyPoints int  `json:"modify_penalty_points" gorm:"default:2"`     // 每次修改扣分
	MaxModifyPenalty    int  `json:"max_modify_penalty" gorm:"default:6"`        // 最大修改惩罚

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联运动类型
	SportType *SportType `json:"sport_type,omitempty" gorm:"foreignKey:SportTypeID"`
}

// IsValidCategory 检查运动类别是否有效
func IsValidCategory(category string) bool {
	return category == string(SportCategoryEsports) || category == string(SportCategoryTraditional)
}

// IsEsports 检查是否为电子竞技
func (st *SportType) IsEsports() bool {
	return st.Category == SportCategoryEsports
}

// IsTraditional 检查是否为传统体育
func (st *SportType) IsTraditional() bool {
	return st.Category == SportCategoryTraditional
}

// GetDisplayName 获取显示名称
func (st *SportType) GetDisplayName() string {
	if st.Name != "" {
		return st.Name
	}
	return st.Code
}

// HasConfiguration 检查是否有配置
func (st *SportType) HasConfiguration() bool {
	return st.Configuration != nil
}

// IsFeatureEnabled 检查功能是否启用
func (sc *SportConfiguration) IsFeatureEnabled(feature string) bool {
	switch feature {
	case "realtime":
		return sc.EnableRealtime
	case "chat":
		return sc.EnableChat
	case "voting":
		return sc.EnableVoting
	case "prediction":
		return sc.EnablePrediction
	case "leaderboard":
		return sc.EnableLeaderboard
	default:
		return false
	}
}

// CanModifyPrediction 检查是否可以修改预测
func (sc *SportConfiguration) CanModifyPrediction(modificationCount int) bool {
	if !sc.AllowModification {
		return false
	}
	return modificationCount < sc.MaxModifications
}

// CanVote 检查是否可以投票
func (sc *SportConfiguration) CanVote(userVoteCount int, isSelfVote bool) bool {
	if !sc.EnableVoting {
		return false
	}
	if isSelfVote && !sc.EnableSelfVoting {
		return false
	}
	return userVoteCount < sc.MaxVotesPerUser
}