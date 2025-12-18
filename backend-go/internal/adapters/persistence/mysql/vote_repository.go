package mysql

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/prediction"
	"backend-go/pkg/response"
	"gorm.io/gorm"
)

// VoteRepository 投票仓储 MySQL 实现
type VoteRepository struct {
	db *gorm.DB
}

// NewVoteRepository 创建投票仓储
func NewVoteRepository(db *gorm.DB) prediction.VoteRepository {
	return &VoteRepository{
		db: db,
	}
}

// CreateVote 创建投票
func (r *VoteRepository) CreateVote(ctx context.Context, vote *prediction.Vote) error {
	if err := r.db.WithContext(ctx).Create(vote).Error; err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}
	return nil
}

// GetVote 获取投票
func (r *VoteRepository) GetVote(ctx context.Context, userID, predictionID uint) (*prediction.Vote, error) {
	var vote prediction.Vote
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Prediction").
		Where("user_id = ? AND prediction_id = ?", userID, predictionID).
		First(&vote).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NewNotFoundError("投票不存在")
		}
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}

	return &vote, nil
}

// DeleteVote 删除投票
func (r *VoteRepository) DeleteVote(ctx context.Context, userID, predictionID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND prediction_id = ?", userID, predictionID).
		Delete(&prediction.Vote{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete vote: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return response.NewNotFoundError("投票不存在")
	}

	return nil
}

// GetVotesByPrediction 获取预测的所有投票
func (r *VoteRepository) GetVotesByPrediction(ctx context.Context, predictionID uint) ([]prediction.Vote, error) {
	var votes []prediction.Vote

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("prediction_id = ?", predictionID).
		Order("created_at DESC").
		Find(&votes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get votes by prediction: %w", err)
	}

	return votes, nil
}

// GetVotesByUser 获取用户的所有投票
func (r *VoteRepository) GetVotesByUser(ctx context.Context, userID uint) ([]prediction.Vote, error) {
	var votes []prediction.Vote

	err := r.db.WithContext(ctx).
		Preload("Prediction").
		Preload("Prediction.Match").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&votes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get votes by user: %w", err)
	}

	return votes, nil
}

// ExistsVote 检查投票是否存在
func (r *VoteRepository) ExistsVote(ctx context.Context, userID, predictionID uint) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&prediction.Vote{}).
		Where("user_id = ? AND prediction_id = ?", userID, predictionID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check vote existence: %w", err)
	}

	return count > 0, nil
}

// GetVoteCount 获取预测的投票数
func (r *VoteRepository) GetVoteCount(ctx context.Context, predictionID uint) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&prediction.Vote{}).
		Where("prediction_id = ?", predictionID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}

	return int(count), nil
}

// CreateVoteWithCount 创建投票并更新计数（事务性操作）
func (r *VoteRepository) CreateVoteWithCount(ctx context.Context, vote *prediction.Vote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建投票
		if err := tx.Create(vote).Error; err != nil {
			return fmt.Errorf("failed to create vote: %w", err)
		}

		// 更新预测投票数
		if err := tx.Model(&prediction.Prediction{}).
			Where("id = ?", vote.PredictionID).
			UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error; err != nil {
			return fmt.Errorf("failed to update vote count: %w", err)
		}

		return nil
	})
}

// DeleteVoteWithCount 删除投票并更新计数（事务性操作）
func (r *VoteRepository) DeleteVoteWithCount(ctx context.Context, userID, predictionID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除投票
		result := tx.Where("user_id = ? AND prediction_id = ?", userID, predictionID).
			Delete(&prediction.Vote{})

		if result.Error != nil {
			return fmt.Errorf("failed to delete vote: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return response.NewNotFoundError("投票不存在")
		}

		// 更新预测投票数
		if err := tx.Model(&prediction.Prediction{}).
			Where("id = ?", predictionID).
			UpdateColumn("vote_count", gorm.Expr("vote_count - 1")).Error; err != nil {
			return fmt.Errorf("failed to update vote count: %w", err)
		}

		return nil
	})
}

// GetVoteStats 获取投票统计
func (r *VoteRepository) GetVoteStats(ctx context.Context, predictionIDs []uint) ([]prediction.VoteStats, error) {
	var stats []prediction.VoteStats

	err := r.db.WithContext(ctx).
		Model(&prediction.Vote{}).
		Select("prediction_id, COUNT(*) as vote_count").
		Where("prediction_id IN ?", predictionIDs).
		Group("prediction_id").
		Find(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get vote stats: %w", err)
	}

	// 设置精选状态
	for i := range stats {
		stats[i].IsFeatured = stats[i].VoteCount >= prediction.GetVoteThreshold()
	}

	return stats, nil
}

// GetTopVotedPredictions 获取投票数最高的预测
func (r *VoteRepository) GetTopVotedPredictions(ctx context.Context, limit int) ([]prediction.VoteStats, error) {
	var stats []prediction.VoteStats

	err := r.db.WithContext(ctx).
		Model(&prediction.Vote{}).
		Select("prediction_id, COUNT(*) as vote_count").
		Group("prediction_id").
		Order("vote_count DESC").
		Limit(limit).
		Find(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get top voted predictions: %w", err)
	}

	// 设置精选状态
	for i := range stats {
		stats[i].IsFeatured = stats[i].VoteCount >= prediction.GetVoteThreshold()
	}

	return stats, nil
}
