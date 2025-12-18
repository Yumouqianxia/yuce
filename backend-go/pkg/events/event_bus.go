package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// InMemoryEventBus 内存事件总线实现
type InMemoryEventBus struct {
	handlers map[string][]shared.EventHandler
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus(logger *logrus.Logger) shared.EventBus {
	if logger == nil {
		logger = logrus.New()
	}

	return &InMemoryEventBus{
		handlers: make(map[string][]shared.EventHandler),
		logger:   logger,
	}
}

// Publish 发布事件
func (bus *InMemoryEventBus) Publish(event shared.Event) error {
	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.GetType()]
	bus.mutex.RUnlock()

	if !exists {
		bus.logger.WithField("event_type", event.GetType()).Debug("No handlers registered for event type")
		return nil
	}

	// 异步处理事件
	go func() {
		for _, handler := range handlers {
			func(h shared.EventHandler) {
				defer func() {
					if r := recover(); r != nil {
						bus.logger.WithFields(logrus.Fields{
							"event_type": event.GetType(),
							"panic":      r,
						}).Error("Event handler panicked")
					}
				}()

				if err := h.Handle(event); err != nil {
					bus.logger.WithFields(logrus.Fields{
						"event_type": event.GetType(),
						"error":      err,
					}).Error("Event handler failed")
				}
			}(handler)
		}
	}()

	bus.logger.WithFields(logrus.Fields{
		"event_type": event.GetType(),
		"handlers":   len(handlers),
	}).Debug("Event published")

	return nil
}

// Subscribe 订阅事件
func (bus *InMemoryEventBus) Subscribe(eventType string, handler shared.EventHandler) error {
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

// Unsubscribe 取消订阅事件
func (bus *InMemoryEventBus) Unsubscribe(eventType string, handler shared.EventHandler) error {
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
			bus.logger.WithField("event_type", eventType).Debug("Event handler unsubscribed")
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// AsyncEventBus 异步事件总线（带缓冲队列）
type AsyncEventBus struct {
	eventQueue chan shared.Event
	handlers   map[string][]shared.EventHandler
	mutex      sync.RWMutex
	logger     *logrus.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewAsyncEventBus 创建异步事件总线
func NewAsyncEventBus(queueSize int, logger *logrus.Logger) *AsyncEventBus {
	if logger == nil {
		logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bus := &AsyncEventBus{
		eventQueue: make(chan shared.Event, queueSize),
		handlers:   make(map[string][]shared.EventHandler),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 启动事件处理协程
	bus.wg.Add(1)
	go bus.processEvents()

	return bus
}

// processEvents 处理事件队列
func (bus *AsyncEventBus) processEvents() {
	defer bus.wg.Done()

	for {
		select {
		case <-bus.ctx.Done():
			bus.logger.Info("Event bus shutting down")
			return
		case event := <-bus.eventQueue:
			bus.handleEvent(event)
		}
	}
}

// handleEvent 处理单个事件
func (bus *AsyncEventBus) handleEvent(event shared.Event) {
	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.GetType()]
	bus.mutex.RUnlock()

	if !exists {
		bus.logger.WithField("event_type", event.GetType()).Debug("No handlers registered for event type")
		return
	}

	// 并发处理所有处理器
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h shared.EventHandler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					bus.logger.WithFields(logrus.Fields{
						"event_type": event.GetType(),
						"panic":      r,
					}).Error("Event handler panicked")
				}
			}()

			start := time.Now()
			if err := h.Handle(event); err != nil {
				bus.logger.WithFields(logrus.Fields{
					"event_type": event.GetType(),
					"error":      err,
					"duration":   time.Since(start),
				}).Error("Event handler failed")
			} else {
				bus.logger.WithFields(logrus.Fields{
					"event_type": event.GetType(),
					"duration":   time.Since(start),
				}).Debug("Event handler completed")
			}
		}(handler)
	}

	wg.Wait()
}

// Publish 发布事件
func (bus *AsyncEventBus) Publish(event shared.Event) error {
	select {
	case bus.eventQueue <- event:
		bus.logger.WithField("event_type", event.GetType()).Debug("Event queued")
		return nil
	case <-bus.ctx.Done():
		return fmt.Errorf("event bus is shutting down")
	default:
		return fmt.Errorf("event queue is full")
	}
}

// Subscribe 订阅事件
func (bus *AsyncEventBus) Subscribe(eventType string, handler shared.EventHandler) error {
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

// Unsubscribe 取消订阅事件
func (bus *AsyncEventBus) Unsubscribe(eventType string, handler shared.EventHandler) error {
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
			bus.logger.WithField("event_type", eventType).Debug("Event handler unsubscribed")
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// Shutdown 关闭事件总线
func (bus *AsyncEventBus) Shutdown() {
	bus.cancel()
	close(bus.eventQueue)
	bus.wg.Wait()
	bus.logger.Info("Event bus shutdown completed")
}
