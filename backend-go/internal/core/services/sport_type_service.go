package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"backend-go/internal/core/domain/sport"
	"backend-go/internal/core/ports"
	"github.com/sirupsen/logrus"
)

// SportTypeService 运动类型服务实现
type SportTypeService struct {
	sportTypeRepo ports.SportTypeRepository
	logger        *logrus.Logger
}

// NewSportTypeService 创建运动类型服务实例
func NewSportTypeService(sportTypeRepo ports.SportTypeRepository, logger *logrus.Logger) *SportTypeService {
	return &SportTypeService{
		sportTypeRepo: sportTypeRepo,
		logger:        logger,
	}
}

// CreateSportType 创建运动类型
func (s *SportTypeService) CreateSportType(ctx context.Context, req *ports.CreateSportTypeRequest) (*sport.SportType, error) {
	s.logger.WithFields(logrus.Fields{
		"name":     req.Name,
		"code":     req.Code,
		"category": req.Category,
	}).Info("Creating sport type")

	// 验证运动类别
	if !sport.IsValidCategory(string(req.Category)) {
		return nil, fmt.Errorf("invalid sport category: %s", req.Category)
	}

	// 标准化代码（转小写）
	code := strings.ToLower(strings.TrimSpace(req.Code))
	
	// 创建运动类型实体
	sportType := &sport.SportType{
		Name:        strings.TrimSpace(req.Name),
		Code:        code,
		Category:    req.Category,
		Icon:        strings.TrimSpace(req.Icon),
		Banner:      strings.TrimSpace(req.Banner),
		Description: strings.TrimSpace(req.Description),
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	// 保存到数据库
	if err := s.sportTypeRepo.Create(ctx, sportType); err != nil {
		s.logger.WithError(err).Error("Failed to create sport type")
		return nil, fmt.Errorf("failed to create sport type: %w", err)
	}

	// 创建默认配置
	defaultConfig := s.createDefaultConfiguration(sportType.ID, req.Category)
	if err := s.sportTypeRepo.CreateConfiguration(ctx, defaultConfig); err != nil {
		s.logger.WithError(err).Error("Failed to create default configuration")
		// 这里不返回错误，因为运动类型已经创建成功
	}

	s.logger.WithField("sport_type_id", sportType.ID).Info("Sport type created successfully")
	return sportType, nil
}

// GetSportType 获取运动类型
func (s *SportTypeService) GetSportType(ctx context.Context, id uint) (*sport.SportType, error) {
	sportType, err := s.sportTypeRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get sport type")
		return nil, err
	}
	
	return sportType, nil
}

// GetSportTypeByCode 根据代码获取运动类型
func (s *SportTypeService) GetSportTypeByCode(ctx context.Context, code string) (*sport.SportType, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	sportType, err := s.sportTypeRepo.GetByCode(ctx, code)
	if err != nil {
		s.logger.WithError(err).WithField("code", code).Error("Failed to get sport type by code")
		return nil, err
	}
	
	return sportType, nil
}

// UpdateSportType 更新运动类型
func (s *SportTypeService) UpdateSportType(ctx context.Context, id uint, req *ports.UpdateSportTypeRequest) (*sport.SportType, error) {
	s.logger.WithField("id", id).Info("Updating sport type")

	// 获取现有运动类型
	sportType, err := s.sportTypeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		sportType.Name = strings.TrimSpace(*req.Name)
	}
	if req.Code != nil {
		sportType.Code = strings.ToLower(strings.TrimSpace(*req.Code))
	}
	if req.Category != nil {
		if !sport.IsValidCategory(string(*req.Category)) {
			return nil, fmt.Errorf("invalid sport category: %s", *req.Category)
		}
		sportType.Category = *req.Category
	}
	if req.Icon != nil {
		sportType.Icon = strings.TrimSpace(*req.Icon)
	}
	if req.Banner != nil {
		sportType.Banner = strings.TrimSpace(*req.Banner)
	}
	if req.Description != nil {
		sportType.Description = strings.TrimSpace(*req.Description)
	}
	if req.IsActive != nil {
		sportType.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		sportType.SortOrder = *req.SortOrder
	}

	// 保存更新
	if err := s.sportTypeRepo.Update(ctx, sportType); err != nil {
		s.logger.WithError(err).Error("Failed to update sport type")
		return nil, fmt.Errorf("failed to update sport type: %w", err)
	}

	s.logger.WithField("sport_type_id", sportType.ID).Info("Sport type updated successfully")
	return sportType, nil
}

// DeleteSportType 删除运动类型
func (s *SportTypeService) DeleteSportType(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("Deleting sport type")

	if err := s.sportTypeRepo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete sport type")
		return fmt.Errorf("failed to delete sport type: %w", err)
	}

	s.logger.WithField("sport_type_id", id).Info("Sport type deleted successfully")
	return nil
}

