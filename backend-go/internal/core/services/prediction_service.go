package services

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/shared"
	"backend-go/internal/core/domain/user"
	"backend-go/pkg/response"
)

// PredictionService 预测服务实现
type PredictionService struct {
	predictionRepo  prediction.Repository
	voteRepo        prediction.VoteRepository
	matchRepo       match.Repository
	userRepo        user.Repository
	scoringRuleRepo prediction.ScoringRuleRepository
	eventBus        shared.EventBus
}

// NewPredictionService 创建预测服务
func NewPredictionService(
	predictionRepo prediction.Repository,
	voteRepo prediction.VoteRepository,
	matchRepo match.Repository,
	userRepo user.Repository,
	scoringRuleRepo prediction.ScoringRuleRepository,
	eventBus shared.EventBus,
) prediction.Service {
	return &PredictionService{
		predictionRepo:  predictionRepo,
		voteRepo:        voteRepo,
		matchRepo:       matchRepo,
		userRepo:        userRepo,
		scoringRuleRepo: scoringRuleRepo,
		eventBus:        eventBus,
	}
}

// CreatePrediction 创建预测
func (s *PredictionService) CreatePrediction(ctx context.Context, userID uint, req *prediction.CreatePredictionRequest) (*prediction.Prediction, error) {
	// 检查比赛是否存在
	matchEntity, err := s.matchRepo.GetByID(ctx, req.MatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	// 检查是否可以预测
	if !matchEntity.CanAcceptPredictions() {
		return nil, response.NewMatchStartedError(req.MatchID)
	}

	// 检查用户是否已经有预测
	existingPrediction, err := s.predictionRepo.GetPredictionByUserAndMatch(ctx, userID, req.MatchID)
	if err == nil && existingPrediction != nil {
		return nil, response.NewPredictionExistsError(userID, req.MatchID)
	}

	// 创建预测
	pred := &prediction.Prediction{
		UserID:          userID,
		MatchID:         req.MatchID,
		PredictedWinner: string(req.PredictedWinner),
		PredictedScoreA: req.PredictedScoreA,
		PredictedScoreB: req.PredictedScoreB,
	}

	if err := s.predictionRepo.CreatePrediction(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}

	// 加载关联数据
	pred.Match = matchEntity
	return pred, nil
}

// UpdatePrediction 更新预测
func (s *PredictionService) UpdatePrediction(ctx context.Context, userID uint, predictionID uint, req *prediction.UpdatePredictionRequest) (*prediction.Prediction, error) {
	// 获取预测
	pred, err := s.predictionRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	// 检查权限
	if pred.UserID != userID {
		return nil, response.NewForbiddenError("无权修改此预测")
	}

	// 检查比赛状态
	matchEntity, err := s.matchRepo.GetByID(ctx, pred.MatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if !matchEntity.CanAcceptPredictions() {
		return nil, response.NewMatchStartedError(pred.MatchID)
	}

	// 更新预测
	pred.PredictedWinner = string(req.PredictedWinner)
	pred.PredictedScoreA = req.PredictedScoreA
	pred.PredictedScoreB = req.PredictedScoreB
	pred.IncrementModificationCount()

	if err := s.predictionRepo.UpdatePrediction(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed to update prediction: %w", err)
	}

	pred.Match = matchEntity
	return pred, nil
}

// GetPrediction 获取预测详情
func (s *PredictionService) GetPrediction(ctx context.Context, id uint) (*prediction.Prediction, error) {
	pred, err := s.predictionRepo.GetPredictionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}
	return pred, nil
}

// GetPredictionsByMatch 获取比赛的所有预测
func (s *PredictionService) GetPredictionsByMatch(ctx context.Context, matchID uint, userID *uint) ([]prediction.PredictionWithVotes, error) {
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get predictions by match: %w", err)
	}
	return predictions, nil
}

// GetUserPredictions 获取用户的所有预测
func (s *PredictionService) GetUserPredictions(ctx context.Context, userID uint) ([]prediction.Prediction, error) {
	predictions, err := s.predictionRepo.GetPredictionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user predictions: %w", err)
	}
	return predictions, nil
}

// VotePrediction 投票支持预测
func (s *PredictionService) VotePrediction(ctx context.Context, userID uint, predictionID uint) error {
	// 获取预测
	pred, err := s.predictionRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		return fmt.Errorf("failed to get prediction: %w", err)
	}

	// 检查是否可以投票
	if err := prediction.CanVote(userID, pred); err != nil {
		return err
	}

	// 检查是否已经投票
	exists, err := s.voteRepo.ExistsVote(ctx, userID, predictionID)
	if err != nil {
		return fmt.Errorf("failed to check vote existence: %w", err)
	}
	if exists {
		return response.NewVoteExistsError(userID, predictionID)
	}

	// 创建投票并更新计数（事务性操作）
	vote := prediction.NewVote(userID, predictionID)
	if err := s.voteRepo.CreateVoteWithCount(ctx, vote); err != nil {
		return fmt.Errorf("failed to create vote with count: %w", err)
	}

	// 获取更新后的投票数
	updatedPred, err := s.predictionRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		// 记录错误但不影响投票操作
		fmt.Printf("Warning: failed to get updated prediction for event: %v", err)
	} else {
		// 发布投票事件
		if s.eventBus != nil {
			event := shared.NewEvent(shared.EventPredictionVoted, &shared.PredictionVotedPayload{
				PredictionID: predictionID,
				UserID:       userID,
				VoteCount:    updatedPred.VoteCount,
			})

			if err := s.eventBus.Publish(event); err != nil {
				fmt.Printf("Warning: failed to publish vote event: %v", err)
			}
		}
	}

	return nil
}

