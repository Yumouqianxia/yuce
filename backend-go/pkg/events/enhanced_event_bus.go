package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	QueueSize      int           `json:"queue_size"`
	WorkerCount    int           `json:"worker_count"`
	RetryAttempts  int           `json:"retry_attempts"`
	RetryDelay     time.Duration `json:"retry_delay"`
	HandlerTimeout time.Duration `json:"handler_timeout"`
	EnableMetrics  bool          `json:"enable_metrics"`
}

// DefaultEventBusConfig 默认配置
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		QueueSize:      1000,
		WorkerCount:    5,
		RetryAttempts:  3,
		RetryDelay:     time.Second,
		HandlerTimeout: 30 * time.Second,
		EnableMetrics:  true,
	}
}

// EventMetrics 事件指标
type EventMetrics struct {
	PublishedEvents    int64         `json:"published_events"`
	ProcessedEvents    int64         `json:"processed_events"`
	FailedEvents       int64         `json:"failed_events"`
	RetryEvents        int64         `json:"retry_events"`
	HandlerPanics      int64         `json:"handler_panics"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	mutex              sync.RWMutex
}

// IncrementPublished 增加发布事件计数
func (m *EventMetrics) IncrementPublished() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.PublishedEvents++
}

// IncrementProcessed 增加处理事件计数
func (m *EventMetrics) IncrementProcessed() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ProcessedEvents++
}

// IncrementFailed 增加失败事件计数
func (m *EventMetrics) IncrementFailed() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.FailedEvents++
}

// IncrementRetry 增加重试事件计数
func (m *EventMetrics) IncrementRetry() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.RetryEvents++
}

// IncrementPanics 增加处理器恐慌计数
func (m *EventMetrics) IncrementPanics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.HandlerPanics++
}

// UpdateProcessTime 更新平均处理时间
func (m *EventMetrics) UpdateProcessTime(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// 简单的移动平均
	if m.AverageProcessTime == 0 {
		m.AverageProcessTime = duration
	} else {
		m.AverageProcessTime = (m.AverageProcessTime + duration) / 2
	}
}

// GetMetrics 获取指标快照
func (m *EventMetrics) GetMetrics() EventMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return *m
}

// EventWrapper 事件包装器，用于重试机制
type EventWrapper struct {
	Event       shared.Event
	Attempts    int
	LastError   error
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// EnhancedEventBus 增强的事件总线
type EnhancedEventBus struct {
	config      *EventBusConfig
	eventQueue  chan *EventWrapper
	retryQueue  chan *EventWrapper
	handlers    map[string][]shared.EventHandler
	handlerMeta map[string]*HandlerMetadata
	mutex       sync.RWMutex
	logger      *logrus.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	metrics     *EventMetrics
}

// HandlerMetadata 处理器元数据
type HandlerMetadata struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Timeout     time.Duration `json:"timeout"`
	RetryCount  int           `json:"retry_count"`
	Priority    int           `json:"priority"` // 优先级，数字越小优先级越高
}

// NewEnhancedEventBus 创建增强的事件总线
func NewEnhancedEventBus(config *EventBusConfig, logger *logrus.Logger) *EnhancedEventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bus := &EnhancedEventBus{
		config:      config,
		eventQueue:  make(chan *EventWrapper, config.QueueSize),
		retryQueue:  make(chan *EventWrapper, config.QueueSize/2),
		handlers:    make(map[string][]shared.EventHandler),
		handlerMeta: make(map[string]*HandlerMetadata),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &EventMetrics{},
	}

	// 启动工作协程
	for i := 0; i < config.WorkerCount; i++ {
		bus.wg.Add(1)
		go bus.processEvents(i)
	}

	// 启动重试协程
	bus.wg.Add(1)
	go bus.processRetryEvents()

	// 启动指标报告协程
	if config.EnableMetrics {
		bus.wg.Add(1)
		go bus.reportMetrics()
	}

	return bus
}

// processEvents 处理事件队列
func (bus *EnhancedEventBus) processEvents(workerID int) {
	defer bus.wg.Done()

	bus.logger.WithField("worker_id", workerID).Info("Event worker started")

	for {
		select {
		case <-bus.ctx.Done():
			bus.logger.WithField("worker_id", workerID).Info("Event worker shutting down")
			return
		case eventWrapper := <-bus.eventQueue:
			bus.handleEventWrapper(eventWrapper, workerID)
		}
	}
}

// processRetryEvents 处理重试事件队列
func (bus *EnhancedEventBus) processRetryEvents() {
	defer bus.wg.Done()

	bus.logger.Info("Retry worker started")

	ticker := time.NewTicker(bus.config.RetryDelay)
	defer ticker.Stop()

	for {
		select {
		case <-bus.ctx.Done():
			bus.logger.Info("Retry worker shutting down")
			return
		case <-ticker.C:
			// 处理重试队列中的事件
			select {
			case eventWrapper := <-bus.retryQueue:
				bus.handleEventWrapper(eventWrapper, -1) // -1 表示重试工作器
			default:
				// 重试队列为空，继续等待
			}
		}
	}
}

// handleEventWrapper 处理事件包装器
func (bus *EnhancedEventBus) handleEventWrapper(eventWrapper *EventWrapper, workerID int) {
	start := time.Now()
	event := eventWrapper.Event

	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.GetType()]
	bus.mutex.RUnlock()

	if !exists {
		bus.logger.WithField("event_type", event.GetType()).Debug("No handlers registered for event type")
		return
	}

	// 按优先级排序处理器
	sortedHandlers := bus.sortHandlersByPriority(event.GetType(), handlers)

	success := true
	for _, handler := range sortedHandlers {
		if err := bus.executeHandler(handler, event); err != nil {
			success = false
			eventWrapper.LastError = err
			eventWrapper.Attempts++

			bus.logger.WithFields(logrus.Fields{
				"event_type": event.GetType(),
				"worker_id":  workerID,
				"attempts":   eventWrapper.Attempts,
				"error":      err,
			}).Error("Event handler failed")

			// 检查是否需要重试
			if eventWrapper.Attempts < bus.config.RetryAttempts {
				bus.metrics.IncrementRetry()
				select {
				case bus.retryQueue <- eventWrapper:
					bus.logger.WithFields(logrus.Fields{
						"event_type": event.GetType(),
						"attempts":   eventWrapper.Attempts,
					}).Info("Event queued for retry")
				default:
					bus.logger.WithField("event_type", event.GetType()).Warn("Retry queue is full, dropping event")
					bus.metrics.IncrementFailed()
				}
				return
			} else {
				bus.metrics.IncrementFailed()
				bus.logger.WithFields(logrus.Fields{
					"event_type": event.GetType(),
					"attempts":   eventWrapper.Attempts,
				}).Error("Event processing failed after all retry attempts")
			}
		}
	}

	if success {
		now := time.Now()
		eventWrapper.ProcessedAt = &now
		bus.metrics.IncrementProcessed()
		bus.metrics.UpdateProcessTime(time.Since(start))

		bus.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"worker_id":  workerID,
			"duration":   time.Since(start),
		}).Debug("Event processed successfully")
	}
}

// executeHandler 执行单个处理器
func (bus *EnhancedEventBus) executeHandler(handler shared.EventHandler, event shared.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			bus.metrics.IncrementPanics()
			err = fmt.Errorf("handler panicked: %v", r)
			bus.logger.WithFields(logrus.Fields{
				"event_type": event.GetType(),
				"panic":      r,
			}).Error("Event handler panicked")
		}
	}()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(bus.ctx, bus.config.HandlerTimeout)
	defer cancel()

	// 在协程中执行处理器
	done := make(chan error, 1)
	go func() {
		done <- handler.Handle(event)
	}()

	select {
	case err = <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("handler timeout after %v", bus.config.HandlerTimeout)
	}
}

// sortHandlersByPriority 按优先级排序处理器
func (bus *EnhancedEventBus) sortHandlersByPriority(eventType string, handlers []shared.EventHandler) []shared.EventHandler {
	// 简单实现：返回原始顺序，可以根据需要实现优先级排序
	return handlers
}

// reportMetrics 报告指标
func (bus *EnhancedEventBus) reportMetrics() {
	defer bus.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case <-ticker.C:
			metrics := bus.metrics.GetMetrics()
			bus.logger.WithFields(logrus.Fields{
				"published_events":     metrics.PublishedEvents,
				"processed_events":     metrics.ProcessedEvents,
				"failed_events":        metrics.FailedEvents,
				"retry_events":         metrics.RetryEvents,
				"handler_panics":       metrics.HandlerPanics,
				"average_process_time": metrics.AverageProcessTime,
			}).Info("Event bus metrics")
		}
	}
}

// Publish 发布事件
func (bus *EnhancedEventBus) Publish(event shared.Event) error {
	eventWrapper := &EventWrapper{
		Event:     event,
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	select {
	case bus.eventQueue <- eventWrapper:
		bus.metrics.IncrementPublished()
		bus.logger.WithField("event_type", event.GetType()).Debug("Event published")
		return nil
	case <-bus.ctx.Done():
		return fmt.Errorf("event bus is shutting down")
	default:
		return fmt.Errorf("event queue is full")
	}
}

// Subscribe 订阅事件
func (bus *EnhancedEventBus) Subscribe(eventType string, handler shared.EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.handlers[eventType] == nil {
		bus.handlers[eventType] = make([]shared.EventHandler, 0)
	}

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)

	bus.logger.WithFields(logrus.Fields{
		"event_type": eventType,
		"handlers":   len(bus.handlers[eventType]),
	}).Debug("Event handler subscribed")

	return nil
}

// SubscribeWithMetadata 订阅事件并设置元数据
func (bus *EnhancedEventBus) SubscribeWithMetadata(eventType string, handler shared.EventHandler, metadata *HandlerMetadata) error {
	if err := bus.Subscribe(eventType, handler); err != nil {
		return err
	}

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// 生成处理器键
	handlerKey := fmt.Sprintf("%s:%p", eventType, handler)
	bus.handlerMeta[handlerKey] = metadata

	return nil
}

// Unsubscribe 取消订阅事件
func (bus *EnhancedEventBus) Unsubscribe(eventType string, handler shared.EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers, exists := bus.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type: %s", eventType)
	}

	// 查找并移除处理器
	for i, h := range handlers {
		if h == handler {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)

			// 移除元数据
			handlerKey := fmt.Sprintf("%s:%p", eventType, handler)
			delete(bus.handlerMeta, handlerKey)

			bus.logger.WithField("event_type", eventType).Debug("Event handler unsubscribed")
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// GetMetrics 获取事件总线指标
func (bus *EnhancedEventBus) GetMetrics() EventMetrics {
	return bus.metrics.GetMetrics()
}

// GetHandlerCount 获取处理器数量
func (bus *EnhancedEventBus) GetHandlerCount(eventType string) int {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	if handlers, exists := bus.handlers[eventType]; exists {
		return len(handlers)
	}
	return 0
}

// ListEventTypes 列出所有注册的事件类型
func (bus *EnhancedEventBus) ListEventTypes() []string {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	eventTypes := make([]string, 0, len(bus.handlers))
	for eventType := range bus.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// Shutdown 关闭事件总线
func (bus *EnhancedEventBus) Shutdown() {
	bus.logger.Info("Shutting down enhanced event bus")

	bus.cancel()

	// 关闭队列
	close(bus.eventQueue)
	close(bus.retryQueue)

	// 等待所有工作协程完成
	bus.wg.Wait()

	bus.logger.Info("Enhanced event bus shutdown completed")
}
