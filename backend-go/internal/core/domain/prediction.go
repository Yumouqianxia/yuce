package domain

import (
	"fmt"
	"time"

	"backend-go/internal/core/domain/user"
)

// Prediction 预测领域实体
type Prediction struct {
	ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint       `gorm:"column:userId;not null;index:idx_user_match,unique" json:"userId"`
	MatchID           uint       `gorm:"column:matchId;not null;index:idx_user_match,unique" json:"matchId"`
	PredictedWinner   string     `gorm:"column:predictedWinner;size:1;not null" json:"predictedWinner"` // 'A' 或 'B'
	PredictedScoreA   int        `gorm:"column:predictedScoreA;not null" json:"predictedScoreA"`
	PredictedScoreB   int        `gorm:"column:predictedScoreB;not null" json:"predictedScoreB"`
	IsVerified        bool       `gorm:"column:isVerified;default:false;not null" json:"isVerified"`
	IsCorrect         bool       `gorm:"column:isCorrect;default:false;not null" json:"isCorrect"`
	EarnedPoints      int        `gorm:"column:earnedPoints;default:0;not null" json:"pointsEarned"`
	IsProcessed       bool       `gorm:"column:isProcessed;default:false;not null" json:"isProcessed"`
	ModificationCount int        `gorm:"column:modification_count;default:0;not null" json:"modificationCount"`
	LastModifiedAt    *time.Time `gorm:"column:last_modified_at;type:datetime" json:"lastModifiedAt"`
	CreatedAt         time.Time  `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`

	// 关联关系
	User          user.User                `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Match         Match                    `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"match,omitempty"`
	Modifications []PredictionModification `gorm:"foreignKey:PredictionID;constraint:OnDelete:CASCADE" json:"modifications,omitempty"`

	// 虚拟字段（不存储在数据库中）
	VoteCount    int  `gorm:"-" json:"voteCount,omitempty"`
	HasUserVoted bool `gorm:"-" json:"hasUserVoted,omitempty"`
	IsFeatured   bool `gorm:"-" json:"isFeatured,omitempty"`
}

// TableName 指定表名
func (Prediction) TableName() string {
	return "predictions"
}

// CanModify 检查是否可以修改预测
func (p *Prediction) CanModify() bool {
	// 如果比赛已经开始或完成，不能修改
	if p.Match.Status != MatchStatusUpcoming {
		return false
	}

	// 如果已经处理过，不能修改
	if p.IsProcessed {
		return false
	}

	return true
}

// Modify 修改预测
func (p *Prediction) Modify(winner string, scoreA, scoreB int) error {
	if !p.CanModify() {
		return ErrCannotModifyPrediction
	}

	if winner != "A" && winner != "B" {
		return ErrInvalidWinner
	}

	// 记录修改
	oldWinner := p.PredictedWinner
	oldScoreA := p.PredictedScoreA
	oldScoreB := p.PredictedScoreB

	p.PredictedWinner = winner
	p.PredictedScoreA = scoreA
	p.PredictedScoreB = scoreB
	p.ModificationCount++
	now := time.Now()
	p.LastModifiedAt = &now

	// 创建修改记录
	modification := PredictionModification{
		UserID:           p.UserID,
		MatchID:          p.MatchID,
		PredictionID:     p.ID,
		OriginalWinner:   oldWinner,
		OriginalScore:    formatScore(oldScoreA, oldScoreB),
		NewWinner:        winner,
		NewScore:         formatScore(scoreA, scoreB),
		ModificationType: getModificationType(oldWinner, winner, oldScoreA, oldScoreB, scoreA, scoreB),
	}

	p.Modifications = append(p.Modifications, modification)

	return nil
}

// CalculatePoints 计算预测得分
func (p *Prediction) CalculatePoints(match *Match) int {
	if !match.IsFinished() {
		return 0
	}

	points := 0

	// 预测获胜者正确：基础分 10 分
	if p.PredictedWinner == match.Winner {
		points += 10
		p.IsCorrect = true

		// 预测比分完全正确：额外 20 分
		if p.PredictedScoreA == match.ScoreA && p.PredictedScoreB == match.ScoreB {
			points += 20
		} else {
			// 预测比分差距正确：额外 10 分
			predictedDiff := abs(p.PredictedScoreA - p.PredictedScoreB)
			actualDiff := abs(match.ScoreA - match.ScoreB)
			if predictedDiff == actualDiff {
				points += 10
			}
		}
	}

	// 根据修改次数减分（跳车惩罚）
	penalty := p.ModificationCount * 2
	points -= penalty

	if points < 0 {
		points = 0
	}

	p.EarnedPoints = points
	return points
}

// IsExactMatch 检查是否完全匹配
func (p *Prediction) IsExactMatch(match *Match) bool {
	return p.PredictedWinner == match.Winner &&
		p.PredictedScoreA == match.ScoreA &&
		p.PredictedScoreB == match.ScoreB
}

// GetPredictedScore 获取预测比分字符串
func (p *Prediction) GetPredictedScore() string {
	return formatScore(p.PredictedScoreA, p.PredictedScoreB)
}

// PredictionModification 预测修改记录
type PredictionModification struct {
	ID               uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint             `gorm:"column:user_id;not null" json:"userId"`
	MatchID          uint             `gorm:"column:match_id;not null" json:"matchId"`
	PredictionID     uint             `gorm:"column:prediction_id;not null" json:"predictionId"`
	OriginalWinner   string           `gorm:"column:original_winner;size:1" json:"originalWinner"`
	OriginalScore    string           `gorm:"column:original_score;size:20" json:"originalScore"`
	NewWinner        string           `gorm:"column:new_winner;size:1" json:"newWinner"`
	NewScore         string           `gorm:"column:new_score;size:20" json:"newScore"`
	ModificationType ModificationType `gorm:"column:modification_type;size:10;not null" json:"modificationType"`
	CreatedAt        time.Time        `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	// 关联关系
	User       user.User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Match      Match      `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"match,omitempty"`
	Prediction Prediction `gorm:"foreignKey:PredictionID;constraint:OnDelete:CASCADE" json:"prediction,omitempty"`
}

// TableName 指定表名
func (PredictionModification) TableName() string {
	return "prediction_modifications"
}

// ModificationType 修改类型枚举
type ModificationType string

const (
	ModificationWinner ModificationType = "winner"
	ModificationScore  ModificationType = "score"
	ModificationBoth   ModificationType = "both"
)

// 辅助函数
func formatScore(scoreA, scoreB int) string {
	return fmt.Sprintf("%d-%d", scoreA, scoreB)
}

func getModificationType(oldWinner, newWinner string, oldScoreA, oldScoreB, newScoreA, newScoreB int) ModificationType {
	winnerChanged := oldWinner != newWinner
	scoreChanged := oldScoreA != newScoreA || oldScoreB != newScoreB

	if winnerChanged && scoreChanged {
		return ModificationBoth
	} else if winnerChanged {
		return ModificationWinner
	} else if scoreChanged {
		return ModificationScore
	}

	return ModificationScore // 默认
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
