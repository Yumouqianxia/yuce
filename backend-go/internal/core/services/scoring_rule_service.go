package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/ports"
	"backend-go/internal/core/types"
	"github.com/sirupsen/logrus"
)

// ScoringRuleService 积分规则服务实现
type ScoringRuleService struct {
	scoringRuleRepo ports.ScoringRuleRepository
	scoreCalculator ScoreCalculator
	logger          *logrus.Logger
}

// NewScoringRuleService 创建积分规则服务实例
func NewScoringRuleService(
	scoringRuleRepo ports.ScoringRuleRepository,
	scoreCalculator ScoreCalculator,
	logger *logrus.Logger,
) *ScoringRuleService {
	return &ScoringRuleService{
		scoringRuleRepo: scoringRuleRepo,
		scoreCalculator: scoreCalculator,
		logger:          logger,
	}
}

// CreateScoringRule 创建积分规则
func (s *ScoringRuleService) CreateScoringRule(ctx context.Context, req *ports.CreateScoringRuleRequest) (*sport.ScoringRule, error) {
	s.logger.WithFields(logrus.Fields{
		"sport_type_id": req.SportTypeID,
		"name":          req.Name,
	}).Info("Creating scoring rule")

	// 创建积分规则实体
	rule := &sport.ScoringRule{
		SportTypeID:  req.SportTypeID,
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		IsActive:     req.IsActive,

		// 基础积分设置
		BasePoints:           req.BasePoints,
		EnableDifficulty:     req.EnableDifficulty,
		DifficultyMultiplier: req.DifficultyMultiplier,

		// 奖励组件
		EnableVoteReward: req.EnableVoteReward,
		VoteRewardPoints: req.VoteRewardPoints,
		MaxVoteReward:    req.MaxVoteReward,

		EnableTimeReward: req.EnableTimeReward,
		TimeRewardPoints: req.TimeRewardPoints,
		TimeRewardHours:  req.TimeRewardHours,

		// 惩罚组件
		EnableModifyPenalty: req.EnableModifyPenalty,
		ModifyPenaltyPoints: req.ModifyPenaltyPoints,
		MaxModifyPenalty:    req.MaxModifyPenalty,
	}

	// 验证业务规则
	if err := s.validateScoringRule(rule); err != nil {
		return nil, fmt.Errorf("invalid scoring rule: %w", err)
	}

	// 如果设置为激活状态，需要先将同运动类型的其他规则设为非激活
	if req.IsActive {
		// 这里先创建规则，然后再设置激活状态
		rule.IsActive = false
	}

	// 保存到数据库
	if err := s.scoringRuleRepo.Create(ctx, rule); err != nil {
		s.logger.WithError(err).Error("Failed to create scoring rule")
		return nil, fmt.Errorf("failed to create scoring rule: %w", err)
	}

	// 如果需要激活，设置为激活状态
	if req.IsActive {
		if err := s.scoringRuleRepo.SetActive(ctx, rule.ID); err != nil {
			s.logger.WithError(err).Error("Failed to set scoring rule as active")
			// 这里不返回错误，因为规则已经创建成功
		} else {
			rule.IsActive = true
		}
	}

	s.logger.WithField("scoring_rule_id", rule.ID).Info("Scoring rule created successfully")
	return rule, nil
}

// GetScoringRule 获取积分规则
func (s *ScoringRuleService) GetScoringRule(ctx context.Context, id uint) (*sport.ScoringRule, error) {
	rule, err := s.scoringRuleRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get scoring rule")
		return nil, err
	}
	
	return rule, nil
}

