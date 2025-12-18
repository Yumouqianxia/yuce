package services

import (
	"context"
	"fmt"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// MatchFinishHandler 比赛结束事件处理器
type MatchFinishHandler struct {
	asyncPointsService *AsyncPointsService
	logger             *logrus.Logger
}

// NewMatchFinishHandler 创建比赛结束事件处理器
func NewMatchFinishHandler(asyncPointsService *AsyncPointsService, logger *logrus.Logger) *MatchFinishHandler {
	if logger == nil {
		logger = logrus.New()
	}

	return &MatchFinishHandler{
		asyncPointsService: asyncPointsService,
		logger:             logger,
	}
}

// Handle 处理比赛结束事件
func (h *MatchFinishHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventMatchFinished {
		return fmt.Errorf("unexpected event type: %s", event.GetType())
	}

	payload, ok := event.GetPayload().(shared.MatchFinishedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match finished event")
	}

	logger := h.logger.WithFields(logrus.Fields{
		"match_id": payload.MatchID,
		"winner":   payload.Winner,
		"score_a":  payload.ScoreA,
		"score_b":  payload.ScoreB,
	})

	logger.Info("Handling match finished event")

	// 异步计算积分
	taskID, err := h.asyncPointsService.QueuePointsCalculation(payload.MatchID, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to queue points calculation")
		return fmt.Errorf("failed to queue points calculation: %w", err)
	}

	logger.WithField("task_id", taskID).Info("Points calculation task queued for finished match")
	return nil
}

// LeaderboardUpdateHandler 排行榜更新事件处理器
type LeaderboardUpdateHandler struct {
	leaderboardCacheService LeaderboardCacheService
	logger                  *logrus.Logger
}

// NewLeaderboardUpdateHandler 创建排行榜更新事件处理器
func NewLeaderboardUpdateHandler(leaderboardCacheService LeaderboardCacheService, logger *logrus.Logger) *LeaderboardUpdateHandler {
	if logger == nil {
		logger = logrus.New()
	}

	return &LeaderboardUpdateHandler{
		leaderboardCacheService: leaderboardCacheService,
		logger:                  logger,
	}
}

// Handle 处理积分计算完成事件，更新排行榜缓存
func (h *LeaderboardUpdateHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventPointsCalculated {
		return fmt.Errorf("unexpected event type: %s", event.GetType())
	}

	payload, ok := event.GetPayload().(shared.PointsCalculatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for points calculated event")
	}

	logger := h.logger.WithFields(logrus.Fields{
		"match_id":    payload.MatchID,
		"predictions": len(payload.Predictions),
	})

	logger.Info("Handling points calculated event")

	// 预热排行榜缓存
	tournaments := []string{"SPRING", "SUMMER", "AUTUMN", "WINTER"}
	for _, tournament := range tournaments {
		go func(t string) {
			if err := h.leaderboardCacheService.RefreshCache(context.Background(), t); err != nil {
				h.logger.WithFields(logrus.Fields{
					"tournament": t,
					"error":      err,
				}).Warn("Failed to warmup leaderboard cache")
			} else {
				h.logger.WithField("tournament", t).Debug("Leaderboard cache warmed up")
			}
		}(tournament)
	}

	logger.Info("Leaderboard cache warmup initiated")
	return nil
}

// PointsNotificationHandler 积分通知事件处理器
type PointsNotificationHandler struct {
	logger *logrus.Logger
}

// NewPointsNotificationHandler 创建积分通知事件处理器
func NewPointsNotificationHandler(logger *logrus.Logger) *PointsNotificationHandler {
	if logger == nil {
		logger = logrus.New()
	}

	return &PointsNotificationHandler{
		logger: logger,
	}
}

// Handle 处理积分计算完成事件，发送通知
func (h *PointsNotificationHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventPointsCalculated {
		return fmt.Errorf("unexpected event type: %s", event.GetType())
	}

	payload, ok := event.GetPayload().(shared.PointsCalculatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for points calculated event")
	}

	logger := h.logger.WithFields(logrus.Fields{
		"match_id":    payload.MatchID,
		"predictions": len(payload.Predictions),
	})

	logger.Info("Processing points calculation notifications")

	// 这里可以实现具体的通知逻辑
	// 例如：发送邮件、推送通知等
	for _, pred := range payload.Predictions {
		if pred.Points > 0 {
			// 发送积分获得通知
			h.logger.WithFields(logrus.Fields{
				"user_id":       pred.UserID,
				"prediction_id": pred.PredictionID,
				"points":        pred.Points,
				"is_correct":    pred.IsCorrect,
			}).Debug("User earned points notification")
		}
	}

	logger.Info("Points calculation notifications processed")
	return nil
}
