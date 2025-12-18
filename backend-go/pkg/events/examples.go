package events

import (
	"fmt"
	"time"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// UserRegistrationHandler 用户注册事件处理器示例
type UserRegistrationHandler struct {
	logger *logrus.Logger
}

// NewUserRegistrationHandler 创建用户注册处理器
func NewUserRegistrationHandler(logger *logrus.Logger) *UserRegistrationHandler {
	return &UserRegistrationHandler{logger: logger}
}

// Handle 处理用户注册事件
func (h *UserRegistrationHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventUserRegistered {
		return nil
	}

	payload, ok := event.GetPayload().(*shared.UserRegisteredPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for user registered event")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  payload.UserID,
		"username": payload.Username,
		"email":    payload.Email,
	}).Info("Processing user registration")

	// 这里可以添加具体的业务逻辑，比如：
	// 1. 发送欢迎邮件
	// 2. 创建用户配置文件
	// 3. 初始化用户统计数据
	// 4. 记录用户行为日志

	return nil
}

// MatchStatusHandler 比赛状态变更处理器示例
type MatchStatusHandler struct {
	logger *logrus.Logger
}

// NewMatchStatusHandler 创建比赛状态处理器
func NewMatchStatusHandler(logger *logrus.Logger) *MatchStatusHandler {
	return &MatchStatusHandler{logger: logger}
}

// Handle 处理比赛状态变更事件
func (h *MatchStatusHandler) Handle(event shared.Event) error {
	switch event.GetType() {
	case shared.EventMatchStarted:
		return h.handleMatchStarted(event)
	case shared.EventMatchFinished:
		return h.handleMatchFinished(event)
	case shared.EventMatchStatusChanged:
		return h.handleMatchStatusChanged(event)
	default:
		return nil
	}
}

// handleMatchStarted 处理比赛开始事件
func (h *MatchStatusHandler) handleMatchStarted(event shared.Event) error {
	payload, ok := event.GetPayload().(*shared.MatchStartedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match started event")
	}

	h.logger.WithFields(logrus.Fields{
		"match_id":   payload.MatchID,
		"start_time": payload.StartTime,
	}).Info("Match started - locking predictions")

	// 业务逻辑：
	// 1. 锁定该比赛的所有预测
	// 2. 通知用户比赛已开始
	// 3. 开始实时比分更新

	return nil
}

// handleMatchFinished 处理比赛结束事件
func (h *MatchStatusHandler) handleMatchFinished(event shared.Event) error {
	payload, ok := event.GetPayload().(*shared.MatchFinishedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match finished event")
	}

	h.logger.WithFields(logrus.Fields{
		"match_id": payload.MatchID,
		"winner":   payload.Winner,
		"score_a":  payload.ScoreA,
		"score_b":  payload.ScoreB,
	}).Info("Match finished - calculating points")

	// 业务逻辑：
	// 1. 计算所有预测的积分
	// 2. 更新用户积分
	// 3. 更新排行榜
	// 4. 发送结果通知

	return nil
}

// handleMatchStatusChanged 处理比赛状态变更事件
func (h *MatchStatusHandler) handleMatchStatusChanged(event shared.Event) error {
	payload, ok := event.GetPayload().(*shared.MatchStatusChangedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match status changed event")
	}

	h.logger.WithFields(logrus.Fields{
		"match_id":   payload.MatchID,
		"old_status": payload.OldStatus,
		"new_status": payload.NewStatus,
	}).Info("Match status changed")

	// 业务逻辑：
	// 1. 广播状态变更给所有订阅者
	// 2. 根据新状态执行相应操作
	// 3. 更新缓存

	return nil
}

// PredictionVoteHandler 预测投票处理器示例
type PredictionVoteHandler struct {
	logger *logrus.Logger
}

// NewPredictionVoteHandler 创建预测投票处理器
func NewPredictionVoteHandler(logger *logrus.Logger) *PredictionVoteHandler {
	return &PredictionVoteHandler{logger: logger}
}

// Handle 处理预测投票事件
func (h *PredictionVoteHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventPredictionVoted {
		return nil
	}

	payload, ok := event.GetPayload().(*shared.PredictionVotedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for prediction voted event")
	}

	h.logger.WithFields(logrus.Fields{
		"prediction_id": payload.PredictionID,
		"user_id":       payload.UserID,
		"vote_count":    payload.VoteCount,
	}).Info("Processing prediction vote")

	// 业务逻辑：
	// 1. 更新预测的投票数
	// 2. 检查是否成为热门预测
	// 3. 实时广播投票更新
	// 4. 更新用户活跃度

	return nil
}

// LeaderboardUpdateHandler 排行榜更新处理器示例
type LeaderboardUpdateHandler struct {
	logger *logrus.Logger
}

// NewLeaderboardUpdateHandler 创建排行榜更新处理器
func NewLeaderboardUpdateHandler(logger *logrus.Logger) *LeaderboardUpdateHandler {
	return &LeaderboardUpdateHandler{logger: logger}
}

// Handle 处理排行榜更新事件
func (h *LeaderboardUpdateHandler) Handle(event shared.Event) error {
	if event.GetType() != shared.EventLeaderboardUpdated {
		return nil
	}

	h.logger.WithField("event_type", event.GetType()).Info("Processing leaderboard update")

	// 业务逻辑：
	// 1. 清除排行榜缓存
	// 2. 预热新的排行榜数据
	// 3. 通知排行榜变更
	// 4. 发送排名变化通知给用户

	return nil
}

