package mysql

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/prediction"
	"backend-go/pkg/response"
	"gorm.io/gorm"
)

// PredictionRepository 预测仓储 MySQL 实现
type PredictionRepository struct {
	db *gorm.DB
}

// NewPredictionRepository 创建预测仓储
func NewPredictionRepository(db *gorm.DB) prediction.Repository {
	return &PredictionRepository{
		db: db,
	}
}

// CreatePrediction 创建预测
func (r *PredictionRepository) CreatePrediction(ctx context.Context, pred *prediction.Prediction) error {
	if err := r.db.WithContext(ctx).Create(pred).Error; err != nil {
		return fmt.Errorf("failed to create prediction: %w", err)
	}
	return nil
}

// GetPredictionByID 根据 ID 获取预测
func (r *PredictionRepository) GetPredictionByID(ctx context.Context, id uint) (*prediction.Prediction, error) {
	var pred prediction.Prediction
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Match").
		First(&pred, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NewNotFoundError("预测不存在")
		}
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	return &pred, nil
}

// GetPredictionByUserAndMatch 根据用户和比赛获取预测
func (r *PredictionRepository) GetPredictionByUserAndMatch(ctx context.Context, userID, matchID uint) (*prediction.Prediction, error) {
	var pred prediction.Prediction
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Match").
		Where("user_id = ? AND match_id = ?", userID, matchID).
		First(&pred).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NewNotFoundError("预测不存在")
		}
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	return &pred, nil
}

// UpdatePrediction 更新预测
func (r *PredictionRepository) UpdatePrediction(ctx context.Context, pred *prediction.Prediction) error {
	if err := r.db.WithContext(ctx).Save(pred).Error; err != nil {
		return fmt.Errorf("failed to update prediction: %w", err)
	}
	return nil
}

// GetPredictionsByMatch 获取比赛的所有预测
func (r *PredictionRepository) GetPredictionsByMatch(ctx context.Context, matchID uint, userID *uint) ([]prediction.PredictionWithVotes, error) {
	var predictions []prediction.Prediction

	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("Match").
		Where("matchId = ?", matchID).
		Order("vote_count DESC, createdAt ASC")

	if err := query.Find(&predictions).Error; err != nil {
		return nil, fmt.Errorf("failed to get predictions by match: %w", err)
	}

	// 转换为 PredictionWithVotes 并检查用户是否已投票
	result := make([]prediction.PredictionWithVotes, len(predictions))
	for i, pred := range predictions {
		hasUserVoted := false

		if userID != nil {
			var voteCount int64
			r.db.WithContext(ctx).
				Model(&prediction.Vote{}).
				Where("user_id = ? AND prediction_id = ?", *userID, pred.ID).
				Count(&voteCount)
			hasUserVoted = voteCount > 0
		}

		result[i] = prediction.PredictionWithVotes{
			Prediction:   &pred,
			HasUserVoted: hasUserVoted,
		}
	}

	return result, nil
}

// GetPredictionsByUser 获取用户的所有预测
func (r *PredictionRepository) GetPredictionsByUser(ctx context.Context, userID uint) ([]prediction.Prediction, error) {
	var predictions []prediction.Prediction

	err := r.db.WithContext(ctx).
		Preload("Match").
		Where("userId = ?", userID).
		Order("createdAt DESC").
		Find(&predictions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get predictions by user: %w", err)
	}

	return predictions, nil
}

// UpdatePredictionPoints 更新预测积分
func (r *PredictionRepository) UpdatePredictionPoints(ctx context.Context, predictionID uint, points int, isCorrect bool) error {
	err := r.db.WithContext(ctx).
		Model(&prediction.Prediction{}).
		Where("id = ?", predictionID).
		Updates(map[string]interface{}{
			"earnedPoints": points,
			"isCorrect":    isCorrect,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update prediction points: %w", err)
	}

	return nil
}

// GetFeaturedPredictions 获取精选预测
func (r *PredictionRepository) GetFeaturedPredictions(ctx context.Context, limit int) ([]prediction.PredictionWithVotes, error) {
	var predictions []prediction.Prediction

	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Match").
		Where("is_featured = ?", true).
		Order("vote_count DESC, created_at DESC").
		Limit(limit).
		Find(&predictions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get featured predictions: %w", err)
	}

	// 转换为 PredictionWithVotes
	result := make([]prediction.PredictionWithVotes, len(predictions))
	for i, pred := range predictions {
		result[i] = prediction.PredictionWithVotes{
			Prediction:   &pred,
			HasUserVoted: false, // 精选预测不需要检查用户投票状态
		}
	}

	return result, nil
}

// SetFeatured 设置精选状态
func (r *PredictionRepository) SetFeatured(ctx context.Context, predictionID uint, featured bool) error {
	err := r.db.WithContext(ctx).
		Model(&prediction.Prediction{}).
		Where("id = ?", predictionID).
		Update("is_featured", featured).Error

	if err != nil {
		return fmt.Errorf("failed to set featured status: %w", err)
	}

	return nil
}

// DeletePrediction 删除预测
func (r *PredictionRepository) DeletePrediction(ctx context.Context, id uint) error {
	// 先删除相关的投票
	if err := r.db.WithContext(ctx).Where("prediction_id = ?", id).Delete(&prediction.Vote{}).Error; err != nil {
		return fmt.Errorf("failed to delete related votes: %w", err)
	}

	// 删除预测
	if err := r.db.WithContext(ctx).Delete(&prediction.Prediction{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete prediction: %w", err)
	}

	return nil
}
