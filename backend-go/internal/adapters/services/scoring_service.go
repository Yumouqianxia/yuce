package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/scoring"
	"backend-go/internal/core/domain/user"
)

// scoringService 积分计算服务实现
type scoringService struct {
	predictionRepo     prediction.Repository
	predictionRuleRepo prediction.ScoringRuleRepository
	userRepo           user.Repository
	matchRepo          match.Repository
	scoringRepo        scoring.Repository
	calculator         scoring.Calculator
	logger             *logrus.Logger
}

// NewScoringService 创建积分计算服务
func NewScoringService(
	predictionRepo prediction.Repository,
	predictionRuleRepo prediction.ScoringRuleRepository,
	userRepo user.Repository,
	matchRepo match.Repository,
	scoringRepo scoring.Repository,
	calculator scoring.Calculator,
	logger *logrus.Logger,
) scoring.Service {
	return &scoringService{
		predictionRepo:     predictionRepo,
		predictionRuleRepo: predictionRuleRepo,
		userRepo:           userRepo,
		matchRepo:          matchRepo,
		scoringRepo:        scoringRepo,
		calculator:         calculator,
		logger:             logger,
	}
}

// CalculateMatchPoints 计算比赛结束后的所有预测积分
func (s *scoringService) CalculateMatchPoints(ctx context.Context, matchID uint) (*scoring.MatchPointsCalculation, error) {
	return s.CalculateMatchPointsWithRule(ctx, matchID, nil)
}

// CalculateMatchPointsWithRule 使用指定规则计算比赛积分
func (s *scoringService) CalculateMatchPointsWithRule(ctx context.Context, matchID uint, ruleID *uint) (*scoring.MatchPointsCalculation, error) {
	s.logger.WithFields(logrus.Fields{
		"match_id": matchID,
		"rule_id":  ruleID,
	}).Info("开始计算比赛积分")

	// 检查比赛是否已处理
	processed, err := s.scoringRepo.IsMatchProcessed(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("检查比赛处理状态失败: %w", err)
	}
	if processed {
		s.logger.WithField("match_id", matchID).Info("比赛积分已处理，返回现有结果")
		return s.scoringRepo.GetMatchCalculation(ctx, matchID)
	}

	// 获取比赛信息
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("获取比赛信息失败: %w", err)
	}

	if !match.IsFinished() {
		return nil, fmt.Errorf("比赛尚未结束，无法计算积分")
	}

	// 获取积分规则
	var rule *prediction.ScoringRule
	if ruleID != nil {
		rule, err = s.predictionRuleRepo.GetScoringRuleByID(ctx, *ruleID)
		if err != nil {
			return nil, fmt.Errorf("获取积分规则失败: %w", err)
		}
	} else {
		rule, err = s.predictionRuleRepo.GetActiveScoringRule(ctx)
		if err != nil {
			s.logger.WithError(err).Warn("获取激活积分规则失败，使用默认规则")
			rule = nil // 使用默认规则
		}
	}

	// 获取比赛的所有预测
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, nil)
	if err != nil {
		return nil, fmt.Errorf("获取比赛预测失败: %w", err)
	}

	// 计算每个预测的积分
	var results []scoring.PointsCalculationResult
	totalPoints := 0

	for _, predWithVotes := range predictions {
		pred := predWithVotes.Prediction
		result := s.calculator.Calculate(pred, rule)
		results = append(results, *result)
		totalPoints += result.Points

		// 更新预测的积分和正确性
		err = s.predictionRepo.UpdatePredictionPoints(ctx, pred.ID, result.Points, result.IsCorrect)
		if err != nil {
			s.logger.WithError(err).WithField("prediction_id", pred.ID).Error("更新预测积分失败")
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"prediction_id": pred.ID,
			"user_id":       pred.UserID,
			"points":        result.Points,
			"is_correct":    result.IsCorrect,
		}).Debug("预测积分计算完成")
	}

	// 创建计算结果
	calculation := &scoring.MatchPointsCalculation{
		MatchID:     matchID,
		Results:     results,
		TotalPoints: totalPoints,
		ProcessedAt: time.Now(),
	}

	// 保存计算结果
	err = s.scoringRepo.SavePointsCalculation(ctx, calculation)
	if err != nil {
		return nil, fmt.Errorf("保存积分计算结果失败: %w", err)
	}

	// 处理积分更新（更新用户积分）
	err = s.ProcessPointsUpdate(ctx, results, string(match.Tournament))
	if err != nil {
		s.logger.WithError(err).Error("处理积分更新失败")
		// 不返回错误，因为积分已经计算完成
	}

	s.logger.WithFields(logrus.Fields{
		"match_id":     matchID,
		"predictions":  len(results),
		"total_points": totalPoints,
	}).Info("比赛积分计算完成")

	return calculation, nil
}

