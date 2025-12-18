package handlers

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/adapters/events/types"
	"backend-go/internal/core/domain/shared"
	"backend-go/pkg/redis"
	"github.com/sirupsen/logrus"
)

// NotificationService 通知服务接口
type NotificationService interface {
	SendWelcomeEmail(email, username string) error
	SendRankingChangeNotification(userID uint, oldRank, newRank int) error
	SendPredictionVoteNotification(userID uint, predictionID uint, voterUsername string) error
	SendMatchStartNotification(userID uint, matchID uint, teamA, teamB string) error
	SendPointsEarnedNotification(userID uint, points int, matchID uint) error
}

// RankingChangedPayload 排名变化事件载荷
type RankingChangedPayload struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	OldRank    int    `json:"old_rank"`
	NewRank    int    `json:"new_rank"`
	RankChange int    `json:"rank_change"`
	Tournament string `json:"tournament"`
}

// NotificationHandler 通知事件处理器
type NotificationHandler struct {
	notificationService NotificationService
	redisClient         *redis.Client
	logger              *logrus.Logger
}

// NewNotificationHandler 创建通知事件处理器
func NewNotificationHandler(
	notificationService NotificationService,
	redisClient *redis.Client,
	logger *logrus.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		redisClient:         redisClient,
		logger:              logger,
	}
}

// Handle 处理事件
func (h *NotificationHandler) Handle(event shared.Event) error {
	ctx := context.Background()

	switch event.GetType() {
	case "user.registered":
		return h.handleUserRegistered(ctx, event)
	case "vote.cast":
		return h.handleVoteCast(ctx, event)
	case "ranking.changed":
		return h.handleRankingChanged(ctx, event)
	case "match.viewed":
		return h.handleMatchSubscription(ctx, event)
	default:
		h.logger.WithField("event_type", event.GetType()).Debug("Unhandled event type in notification handler")
		return nil
	}
}



// handleUserRegistered 处理用户注册通知
func (h *NotificationHandler) handleUserRegistered(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*types.UserRegisteredPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for user registered event")
	}

	// 发送欢迎邮件
	if err := h.notificationService.SendWelcomeEmail(payload.Email, payload.Username); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id":  payload.UserID,
			"email":    payload.Email,
			"username": payload.Username,
			"error":    err,
		}).Error("Failed to send welcome email")
		return err
	}

	// 记录通知发送
	notificationKey := fmt.Sprintf("notifications:sent:welcome:%d", payload.UserID)
	if err := h.redisClient.Set(ctx, notificationKey, time.Now().Unix(), 24*time.Hour); err != nil {
		h.logger.WithError(err).Warn("Failed to record welcome notification")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  payload.UserID,
		"username": payload.Username,
		"email":    payload.Email,
	}).Info("Welcome notification sent")

	return nil
}

// handleVoteCast 处理投票通知
func (h *NotificationHandler) handleVoteCast(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*types.VoteCastPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for vote cast event")
	}

	// 检查是否需要发送通知（避免频繁通知）
	notificationKey := fmt.Sprintf("notifications:vote_throttle:%d", payload.PredictionID)
	exists, err := h.redisClient.Exists(ctx, notificationKey)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to check vote notification throttle")
	}

	if exists > 0 {
		h.logger.WithField("prediction_id", payload.PredictionID).Debug("Vote notification throttled")
		return nil
	}

	// 获取投票者用户名（这里简化处理，实际应该从用户服务获取）
	voterUsername := fmt.Sprintf("User_%d", payload.VoterID)

	// 发送投票通知给预测创建者
	if err := h.notificationService.SendPredictionVoteNotification(
		payload.UserID,
		payload.PredictionID,
		voterUsername,
	); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id":       payload.UserID,
			"prediction_id": payload.PredictionID,
			"voter_id":      payload.VoterID,
			"error":         err,
		}).Error("Failed to send vote notification")
		return err
	}

	// 设置通知节流（5分钟内不重复发送）
	if err := h.redisClient.Set(ctx, notificationKey, time.Now().Unix(), 5*time.Minute); err != nil {
		h.logger.WithError(err).Warn("Failed to set vote notification throttle")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":        payload.UserID,
		"prediction_id":  payload.PredictionID,
		"voter_id":       payload.VoterID,
		"new_vote_count": payload.NewVoteCount,
	}).Info("Vote notification sent")

	return nil
}

// handleRankingChanged 处理排名变化通知
func (h *NotificationHandler) handleRankingChanged(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*RankingChangedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for ranking changed event")
	}

	// 只有排名提升时才发送通知
	if payload.RankChange <= 0 {
		return nil
	}

	// 检查是否是显著的排名变化（提升超过5名或进入前10）
	significantChange := payload.RankChange >= 5 || payload.NewRank <= 10
	if !significantChange {
		return nil
	}

	// 发送排名变化通知
	if err := h.notificationService.SendRankingChangeNotification(
		payload.UserID,
		payload.OldRank,
		payload.NewRank,
	); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id":  payload.UserID,
			"old_rank": payload.OldRank,
			"new_rank": payload.NewRank,
			"error":    err,
		}).Error("Failed to send ranking change notification")
		return err
	}

	// 记录通知发送
	notificationKey := fmt.Sprintf("notifications:sent:ranking:%d:%s", payload.UserID, payload.Tournament)
	if err := h.redisClient.Set(ctx, notificationKey, time.Now().Unix(), 24*time.Hour); err != nil {
		h.logger.WithError(err).Warn("Failed to record ranking notification")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     payload.UserID,
		"username":    payload.Username,
		"tournament":  payload.Tournament,
		"old_rank":    payload.OldRank,
		"new_rank":    payload.NewRank,
		"rank_change": payload.RankChange,
	}).Info("Ranking change notification sent")

	return nil
}

