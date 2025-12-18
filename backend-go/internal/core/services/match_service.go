package services

import (
	"context"
	"time"

	"backend-go/internal/core/domain"
	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// MatchService 比赛服务实现
type MatchService struct {
	matchRepo    match.Repository
	cacheService *MatchCacheService
	eventBus     shared.EventBus
	logger       *logrus.Logger
}

// NewMatchService 创建比赛服务实例
func NewMatchService(matchRepo match.Repository, cacheService *MatchCacheService, eventBus shared.EventBus, logger *logrus.Logger) match.Service {
	if logger == nil {
		logger = logrus.New()
	}

	return &MatchService{
		matchRepo:    matchRepo,
		cacheService: cacheService,
		eventBus:     eventBus,
		logger:       logger,
	}
}

// CreateMatch 创建比赛
func (s *MatchService) CreateMatch(ctx context.Context, req *match.CreateMatchRequest) (*match.Match, error) {
	// 验证输入
	if req.TeamA == "" || req.TeamB == "" {
		return nil, domain.ErrInvalidInput
	}

	// 允许少量偏移，避免前端/服务器时钟秒级差异导致误判
	if req.StartTime.Before(time.Now().Add(-5 * time.Minute)) {
		return nil, domain.ErrInvalidStartTime
	}

	if !domain.IsValidTournament(string(req.Tournament)) {
		return nil, domain.ErrInvalidTournament
	}

	// 创建比赛实体
	m := &match.Match{
		TeamA:      req.TeamA,
		TeamB:      req.TeamB,
		Tournament: req.Tournament,
		StartTime:  req.StartTime,
		Status:     match.MatchStatusUpcoming,
		ScoreA:     0,
		ScoreB:     0,
	}

	// 保存到数据库
	err := s.matchRepo.Create(ctx, m)
	if err != nil {
		return nil, err
	}

	// 使比赛列表缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	return m, nil
}

// GetMatch 获取比赛详情
func (s *MatchService) GetMatch(ctx context.Context, id uint) (*match.Match, error) {
	var m *match.Match
	var err error
	
	if s.cacheService != nil {
		m, err = s.cacheService.GetMatch(ctx, id)
	} else {
		m, err = s.matchRepo.GetByID(ctx, id)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 填充计算字段
	m.FillComputedFields()
	return m, nil
}

// ListMatches 获取比赛列表
func (s *MatchService) ListMatches(ctx context.Context, filter match.ListFilter) ([]match.Match, error) {
	var matches []match.Match
	var err error
	
	if s.cacheService != nil {
		matches, err = s.cacheService.ListMatches(ctx, filter)
	} else {
		matches, err = s.matchRepo.List(ctx, filter)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 填充计算字段
	for i := range matches {
		matches[i].FillComputedFields()
	}
	
	return matches, nil
}

// UpdateMatch 更新比赛信息
func (s *MatchService) UpdateMatch(ctx context.Context, id uint, req *match.UpdateMatchRequest) (*match.Match, error) {
	// 获取现有比赛
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查比赛状态
	if m.Status != match.MatchStatusUpcoming {
		return nil, domain.ErrMatchAlreadyStarted
	}

	// 更新字段
	if req.TeamA != "" {
		m.TeamA = req.TeamA
	}

	if req.TeamB != "" {
		m.TeamB = req.TeamB
	}

	if req.Tournament != "" {
		if !domain.IsValidTournament(string(req.Tournament)) {
			return nil, domain.ErrInvalidTournament
		}
		m.Tournament = req.Tournament
	}

	if req.StartTime != nil {
		if req.StartTime.Before(time.Now()) {
			return nil, domain.ErrInvalidStartTime
		}
		m.StartTime = *req.StartTime
	}

	// 保存更新
	err = s.matchRepo.Update(ctx, m)
	if err != nil {
		return nil, err
	}

	// 使缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatch(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match cache")
		}
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	return m, nil
}

// StartMatch 开始比赛
func (s *MatchService) StartMatch(ctx context.Context, id uint) error {
	// 获取比赛
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查状态
	if m.Status != match.MatchStatusUpcoming {
		return domain.ErrInvalidMatchStatus
	}

	oldStatus := m.Status

	// 更新状态
	err = s.matchRepo.UpdateStatus(ctx, id, match.MatchStatusLive)
	if err != nil {
		return err
	}

	// 使缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatch(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match cache")
		}
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	// 发布比赛开始事件
	if s.eventBus != nil {
		startedPayload := shared.MatchStartedPayload{
			MatchID:   id,
			StartTime: time.Now(),
		}

		statusChangedPayload := shared.MatchStatusChangedPayload{
			MatchID:   id,
			OldStatus: string(oldStatus),
			NewStatus: string(match.MatchStatusLive),
		}

		startedEvent := shared.NewEvent(shared.EventMatchStarted, startedPayload)
		statusEvent := shared.NewEvent(shared.EventMatchStatusChanged, statusChangedPayload)

		if err := s.eventBus.Publish(startedEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match started event")
		}

		if err := s.eventBus.Publish(statusEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match status changed event")
		}

		s.logger.WithFields(logrus.Fields{
			"match_id":   id,
			"old_status": oldStatus,
			"new_status": match.MatchStatusLive,
		}).Info("Match started events published")
	}

	return nil
}

// SetResult 设置比赛结果
func (s *MatchService) SetResult(ctx context.Context, id uint, req *match.SetResultRequest) error {
	// 获取比赛
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 允许已结束比赛重新设置结果（覆盖），不再阻断

	// 验证获胜者
	if req.Winner != "" && req.Winner != "A" && req.Winner != "B" {
		return domain.ErrInvalidWinner
	}

	oldStatus := m.Status
	oldScoreA := m.ScoreA
	oldScoreB := m.ScoreB

	// 设置结果
	err = s.matchRepo.SetResult(ctx, id, req.ScoreA, req.ScoreB, req.Winner)
	if err != nil {
		return err
	}

	// 使缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatch(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match cache")
		}
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	// 发布比赛结束事件
	if s.eventBus != nil {
		finishedPayload := shared.MatchFinishedPayload{
			MatchID: id,
			Winner:  req.Winner,
			ScoreA:  req.ScoreA,
			ScoreB:  req.ScoreB,
		}

		statusChangedPayload := shared.MatchStatusChangedPayload{
			MatchID:   id,
			OldStatus: string(oldStatus),
			NewStatus: string(match.MatchStatusFinished),
		}

		scoreUpdatedPayload := shared.MatchScoreUpdatedPayload{
			MatchID:   id,
			OldScoreA: oldScoreA,
			OldScoreB: oldScoreB,
			NewScoreA: req.ScoreA,
			NewScoreB: req.ScoreB,
		}

		finishedEvent := shared.NewEvent(shared.EventMatchFinished, finishedPayload)
		statusEvent := shared.NewEvent(shared.EventMatchStatusChanged, statusChangedPayload)
		scoreEvent := shared.NewEvent(shared.EventMatchScoreUpdated, scoreUpdatedPayload)

		if err := s.eventBus.Publish(finishedEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match finished event")
		}

		if err := s.eventBus.Publish(statusEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match status changed event")
		}

		if err := s.eventBus.Publish(scoreEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match score updated event")
		}

		s.logger.WithFields(logrus.Fields{
			"match_id": id,
			"winner":   req.Winner,
			"score_a":  req.ScoreA,
			"score_b":  req.ScoreB,
		}).Info("Match finished events published")
	}

	return nil
}

