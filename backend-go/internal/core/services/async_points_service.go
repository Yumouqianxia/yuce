package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/scoring"
	"backend-go/internal/core/domain/shared"
	"backend-go/internal/core/domain/user"
	"backend-go/pkg/cache"
	"github.com/sirupsen/logrus"
)

// PointsCalculationTask 积分计算任务
type PointsCalculationTask struct {
	ID        string    `json:"id"`
	MatchID   uint      `json:"match_id"`
	RuleID    *uint     `json:"rule_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	Error     string    `json:"error,omitempty"`
}

// TaskStatus 任务状态
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// AsyncPointsService 异步积分计算服务
type AsyncPointsService struct {
	predictionRepo  prediction.Repository
	scoringRuleRepo prediction.ScoringRuleRepository
	matchRepo       match.Repository
	userRepo        user.Repository
	cacheService    cache.CacheService
	eventBus        shared.EventBus
	logger          *logrus.Logger

	// 任务队列和处理
	taskQueue   chan *PointsCalculationTask
	activeTasks map[string]*PointsCalculationTask
	taskMutex   sync.RWMutex

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 配置
	maxWorkers int
	queueSize  int
}

// NewAsyncPointsService 创建异步积分计算服务
func NewAsyncPointsService(
	predictionRepo prediction.Repository,
	scoringRuleRepo prediction.ScoringRuleRepository,
	matchRepo match.Repository,
	userRepo user.Repository,
	cacheService cache.CacheService,
	eventBus shared.EventBus,
	logger *logrus.Logger,
) *AsyncPointsService {
	if logger == nil {
		logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &AsyncPointsService{
		predictionRepo:  predictionRepo,
		scoringRuleRepo: scoringRuleRepo,
		matchRepo:       matchRepo,
		userRepo:        userRepo,
		cacheService:    cacheService,
		eventBus:        eventBus,
		logger:          logger,
		taskQueue:       make(chan *PointsCalculationTask, 100), // 队列大小100
		activeTasks:     make(map[string]*PointsCalculationTask),
		ctx:             ctx,
		cancel:          cancel,
		maxWorkers:      5, // 最大5个工作协程
		queueSize:       100,
	}

	// 启动工作协程
	service.startWorkers()

	return service
}

// startWorkers 启动工作协程
func (s *AsyncPointsService) startWorkers() {
	for i := 0; i < s.maxWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	s.logger.WithField("workers", s.maxWorkers).Info("Async points calculation workers started")
}

// worker 工作协程
func (s *AsyncPointsService) worker(workerID int) {
	defer s.wg.Done()

	logger := s.logger.WithField("worker_id", workerID)
	logger.Debug("Points calculation worker started")

	for {
		select {
		case <-s.ctx.Done():
			logger.Debug("Points calculation worker shutting down")
			return
		case task := <-s.taskQueue:
			s.processTask(task, logger)
		}
	}
}

// processTask 处理积分计算任务
func (s *AsyncPointsService) processTask(task *PointsCalculationTask, logger *logrus.Entry) {
	// 更新任务状态
	s.updateTaskStatus(task.ID, TaskStatusProcessing, "")

	logger = logger.WithFields(logrus.Fields{
		"task_id":  task.ID,
		"match_id": task.MatchID,
	})

	logger.Info("Processing points calculation task")
	start := time.Now()

	// 执行积分计算
	result, err := s.calculatePointsForMatch(context.Background(), task.MatchID, task.RuleID)
	if err != nil {
		logger.WithError(err).Error("Points calculation failed")
		s.updateTaskStatus(task.ID, TaskStatusFailed, err.Error())
		return
	}

	// 更新缓存
	if err := s.updateCacheAfterCalculation(context.Background(), task.MatchID, result); err != nil {
		logger.WithError(err).Warn("Failed to update cache after points calculation")
	}

	// 发布积分计算完成事件
	if err := s.publishPointsCalculatedEvent(result); err != nil {
		logger.WithError(err).Warn("Failed to publish points calculated event")
	}

	// 更新任务状态
	s.updateTaskStatus(task.ID, TaskStatusCompleted, "")

	duration := time.Since(start)
	logger.WithFields(logrus.Fields{
		"duration":     duration,
		"predictions":  len(result.Results),
		"total_points": result.TotalPoints,
	}).Info("Points calculation completed")
}

// calculatePointsForMatch 计算比赛积分
func (s *AsyncPointsService) calculatePointsForMatch(ctx context.Context, matchID uint, ruleID *uint) (*scoring.MatchPointsCalculation, error) {
	// 获取比赛信息
	matchEntity, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if !matchEntity.IsFinished() {
		return nil, fmt.Errorf("match %d is not finished", matchID)
	}

	// 获取积分规则
	var rule *prediction.ScoringRule
	if ruleID != nil {
		rule, err = s.scoringRuleRepo.GetScoringRuleByID(ctx, *ruleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get scoring rule: %w", err)
		}
	} else {
		rule, err = s.scoringRuleRepo.GetActiveScoringRule(ctx)
		if err != nil {
			s.logger.WithError(err).Debug("No active scoring rule found, using default calculation")
		}
	}

	// 获取比赛的所有预测
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get predictions: %w", err)
	}

	// 计算结果
	result := &scoring.MatchPointsCalculation{
		MatchID:     matchID,
		Results:     make([]scoring.PointsCalculationResult, 0, len(predictions)),
		TotalPoints: 0,
		ProcessedAt: time.Now(),
	}

	// 批量更新用户积分的映射
	userPointsUpdates := make(map[uint]int)

	// 计算每个预测的积分
	for _, predWithVotes := range predictions {
		pred := predWithVotes.Prediction
		pred.Match = matchEntity

		// 计算积分
		var points int
		var accuracy scoring.PredictionAccuracy
		var reason string

		if rule != nil {
			points = pred.CalculatePointsWithRule(rule)
			accuracy = scoring.GetPredictionAccuracy(pred, matchEntity)
		} else {
			points = pred.CalculatePoints()
			accuracy = scoring.GetPredictionAccuracy(pred, matchEntity)
		}

		// 计算热门奖励
		popularityBonus := scoring.CalculatePopularityBonus(pred.VoteCount)
		points += popularityBonus.Bonus

		// 构建原因说明
		reason = scoring.BuildPointsReason(accuracy, points-popularityBonus.Bonus, popularityBonus)

		// 更新预测积分
		if err := s.predictionRepo.UpdatePredictionPoints(ctx, pred.ID, points, pred.IsCorrect); err != nil {
			return nil, fmt.Errorf("failed to update prediction points: %w", err)
		}

		// 累计用户积分更新
		userPointsUpdates[pred.UserID] += points

		// 添加到结果
		result.Results = append(result.Results, scoring.PointsCalculationResult{
			PredictionID: pred.ID,
			UserID:       pred.UserID,
			MatchID:      matchID,
			Points:       points,
			IsCorrect:    pred.IsCorrect,
			Reason:       reason,
		})

		result.TotalPoints += points
	}

	// 批量更新用户积分
	if err := s.batchUpdateUserPoints(ctx, userPointsUpdates); err != nil {
		return nil, fmt.Errorf("failed to batch update user points: %w", err)
	}

	return result, nil
}

// batchUpdateUserPoints 批量更新用户积分
func (s *AsyncPointsService) batchUpdateUserPoints(ctx context.Context, updates map[uint]int) error {
	for userID, pointsChange := range updates {
		userEntity, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err,
			}).Error("Failed to get user for points update")
			continue
		}

		oldPoints := userEntity.Points
		userEntity.Points += pointsChange

		if err := s.userRepo.Update(ctx, userEntity); err != nil {
			s.logger.WithFields(logrus.Fields{
				"user_id":       userID,
				"old_points":    oldPoints,
				"points_change": pointsChange,
				"error":         err,
			}).Error("Failed to update user points")
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"user_id":       userID,
			"old_points":    oldPoints,
			"new_points":    userEntity.Points,
			"points_change": pointsChange,
		}).Debug("User points updated")
	}

	return nil
}

// updateCacheAfterCalculation 计算完成后更新缓存
func (s *AsyncPointsService) updateCacheAfterCalculation(ctx context.Context, matchID uint, result *scoring.MatchPointsCalculation) error {
	if s.cacheService == nil {
		return nil
	}

	// 使排行榜缓存失效
	tournaments := []string{"SPRING", "SUMMER", "AUTUMN", "WINTER"}
	for _, tournament := range tournaments {
		cacheKey := fmt.Sprintf("leaderboard:%s", tournament)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			s.logger.WithFields(logrus.Fields{
				"cache_key": cacheKey,
				"error":     err,
			}).Warn("Failed to invalidate leaderboard cache")
		}
	}

	// 使用户相关缓存失效
	for _, pointsResult := range result.Results {
		userCacheKey := fmt.Sprintf("user:%d", pointsResult.UserID)
		if err := s.cacheService.Delete(ctx, userCacheKey); err != nil {
			s.logger.WithFields(logrus.Fields{
				"cache_key": userCacheKey,
				"error":     err,
			}).Debug("Failed to invalidate user cache")
		}
	}

	s.logger.WithField("match_id", matchID).Debug("Cache invalidated after points calculation")
	return nil
}

// publishPointsCalculatedEvent 发布积分计算完成事件
func (s *AsyncPointsService) publishPointsCalculatedEvent(result *scoring.MatchPointsCalculation) error {
	if s.eventBus == nil {
		return nil
	}

	// 转换结果格式
	predictions := make([]shared.PredictionPointsInfo, len(result.Results))
	for i, r := range result.Results {
		predictions[i] = shared.PredictionPointsInfo{
			PredictionID: r.PredictionID,
			UserID:       r.UserID,
			Points:       r.Points,
			IsCorrect:    r.IsCorrect,
		}
	}

	payload := shared.PointsCalculatedPayload{
		MatchID:     result.MatchID,
		Predictions: predictions,
	}

	event := shared.NewEvent(shared.EventPointsCalculated, payload)
	return s.eventBus.Publish(event)
}

// updateTaskStatus 更新任务状态
func (s *AsyncPointsService) updateTaskStatus(taskID, status, errorMsg string) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	if task, exists := s.activeTasks[taskID]; exists {
		task.Status = status
		task.Error = errorMsg

		// 如果任务完成或失败，从活跃任务中移除
		if status == TaskStatusCompleted || status == TaskStatusFailed {
			delete(s.activeTasks, taskID)
		}
	}
}

// QueuePointsCalculation 将积分计算任务加入队列
func (s *AsyncPointsService) QueuePointsCalculation(matchID uint, ruleID *uint) (string, error) {
	taskID := fmt.Sprintf("points_%d_%d", matchID, time.Now().UnixNano())

	task := &PointsCalculationTask{
		ID:        taskID,
		MatchID:   matchID,
		RuleID:    ruleID,
		CreatedAt: time.Now(),
		Status:    TaskStatusPending,
	}

	// 添加到活跃任务
	s.taskMutex.Lock()
	s.activeTasks[taskID] = task
	s.taskMutex.Unlock()

	// 尝试加入队列
	select {
	case s.taskQueue <- task:
		s.logger.WithFields(logrus.Fields{
			"task_id":  taskID,
			"match_id": matchID,
		}).Info("Points calculation task queued")
		return taskID, nil
	default:
		// 队列满了
		s.taskMutex.Lock()
		delete(s.activeTasks, taskID)
		s.taskMutex.Unlock()
		return "", fmt.Errorf("task queue is full")
	}
}

// GetTaskStatus 获取任务状态
func (s *AsyncPointsService) GetTaskStatus(taskID string) (*PointsCalculationTask, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	task, exists := s.activeTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// 返回副本
	taskCopy := *task
	return &taskCopy, nil
}

// GetQueueStatus 获取队列状态
func (s *AsyncPointsService) GetQueueStatus() map[string]interface{} {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	return map[string]interface{}{
		"queue_length":   len(s.taskQueue),
		"queue_capacity": s.queueSize,
		"active_tasks":   len(s.activeTasks),
		"max_workers":    s.maxWorkers,
	}
}

// Shutdown 关闭服务
func (s *AsyncPointsService) Shutdown() {
	s.logger.Info("Shutting down async points service")

	s.cancel()
	close(s.taskQueue)
	s.wg.Wait()

	s.logger.Info("Async points service shutdown completed")
}