// handleMatchSubscription 处理比赛订阅通知
func (h *NotificationHandler) handleMatchSubscription(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*MatchViewedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match viewed event")
	}

	// 如果用户查看比赛超过30秒，自动订阅比赛通知
	if payload.ViewDuration < 30*time.Second {
		return nil
	}

	// 检查是否已经订阅
	subscriptionKey := fmt.Sprintf("subscriptions:match:%d:user:%d", payload.MatchID, payload.UserID)
	exists, err := h.redisClient.Exists(ctx, subscriptionKey)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to check match subscription")
		return nil
	}

	if exists > 0 {
		return nil // 已经订阅
	}

	// 添加订阅
	if err := h.redisClient.Set(ctx, subscriptionKey, time.Now().Unix(), 7*24*time.Hour); err != nil {
		h.logger.WithError(err).Error("Failed to add match subscription")
		return err
	}

	// 添加到比赛订阅者列表
	subscribersKey := fmt.Sprintf("subscriptions:match:%d:subscribers", payload.MatchID)
	if err := h.redisClient.SAdd(ctx, subscribersKey, payload.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to add user to match subscribers")
	}
	h.redisClient.Expire(ctx, subscribersKey, 7*24*time.Hour)

	h.logger.WithFields(logrus.Fields{
		"user_id":       payload.UserID,
		"match_id":      payload.MatchID,
		"view_duration": payload.ViewDuration,
	}).Info("User auto-subscribed to match notifications")

	return nil
}

// SendMatchStartNotifications 发送比赛开始通知给所有订阅者
func (h *NotificationHandler) SendMatchStartNotifications(ctx context.Context, matchID uint, teamA, teamB string) error {
	subscribersKey := fmt.Sprintf("subscriptions:match:%d:subscribers", matchID)

	subscribers, err := h.redisClient.SMembers(ctx, subscribersKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get match subscribers")
		return err
	}

	for _, subscriberStr := range subscribers {
		var userID uint
		if _, err := fmt.Sscanf(subscriberStr, "%d", &userID); err != nil {
			h.logger.WithError(err).Warn("Invalid subscriber ID format")
			continue
		}

		// 异步发送通知
		go func(uid uint) {
			if err := h.notificationService.SendMatchStartNotification(uid, matchID, teamA, teamB); err != nil {
				h.logger.WithFields(logrus.Fields{
					"user_id":  uid,
					"match_id": matchID,
					"error":    err,
				}).Error("Failed to send match start notification")
			}
		}(userID)
	}

	h.logger.WithFields(logrus.Fields{
		"match_id":         matchID,
		"team_a":           teamA,
		"team_b":           teamB,
		"subscriber_count": len(subscribers),
	}).Info("Match start notifications sent")

	return nil
}

// SendPointsEarnedNotifications 发送积分获得通知
func (h *NotificationHandler) SendPointsEarnedNotifications(ctx context.Context, matchID uint, predictions []shared.PredictionPointsInfo) error {
	for _, pred := range predictions {
		if pred.Points <= 0 {
			continue // 只通知获得积分的用户
		}

		// 异步发送通知
		go func(userID uint, points int) {
			if err := h.notificationService.SendPointsEarnedNotification(userID, points, matchID); err != nil {
				h.logger.WithFields(logrus.Fields{
					"user_id":  userID,
					"match_id": matchID,
					"points":   points,
					"error":    err,
				}).Error("Failed to send points earned notification")
			}
		}(pred.UserID, pred.Points)
	}

	h.logger.WithFields(logrus.Fields{
		"match_id":         matchID,
		"prediction_count": len(predictions),
	}).Info("Points earned notifications sent")

	return nil
}

// MockNotificationService 模拟通知服务实现
type MockNotificationService struct {
	logger *logrus.Logger
}

// NewMockNotificationService 创建模拟通知服务
func NewMockNotificationService(logger *logrus.Logger) NotificationService {
	return &MockNotificationService{
		logger: logger,
	}
}

// SendWelcomeEmail 发送欢迎邮件
func (s *MockNotificationService) SendWelcomeEmail(email, username string) error {
	s.logger.WithFields(logrus.Fields{
		"email":    email,
		"username": username,
	}).Info("Mock: Welcome email sent")
	return nil
}

// SendRankingChangeNotification 发送排名变化通知
func (s *MockNotificationService) SendRankingChangeNotification(userID uint, oldRank, newRank int) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"old_rank": oldRank,
		"new_rank": newRank,
	}).Info("Mock: Ranking change notification sent")
	return nil
}

// SendPredictionVoteNotification 发送预测投票通知
func (s *MockNotificationService) SendPredictionVoteNotification(userID uint, predictionID uint, voterUsername string) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"prediction_id":  predictionID,
		"voter_username": voterUsername,
	}).Info("Mock: Prediction vote notification sent")
	return nil
}

// SendMatchStartNotification 发送比赛开始通知
func (s *MockNotificationService) SendMatchStartNotification(userID uint, matchID uint, teamA, teamB string) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"match_id": matchID,
		"team_a":   teamA,
		"team_b":   teamB,
	}).Info("Mock: Match start notification sent")
	return nil
}

// SendPointsEarnedNotification 发送积分获得通知
func (s *MockNotificationService) SendPointsEarnedNotification(userID uint, points int, matchID uint) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"points":   points,
		"match_id": matchID,
	}).Info("Mock: Points earned notification sent")
	return nil
}