// ListSportTypes 获取运动类型列表
func (s *SportTypeService) ListSportTypes(ctx context.Context, req *ports.ListSportTypesRequest) (*ports.ListSportTypesResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 构建查询选项
	options := &ports.ListSportTypesOptions{
		Category: req.Category,
		IsActive: req.IsActive,
		OrderBy:  req.OrderBy,
		Limit:    req.PageSize,
		Offset:   (req.Page - 1) * req.PageSize,
	}

	// 获取总数
	total, err := s.sportTypeRepo.Count(ctx, options)
	if err != nil {
		s.logger.WithError(err).Error("Failed to count sport types")
		return nil, fmt.Errorf("failed to count sport types: %w", err)
	}

	// 获取列表
	sportTypes, err := s.sportTypeRepo.List(ctx, options)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list sport types")
		return nil, fmt.Errorf("failed to list sport types: %w", err)
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &ports.ListSportTypesResponse{
		SportTypes: sportTypes,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetSportConfiguration 获取运动配置
func (s *SportTypeService) GetSportConfiguration(ctx context.Context, sportTypeID uint) (*sport.SportConfiguration, error) {
	config, err := s.sportTypeRepo.GetConfiguration(ctx, sportTypeID)
	if err != nil {
		s.logger.WithError(err).WithField("sport_type_id", sportTypeID).Error("Failed to get sport configuration")
		return nil, err
	}
	
	return config, nil
}

// UpdateSportConfiguration 更新运动配置
func (s *SportTypeService) UpdateSportConfiguration(ctx context.Context, sportTypeID uint, req *ports.UpdateSportConfigurationRequest) (*sport.SportConfiguration, error) {
	s.logger.WithField("sport_type_id", sportTypeID).Info("Updating sport configuration")

	// 获取现有配置
	config, err := s.sportTypeRepo.GetConfiguration(ctx, sportTypeID)
	if err != nil {
		return nil, err
	}

	// 更新配置字段
	if req.EnableRealtime != nil {
		config.EnableRealtime = *req.EnableRealtime
	}
	if req.EnableChat != nil {
		config.EnableChat = *req.EnableChat
	}
	if req.EnableVoting != nil {
		config.EnableVoting = *req.EnableVoting
	}
	if req.EnablePrediction != nil {
		config.EnablePrediction = *req.EnablePrediction
	}
	if req.EnableLeaderboard != nil {
		config.EnableLeaderboard = *req.EnableLeaderboard
	}
	if req.AllowModification != nil {
		config.AllowModification = *req.AllowModification
	}
	if req.MaxModifications != nil {
		config.MaxModifications = *req.MaxModifications
	}
	if req.ModificationDeadline != nil {
		config.ModificationDeadline = *req.ModificationDeadline
	}
	if req.EnableSelfVoting != nil {
		config.EnableSelfVoting = *req.EnableSelfVoting
	}
	if req.MaxVotesPerUser != nil {
		config.MaxVotesPerUser = *req.MaxVotesPerUser
	}
	if req.VotingDeadline != nil {
		config.VotingDeadline = *req.VotingDeadline
	}

	// 保存更新
	if err := s.sportTypeRepo.UpdateConfiguration(ctx, config); err != nil {
		s.logger.WithError(err).Error("Failed to update sport configuration")
		return nil, fmt.Errorf("failed to update sport configuration: %w", err)
	}

	s.logger.WithField("sport_type_id", sportTypeID).Info("Sport configuration updated successfully")
	return config, nil
}

// BatchUpdateConfiguration 批量更新配置
func (s *SportTypeService) BatchUpdateConfiguration(ctx context.Context, req *ports.BatchUpdateConfigRequest) error {
	s.logger.WithFields(logrus.Fields{
		"sport_type_ids": req.SportTypeIDs,
		"config_count":   len(req.SportTypeIDs),
	}).Info("Batch updating sport configurations")

	for _, sportTypeID := range req.SportTypeIDs {
		_, err := s.UpdateSportConfiguration(ctx, sportTypeID, req.Config)
		if err != nil {
			s.logger.WithError(err).WithField("sport_type_id", sportTypeID).Error("Failed to update configuration in batch")
			return fmt.Errorf("failed to update configuration for sport type %d: %w", sportTypeID, err)
		}
	}

	s.logger.Info("Batch configuration update completed successfully")
	return nil
}

// GetSportTypeStats 获取运动类型统计信息
func (s *SportTypeService) GetSportTypeStats(ctx context.Context, sportTypeID uint) (*ports.SportTypeStats, error) {
	// 这里需要实现统计逻辑，暂时返回基础信息
	// TODO: 实现完整的统计功能
	return &ports.SportTypeStats{
		SportTypeID:     sportTypeID,
		MatchCount:      0,
		PredictionCount: 0,
		UserCount:       0,
		LastMatchTime:   nil,
	}, nil
}

// createDefaultConfiguration 创建默认配置
func (s *SportTypeService) createDefaultConfiguration(sportTypeID uint, category sport.SportCategory) *sport.SportConfiguration {
	config := &sport.SportConfiguration{
		SportTypeID:          sportTypeID,
		EnableRealtime:       true,
		EnableChat:           false,
		EnableVoting:         true,
		EnablePrediction:     true,
		EnableLeaderboard:    true,
		AllowModification:    true,
		MaxModifications:     3,
		ModificationDeadline: 30,
		EnableSelfVoting:     false,
		MaxVotesPerUser:      10,
		VotingDeadline:       0,
	}

	// 根据运动类别调整默认配置
	if category == sport.SportCategoryEsports {
		config.EnableChat = true // 电竞默认开启聊天
		config.MaxVotesPerUser = 15
	} else {
		config.AllowModification = false // 传统体育默认不允许修改
		config.MaxModifications = 0
		config.MaxVotesPerUser = 5
	}

	return config
}