// UpdateScoringRule 更新积分规则
func (s *ScoringRuleService) UpdateScoringRule(ctx context.Context, id uint, req *ports.UpdateScoringRuleRequest) (*sport.ScoringRule, error) {
	s.logger.WithField("id", id).Info("Updating scoring rule")

	// 获取现有规则
	rule, err := s.scoringRuleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		rule.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		rule.Description = strings.TrimSpace(*req.Description)
	}
	if req.BasePoints != nil {
		rule.BasePoints = *req.BasePoints
	}
	if req.EnableDifficulty != nil {
		rule.EnableDifficulty = *req.EnableDifficulty
	}
	if req.DifficultyMultiplier != nil {
		rule.DifficultyMultiplier = *req.DifficultyMultiplier
	}
	if req.EnableVoteReward != nil {
		rule.EnableVoteReward = *req.EnableVoteReward
	}
	if req.VoteRewardPoints != nil {
		rule.VoteRewardPoints = *req.VoteRewardPoints
	}
	if req.MaxVoteReward != nil {
		rule.MaxVoteReward = *req.MaxVoteReward
	}
	if req.EnableTimeReward != nil {
		rule.EnableTimeReward = *req.EnableTimeReward
	}
	if req.TimeRewardPoints != nil {
		rule.TimeRewardPoints = *req.TimeRewardPoints
	}
	if req.TimeRewardHours != nil {
		rule.TimeRewardHours = *req.TimeRewardHours
	}
	if req.EnableModifyPenalty != nil {
		rule.EnableModifyPenalty = *req.EnableModifyPenalty
	}
	if req.ModifyPenaltyPoints != nil {
		rule.ModifyPenaltyPoints = *req.ModifyPenaltyPoints
	}
	if req.MaxModifyPenalty != nil {
		rule.MaxModifyPenalty = *req.MaxModifyPenalty
	}

	// 验证业务规则
	if err := s.validateScoringRule(rule); err != nil {
		return nil, fmt.Errorf("invalid scoring rule: %w", err)
	}

	// 处理激活状态变更
	if req.IsActive != nil && *req.IsActive != rule.IsActive {
		if *req.IsActive {
			// 设置为激活状态
			if err := s.scoringRuleRepo.SetActive(ctx, rule.ID); err != nil {
				return nil, fmt.Errorf("failed to set scoring rule as active: %w", err)
			}
			rule.IsActive = true
		} else {
			// 设置为非激活状态
			rule.IsActive = false
		}
	}

	// 保存更新
	if err := s.scoringRuleRepo.Update(ctx, rule); err != nil {
		s.logger.WithError(err).Error("Failed to update scoring rule")
		return nil, fmt.Errorf("failed to update scoring rule: %w", err)
	}

	s.logger.WithField("scoring_rule_id", rule.ID).Info("Scoring rule updated successfully")
	return rule, nil
}

// DeleteScoringRule 删除积分规则
func (s *ScoringRuleService) DeleteScoringRule(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("Deleting scoring rule")

	if err := s.scoringRuleRepo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete scoring rule")
		return fmt.Errorf("failed to delete scoring rule: %w", err)
	}

	s.logger.WithField("scoring_rule_id", id).Info("Scoring rule deleted successfully")
	return nil
}

// ListScoringRules 获取积分规则列表
func (s *ScoringRuleService) ListScoringRules(ctx context.Context, req *ports.ListScoringRulesRequest) (*ports.ListScoringRulesResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 构建查询选项
	options := &ports.ListScoringRulesOptions{
		SportTypeID: req.SportTypeID,
		IsActive:    req.IsActive,
		OrderBy:     req.OrderBy,
		Limit:       req.PageSize,
		Offset:      (req.Page - 1) * req.PageSize,
	}

	// 获取总数
	total, err := s.scoringRuleRepo.Count(ctx, options)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count scoring rules")
		return nil, fmt.Errorf("failed to count scoring rules: %w", err)
	}

	// 获取列表
	rules, err := s.scoringRuleRepo.List(ctx, options)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list scoring rules")
		return nil, fmt.Errorf("failed to list scoring rules: %w", err)
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &ports.ListScoringRulesResponse{
		ScoringRules: rules,
		Total:        total,
		Page:         req.Page,
		PageSize:     req.PageSize,
		TotalPages:   totalPages,
	}, nil
}

// GetActiveScoringRule 获取激活的积分规则
func (s *ScoringRuleService) GetActiveScoringRule(ctx context.Context, sportTypeID uint) (*sport.ScoringRule, error) {
	rule, err := s.scoringRuleRepo.GetActiveBySportTypeID(ctx, sportTypeID)
	if err != nil {
		s.logger.WithError(err).WithField("sport_type_id", sportTypeID).Error("Failed to get active scoring rule")
		return nil, err
	}
	
	return rule, nil
}

