package events

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/adapters/events/handlers"
	"backend-go/internal/adapters/events/monitoring"
	"backend-go/internal/adapters/events/persistence"
	"backend-go/internal/adapters/events/replay"
	"backend-go/internal/core/domain/shared"
	"backend-go/pkg/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventManager 事件管理器
type EventManager struct {
	eventBus            shared.EventBus
	statisticsHandler   *handlers.StatisticsHandler
	notificationHandler *handlers.NotificationHandler
	persistentHandler   *persistence.PersistentEventHandler
	metricsCollector    *monitoring.MetricsCollector
	eventReplayer       *replay.EventReplayer
	logger              *logrus.Logger
}

// NewEventManager 创建事件管理器
func NewEventManager(
	eventBus shared.EventBus,
	db *gorm.DB,
	redisClient *redis.Client,
	logger *logrus.Logger,
) *EventManager {
	// 创建事件存储
	mysqlStore := persistence.NewMySQLEventStore(db, logger)
	redisStore := persistence.NewRedisEventStore(redisClient, logger)

	// 创建事件处理器
	statisticsHandler := handlers.NewStatisticsHandler(redisClient, logger)
	notificationService := handlers.NewMockNotificationService(logger)
	notificationHandler := handlers.NewNotificationHandler(notificationService, redisClient, logger)
	persistentHandler := persistence.NewPersistentEventHandler(mysqlStore, redisStore, logger)
	metricsCollector := monitoring.NewMetricsCollector(redisClient, logger)
	eventReplayer := replay.NewEventReplayer(mysqlStore, eventBus, logger)

	manager := &EventManager{
		eventBus:            eventBus,
		statisticsHandler:   statisticsHandler,
		notificationHandler: notificationHandler,
		persistentHandler:   persistentHandler,
		metricsCollector:    metricsCollector,
		eventReplayer:       eventReplayer,
		logger:              logger,
	}

	// 注册事件处理器
	manager.registerEventHandlers()

	return manager
}

// registerEventHandlers 注册事件处理器
func (m *EventManager) registerEventHandlers() {
	// 注册统计处理器
	m.eventBus.Subscribe(EventUserRegistered, m.statisticsHandler)
	m.eventBus.Subscribe(EventUserLoggedIn, m.statisticsHandler)
	m.eventBus.Subscribe(EventPredictionCreated, m.statisticsHandler)
	m.eventBus.Subscribe(EventVoteCast, m.statisticsHandler)
	m.eventBus.Subscribe(EventMatchViewed, m.statisticsHandler)
	m.eventBus.Subscribe(EventLeaderboardViewed, m.statisticsHandler)
	m.eventBus.Subscribe(EventPageViewed, m.statisticsHandler)
	m.eventBus.Subscribe(EventFeatureUsed, m.statisticsHandler)
	m.eventBus.Subscribe(EventErrorEncountered, m.statisticsHandler)

	// 注册通知处理器
	m.eventBus.Subscribe(EventUserRegistered, m.notificationHandler)
	m.eventBus.Subscribe(EventVoteCast, m.notificationHandler)
	m.eventBus.Subscribe(EventRankingChanged, m.notificationHandler)
	m.eventBus.Subscribe(EventMatchViewed, m.notificationHandler)

	// 注册持久化处理器（处理所有事件）
	eventTypes := []string{
		EventUserRegistered, EventUserLoggedIn, EventUserLoggedOut,
		EventUserProfileUpdated, EventUserPasswordChanged, EventUserDeactivated,
		EventPredictionCreated, EventPredictionUpdated, EventPredictionDeleted, EventPredictionViewed,
		EventVoteCast, EventVoteRemoved, EventVoteChanged,
		EventMatchViewed, EventMatchSubscribed, EventMatchUnsubscribed,
		EventLeaderboardViewed, EventRankingChanged,
		EventPageViewed, EventFeatureUsed, EventErrorEncountered, EventSearchPerformed,
	}

	for _, eventType := range eventTypes {
		m.eventBus.Subscribe(eventType, m.persistentHandler)
		m.eventBus.Subscribe(eventType, m.metricsCollector)
	}

	m.logger.Info("Event handlers registered successfully")
}

