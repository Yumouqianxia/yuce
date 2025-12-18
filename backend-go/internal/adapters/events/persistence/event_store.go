package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend-go/internal/core/domain/shared"
	"backend-go/pkg/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventRecord 事件记录模型
type EventRecord struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EventType    string     `gorm:"size:100;index" json:"event_type"`
	EventID      string     `gorm:"size:100;uniqueIndex" json:"event_id"`
	UserID       uint       `gorm:"index" json:"user_id,omitempty"`
	Payload      string     `gorm:"type:text" json:"payload"`
	Metadata     string     `gorm:"type:text" json:"metadata,omitempty"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	Status       string     `gorm:"size:20;default:pending" json:"status"` // pending, processed, failed
	RetryCount   int        `gorm:"default:0" json:"retry_count"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
}

// EventStore 事件存储接口
type EventStore interface {
	Store(ctx context.Context, event shared.Event) error
	GetEvents(ctx context.Context, eventType string, limit int, offset int) ([]EventRecord, error)
	GetEventsByUser(ctx context.Context, userID uint, limit int, offset int) ([]EventRecord, error)
	GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int, offset int) ([]EventRecord, error)
	MarkAsProcessed(ctx context.Context, eventID string) error
	MarkAsFailed(ctx context.Context, eventID string, errorMessage string) error
	GetFailedEvents(ctx context.Context, limit int) ([]EventRecord, error)
	ReplayEvents(ctx context.Context, eventType string, start, end time.Time) ([]shared.Event, error)
}

// MySQLEventStore MySQL 事件存储实现
type MySQLEventStore struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewMySQLEventStore 创建 MySQL 事件存储
func NewMySQLEventStore(db *gorm.DB, logger *logrus.Logger) EventStore {
	return &MySQLEventStore{
		db:     db,
		logger: logger,
	}
}

// Store 存储事件
func (s *MySQLEventStore) Store(ctx context.Context, event shared.Event) error {
	payloadBytes, err := json.Marshal(event.GetPayload())
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// 生成事件ID
	eventID := fmt.Sprintf("%s_%d_%d", event.GetType(), time.Now().UnixNano(), time.Now().Unix())

	// 提取用户ID（如果存在）
	var userID uint
	if userEvent, ok := event.(*shared.BaseEvent); ok {
		if payload, ok := userEvent.Payload.(map[string]interface{}); ok {
			if uid, exists := payload["user_id"]; exists {
				if uidFloat, ok := uid.(float64); ok {
					userID = uint(uidFloat)
				}
			}
		}
	}

	record := EventRecord{
		EventType: event.GetType(),
		EventID:   eventID,
		UserID:    userID,
		Payload:   string(payloadBytes),
		CreatedAt: event.GetTimestamp(),
		Status:    "pending",
	}

	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		s.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"event_id":   eventID,
			"error":      err,
		}).Error("Failed to store event")
		return fmt.Errorf("failed to store event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"event_type": event.GetType(),
		"event_id":   eventID,
		"user_id":    userID,
	}).Debug("Event stored successfully")

	return nil
}

// GetEvents 获取指定类型的事件
func (s *MySQLEventStore) GetEvents(ctx context.Context, eventType string, limit int, offset int) ([]EventRecord, error) {
	var records []EventRecord

	query := s.db.WithContext(ctx).Where("event_type = ?", eventType)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return records, nil
}

// GetEventsByUser 获取指定用户的事件
func (s *MySQLEventStore) GetEventsByUser(ctx context.Context, userID uint, limit int, offset int) ([]EventRecord, error) {
	var records []EventRecord

	query := s.db.WithContext(ctx).Where("user_id = ?", userID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}

	return records, nil
}

// GetEventsByTimeRange 获取指定时间范围的事件
func (s *MySQLEventStore) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int, offset int) ([]EventRecord, error) {
	var records []EventRecord

	query := s.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", start, end)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get events by time range: %w", err)
	}

	return records, nil
}

// MarkAsProcessed 标记事件为已处理
func (s *MySQLEventStore) MarkAsProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&EventRecord{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       "processed",
			"processed_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark event as processed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	return nil
}

// MarkAsFailed 标记事件为处理失败
func (s *MySQLEventStore) MarkAsFailed(ctx context.Context, eventID string, errorMessage string) error {
	result := s.db.WithContext(ctx).Model(&EventRecord{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errorMessage,
			"retry_count":   gorm.Expr("retry_count + 1"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark event as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	return nil
}

// GetFailedEvents 获取处理失败的事件
func (s *MySQLEventStore) GetFailedEvents(ctx context.Context, limit int) ([]EventRecord, error) {
	var records []EventRecord

	query := s.db.WithContext(ctx).Where("status = ? AND retry_count < ?", "failed", 3)
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed events: %w", err)
	}

	return records, nil
}

// ReplayEvents 重放事件
func (s *MySQLEventStore) ReplayEvents(ctx context.Context, eventType string, start, end time.Time) ([]shared.Event, error) {
	var records []EventRecord

	query := s.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", start, end)
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if err := query.Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get events for replay: %w", err)
	}

	events := make([]shared.Event, 0, len(records))
	for _, record := range records {
		var payload interface{}
		if err := json.Unmarshal([]byte(record.Payload), &payload); err != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id": record.EventID,
				"error":    err,
			}).Warn("Failed to unmarshal event payload for replay")
			continue
		}

		event := &shared.BaseEvent{
			Type:      record.EventType,
			Payload:   payload,
			Timestamp: record.CreatedAt,
		}

		events = append(events, event)
	}

	return events, nil
}

