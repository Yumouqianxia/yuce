package prediction

import (
	"errors"
	"time"

	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/user"
)

// 投票相关错误
var (
	ErrCannotVoteOwnPrediction = errors.New("cannot vote for own prediction")
	ErrMatchAlreadyStarted     = errors.New("match already started")
)

// Prediction 预测实体
type Prediction struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"userId" gorm:"column:userId;index;not null"`
	MatchID           uint      `json:"matchId" gorm:"column:matchId;index;not null"`
	PredictedWinner   string    `json:"predictedWinner" gorm:"column:predictedWinner;size:10;not null"`
	PredictedScoreA   int       `json:"predictedScoreA" gorm:"column:predictedScoreA;not null"`
	PredictedScoreB   int       `json:"predictedScoreB" gorm:"column:predictedScoreB;not null"`
	IsCorrect         bool      `json:"isCorrect" gorm:"column:isCorrect;default:false"`
	EarnedPoints      int       `json:"pointsEarned" gorm:"column:earnedPoints;default:0"`
	ModificationCount int       `json:"modificationCount" gorm:"column:modification_count;default:0"`
	VoteCount         int       `json:"voteCount" gorm:"column:vote_count;default:0"`
	IsFeatured        bool      `json:"isFeatured" gorm:"column:is_featured;default:false"`
	CreatedAt         time.Time `json:"createdAt" gorm:"column:createdAt"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"column:updatedAt"`

	// 关联关系
	User  *user.User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Match *match.Match `json:"match,omitempty" gorm:"foreignKey:MatchID"`
	Votes []Vote       `json:"votes,omitempty" gorm:"foreignKey:PredictionID"`
}

// Vote 投票实体
type Vote struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"userId" gorm:"column:user_id;not null;index:idx_user_prediction,unique"`
	PredictionID uint      `json:"predictionId" gorm:"column:prediction_id;not null;index:idx_user_prediction,unique"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`

	// 关联关系
	User       *user.User  `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Prediction *Prediction `json:"prediction,omitempty" gorm:"foreignKey:PredictionID;constraint:OnDelete:CASCADE"`
}

// TableName 指定预测表名
func (Prediction) TableName() string {
	return "predictions"
}

// TableName 指定投票表名
func (Vote) TableName() string {
	return "votes"
}

// NewVote 创建新投票
func NewVote(userID, predictionID uint) *Vote {
	return &Vote{
		UserID:       userID,
		PredictionID: predictionID,
	}
}

// CanVote 检查用户是否可以对预测投票
func CanVote(userID uint, prediction *Prediction) error {
	// 不能给自己的预测投票
	if userID == prediction.UserID {
		return ErrCannotVoteOwnPrediction
	}

	// 比赛必须还未开始
	if prediction.Match != nil && !prediction.Match.CanPredict() {
		return ErrMatchAlreadyStarted
	}

	return nil
}

// GetPredictedWinner 获取预测的获胜者
func (p *Prediction) GetPredictedWinner() match.Winner {
	return match.Winner(p.PredictedWinner)
}

// SetPredictedWinner 设置预测的获胜者
func (p *Prediction) SetPredictedWinner(winner match.Winner) {
	p.PredictedWinner = string(winner)
}

// CanModify 检查是否可以修改预测
func (p *Prediction) CanModify() bool {
	if p.Match == nil {
		return false
	}
	return p.Match.CanPredict()
}

// CalculatePoints 使用默认规则计算预测积分（向后兼容）
func (p *Prediction) CalculatePoints() int {
	if p.Match == nil || !p.Match.IsFinished() {
		return 0
	}

	points := 0

	// 预测获胜者正确：基础分 10 分
	if p.PredictedWinner == p.Match.Winner {
		points += 10
		p.IsCorrect = true

		// 预测比分完全正确：额外 20 分
		if p.PredictedScoreA == p.Match.ScoreA && p.PredictedScoreB == p.Match.ScoreB {
			points += 20
		}
	}

	// 根据投票数给予热门奖励
	if p.VoteCount >= 10 {
		points += 5
	}

	return points
}

// CalculatePointsWithRule 使用指定规则计算预测积分
func (p *Prediction) CalculatePointsWithRule(rule *ScoringRule) int {
	if p.Match == nil || !p.Match.IsFinished() || rule == nil {
		return 0
	}

	// 使用规则计算基础积分
	points := rule.CalculatePoints(p)

	// 根据投票数给予热门奖励（保持原有逻辑）
	if p.VoteCount >= 10 {
		points += 5
	}

	// 更新预测正确性标记
	teamCorrect := p.PredictedWinner == p.Match.Winner
	scoreCorrect := p.PredictedScoreA == p.Match.ScoreA && p.PredictedScoreB == p.Match.ScoreB
	p.IsCorrect = teamCorrect || scoreCorrect

	return points
}

// IncrementVoteCount 增加投票数
func (p *Prediction) IncrementVoteCount() {
	p.VoteCount++
}

// DecrementVoteCount 减少投票数
func (p *Prediction) DecrementVoteCount() {
	if p.VoteCount > 0 {
		p.VoteCount--
	}
}

// IncrementModificationCount 增加修改次数
func (p *Prediction) IncrementModificationCount() {
	p.ModificationCount++
}

// PredictionWithVotes 带投票信息的预测
type PredictionWithVotes struct {
	*Prediction
	HasUserVoted bool `json:"has_user_voted"`
}

// VoteStats 投票统计
type VoteStats struct {
	PredictionID uint `json:"prediction_id"`
	VoteCount    int  `json:"vote_count"`
	IsFeatured   bool `json:"is_featured"`
}

// GetVoteThreshold 获取热门预测的投票阈值
func GetVoteThreshold() int {
	return 5 // 5票以上为热门预测
}

// IsFeaturedByVotes 根据投票数判断是否为精选预测
func (p *Prediction) IsFeaturedByVotes() bool {
	return p.VoteCount >= GetVoteThreshold()
}

// UpdateFeaturedStatus 根据投票数更新精选状态
func (p *Prediction) UpdateFeaturedStatus() {
	p.IsFeatured = p.IsFeaturedByVotes()
}