// SetActiveScoringRule 设置激活的积分规则
func (s *ScoringRuleService) SetActiveScoringRule(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("Setting scoring rule as active")

	if err := s.scoringRuleRepo.SetActive(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to set scoring rule as active")
		return fmt.Errorf("failed to set scoring rule as active: %w", err)
	}

	s.logger.WithField("scoring_rule_id", id).Info("Scoring rule set as active successfully")
	return nil
}

// GetScoringRulesBySportType 根据运动类型获取积分规则
func (s *ScoringRuleService) GetScoringRulesBySportType(ctx context.Context, sportTypeID uint) ([]*sport.ScoringRule, error) {
	rules, err := s.scoringRuleRepo.GetBySportTypeID(ctx, sportTypeID)
	if err != nil {
		s.logger.WithError(err).WithField("sport_type_id", sportTypeID).Error("Failed to get scoring rules by sport type")
		return nil, err
	}
	
	return rules, nil
}

// CalculateScore 计算积分
func (s *ScoringRuleService) CalculateScore(ctx context.Context, predictionID uint) (*types.ScoreBreakdown, error) {
	// TODO: 实现从数据库获取预测和比赛信息，然后计算积分
	// 这里需要集成到现有的预测系统
	return nil, fmt.Errorf("not implemented yet")
}

// PreviewScore 预览积分计算
func (s *ScoringRuleService) PreviewScore(ctx context.Context, req *types.PreviewScoreRequest) (*types.ScoreBreakdown, error) {
	return s.scoreCalculator.PreviewScore(ctx, req)
}

// RecalculateScores 批量重算积分
func (s *ScoringRuleService) RecalculateScores(ctx context.Context, sportTypeID uint, ruleID uint) (*ports.RecalculateResult, error) {
	s.logger.WithFields(logrus.Fields{
		"sport_type_id": sportTypeID,
		"rule_id":       ruleID,
	}).Info("Starting score recalculation")

	// TODO: 实现批量重算逻辑
	// 1. 获取指定运动类型的所有已结束比赛的预测
	// 2. 使用新的积分规则重新计算积分
	// 3. 更新预测的积分和用户总积分
	// 4. 返回重算结果统计

	result := &ports.RecalculateResult{
		TotalPredictions:   0,
		UpdatedPredictions: 0,
		FailedPredictions:  0,
		TotalPointsChanged: 0,
	}

	s.logger.WithFields(logrus.Fields{
		"total_predictions":   result.TotalPredictions,
		"updated_predictions": result.UpdatedPredictions,
		"failed_predictions":  result.FailedPredictions,
	}).Info("Score recalculation completed")

	return result, nil
}

// validateScoringRule 验证积分规则
func (s *ScoringRuleService) validateScoringRule(rule *sport.ScoringRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.BasePoints <= 0 {
		return fmt.Errorf("base points must be greater than 0")
	}

	if rule.EnableDifficulty && rule.DifficultyMultiplier <= 0 {
		return fmt.Errorf("difficulty multiplier must be greater than 0 when difficulty is enabled")
	}

	if rule.EnableVoteReward {
		if rule.VoteRewardPoints < 0 {
			return fmt.Errorf("vote reward points cannot be negative")
		}
		if rule.MaxVoteReward < 0 {
			return fmt.Errorf("max vote reward cannot be negative")
		}
		if rule.MaxVoteReward > 0 && rule.VoteRewardPoints > rule.MaxVoteReward {
			return fmt.Errorf("vote reward points cannot exceed max vote reward")
		}
	}

	if rule.EnableTimeReward {
		if rule.TimeRewardPoints < 0 {
			return fmt.Errorf("time reward points cannot be negative")
		}
		if rule.TimeRewardHours <= 0 {
			return fmt.Errorf("time reward hours must be greater than 0")
		}
	}

	if rule.EnableModifyPenalty {
		if rule.ModifyPenaltyPoints < 0 {
			return fmt.Errorf("modify penalty points cannot be negative")
		}
		if rule.MaxModifyPenalty < 0 {
			return fmt.Errorf("max modify penalty cannot be negative")
		}
		if rule.MaxModifyPenalty > 0 && rule.ModifyPenaltyPoints > rule.MaxModifyPenalty {
			return fmt.Errorf("modify penalty points cannot exceed max modify penalty")
		}
	}

	return nil
}