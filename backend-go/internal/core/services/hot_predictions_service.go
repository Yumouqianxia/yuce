package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/shared"
	"backend-go/internal/shared/logger"
)

// HotPredictionsService 热门预测服务
type HotPredictionsService struct {
	predictionRepo prediction.Repository
	eventBus       shared.EventBus

	// 缓存热门预测数据
	hotPredictionsCache map[uint][]prediction.PredictionWithVotes
	cacheMutex          sync.RWMutex
	cacheExpiry         map[uint]time.Time

	// 配置
	cacheTimeout    time.Duration
	updateThreshold int // 投票数变化阈值，超过此值才触发更新
}

// NewHotPredictionsService 创建热门预测服务
func NewHotPredictionsService(
	predictionRepo prediction.Repository,
	eventBus shared.EventBus,
) *HotPredictionsService {
	service := &HotPredictionsService{
		predictionRepo:      predictionRepo,
		eventBus:            eventBus,
		hotPredictionsCache: make(map[uint][]prediction.PredictionWithVotes),
		cacheExpiry:         make(map[uint]time.Time),
		cacheTimeout:        5 * time.Minute, // 缓存5分钟
		updateThreshold:     1,               // 每1票变化就更新
	}

	// 订阅投票事件
	service.subscribeToEvents()

	return service
}

// subscribeToEvents 订阅投票相关事件
func (s *HotPredictionsService) subscribeToEvents() {
	voteHandler := &voteEventHandler{service: s}
	s.eventBus.Subscribe(shared.EventPredictionVoted, voteHandler)
	s.eventBus.Subscribe(shared.EventPredictionUnvoted, voteHandler)

	logger.Info("Hot predictions service subscribed to vote events")
}

// handleVoteEvent 处理投票事件
func (s *HotPredictionsService) handleVoteEvent(event shared.Event) error {
	payload, ok := event.GetPayload().(*shared.PredictionVotedPayload)
	if !ok {
		return fmt.Errorf("invalid vote event payload")
	}

	// 获取预测信息以确定比赛ID
	ctx := context.Background()
	pred, err := s.predictionRepo.GetPredictionByID(ctx, payload.PredictionID)
	if err != nil {
		logger.Error("Failed to get prediction %d for hot predictions update: %v", payload.PredictionID, err)
		return err
	}

	// 异步更新热门预测
	go s.updateHotPredictionsForMatch(pred.MatchID)

	return nil
}

// GetHotPredictions 获取比赛的热门预测
func (s *HotPredictionsService) GetHotPredictions(ctx context.Context, matchID uint, limit int) ([]prediction.PredictionWithVotes, error) {
	s.cacheMutex.RLock()

	// 检查缓存是否有效
	if cachedPredictions, exists := s.hotPredictionsCache[matchID]; exists {
		if expiry, hasExpiry := s.cacheExpiry[matchID]; hasExpiry && time.Now().Before(expiry) {
			s.cacheMutex.RUnlock()

			// 返回限制数量的结果
			if len(cachedPredictions) > limit && limit > 0 {
				return cachedPredictions[:limit], nil
			}
			return cachedPredictions, nil
		}
	}

	s.cacheMutex.RUnlock()

	// 缓存无效，重新获取数据
	return s.refreshHotPredictions(ctx, matchID, limit)
}

// refreshHotPredictions 刷新热门预测数据
func (s *HotPredictionsService) refreshHotPredictions(ctx context.Context, matchID uint, limit int) ([]prediction.PredictionWithVotes, error) {
	// 获取比赛的所有预测
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get predictions for match %d: %w", matchID, err)
	}

	// 按投票数排序
	sortedPredictions := s.sortPredictionsByVotes(predictions)

	// 更新缓存
	s.cacheMutex.Lock()
	s.hotPredictionsCache[matchID] = sortedPredictions
	s.cacheExpiry[matchID] = time.Now().Add(s.cacheTimeout)
	s.cacheMutex.Unlock()

	// 返回限制数量的结果
	if len(sortedPredictions) > limit && limit > 0 {
		return sortedPredictions[:limit], nil
	}

	return sortedPredictions, nil
}

// updateHotPredictionsForMatch 更新指定比赛的热门预测
func (s *HotPredictionsService) updateHotPredictionsForMatch(matchID uint) {
	ctx := context.Background()

	// 刷新热门预测数据
	hotPredictions, err := s.refreshHotPredictions(ctx, matchID, 0) // 获取所有预测
	if err != nil {
		logger.Error("Failed to refresh hot predictions for match %d: %v", matchID, err)
		return
	}

	// 发布热门预测更新事件
	event := shared.NewEvent("hot_predictions_updated", map[string]interface{}{
		"match_id":        matchID,
		"hot_predictions": hotPredictions[:min(len(hotPredictions), 10)], // 只发布前10个
		"total_count":     len(hotPredictions),
		"updated_at":      time.Now(),
	})

	if err := s.eventBus.Publish(event); err != nil {
		logger.Error("Failed to publish hot predictions update event: %v", err)
	}

	logger.Debug("Hot predictions updated for match %d", matchID)
}

// sortPredictionsByVotes 按投票数排序预测
func (s *HotPredictionsService) sortPredictionsByVotes(predictions []prediction.PredictionWithVotes) []prediction.PredictionWithVotes {
	// 创建副本以避免修改原始数据
	sorted := make([]prediction.PredictionWithVotes, len(predictions))
	copy(sorted, predictions)

	// 按投票数降序排序，投票数相同时按创建时间升序排序
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].VoteCount == sorted[j].VoteCount {
			return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		}
		return sorted[i].VoteCount > sorted[j].VoteCount
	})

	return sorted
}

