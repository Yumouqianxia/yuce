package events

import (
	"encoding/json"
	"fmt"
	"time"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// LoggingEventHandler 日志事件处理器
type LoggingEventHandler struct {
	logger *logrus.Logger
	level  logrus.Level
}

// NewLoggingEventHandler 创建日志事件处理器
func NewLoggingEventHandler(logger *logrus.Logger, level logrus.Level) *LoggingEventHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &LoggingEventHandler{
		logger: logger,
		level:  level,
	}
}

// Handle 处理事件
func (h *LoggingEventHandler) Handle(event shared.Event) error {
	entry := h.logger.WithFields(logrus.Fields{
		"event_type": event.GetType(),
		"timestamp":  event.GetTimestamp(),
		"payload":    event.GetPayload(),
	})

	switch h.level {
	case logrus.DebugLevel:
		entry.Debug("Event received")
	case logrus.InfoLevel:
		entry.Info("Event received")
	case logrus.WarnLevel:
		entry.Warn("Event received")
	case logrus.ErrorLevel:
		entry.Error("Event received")
	default:
		entry.Info("Event received")
	}

	return nil
}

// MetricsEventHandler 指标事件处理器
type MetricsEventHandler struct {
	metrics map[string]*EventTypeMetrics
	logger  *logrus.Logger
}

// EventTypeMetrics 事件类型指标
type EventTypeMetrics struct {
	Count       int64         `json:"count"`
	LastSeen    time.Time     `json:"last_seen"`
	TotalTime   time.Duration `json:"total_time"`
	AverageTime time.Duration `json:"average_time"`
}

// NewMetricsEventHandler 创建指标事件处理器
func NewMetricsEventHandler(logger *logrus.Logger) *MetricsEventHandler {
	return &MetricsEventHandler{
		metrics: make(map[string]*EventTypeMetrics),
		logger:  logger,
	}
}

// Handle 处理事件
func (h *MetricsEventHandler) Handle(event shared.Event) error {
	start := time.Now()
	eventType := event.GetType()

	// 更新指标
	if h.metrics[eventType] == nil {
		h.metrics[eventType] = &EventTypeMetrics{}
	}

	metrics := h.metrics[eventType]
	metrics.Count++
	metrics.LastSeen = event.GetTimestamp()

	processingTime := time.Since(start)
	metrics.TotalTime += processingTime
	metrics.AverageTime = time.Duration(int64(metrics.TotalTime) / metrics.Count)

	h.logger.WithFields(logrus.Fields{
		"event_type":      eventType,
		"count":           metrics.Count,
		"processing_time": processingTime,
		"average_time":    metrics.AverageTime,
	}).Debug("Event metrics updated")

	return nil
}

// GetMetrics 获取指标
func (h *MetricsEventHandler) GetMetrics() map[string]*EventTypeMetrics {
	result := make(map[string]*EventTypeMetrics)
	for eventType, metrics := range h.metrics {
		metricsCopy := *metrics
		result[eventType] = &metricsCopy
	}
	return result
}

// PersistenceEventHandler 持久化事件处理器
type PersistenceEventHandler struct {
	storage EventStorage
	logger  *logrus.Logger
}

// EventStorage 事件存储接口
type EventStorage interface {
	Store(event shared.Event) error
	Retrieve(eventType string, limit int) ([]shared.Event, error)
	Delete(eventType string, before time.Time) error
}

// NewPersistenceEventHandler 创建持久化事件处理器
func NewPersistenceEventHandler(storage EventStorage, logger *logrus.Logger) *PersistenceEventHandler {
	return &PersistenceEventHandler{
		storage: storage,
		logger:  logger,
	}
}

// Handle 处理事件
func (h *PersistenceEventHandler) Handle(event shared.Event) error {
	if err := h.storage.Store(event); err != nil {
		h.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"error":      err,
		}).Error("Failed to store event")
		return fmt.Errorf("failed to store event: %w", err)
	}

	h.logger.WithField("event_type", event.GetType()).Debug("Event stored successfully")
	return nil
}

// NotificationEventHandler 通知事件处理器
type NotificationEventHandler struct {
	notifier EventNotifier
	filters  []EventFilter
	logger   *logrus.Logger
}

// EventNotifier 事件通知器接口
type EventNotifier interface {
	Notify(event shared.Event) error
}

// EventFilter 事件过滤器接口
type EventFilter interface {
	ShouldNotify(event shared.Event) bool
}

// NewNotificationEventHandler 创建通知事件处理器
func NewNotificationEventHandler(notifier EventNotifier, logger *logrus.Logger) *NotificationEventHandler {
	return &NotificationEventHandler{
		notifier: notifier,
		filters:  make([]EventFilter, 0),
		logger:   logger,
	}
}

// AddFilter 添加过滤器
func (h *NotificationEventHandler) AddFilter(filter EventFilter) {
	h.filters = append(h.filters, filter)
}

// Handle 处理事件
func (h *NotificationEventHandler) Handle(event shared.Event) error {
	// 检查过滤器
	for _, filter := range h.filters {
		if !filter.ShouldNotify(event) {
			h.logger.WithField("event_type", event.GetType()).Debug("Event filtered out by notification filter")
			return nil
		}
	}

	if err := h.notifier.Notify(event); err != nil {
		h.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"error":      err,
		}).Error("Failed to send notification")
		return fmt.Errorf("failed to send notification: %w", err)
	}

	h.logger.WithField("event_type", event.GetType()).Debug("Notification sent successfully")
	return nil
}

