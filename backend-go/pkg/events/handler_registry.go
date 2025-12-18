package events

import (
	"fmt"
	"reflect"
	"sync"

	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// HandlerRegistry 事件处理器注册表
type HandlerRegistry struct {
	eventBus    shared.EventBus
	handlers    map[string][]RegisteredHandler
	handlerInfo map[string]*HandlerInfo
	mutex       sync.RWMutex
	logger      *logrus.Logger
}

// RegisteredHandler 注册的处理器
type RegisteredHandler struct {
	Handler  shared.EventHandler
	Metadata *HandlerMetadata
	Info     *HandlerInfo
}

// HandlerInfo 处理器信息
type HandlerInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Enabled     bool   `json:"enabled"`
}

// NewHandlerRegistry 创建处理器注册表
func NewHandlerRegistry(eventBus shared.EventBus, logger *logrus.Logger) *HandlerRegistry {
	if logger == nil {
		logger = logrus.New()
	}

	return &HandlerRegistry{
		eventBus:    eventBus,
		handlers:    make(map[string][]RegisteredHandler),
		handlerInfo: make(map[string]*HandlerInfo),
		logger:      logger,
	}
}

// RegisterHandler 注册事件处理器
func (r *HandlerRegistry) RegisterHandler(eventType string, handler shared.EventHandler, metadata *HandlerMetadata, info *HandlerInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 生成处理器ID
	handlerID := r.generateHandlerID(handler)

	// 检查处理器是否已注册
	if _, exists := r.handlerInfo[handlerID]; exists {
		return fmt.Errorf("handler already registered: %s", handlerID)
	}

	// 设置默认信息
	if info == nil {
		info = &HandlerInfo{
			Name:    handlerID,
			Type:    reflect.TypeOf(handler).String(),
			Enabled: true,
		}
	}

	// 设置默认元数据
	if metadata == nil {
		metadata = &HandlerMetadata{
			Name:     info.Name,
			Priority: 100, // 默认优先级
		}
	}

	// 创建注册的处理器
	registeredHandler := RegisteredHandler{
		Handler:  handler,
		Metadata: metadata,
		Info:     info,
	}

	// 添加到注册表
	if r.handlers[eventType] == nil {
		r.handlers[eventType] = make([]RegisteredHandler, 0)
	}
	r.handlers[eventType] = append(r.handlers[eventType], registeredHandler)
	r.handlerInfo[handlerID] = info

	// 订阅事件总线
	if info.Enabled {
		if err := r.eventBus.Subscribe(eventType, handler); err != nil {
			// 回滚注册
			r.handlers[eventType] = r.handlers[eventType][:len(r.handlers[eventType])-1]
			delete(r.handlerInfo, handlerID)
			return fmt.Errorf("failed to subscribe to event bus: %w", err)
		}
	}

	r.logger.WithFields(logrus.Fields{
		"event_type":   eventType,
		"handler_id":   handlerID,
		"handler_name": info.Name,
		"enabled":      info.Enabled,
	}).Info("Event handler registered")

	return nil
}

// UnregisterHandler 取消注册事件处理器
func (r *HandlerRegistry) UnregisterHandler(eventType string, handler shared.EventHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	handlerID := r.generateHandlerID(handler)

	// 查找并移除处理器
	handlers, exists := r.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type: %s", eventType)
	}

	for i, registeredHandler := range handlers {
		if registeredHandler.Handler == handler {
			// 从事件总线取消订阅
			if err := r.eventBus.Unsubscribe(eventType, handler); err != nil {
				r.logger.WithError(err).Warn("Failed to unsubscribe from event bus")
			}

			// 从注册表移除
			r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			delete(r.handlerInfo, handlerID)

			r.logger.WithFields(logrus.Fields{
				"event_type": eventType,
				"handler_id": handlerID,
			}).Info("Event handler unregistered")

			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// EnableHandler 启用处理器
func (r *HandlerRegistry) EnableHandler(handlerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	info, exists := r.handlerInfo[handlerID]
	if !exists {
		return fmt.Errorf("handler not found: %s", handlerID)
	}

	if info.Enabled {
		return nil // 已经启用
	}

	// 查找处理器并订阅事件总线
	for eventType, handlers := range r.handlers {
		for _, registeredHandler := range handlers {
			if r.generateHandlerID(registeredHandler.Handler) == handlerID {
				if err := r.eventBus.Subscribe(eventType, registeredHandler.Handler); err != nil {
					return fmt.Errorf("failed to subscribe to event bus: %w", err)
				}
				info.Enabled = true
				r.logger.WithFields(logrus.Fields{
					"handler_id": handlerID,
					"event_type": eventType,
				}).Info("Event handler enabled")
				return nil
			}
		}
	}

	return fmt.Errorf("handler implementation not found: %s", handlerID)
}

// DisableHandler 禁用处理器
func (r *HandlerRegistry) DisableHandler(handlerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	info, exists := r.handlerInfo[handlerID]
	if !exists {
		return fmt.Errorf("handler not found: %s", handlerID)
	}

	if !info.Enabled {
		return nil // 已经禁用
	}

	// 查找处理器并从事件总线取消订阅
	for eventType, handlers := range r.handlers {
		for _, registeredHandler := range handlers {
			if r.generateHandlerID(registeredHandler.Handler) == handlerID {
				if err := r.eventBus.Unsubscribe(eventType, registeredHandler.Handler); err != nil {
					r.logger.WithError(err).Warn("Failed to unsubscribe from event bus")
				}
				info.Enabled = false
				r.logger.WithFields(logrus.Fields{
					"handler_id": handlerID,
					"event_type": eventType,
				}).Info("Event handler disabled")
				return nil
			}
		}
	}

	return fmt.Errorf("handler implementation not found: %s", handlerID)
}

// GetHandlerInfo 获取处理器信息
func (r *HandlerRegistry) GetHandlerInfo(handlerID string) (*HandlerInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	info, exists := r.handlerInfo[handlerID]
	if !exists {
		return nil, fmt.Errorf("handler not found: %s", handlerID)
	}

	// 返回副本
	infoCopy := *info
	return &infoCopy, nil
}

// ListHandlers 列出所有处理器
func (r *HandlerRegistry) ListHandlers() map[string]*HandlerInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]*HandlerInfo)
	for id, info := range r.handlerInfo {
		infoCopy := *info
		result[id] = &infoCopy
	}
	return result
}

