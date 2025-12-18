package services

import (
	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/prediction"
	"backend-go/internal/core/domain/shared"
	"backend-go/internal/core/domain/user"
	"backend-go/pkg/cache"
	"backend-go/pkg/events"
	"github.com/sirupsen/logrus"
)

// AsyncPointsIntegration 异步积分计算集成服务
type AsyncPointsIntegration struct {
	EventBus            shared.EventBus
	AsyncPointsService  *AsyncPointsService
	MatchFinishHandler  *MatchFinishHandler
	LeaderboardHandler  *LeaderboardUpdateHandler
	NotificationHandler *PointsNotificationHandler
}

// NewAsyncPointsIntegration 创建异步积分计算集成服务
func NewAsyncPointsIntegration(
	predictionRepo prediction.Repository,
	scoringRuleRepo prediction.ScoringRuleRepository,
	matchRepo match.Repository,
	userRepo user.Repository,
	cacheService cache.CacheService,
	leaderboardCacheService LeaderboardCacheService,
	logger *logrus.Logger,
) *AsyncPointsIntegration {
	// 创建事件总线
	eventBus := events.NewAsyncEventBus(1000, logger) // 队列大小1000

	// 创建异步积分服务
	asyncPointsService := NewAsyncPointsService(
		predictionRepo,
		scoringRuleRepo,
		matchRepo,
		userRepo,
		cacheService,
		eventBus,
		logger,
	)

	// 创建事件处理器
	matchFinishHandler := NewMatchFinishHandler(asyncPointsService, logger)
	leaderboardHandler := NewLeaderboardUpdateHandler(leaderboardCacheService, logger)
	notificationHandler := NewPointsNotificationHandler(logger)

	// 注册事件处理器
	eventBus.Subscribe(shared.EventMatchFinished, matchFinishHandler)
	eventBus.Subscribe(shared.EventPointsCalculated, leaderboardHandler)
	eventBus.Subscribe(shared.EventPointsCalculated, notificationHandler)

	logger.Info("Async points calculation integration initialized")

	return &AsyncPointsIntegration{
		EventBus:            eventBus,
		AsyncPointsService:  asyncPointsService,
		MatchFinishHandler:  matchFinishHandler,
		LeaderboardHandler:  leaderboardHandler,
		NotificationHandler: notificationHandler,
	}
}

// Shutdown 关闭集成服务
func (integration *AsyncPointsIntegration) Shutdown() {
	if integration.AsyncPointsService != nil {
		integration.AsyncPointsService.Shutdown()
	}

	if asyncBus, ok := integration.EventBus.(*events.AsyncEventBus); ok {
		asyncBus.Shutdown()
	}
}

// GetEventBus 获取事件总线
func (integration *AsyncPointsIntegration) GetEventBus() shared.EventBus {
	return integration.EventBus
}

// GetAsyncPointsService 获取异步积分服务
func (integration *AsyncPointsIntegration) GetAsyncPointsService() *AsyncPointsService {
	return integration.AsyncPointsService
}

// ManualTriggerPointsCalculation 手动触发积分计算
func (integration *AsyncPointsIntegration) ManualTriggerPointsCalculation(matchID uint, ruleID *uint) (string, error) {
	return integration.AsyncPointsService.QueuePointsCalculation(matchID, ruleID)
}

// GetCalculationStatus 获取计算状态
func (integration *AsyncPointsIntegration) GetCalculationStatus() map[string]interface{} {
	return integration.AsyncPointsService.GetQueueStatus()
}