// WebSocket event handler removed - real-time features not needed

// ConditionalEventHandler 条件事件处理器
type ConditionalEventHandler struct {
	condition EventCondition
	handler   shared.EventHandler
	logger    *logrus.Logger
}

// EventCondition 事件条件接口
type EventCondition interface {
	ShouldHandle(event shared.Event) bool
}

// NewConditionalEventHandler 创建条件事件处理器
func NewConditionalEventHandler(condition EventCondition, handler shared.EventHandler, logger *logrus.Logger) *ConditionalEventHandler {
	return &ConditionalEventHandler{
		condition: condition,
		handler:   handler,
		logger:    logger,
	}
}

// Handle 处理事件
func (h *ConditionalEventHandler) Handle(event shared.Event) error {
	if !h.condition.ShouldHandle(event) {
		h.logger.WithField("event_type", event.GetType()).Debug("Event condition not met, skipping handler")
		return nil
	}

	return h.handler.Handle(event)
}

// ChainEventHandler 链式事件处理器
type ChainEventHandler struct {
	handlers []shared.EventHandler
	logger   *logrus.Logger
}

// NewChainEventHandler 创建链式事件处理器
func NewChainEventHandler(handlers []shared.EventHandler, logger *logrus.Logger) *ChainEventHandler {
	return &ChainEventHandler{
		handlers: handlers,
		logger:   logger,
	}
}

// Handle 处理事件
func (h *ChainEventHandler) Handle(event shared.Event) error {
	for i, handler := range h.handlers {
		if err := handler.Handle(event); err != nil {
			h.logger.WithFields(logrus.Fields{
				"event_type":    event.GetType(),
				"handler_index": i,
				"error":         err,
			}).Error("Handler in chain failed")
			return fmt.Errorf("handler %d in chain failed: %w", i, err)
		}
	}
	return nil
}

// AsyncEventHandler 异步事件处理器包装器
type AsyncEventHandler struct {
	handler shared.EventHandler
	logger  *logrus.Logger
}

// NewAsyncEventHandler 创建异步事件处理器
func NewAsyncEventHandler(handler shared.EventHandler, logger *logrus.Logger) *AsyncEventHandler {
	return &AsyncEventHandler{
		handler: handler,
		logger:  logger,
	}
}

// Handle 处理事件
func (h *AsyncEventHandler) Handle(event shared.Event) error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.WithFields(logrus.Fields{
					"event_type": event.GetType(),
					"panic":      r,
				}).Error("Async event handler panicked")
			}
		}()

		if err := h.handler.Handle(event); err != nil {
			h.logger.WithFields(logrus.Fields{
				"event_type": event.GetType(),
				"error":      err,
			}).Error("Async event handler failed")
		}
	}()

	return nil
}

// 实用工具函数

// EventTypeFilter 事件类型过滤器
type EventTypeFilter struct {
	allowedTypes map[string]bool
}

// NewEventTypeFilter 创建事件类型过滤器
func NewEventTypeFilter(allowedTypes []string) *EventTypeFilter {
	typeMap := make(map[string]bool)
	for _, eventType := range allowedTypes {
		typeMap[eventType] = true
	}
	return &EventTypeFilter{allowedTypes: typeMap}
}

// ShouldNotify 检查是否应该通知
func (f *EventTypeFilter) ShouldNotify(event shared.Event) bool {
	return f.allowedTypes[event.GetType()]
}

// ShouldHandle 检查是否应该处理
func (f *EventTypeFilter) ShouldHandle(event shared.Event) bool {
	return f.allowedTypes[event.GetType()]
}

// PayloadCondition 载荷条件
type PayloadCondition struct {
	fieldPath string
	expected  interface{}
}

// NewPayloadCondition 创建载荷条件
func NewPayloadCondition(fieldPath string, expected interface{}) *PayloadCondition {
	return &PayloadCondition{
		fieldPath: fieldPath,
		expected:  expected,
	}
}

// ShouldHandle 检查是否应该处理
func (c *PayloadCondition) ShouldHandle(event shared.Event) bool {
	payload := event.GetPayload()

	// 简单实现：将载荷转换为 JSON 并检查字段
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return false
	}

	value, exists := payloadMap[c.fieldPath]
	if !exists {
		return false
	}

	return value == c.expected
}

// TimeWindowCondition 时间窗口条件
type TimeWindowCondition struct {
	startTime time.Time
	endTime   time.Time
}

// NewTimeWindowCondition 创建时间窗口条件
func NewTimeWindowCondition(startTime, endTime time.Time) *TimeWindowCondition {
	return &TimeWindowCondition{
		startTime: startTime,
		endTime:   endTime,
	}
}

// ShouldHandle 检查是否应该处理
func (c *TimeWindowCondition) ShouldHandle(event shared.Event) bool {
	eventTime := event.GetTimestamp()
	return eventTime.After(c.startTime) && eventTime.Before(c.endTime)
}

// ShouldNotify 检查是否应该通知
func (c *TimeWindowCondition) ShouldNotify(event shared.Event) bool {
	return c.ShouldHandle(event)
}