// ListHandlersByEventType 按事件类型列出处理器
func (r *HandlerRegistry) ListHandlersByEventType(eventType string) []RegisteredHandler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	handlers, exists := r.handlers[eventType]
	if !exists {
		return nil
	}

	// 返回副本
	result := make([]RegisteredHandler, len(handlers))
	copy(result, handlers)
	return result
}

// GetHandlerCount 获取处理器数量
func (r *HandlerRegistry) GetHandlerCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.handlerInfo)
}

// GetEventTypeCount 获取事件类型数量
func (r *HandlerRegistry) GetEventTypeCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.handlers)
}

// ListEventTypes 列出所有事件类型
func (r *HandlerRegistry) ListEventTypes() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	eventTypes := make([]string, 0, len(r.handlers))
	for eventType := range r.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// generateHandlerID 生成处理器ID
func (r *HandlerRegistry) generateHandlerID(handler shared.EventHandler) string {
	return fmt.Sprintf("%s_%p", reflect.TypeOf(handler).String(), handler)
}

// BatchRegisterHandlers 批量注册处理器
func (r *HandlerRegistry) BatchRegisterHandlers(registrations []HandlerRegistration) error {
	var errors []error

	for _, registration := range registrations {
		if err := r.RegisterHandler(
			registration.EventType,
			registration.Handler,
			registration.Metadata,
			registration.Info,
		); err != nil {
			errors = append(errors, fmt.Errorf("failed to register handler for %s: %w", registration.EventType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch registration failed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// HandlerRegistration 处理器注册信息
type HandlerRegistration struct {
	EventType string
	Handler   shared.EventHandler
	Metadata  *HandlerMetadata
	Info      *HandlerInfo
}

// ValidateHandlers 验证所有处理器
func (r *HandlerRegistry) ValidateHandlers() []ValidationError {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var errors []ValidationError

	for eventType, handlers := range r.handlers {
		for _, registeredHandler := range handlers {
			if registeredHandler.Handler == nil {
				errors = append(errors, ValidationError{
					EventType: eventType,
					HandlerID: r.generateHandlerID(registeredHandler.Handler),
					Error:     "handler is nil",
				})
			}

			if registeredHandler.Info == nil {
				errors = append(errors, ValidationError{
					EventType: eventType,
					HandlerID: r.generateHandlerID(registeredHandler.Handler),
					Error:     "handler info is nil",
				})
			}
		}
	}

	return errors
}

// ValidationError 验证错误
type ValidationError struct {
	EventType string `json:"event_type"`
	HandlerID string `json:"handler_id"`
	Error     string `json:"error"`
}

// GetRegistryStats 获取注册表统计信息
func (r *HandlerRegistry) GetRegistryStats() RegistryStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := RegistryStats{
		TotalHandlers:   len(r.handlerInfo),
		TotalEventTypes: len(r.handlers),
		EnabledHandlers: 0,
		EventTypeStats:  make(map[string]int),
	}

	for _, info := range r.handlerInfo {
		if info.Enabled {
			stats.EnabledHandlers++
		}
	}

	for eventType, handlers := range r.handlers {
		stats.EventTypeStats[eventType] = len(handlers)
	}

	return stats
}

// RegistryStats 注册表统计信息
type RegistryStats struct {
	TotalHandlers   int            `json:"total_handlers"`
	TotalEventTypes int            `json:"total_event_types"`
	EnabledHandlers int            `json:"enabled_handlers"`
	EventTypeStats  map[string]int `json:"event_type_stats"`
}