// CancelMatch 取消比赛
func (s *MatchService) CancelMatch(ctx context.Context, id uint) error {
	// 获取比赛
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查状态
	if m.Status == match.MatchStatusFinished {
		return domain.ErrMatchAlreadyFinished
	}

	oldStatus := m.Status

	// 更新状态
	err = s.matchRepo.UpdateStatus(ctx, id, match.MatchStatusCancelled)
	if err != nil {
		return err
	}

	// 使缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatch(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match cache")
		}
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	// 发布比赛取消事件
	if s.eventBus != nil {
		cancelledPayload := shared.MatchCancelledPayload{
			MatchID: id,
			Reason:  "Match cancelled by administrator",
		}

		statusChangedPayload := shared.MatchStatusChangedPayload{
			MatchID:   id,
			OldStatus: string(oldStatus),
			NewStatus: string(match.MatchStatusCancelled),
		}

		cancelledEvent := shared.NewEvent(shared.EventMatchCancelled, cancelledPayload)
		statusEvent := shared.NewEvent(shared.EventMatchStatusChanged, statusChangedPayload)

		if err := s.eventBus.Publish(cancelledEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match cancelled event")
		}

		if err := s.eventBus.Publish(statusEvent); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match status changed event")
		}

		s.logger.WithFields(logrus.Fields{
			"match_id":   id,
			"old_status": oldStatus,
			"new_status": match.MatchStatusCancelled,
		}).Info("Match cancelled events published")
	}

	return nil
}