// RedisEventStore Redis 事件存储实现（用于缓存和快速访问）
type RedisEventStore struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisEventStore 创建 Redis 事件存储
func NewRedisEventStore(client *redis.Client, logger *logrus.Logger) *RedisEventStore {
	return &RedisEventStore{
		client: client,
		logger: logger,
	}
}

// StoreRecentEvent 存储最近的事件（用于快速访问）
func (s *RedisEventStore) StoreRecentEvent(ctx context.Context, event shared.Event) error {
	eventData := map[string]interface{}{
		"type":      event.GetType(),
		"payload":   event.GetPayload(),
		"timestamp": event.GetTimestamp().Unix(),
	}

	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 存储到最近事件列表
	recentKey := "events:recent"
	if err := s.client.LPush(ctx, recentKey, string(eventJSON)); err != nil {
		return fmt.Errorf("failed to store recent event: %w", err)
	}

	// 保持列表长度在1000以内
	if err := s.client.LTrim(ctx, recentKey, 0, 999); err != nil {
		s.logger.WithError(err).Warn("Failed to trim recent events list")
	}

	// 按事件类型存储
	typeKey := fmt.Sprintf("events:type:%s", event.GetType())
	if err := s.client.LPush(ctx, typeKey, string(eventJSON)); err != nil {
		s.logger.WithError(err).Warn("Failed to store event by type")
	}
	s.client.LTrim(ctx, typeKey, 0, 499) // 每种类型保持500个
	s.client.Expire(ctx, typeKey, 24*time.Hour)

	return nil
}

// GetRecentEvents 获取最近的事件
func (s *RedisEventStore) GetRecentEvents(ctx context.Context, limit int) ([]shared.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	eventStrings, err := s.client.LRange(ctx, "events:recent", 0, int64(limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	events := make([]shared.Event, 0, len(eventStrings))
	for _, eventStr := range eventStrings {
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(eventStr), &eventData); err != nil {
			s.logger.WithError(err).Warn("Failed to unmarshal event data")
			continue
		}

		timestamp := time.Unix(int64(eventData["timestamp"].(float64)), 0)
		event := &shared.BaseEvent{
			Type:      eventData["type"].(string),
			Payload:   eventData["payload"],
			Timestamp: timestamp,
		}

		events = append(events, event)
	}

	return events, nil
}

// GetRecentEventsByType 获取指定类型的最近事件
func (s *RedisEventStore) GetRecentEventsByType(ctx context.Context, eventType string, limit int) ([]shared.Event, error) {
	if limit <= 0 {
		limit = 100
	}

	typeKey := fmt.Sprintf("events:type:%s", eventType)
	eventStrings, err := s.client.LRange(ctx, typeKey, 0, int64(limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get recent events by type: %w", err)
	}

	events := make([]shared.Event, 0, len(eventStrings))
	for _, eventStr := range eventStrings {
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(eventStr), &eventData); err != nil {
			s.logger.WithError(err).Warn("Failed to unmarshal event data")
			continue
		}

		timestamp := time.Unix(int64(eventData["timestamp"].(float64)), 0)
		event := &shared.BaseEvent{
			Type:      eventData["type"].(string),
			Payload:   eventData["payload"],
			Timestamp: timestamp,
		}

		events = append(events, event)
	}

	return events, nil
}

// PersistentEventHandler 持久化事件处理器
type PersistentEventHandler struct {
	mysqlStore EventStore
	redisStore *RedisEventStore
	logger     *logrus.Logger
}

// NewPersistentEventHandler 创建持久化事件处理器
func NewPersistentEventHandler(
	mysqlStore EventStore,
	redisStore *RedisEventStore,
	logger *logrus.Logger,
) *PersistentEventHandler {
	return &PersistentEventHandler{
		mysqlStore: mysqlStore,
		redisStore: redisStore,
		logger:     logger,
	}
}

// Handle 处理事件持久化
func (h *PersistentEventHandler) Handle(event shared.Event) error {
	ctx := context.Background()

	// 存储到 MySQL（持久化）
	if err := h.mysqlStore.Store(ctx, event); err != nil {
		h.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"error":      err,
		}).Error("Failed to store event to MySQL")
		return err
	}

	// 存储到 Redis（快速访问）
	if err := h.redisStore.StoreRecentEvent(ctx, event); err != nil {
		h.logger.WithFields(logrus.Fields{
			"event_type": event.GetType(),
			"error":      err,
		}).Warn("Failed to store event to Redis")
		// 不返回错误，因为 Redis 存储失败不应该影响主流程
	}

	return nil
}