// CalculatePredictionPoints 计算单个预测的积分
func (s *scoringService) CalculatePredictionPoints(ctx context.Context, predictionID uint, ruleID *uint) (*scoring.PointsCalculationResult, error) {
	// 获取预测信息
	pred, err := s.predictionRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		return nil, fmt.Errorf("获取预测信息失败: %w", err)
	}

	// 获取积分规则
	var rule *prediction.ScoringRule
	if ruleID != nil {
		rule, err = s.predictionRuleRepo.GetScoringRuleByID(ctx, *ruleID)
		if err != nil {
			return nil, fmt.Errorf("获取积分规则失败: %w", err)
		}
	} else {
		rule, err = s.predictionRuleRepo.GetActiveScoringRule(ctx)
		if err != nil {
			s.logger.WithError(err).Warn("获取激活积分规则失败，使用默认规则")
			rule = nil
		}
	}

	// 计算积分
	result := s.calculator.Calculate(pred, rule)

	return result, nil
}

// RecalculateAllPoints 重新计算所有已结束比赛的积分
func (s *scoringService) RecalculateAllPoints(ctx context.Context, ruleID *uint) error {
	s.logger.Info("开始重新计算所有积分")

	// 获取所有已结束的比赛
	matches, err := s.matchRepo.GetFinishedMatches(ctx)
	if err != nil {
		return fmt.Errorf("获取已结束比赛失败: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, match := range matches {
		_, err := s.CalculateMatchPointsWithRule(ctx, match.ID, ruleID)
		if err != nil {
			s.logger.WithError(err).WithField("match_id", match.ID).Error("重新计算比赛积分失败")
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"total_matches": len(matches),
		"success":       successCount,
		"errors":        errorCount,
	}).Info("重新计算积分完成")

	if errorCount > 0 {
		return fmt.Errorf("重新计算积分时发生 %d 个错误", errorCount)
	}

	return nil
}

// GetPointsHistory 获取用户积分历史
func (s *scoringService) GetPointsHistory(ctx context.Context, userID uint, tournament string) ([]scoring.PointsUpdateEvent, error) {
	return s.scoringRepo.GetPointsHistory(ctx, userID, tournament)
}

// ProcessPointsUpdate 处理积分更新（更新用户积分和排行榜）
func (s *scoringService) ProcessPointsUpdate(ctx context.Context, results []scoring.PointsCalculationResult, tournament string) error {
	s.logger.WithFields(logrus.Fields{
		"results_count": len(results),
		"tournament":    tournament,
	}).Info("开始处理积分更新")

	// 按用户分组积分更新
	userPointsMap := make(map[uint]int)
	var events []scoring.PointsUpdateEvent

	for _, result := range results {
		if result.Points > 0 {
			userPointsMap[result.UserID] += result.Points

			// 创建积分更新事件
			event := scoring.PointsUpdateEvent{
				UserID:       result.UserID,
				MatchID:      result.MatchID,
				PredictionID: result.PredictionID,
				OldPoints:    0, // 将在更新用户积分时填充
				NewPoints:    0, // 将在更新用户积分时填充
				PointsChange: result.Points,
				Tournament:   tournament,
				Timestamp:    time.Now(),
			}
			events = append(events, event)
		}
	}

	// 更新用户积分
	for userID, pointsToAdd := range userPointsMap {
		// 获取用户当前积分
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("获取用户信息失败")
			continue
		}

		oldPoints := user.Points

		// 更新用户积分
		err = s.userRepo.UpdatePoints(ctx, userID, pointsToAdd)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id": userID,
				"points":  pointsToAdd,
			}).Error("更新用户积分失败")
			continue
		}

		newPoints := oldPoints + pointsToAdd

		// 更新事件中的积分信息
		for i := range events {
			if events[i].UserID == userID {
				events[i].OldPoints = oldPoints
				events[i].NewPoints = newPoints
			}
		}

		s.logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"old_points": oldPoints,
			"new_points": newPoints,
			"points_add": pointsToAdd,
		}).Info("用户积分更新成功")
	}

	// 保存积分更新事件
	for _, event := range events {
		err := s.scoringRepo.SavePointsUpdateEvent(ctx, &event)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", event.UserID).Error("保存积分更新事件失败")
		}
	}

	s.logger.WithField("users_updated", len(userPointsMap)).Info("积分更新处理完成")

	return nil
}