// UnvotePrediction 取消投票
func (s *PredictionService) UnvotePrediction(ctx context.Context, userID uint, predictionID uint) error {
	// 检查投票是否存在
	exists, err := s.voteRepo.ExistsVote(ctx, userID, predictionID)
	if err != nil {
		return fmt.Errorf("failed to check vote existence: %w", err)
	}
	if !exists {
		return response.NewNotFoundError("投票不存在")
	}

	// 删除投票并更新计数（事务性操作）
	if err := s.voteRepo.DeleteVoteWithCount(ctx, userID, predictionID); err != nil {
		return fmt.Errorf("failed to delete vote with count: %w", err)
	}

	// 获取更新后的投票数
	updatedPred, err := s.predictionRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		// 记录错误但不影响取消投票操作
		fmt.Printf("Warning: failed to get updated prediction for unvote event: %v", err)
	} else {
		// 发布取消投票事件
		if s.eventBus != nil {
			event := shared.NewEvent(shared.EventPredictionUnvoted, &shared.PredictionVotedPayload{
				PredictionID: predictionID,
				UserID:       userID,
				VoteCount:    updatedPred.VoteCount,
			})

			if err := s.eventBus.Publish(event); err != nil {
				fmt.Printf("Warning: failed to publish unvote event: %v", err)
			}
		}
	}

	return nil
}

// CalculatePoints 计算比赛结束后的积分
func (s *PredictionService) CalculatePoints(ctx context.Context, matchID uint) error {
	return s.CalculatePointsWithCustomRule(ctx, matchID, nil)
}

// CalculatePointsWithCustomRule 使用自定义规则计算积分
func (s *PredictionService) CalculatePointsWithCustomRule(ctx context.Context, matchID uint, ruleID *uint) error {
	// 获取比赛信息
	matchEntity, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	if !matchEntity.IsFinished() {
		return response.NewBadRequestError("比赛尚未结束", nil)
	}

	// 获取积分规则（兼容旧部署：无规则仓库时直接使用默认规则）
	var rule *prediction.ScoringRule
	if s.scoringRuleRepo != nil {
		if ruleID != nil {
			rule, err = s.scoringRuleRepo.GetScoringRuleByID(ctx, *ruleID)
			if err != nil {
				return fmt.Errorf("failed to get scoring rule: %w", err)
			}
		} else {
			rule, err = s.scoringRuleRepo.GetActiveScoringRule(ctx)
			if err != nil && !response.IsErrorCode(err, response.CodeNoActiveScoringRule) {
				return fmt.Errorf("failed to get active scoring rule: %w", err)
			}
		}
	}

	// 获取比赛的所有预测
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, nil)
	if err != nil {
		return fmt.Errorf("failed to get predictions: %w", err)
	}

	// 计算每个预测的积分
	for _, predWithVotes := range predictions {
		pred := predWithVotes.Prediction
		pred.Match = matchEntity

		var points int
		if rule != nil {
			points = pred.CalculatePointsWithRule(rule)
		} else {
			points = pred.CalculatePoints()
		}

		// 更新预测积分
		if err := s.predictionRepo.UpdatePredictionPoints(ctx, pred.ID, points, pred.IsCorrect); err != nil {
			return fmt.Errorf("failed to update prediction points: %w", err)
		}

		// 更新用户总积分
		userEntity, err := s.userRepo.GetByID(ctx, pred.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		userEntity.Points += points
		if err := s.userRepo.Update(ctx, userEntity); err != nil {
			return fmt.Errorf("failed to update user points: %w", err)
		}
	}

	return nil
}

// GetFeaturedPredictions 获取精选预测
func (s *PredictionService) GetFeaturedPredictions(ctx context.Context) ([]prediction.PredictionWithVotes, error) {
	predictions, err := s.predictionRepo.GetFeaturedPredictions(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured predictions: %w", err)
	}
	return predictions, nil
}

// UpdateFeaturedPredictions 更新精选预测
func (s *PredictionService) UpdateFeaturedPredictions(ctx context.Context) error {
	// 这里可以实现更复杂的精选逻辑
	// 例如：根据投票数、准确性等因素选择精选预测
	// 暂时简单实现：将投票数最高的预测设为精选

	// 获取所有预测，按投票数排序
	// 这里需要扩展仓储接口来支持更复杂的查询
	// 暂时返回成功
	return nil
}
