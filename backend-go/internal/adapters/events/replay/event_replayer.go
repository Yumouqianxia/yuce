package replay

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/adapters/events/persistence"
	"backend-go/internal/core/domain/shared"
	"github.com/sirupsen/logrus"
)

// EventReplayer 事件重放器
type EventReplayer struct {
	eventStore persistence.EventStore
	eventBus   shared.EventBus
	logger     *logrus.Logger
}

// NewEventReplayer 创建事件重放器
func NewEventReplayer(
	eventStore persistence.EventStore,
	eventBus shared.EventBus,
	logger *logrus.Logger,
) *EventReplayer {
	return &EventReplayer{
		eventStore: eventStore,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// ReplayOptions 重放选项
type ReplayOptions struct {
	EventType    string        `json:"event_type,omitempty"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	UserID       uint          `json:"user_id,omitempty"`
	BatchSize    int           `json:"batch_size"`
	DelayBetween time.Duration `json:"delay_between"`
	DryRun       bool          `json:"dry_run"`
}

// ReplayResult 重放结果
type ReplayResult struct {
	TotalEvents     int    `json:"total_events"`
	ProcessedEvents int    `json:"processed_events"`
	FailedEvents    int    `json:"failed_events"`
	Duration        int64  `json:"duration" swaggertype:"integer" example:"1500000000"` // Duration in nanoseconds
	Errors          []string `json:"errors,omitempty"`
}

// ReplayEvents 重放事件
func (r *EventReplayer) ReplayEvents(ctx context.Context, options ReplayOptions) (*ReplayResult, error) {
	start := time.Now()
	result := &ReplayResult{
		Errors: make([]string, 0),
	}

	r.logger.WithFields(logrus.Fields{
		"event_type": options.EventType,
		"start_time": options.StartTime,
		"end_time":   options.EndTime,
		"user_id":    options.UserID,
		"batch_size": options.BatchSize,
		"dry_run":    options.DryRun,
	}).Info("Starting event replay")

	// 获取要重放的事件
	events, err := r.getEventsForReplay(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for replay: %w", err)
	}

	result.TotalEvents = len(events)

	if options.DryRun {
		r.logger.WithField("total_events", result.TotalEvents).Info("Dry run completed")
		result.Duration = int64(time.Since(start))
		return result, nil
	}

	// 批量处理事件
	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}

		batch := events[i:end]
		if err := r.processBatch(ctx, batch, options, result); err != nil {
			r.logger.WithError(err).Error("Failed to process batch")
			result.Errors = append(result.Errors, err.Error())
		}

		// 批次间延迟
		if options.DelayBetween > 0 && end < len(events) {
			time.Sleep(options.DelayBetween)
		}

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			r.logger.Info("Event replay cancelled")
			result.Duration = int64(time.Since(start))
			return result, ctx.Err()
		default:
		}
	}

	result.Duration = int64(time.Since(start))

	r.logger.WithFields(logrus.Fields{
		"total_events":     result.TotalEvents,
		"processed_events": result.ProcessedEvents,
		"failed_events":    result.FailedEvents,
		"duration":         result.Duration,
		"error_count":      len(result.Errors),
	}).Info("Event replay completed")

	return result, nil
}

// getEventsForReplay 获取要重放的事件
func (r *EventReplayer) getEventsForReplay(ctx context.Context, options ReplayOptions) ([]shared.Event, error) {
	if options.UserID > 0 {
		// 获取特定用户的事件
		records, err := r.eventStore.GetEventsByUser(ctx, options.UserID, 0, 0)
		if err != nil {
			return nil, err
		}
		return r.convertRecordsToEvents(records, options), nil
	}

	if options.EventType != "" {
		// 获取特定类型的事件
		return r.eventStore.ReplayEvents(ctx, options.EventType, options.StartTime, options.EndTime)
	}

	// 获取时间范围内的所有事件
	return r.eventStore.ReplayEvents(ctx, "", options.StartTime, options.EndTime)
}

// convertRecordsToEvents 将事件记录转换为事件
func (r *EventReplayer) convertRecordsToEvents(records []persistence.EventRecord, options ReplayOptions) []shared.Event {
	events := make([]shared.Event, 0)

	for _, record := range records {
		// 过滤时间范围
		if record.CreatedAt.Before(options.StartTime) || record.CreatedAt.After(options.EndTime) {
			continue
		}

		// 过滤事件类型
		if options.EventType != "" && record.EventType != options.EventType {
			continue
		}

		// 这里简化处理，实际应该根据事件类型反序列化载荷
		event := &shared.BaseEvent{
			Type:      record.EventType,
			Payload:   record.Payload, // 简化处理
			Timestamp: record.CreatedAt,
		}

		events = append(events, event)
	}

	return events
}

// processBatch 处理事件批次
func (r *EventReplayer) processBatch(ctx context.Context, events []shared.Event, options ReplayOptions, result *ReplayResult) error {
	for _, event := range events {
		if err := r.processEvent(ctx, event); err != nil {
			result.FailedEvents++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to process event %s: %v", event.GetType(), err))
			r.logger.WithFields(logrus.Fields{
				"event_type": event.GetType(),
				"error":      err,
			}).Error("Failed to replay event")
		} else {
			result.ProcessedEvents++
		}

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// processEvent 处理单个事件
func (r *EventReplayer) processEvent(ctx context.Context, event shared.Event) error {
	// 发布事件到事件总线
	if err := r.eventBus.Publish(event); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"event_type": event.GetType(),
		"timestamp":  event.GetTimestamp(),
	}).Debug("Event replayed successfully")

	return nil
}

// ReplayFailedEvents 重放失败的事件
func (r *EventReplayer) ReplayFailedEvents(ctx context.Context, limit int) (*ReplayResult, error) {
	start := time.Now()
	result := &ReplayResult{
		Errors: make([]string, 0),
	}

	r.logger.WithField("limit", limit).Info("Starting failed events replay")

	// 获取失败的事件
	failedRecords, err := r.eventStore.GetFailedEvents(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed events: %w", err)
	}

	result.TotalEvents = len(failedRecords)

	for _, record := range failedRecords {
		// 这里简化处理，实际应该根据事件类型反序列化载荷
		event := &shared.BaseEvent{
			Type:      record.EventType,
			Payload:   record.Payload,
			Timestamp: record.CreatedAt,
		}

		if err := r.processEvent(ctx, event); err != nil {
			result.FailedEvents++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to replay event %s: %v", record.EventID, err))

			// 标记为失败
			if markErr := r.eventStore.MarkAsFailed(ctx, record.EventID, err.Error()); markErr != nil {
				r.logger.WithError(markErr).Error("Failed to mark event as failed")
			}
		} else {
			result.ProcessedEvents++

			// 标记为已处理
			if markErr := r.eventStore.MarkAsProcessed(ctx, record.EventID); markErr != nil {
				r.logger.WithError(markErr).Error("Failed to mark event as processed")
			}
		}

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			result.Duration = int64(time.Since(start))
			return result, ctx.Err()
		default:
		}
	}

	result.Duration = int64(time.Since(start))

	r.logger.WithFields(logrus.Fields{
		"total_events":     result.TotalEvents,
		"processed_events": result.ProcessedEvents,
		"failed_events":    result.FailedEvents,
		"duration":         result.Duration,
	}).Info("Failed events replay completed")

	return result, nil
}

// ReplayUserEvents 重放特定用户的事件
func (r *EventReplayer) ReplayUserEvents(ctx context.Context, userID uint, startTime, endTime time.Time) (*ReplayResult, error) {
	options := ReplayOptions{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		BatchSize: 50,
	}

	return r.ReplayEvents(ctx, options)
}

// ReplayEventsByType 重放特定类型的事件
func (r *EventReplayer) ReplayEventsByType(ctx context.Context, eventType string, startTime, endTime time.Time) (*ReplayResult, error) {
	options := ReplayOptions{
		EventType: eventType,
		StartTime: startTime,
		EndTime:   endTime,
		BatchSize: 100,
	}

	return r.ReplayEvents(ctx, options)
}

// ValidateReplay 验证重放操作
func (r *EventReplayer) ValidateReplay(ctx context.Context, options ReplayOptions) error {
	// 验证时间范围
	if options.EndTime.Before(options.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}

	// 验证时间范围不能太大（防止重放过多事件）
	maxDuration := 7 * 24 * time.Hour // 最多7天
	if options.EndTime.Sub(options.StartTime) > maxDuration {
		return fmt.Errorf("time range cannot exceed %v", maxDuration)
	}

	// 验证批次大小
	if options.BatchSize < 0 || options.BatchSize > 1000 {
		return fmt.Errorf("batch size must be between 1 and 1000")
	}

	// 估算事件数量
	events, err := r.getEventsForReplay(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to estimate event count: %w", err)
	}

	maxEvents := 10000 // 最多重放10000个事件
	if len(events) > maxEvents {
		return fmt.Errorf("too many events to replay (%d), maximum allowed is %d", len(events), maxEvents)
	}

	return nil
}

// GetReplayStatus 获取重放状态
func (r *EventReplayer) GetReplayStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// 获取失败事件数量
	failedEvents, err := r.eventStore.GetFailedEvents(ctx, 0)
	if err != nil {
		r.logger.WithError(err).Warn("Failed to get failed events count")
		status["failed_events_count"] = 0
	} else {
		status["failed_events_count"] = len(failedEvents)
	}

	// 获取最近的事件时间戳
	recentEvents, err := r.eventStore.GetEventsByTimeRange(ctx, time.Now().Add(-24*time.Hour), time.Now(), 1, 0)
	if err != nil {
		r.logger.WithError(err).Warn("Failed to get recent events")
	} else if len(recentEvents) > 0 {
		status["last_event_time"] = recentEvents[0].CreatedAt
	}

	status["replay_available"] = true
	status["max_replay_duration"] = "7 days"
	status["max_events_per_replay"] = 10000

	return status, nil
}
