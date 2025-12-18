package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// EventManager 事件管理器
type EventManager struct {
	eventBus        *EnhancedEventBus
	handlerRegistry *HandlerRegistry
	config          *EventManagerConfig
	logger          *logrus.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup

	// 内置处理器
	loggingHandler      *LoggingEventHandler
	metricsHandler      *MetricsEventHandler
	persistenceHandler  *PersistenceEventHandler
	notificationHandler *NotificationEventHandler
	// webSocketHandler removed
}

// EventManagerConfig 事件管理器配置
type EventManagerConfig struct {
	EventBusConfig     *EventBusConfig `json:"event_bus"`
	EnableLogging      bool            `json:"enable_logging"`
	EnableMetrics      bool            `json:"enable_metrics"`
	EnablePersistence  bool            `json:"enable_persistence"`
	EnableNotification bool            `json:"enable_notification"`
	// EnableWebSocket removed
	LogLevel           string          `json:"log_level"`

	// 自动注册标准处理器
	AutoRegisterHandlers bool `json:"auto_register_handlers"`

	// 健康检查配置
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// DefaultEventManagerConfig 默认配置
func DefaultEventManagerConfig() *EventManagerConfig {
	return &EventManagerConfig{
		EventBusConfig:       DefaultEventBusConfig(),
		EnableLogging:        true,
		EnableMetrics:        true,
		EnablePersistence:    false,
		EnableNotification:   false,
		// EnableWebSocket removed
		LogLevel:             "info",
		AutoRegisterHandlers: true,
		HealthCheckInterval:  30 * time.Second,
	}
}

// NewEventManager 创建事件管理器
func NewEventManager(config *EventManagerConfig, logger *logrus.Logger) *EventManager {
	if config == nil {
		config = DefaultEventManagerConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 创建增强的事件总线
	eventBus := NewEnhancedEventBus(config.EventBusConfig, logger)

	// 创建处理器注册表
	handlerRegistry := NewHandlerRegistry(eventBus, logger)

	manager := &EventManager{
		eventBus:        eventBus,
		handlerRegistry: handlerRegistry,
		config:          config,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}

	// 初始化内置处理器
	manager.initializeBuiltinHandlers()

	// 自动注册标准处理器
	if config.AutoRegisterHandlers {
		manager.registerStandardHandlers()
	}

	// 启动健康检查
	if config.HealthCheckInterval > 0 {
		manager.wg.Add(1)
		go manager.healthCheck()
	}

	return manager
}

// initializeBuiltinHandlers 初始化内置处理器
func (m *EventManager) initializeBuiltinHandlers() {
	// 日志处理器
	if m.config.EnableLogging {
		level := logrus.InfoLevel
		if parsedLevel, err := logrus.ParseLevel(m.config.LogLevel); err == nil {
			level = parsedLevel
		}
		m.loggingHandler = NewLoggingEventHandler(m.logger, level)
	}

	// 指标处理器
	if m.config.EnableMetrics {
		m.metricsHandler = NewMetricsEventHandler(m.logger)
	}

	// 其他处理器根据需要初始化
	// 注意：持久化、通知和 WebSocket 处理器需要外部依赖，在这里不初始化
}

// registerStandardHandlers 注册标准处理器
func (m *EventManager) registerStandardHandlers() {
	// 注册日志处理器到所有事件类型
	if m.loggingHandler != nil {
		eventTypes := []string{
			shared.EventUserRegistered,
			shared.EventUserLoggedIn,
			shared.EventMatchCreated,
			shared.EventMatchStarted,
			shared.EventMatchFinished,
			shared.EventMatchStatusChanged,
			shared.EventPredictionCreated,
			shared.EventPredictionVoted,
			shared.EventPointsCalculated,
		}

		for _, eventType := range eventTypes {
			m.RegisterHandler(eventType, m.loggingHandler, &HandlerMetadata{
				Name:        "logging_handler",
				Description: "Standard logging handler for all events",
				Priority:    1000, // 低优先级，最后执行
			}, &HandlerInfo{
				Name:        "Standard Logging Handler",
				Type:        "logging",
				Description: "Logs all events to the configured logger",
				Version:     "1.0.0",
				Author:      "System",
				Enabled:     true,
			})
		}
	}

	// 注册指标处理器
	if m.metricsHandler != nil {
		eventTypes := []string{
			shared.EventUserRegistered,
			shared.EventUserLoggedIn,
			shared.EventMatchCreated,
			shared.EventMatchStarted,
			shared.EventMatchFinished,
			shared.EventPredictionCreated,
			shared.EventPredictionVoted,
			shared.EventPointsCalculated,
		}

		for _, eventType := range eventTypes {
			m.RegisterHandler(eventType, m.metricsHandler, &HandlerMetadata{
				Name:        "metrics_handler",
				Description: "Standard metrics handler for all events",
				Priority:    999, // 低优先级
			}, &HandlerInfo{
				Name:        "Standard Metrics Handler",
				Type:        "metrics",
				Description: "Collects metrics for all events",
				Version:     "1.0.0",
				Author:      "System",
				Enabled:     true,
			})
		}
	}
}

// RegisterHandler 注册事件处理器
func (m *EventManager) RegisterHandler(eventType string, handler shared.EventHandler, metadata *HandlerMetadata, info *HandlerInfo) error {
	return m.handlerRegistry.RegisterHandler(eventType, handler, metadata, info)
}

// UnregisterHandler 取消注册事件处理器
func (m *EventManager) UnregisterHandler(eventType string, handler shared.EventHandler) error {
	return m.handlerRegistry.UnregisterHandler(eventType, handler)
}

// PublishEvent 发布事件
func (m *EventManager) PublishEvent(event shared.Event) error {
	return m.eventBus.Publish(event)
}

// PublishEventSync 同步发布事件（等待处理完成）
func (m *EventManager) PublishEventSync(event shared.Event, timeout time.Duration) error {
	// 创建一个完成通道
	done := make(chan error, 1)

	// 创建一个临时处理器来监听完成
	tempHandler := &syncHandler{done: done}

	// 注册临时处理器
	if err := m.eventBus.Subscribe(event.GetType(), tempHandler); err != nil {
		return err
	}

	// 发布事件
	if err := m.eventBus.Publish(event); err != nil {
		m.eventBus.Unsubscribe(event.GetType(), tempHandler)
		return err
	}

	// 等待完成或超时
	select {
	case err := <-done:
		m.eventBus.Unsubscribe(event.GetType(), tempHandler)
		return err
	case <-time.After(timeout):
		m.eventBus.Unsubscribe(event.GetType(), tempHandler)
		return fmt.Errorf("event processing timeout after %v", timeout)
	}
}

// syncHandler 同步处理器
type syncHandler struct {
	done chan error
}

func (h *syncHandler) Handle(event shared.Event) error {
	h.done <- nil
	return nil
}

// SetPersistenceHandler 设置持久化处理器
func (m *EventManager) SetPersistenceHandler(storage EventStorage) error {
	if !m.config.EnablePersistence {
		return fmt.Errorf("persistence is not enabled in configuration")
	}

	m.persistenceHandler = NewPersistenceEventHandler(storage, m.logger)

	// 注册到所有事件类型
	eventTypes := m.handlerRegistry.ListEventTypes()
	for _, eventType := range eventTypes {
		if err := m.RegisterHandler(eventType, m.persistenceHandler, &HandlerMetadata{
			Name:        "persistence_handler",
			Description: "Persists events to storage",
			Priority:    10, // 高优先级
		}, &HandlerInfo{
			Name:        "Persistence Handler",
			Type:        "persistence",
			Description: "Stores events in persistent storage",
			Version:     "1.0.0",
			Author:      "System",
			Enabled:     true,
		}); err != nil {
			return fmt.Errorf("failed to register persistence handler for %s: %w", eventType, err)
		}
	}

	return nil
}

// SetNotificationHandler 设置通知处理器
func (m *EventManager) SetNotificationHandler(notifier EventNotifier) error {
	if !m.config.EnableNotification {
		return fmt.Errorf("notification is not enabled in configuration")
	}

	m.notificationHandler = NewNotificationEventHandler(notifier, m.logger)

	// 只注册到特定事件类型
	notificationEvents := []string{
		shared.EventUserRegistered,
		shared.EventMatchStarted,
		shared.EventMatchFinished,
		shared.EventPointsCalculated,
	}

	for _, eventType := range notificationEvents {
		if err := m.RegisterHandler(eventType, m.notificationHandler, &HandlerMetadata{
			Name:        "notification_handler",
			Description: "Sends notifications for important events",
			Priority:    50,
		}, &HandlerInfo{
			Name:        "Notification Handler",
			Type:        "notification",
			Description: "Sends notifications via configured channels",
			Version:     "1.0.0",
			Author:      "System",
			Enabled:     true,
		}); err != nil {
			return fmt.Errorf("failed to register notification handler for %s: %w", eventType, err)
		}
	}

	return nil
}

// WebSocket handler methods removed - real-time features not needed

// GetMetrics 获取事件管理器指标
func (m *EventManager) GetMetrics() EventManagerMetrics {
	busMetrics := m.eventBus.GetMetrics()
	registryStats := m.handlerRegistry.GetRegistryStats()

	metrics := EventManagerMetrics{
		EventBusMetrics: busMetrics,
		RegistryStats:   registryStats,
	}

	// 添加内置处理器指标
	if m.metricsHandler != nil {
		metrics.HandlerMetrics = m.metricsHandler.GetMetrics()
	}

	return metrics
}

// EventManagerMetrics 事件管理器指标
type EventManagerMetrics struct {
	EventBusMetrics EventMetrics                 `json:"event_bus_metrics"`
	RegistryStats   RegistryStats                `json:"registry_stats"`
	HandlerMetrics  map[string]*EventTypeMetrics `json:"handler_metrics,omitempty"`
}

// healthCheck 健康检查
func (m *EventManager) healthCheck() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (m *EventManager) performHealthCheck() {
	// 验证处理器
	validationErrors := m.handlerRegistry.ValidateHandlers()
	if len(validationErrors) > 0 {
		m.logger.WithField("errors", validationErrors).Warn("Handler validation errors found")
	}

	// 检查事件总线指标
	metrics := m.eventBus.GetMetrics()
	if metrics.FailedEvents > 0 {
		m.logger.WithFields(logrus.Fields{
			"failed_events": metrics.FailedEvents,
			"total_events":  metrics.PublishedEvents,
			"failure_rate":  float64(metrics.FailedEvents) / float64(metrics.PublishedEvents),
		}).Warn("Event processing failures detected")
	}

	m.logger.WithFields(logrus.Fields{
		"published_events":    metrics.PublishedEvents,
		"processed_events":    metrics.ProcessedEvents,
		"failed_events":       metrics.FailedEvents,
		"handler_panics":      metrics.HandlerPanics,
		"registered_handlers": m.handlerRegistry.GetHandlerCount(),
	}).Debug("Event manager health check completed")
}

// Shutdown 关闭事件管理器
func (m *EventManager) Shutdown() {
	m.logger.Info("Shutting down event manager")

	// 取消上下文
	m.cancel()

	// 等待健康检查协程完成
	m.wg.Wait()

	// 关闭事件总线
	m.eventBus.Shutdown()

	m.logger.Info("Event manager shutdown completed")
}

// GetHandlerRegistry 获取处理器注册表
func (m *EventManager) GetHandlerRegistry() *HandlerRegistry {
	return m.handlerRegistry
}

// GetEventBus 获取事件总线
func (m *EventManager) GetEventBus() *EnhancedEventBus {
	return m.eventBus
}

// EnableHandler 启用处理器
func (m *EventManager) EnableHandler(handlerID string) error {
	return m.handlerRegistry.EnableHandler(handlerID)
}

// DisableHandler 禁用处理器
func (m *EventManager) DisableHandler(handlerID string) error {
	return m.handlerRegistry.DisableHandler(handlerID)
}

// ListHandlers 列出所有处理器
func (m *EventManager) ListHandlers() map[string]*HandlerInfo {
	return m.handlerRegistry.ListHandlers()
}

// GetHandlerInfo 获取处理器信息
func (m *EventManager) GetHandlerInfo(handlerID string) (*HandlerInfo, error) {
	return m.handlerRegistry.GetHandlerInfo(handlerID)
}