// ExampleEventManagerUsage 事件管理器使用示例
func ExampleEventManagerUsage() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// 创建事件管理器
	config := DefaultEventManagerConfig()
	config.EnableLogging = true
	config.EnableMetrics = true

	eventManager := NewEventManager(config, logger)
	defer eventManager.Shutdown()

	// 注册自定义处理器
	userHandler := NewUserRegistrationHandler(logger)
	eventManager.RegisterHandler(shared.EventUserRegistered, userHandler, &HandlerMetadata{
		Name:        "user_registration_handler",
		Description: "Handles user registration events",
		Priority:    10,
	}, &HandlerInfo{
		Name:        "User Registration Handler",
		Type:        "business",
		Description: "Processes user registration and sends welcome emails",
		Version:     "1.0.0",
		Author:      "Business Team",
		Enabled:     true,
	})

	matchHandler := NewMatchStatusHandler(logger)
	eventManager.RegisterHandler(shared.EventMatchStarted, matchHandler, &HandlerMetadata{
		Name:        "match_status_handler",
		Description: "Handles match status changes",
		Priority:    5,
	}, &HandlerInfo{
		Name:        "Match Status Handler",
		Type:        "business",
		Description: "Processes match status changes and updates predictions",
		Version:     "1.0.0",
		Author:      "Business Team",
		Enabled:     true,
	})

	voteHandler := NewPredictionVoteHandler(logger)
	eventManager.RegisterHandler(shared.EventPredictionVoted, voteHandler, &HandlerMetadata{
		Name:        "prediction_vote_handler",
		Description: "Handles prediction voting events",
		Priority:    15,
	}, &HandlerInfo{
		Name:        "Prediction Vote Handler",
		Type:        "business",
		Description: "Processes prediction votes and updates statistics",
		Version:     "1.0.0",
		Author:      "Business Team",
		Enabled:     true,
	})

	// 发布事件示例
	userEvent := shared.NewEvent(shared.EventUserRegistered, &shared.UserRegisteredPayload{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
	})
	eventManager.PublishEvent(userEvent)

	matchEvent := shared.NewEvent(shared.EventMatchStarted, &shared.MatchStartedPayload{
		MatchID:   1,
		StartTime: time.Now(),
	})
	eventManager.PublishEvent(matchEvent)

	voteEvent := shared.NewEvent(shared.EventPredictionVoted, &shared.PredictionVotedPayload{
		PredictionID: 1,
		UserID:       2,
		VoteCount:    5,
	})
	eventManager.PublishEvent(voteEvent)

	// 等待事件处理完成
	time.Sleep(2 * time.Second)

	// 获取指标
	metrics := eventManager.GetMetrics()
	logger.WithFields(logrus.Fields{
		"published_events": metrics.EventBusMetrics.PublishedEvents,
		"processed_events": metrics.EventBusMetrics.ProcessedEvents,
		"failed_events":    metrics.EventBusMetrics.FailedEvents,
		"total_handlers":   metrics.RegistryStats.TotalHandlers,
	}).Info("Event processing completed")
}

// ExampleConditionalHandler 条件处理器使用示例
func ExampleConditionalHandler() {
	logger := logrus.New()

	// 创建只处理特定用户事件的条件
	condition := NewPayloadCondition("user_id", uint(1))

	// 创建基础处理器
	baseHandler := NewUserRegistrationHandler(logger)

	// 创建条件处理器
	conditionalHandler := NewConditionalEventHandler(condition, baseHandler, logger)

	// 创建事件管理器并注册条件处理器
	eventManager := NewEventManager(nil, logger)
	defer eventManager.Shutdown()

	eventManager.RegisterHandler(shared.EventUserRegistered, conditionalHandler, &HandlerMetadata{
		Name:        "conditional_user_handler",
		Description: "Only handles events for user ID 1",
		Priority:    5,
	}, &HandlerInfo{
		Name:        "Conditional User Handler",
		Type:        "conditional",
		Description: "Processes events only for specific users",
		Version:     "1.0.0",
		Author:      "System",
		Enabled:     true,
	})

	// 发布事件 - 这个会被处理
	event1 := shared.NewEvent(shared.EventUserRegistered, &shared.UserRegisteredPayload{
		UserID:   1,
		Username: "user1",
		Email:    "user1@example.com",
	})
	eventManager.PublishEvent(event1)

	// 发布事件 - 这个不会被处理
	event2 := shared.NewEvent(shared.EventUserRegistered, &shared.UserRegisteredPayload{
		UserID:   2,
		Username: "user2",
		Email:    "user2@example.com",
	})
	eventManager.PublishEvent(event2)

	time.Sleep(1 * time.Second)
}

// ExampleChainHandler 链式处理器使用示例
func ExampleChainHandler() {
	logger := logrus.New()

	// 创建多个处理器
	handler1 := NewUserRegistrationHandler(logger)
	handler2 := NewLoggingEventHandler(logger, logrus.InfoLevel)
	handler3 := NewMetricsEventHandler(logger)

	// 创建链式处理器
	chainHandler := NewChainEventHandler([]shared.EventHandler{
		handler1,
		handler2,
		handler3,
	}, logger)

	// 创建事件管理器并注册链式处理器
	eventManager := NewEventManager(nil, logger)
	defer eventManager.Shutdown()

	eventManager.RegisterHandler(shared.EventUserRegistered, chainHandler, &HandlerMetadata{
		Name:        "user_registration_chain",
		Description: "Chain of handlers for user registration",
		Priority:    1,
	}, &HandlerInfo{
		Name:        "User Registration Chain",
		Type:        "chain",
		Description: "Executes multiple handlers in sequence",
		Version:     "1.0.0",
		Author:      "System",
		Enabled:     true,
	})

	// 发布事件
	event := shared.NewEvent(shared.EventUserRegistered, &shared.UserRegisteredPayload{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
	})
	eventManager.PublishEvent(event)

	time.Sleep(1 * time.Second)
}