// PublishUserRegistered 发布用户注册事件
func (m *EventManager) PublishUserRegistered(userID uint, username, email, nickname, source string) error {
	payload := &UserRegisteredPayload{
		UserID:             userID,
		Username:           username,
		Email:              email,
		RegistrationSource: source,
	}

	event := NewUserBehaviorEvent(EventUserRegistered, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishUserLoggedIn 发布用户登录事件
func (m *EventManager) PublishUserLoggedIn(userID uint, username, loginMethod, loginSource string, loginCount int) error {
	payload := &UserLoggedInPayload{
		UserID:      userID,
		Username:    username,
		LoginMethod: loginMethod,
		LoginSource: loginSource,
		LoginCount:  loginCount,
	}

	event := NewUserBehaviorEvent(EventUserLoggedIn, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishUserLoggedOut 发布用户登出事件
func (m *EventManager) PublishUserLoggedOut(userID uint, username string, sessionDuration time.Duration) error {
	payload := &UserLoggedOutPayload{
		UserID:          userID,
		Username:        username,
		SessionDuration: sessionDuration,
		LoggedOutAt:     time.Now(),
	}

	event := NewUserBehaviorEvent(EventUserLoggedOut, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishUserProfileUpdated 发布用户资料更新事件
func (m *EventManager) PublishUserProfileUpdated(userID uint, username string, updatedFields []string, oldValues, newValues map[string]interface{}) error {
	payload := &UserProfileUpdatedPayload{
		UserID:        userID,
		Username:      username,
		UpdatedFields: updatedFields,
		OldValues:     oldValues,
		NewValues:     newValues,
		UpdatedAt:     time.Now(),
	}

	event := NewUserBehaviorEvent(EventUserProfileUpdated, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishPredictionCreated 发布预测创建事件
func (m *EventManager) PublishPredictionCreated(predictionID, userID, matchID uint, predictedWinner string, scoreA, scoreB int, tournament string, timeToStart time.Duration) error {
	payload := &PredictionCreatedPayload{
		PredictionID:     predictionID,
		UserID:           userID,
		MatchID:          matchID,
		Tournament:       tournament,
		TimeToMatchStart: timeToStart,
	}

	event := NewUserBehaviorEvent(EventPredictionCreated, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishPredictionUpdated 发布预测更新事件
func (m *EventManager) PublishPredictionUpdated(predictionID, userID, matchID uint, oldWinner, newWinner string, oldScoreA, newScoreA, oldScoreB, newScoreB int, modCount int, timeToStart time.Duration) error {
	payload := &PredictionUpdatedPayload{
		PredictionID:       predictionID,
		UserID:             userID,
		MatchID:            matchID,
		OldPredictedWinner: oldWinner,
		NewPredictedWinner: newWinner,
		OldPredictedScoreA: oldScoreA,
		NewPredictedScoreA: newScoreA,
		OldPredictedScoreB: oldScoreB,
		NewPredictedScoreB: newScoreB,
		ModificationCount:  modCount,
		TimeToMatchStart:   timeToStart,
		UpdatedAt:          time.Now(),
	}

	event := NewUserBehaviorEvent(EventPredictionUpdated, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishVoteCast 发布投票事件
func (m *EventManager) PublishVoteCast(voteID, userID, predictionID, matchID, voterID uint, newVoteCount int) error {
	payload := &VoteCastPayload{
		VoteID:       voteID,
		UserID:       userID,
		PredictionID: predictionID,
		VoterID:      voterID,
		NewVoteCount: newVoteCount,
	}

	event := NewUserBehaviorEvent(EventVoteCast, voterID, payload)
	return m.eventBus.Publish(event)
}

// PublishVoteRemoved 发布取消投票事件
func (m *EventManager) PublishVoteRemoved(userID, predictionID, matchID, voterID uint, newVoteCount int) error {
	payload := &VoteRemovedPayload{
		UserID:       userID,
		PredictionID: predictionID,
		MatchID:      matchID,
		VoterID:      voterID,
		NewVoteCount: newVoteCount,
		RemovedAt:    time.Now(),
	}

	event := NewUserBehaviorEvent(EventVoteRemoved, voterID, payload)
	return m.eventBus.Publish(event)
}

// PublishMatchViewed 发布比赛查看事件
func (m *EventManager) PublishMatchViewed(matchID, userID uint, teamA, teamB, tournament, status string, viewDuration time.Duration) error {
	payload := &MatchViewedPayload{
		MatchID:      matchID,
		UserID:       userID,
		Tournament:   tournament,
		ViewDuration: viewDuration,
	}

	event := NewUserBehaviorEvent(EventMatchViewed, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishLeaderboardViewed 发布排行榜查看事件
func (m *EventManager) PublishLeaderboardViewed(userID uint, tournament string, userRank, userPoints int, viewDuration time.Duration) error {
	payload := &LeaderboardViewedPayload{
		UserID:     userID,
		Tournament: tournament,
		UserRank:   userRank,
		UserPoints: userPoints,
	}

	event := NewUserBehaviorEvent(EventLeaderboardViewed, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishRankingChanged 发布排名变化事件
func (m *EventManager) PublishRankingChanged(userID uint, username, tournament string, oldRank, newRank, oldPoints, newPoints int) error {
	payload := &RankingChangedPayload{
		UserID:     userID,
		Username:   username,
		Tournament: tournament,
		OldRank:    oldRank,
		NewRank:    newRank,
		RankChange: oldRank - newRank, // 正数表示排名上升
	}

	event := NewUserBehaviorEvent(EventRankingChanged, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishPageViewed 发布页面访问事件
func (m *EventManager) PublishPageViewed(userID uint, pagePath, pageTitle, referrer string, viewDuration time.Duration) error {
	payload := &PageViewedPayload{
		UserID:   userID,
		PagePath: pagePath,
		Referrer: referrer,
	}

	event := NewUserBehaviorEvent(EventPageViewed, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishFeatureUsed 发布功能使用事件
func (m *EventManager) PublishFeatureUsed(userID uint, featureName, action string, parameters map[string]interface{}, success bool, duration time.Duration) error {
	payload := &FeatureUsedPayload{
		UserID:      userID,
		FeatureName: featureName,
		Action:      action,
		Success:     success,
		Duration:    duration,
	}

	event := NewUserBehaviorEvent(EventFeatureUsed, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishErrorEncountered 发布错误遇到事件
func (m *EventManager) PublishErrorEncountered(userID uint, errorType, errorCode, errorMessage, severity string, context map[string]interface{}) error {
	payload := &ErrorEncounteredPayload{
		UserID:       userID,
		ErrorType:    errorType,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Severity:     severity,
	}

	event := NewUserBehaviorEvent(EventErrorEncountered, userID, payload)
	return m.eventBus.Publish(event)
}

// PublishSearchPerformed 发布搜索执行事件
func (m *EventManager) PublishSearchPerformed(userID uint, searchQuery, searchType string, resultCount int, searchDuration time.Duration) error {
	payload := &SearchPerformedPayload{
		UserID:         userID,
		SearchQuery:    searchQuery,
		SearchType:     searchType,
		ResultCount:    resultCount,
		SearchDuration: searchDuration,
		PerformedAt:    time.Now(),
	}

	event := NewUserBehaviorEvent(EventSearchPerformed, userID, payload)
	return m.eventBus.Publish(event)
}

// GetStatistics 获取统计数据
func (m *EventManager) GetStatistics(ctx context.Context, statType, period string) (map[string]interface{}, error) {
	return m.statisticsHandler.GetStatistics(ctx, statType, period)
}

// GetMetrics 获取指标数据
func (m *EventManager) GetMetrics() map[string]*monitoring.EventMetrics {
	return m.metricsCollector.GetMetrics()
}

// GetSystemMetrics 获取系统级指标
func (m *EventManager) GetSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	return m.metricsCollector.GetSystemMetrics(ctx)
}

// ReplayEvents 重放事件
func (m *EventManager) ReplayEvents(ctx context.Context, options replay.ReplayOptions) (*replay.ReplayResult, error) {
	// 验证重放操作
	if err := m.eventReplayer.ValidateReplay(ctx, options); err != nil {
		return nil, fmt.Errorf("replay validation failed: %w", err)
	}

	return m.eventReplayer.ReplayEvents(ctx, options)
}

// ReplayFailedEvents 重放失败的事件
func (m *EventManager) ReplayFailedEvents(ctx context.Context, limit int) (*replay.ReplayResult, error) {
	return m.eventReplayer.ReplayFailedEvents(ctx, limit)
}

// GetReplayStatus 获取重放状态
func (m *EventManager) GetReplayStatus(ctx context.Context) (map[string]interface{}, error) {
	return m.eventReplayer.GetReplayStatus(ctx)
}

// Shutdown 关闭事件管理器
func (m *EventManager) Shutdown() {
	m.logger.Info("Shutting down event manager")

	// 如果事件总线支持关闭，调用关闭方法
	if shutdownable, ok := m.eventBus.(interface{ Shutdown() }); ok {
		shutdownable.Shutdown()
	}

	m.logger.Info("Event manager shutdown completed")
}