// GetUpcomingMatches 获取即将开始的比赛
func (s *MatchService) GetUpcomingMatches(ctx context.Context) ([]match.Match, error) {
	var matches []match.Match
	var err error
	
	if s.cacheService != nil {
		matches, err = s.cacheService.GetUpcomingMatches(ctx, 10)
	} else {
		matches, err = s.matchRepo.GetUpcoming(ctx, 10) // 默认返回10个
	}
	
	if err != nil {
		return nil, err
	}
	
	// 填充计算字段
	for i := range matches {
		matches[i].FillComputedFields()
	}
	
	return matches, nil
}

// GetLiveMatches 获取正在进行的比赛
func (s *MatchService) GetLiveMatches(ctx context.Context) ([]match.Match, error) {
	var matches []match.Match
	var err error
	
	if s.cacheService != nil {
		matches, err = s.cacheService.GetLiveMatches(ctx)
	} else {
		matches, err = s.matchRepo.GetLive(ctx)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 填充计算字段
	for i := range matches {
		matches[i].FillComputedFields()
	}
	
	return matches, nil
}

// GetFinishedMatches 获取已结束的比赛
func (s *MatchService) GetFinishedMatches(ctx context.Context, limit int) ([]match.Match, error) {
	if limit <= 0 {
		limit = 20 // 默认返回20个
	}
	
	var matches []match.Match
	var err error
	
	if s.cacheService != nil {
		matches, err = s.cacheService.GetFinishedMatches(ctx, limit)
	} else {
		matches, err = s.matchRepo.GetFinished(ctx, limit)
	}
	
	if err != nil {
		return nil, err
	}
	
	// 填充计算字段
	for i := range matches {
		matches[i].FillComputedFields()
	}
	
	return matches, nil
}

// UpdateScore 更新比赛比分（用于直播比赛）
func (s *MatchService) UpdateScore(ctx context.Context, id uint, scoreA, scoreB int) error {
	// 获取比赛
	m, err := s.matchRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查状态 - 只有进行中的比赛可以更新比分
	if m.Status != match.MatchStatusLive {
		return domain.ErrInvalidMatchStatus
	}

	oldScoreA := m.ScoreA
	oldScoreB := m.ScoreB

	// 更新比分
	err = s.matchRepo.UpdateScore(ctx, id, scoreA, scoreB)
	if err != nil {
		return err
	}

	// 使缓存失效
	if s.cacheService != nil {
		if err := s.cacheService.InvalidateMatch(ctx, id); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match cache")
		}
		if err := s.cacheService.InvalidateMatchLists(ctx); err != nil {
			s.logger.WithError(err).Warn("Failed to invalidate match lists cache")
		}
	}

	// 发布比分更新事件
	if s.eventBus != nil {
		payload := shared.MatchScoreUpdatedPayload{
			MatchID:   id,
			OldScoreA: oldScoreA,
			OldScoreB: oldScoreB,
			NewScoreA: scoreA,
			NewScoreB: scoreB,
		}

		event := shared.NewEvent(shared.EventMatchScoreUpdated, payload)
		if err := s.eventBus.Publish(event); err != nil {
			s.logger.WithError(err).Warn("Failed to publish match score updated event")
		} else {
			s.logger.WithFields(logrus.Fields{
				"match_id":    id,
				"old_score_a": oldScoreA,
				"old_score_b": oldScoreB,
				"new_score_a": scoreA,
				"new_score_b": scoreB,
			}).Info("Match score updated event published")
		}
	}

	return nil
}