// GetVoteStatistics 获取比赛的投票统计信息
func (s *HotPredictionsService) GetVoteStatistics(ctx context.Context, matchID uint) (map[string]interface{}, error) {
	// 获取比赛的所有预测
	predictions, err := s.predictionRepo.GetPredictionsByMatch(ctx, matchID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get predictions for match %d: %w", matchID, err)
	}

	return s.calculateDetailedStatistics(predictions), nil
}

// calculateDetailedStatistics 计算详细的投票统计信息
func (s *HotPredictionsService) calculateDetailedStatistics(predictions []prediction.PredictionWithVotes) map[string]interface{} {
	totalPredictions := len(predictions)
	totalVotes := 0
	featuredCount := 0
	maxVotes := 0
	minVotes := 0
	voteDistribution := make(map[string]int)
	winnerDistribution := make(map[string]int)

	if totalPredictions > 0 {
		minVotes = predictions[0].VoteCount
	}

	for _, pred := range predictions {
		totalVotes += pred.VoteCount

		if pred.IsFeatured {
			featuredCount++
		}

		if pred.VoteCount > maxVotes {
			maxVotes = pred.VoteCount
		}

		if pred.VoteCount < minVotes {
			minVotes = pred.VoteCount
		}

		// 投票数分布
		voteRange := s.getVoteRange(pred.VoteCount)
		voteDistribution[voteRange]++

		// 预测获胜者分布
		winnerDistribution[pred.PredictedWinner]++
	}

	avgVotes := 0.0
	if totalPredictions > 0 {
		avgVotes = float64(totalVotes) / float64(totalPredictions)
	}

	// 计算投票活跃度（有投票的预测占比）
	activeCount := 0
	for _, pred := range predictions {
		if pred.VoteCount > 0 {
			activeCount++
		}
	}

	activityRate := 0.0
	if totalPredictions > 0 {
		activityRate = float64(activeCount) / float64(totalPredictions) * 100
	}

	return map[string]interface{}{
		"total_predictions":   totalPredictions,
		"total_votes":         totalVotes,
		"featured_count":      featuredCount,
		"max_votes":           maxVotes,
		"min_votes":           minVotes,
		"average_votes":       avgVotes,
		"activity_rate":       activityRate,
		"vote_distribution":   voteDistribution,
		"winner_distribution": winnerDistribution,
		"active_predictions":  activeCount,
	}
}

// getVoteRange 获取投票数范围
func (s *HotPredictionsService) getVoteRange(voteCount int) string {
	switch {
	case voteCount == 0:
		return "0"
	case voteCount <= 2:
		return "1-2"
	case voteCount <= 5:
		return "3-5"
	case voteCount <= 10:
		return "6-10"
	case voteCount <= 20:
		return "11-20"
	case voteCount <= 50:
		return "21-50"
	default:
		return "50+"
	}
}

// ClearCache 清除指定比赛的缓存
func (s *HotPredictionsService) ClearCache(matchID uint) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	delete(s.hotPredictionsCache, matchID)
	delete(s.cacheExpiry, matchID)

	logger.Debug("Cleared hot predictions cache for match %d", matchID)
}

// ClearAllCache 清除所有缓存
func (s *HotPredictionsService) ClearAllCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.hotPredictionsCache = make(map[uint][]prediction.PredictionWithVotes)
	s.cacheExpiry = make(map[uint]time.Time)

	logger.Info("Cleared all hot predictions cache")
}

// GetCacheStats 获取缓存统计信息
func (s *HotPredictionsService) GetCacheStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	validCaches := 0
	expiredCaches := 0
	now := time.Now()

	for matchID, expiry := range s.cacheExpiry {
		if _, exists := s.hotPredictionsCache[matchID]; exists {
			if now.Before(expiry) {
				validCaches++
			} else {
				expiredCaches++
			}
		}
	}

	return map[string]interface{}{
		"total_cached_matches": len(s.hotPredictionsCache),
		"valid_caches":         validCaches,
		"expired_caches":       expiredCaches,
		"cache_timeout":        s.cacheTimeout.String(),
	}
}

// Start 启动热门预测服务
func (s *HotPredictionsService) Start(ctx context.Context) error {
	logger.Info("Hot predictions service started")

	// 启动定期清理过期缓存的协程
	go s.startCacheCleanup(ctx)

	return nil
}

// Stop 停止热门预测服务
func (s *HotPredictionsService) Stop() error {
	// 注意：实际的取消订阅需要保存处理器引用，这里简化处理
	logger.Info("Hot predictions service stopped")
	return nil
}

// startCacheCleanup 启动缓存清理协程
func (s *HotPredictionsService) startCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpiredCache()
		}
	}
}

// cleanupExpiredCache 清理过期缓存
func (s *HotPredictionsService) cleanupExpiredCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	expiredMatches := make([]uint, 0)

	for matchID, expiry := range s.cacheExpiry {
		if now.After(expiry) {
			expiredMatches = append(expiredMatches, matchID)
		}
	}

	for _, matchID := range expiredMatches {
		delete(s.hotPredictionsCache, matchID)
		delete(s.cacheExpiry, matchID)
	}

	if len(expiredMatches) > 0 {
		logger.Debug("Cleaned up %d expired cache entries", len(expiredMatches))
	}
}

// voteEventHandler 投票事件处理器
type voteEventHandler struct {
	service *HotPredictionsService
}

// Handle 实现 EventHandler 接口
func (h *voteEventHandler) Handle(event shared.Event) error {
	return h.service.handleVoteEvent(event)
}

// min 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
