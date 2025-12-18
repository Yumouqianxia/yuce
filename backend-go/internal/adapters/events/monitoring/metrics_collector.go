package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-go/internal/core/domain/shared"
	"backend-go/pkg/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// Event payload types to avoid import cycles
type UserLoggedInPayload struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	LoginMethod string `json:"login_method"`
	LoginSource string `json:"login_source"`
	LoginCount  int    `json:"login_count"`
}

type PredictionCreatedPayload struct {
	PredictionID     uint          `json:"prediction_id"`
	UserID           uint          `json:"user_id"`
	MatchID          uint          `json:"match_id"`
	Tournament       string        `json:"tournament"`
	TimeToMatchStart time.Duration `json:"time_to_match_start"`
}

type VoteCastPayload struct {
	VoteID       uint `json:"vote_id"`
	UserID       uint `json:"user_id"`
	PredictionID uint `json:"prediction_id"`
	VoterID      uint `json:"voter_id"`
	NewVoteCount int  `json:"new_vote_count"`
}

type ErrorEncounteredPayload struct {
	UserID       uint   `json:"user_id"`
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Severity     string `json:"severity"`
}

// Prometheus 指标定义
var (
	// 事件计数器
	eventCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_behavior_events_total",
			Help: "Total number of user behavior events",
		},
		[]string{"event_type", "status"},
	)

	// 事件处理延迟
	eventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds",
			Help:    "Time taken to process events",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type", "handler_type"},
	)

	// 用户活跃度指标
	activeUsersGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
		[]string{"period"}, // daily, weekly, monthly
	)

	// 预测相关指标
	predictionsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "predictions_count",
			Help: "Number of predictions",
		},
		[]string{"tournament", "status"},
	)

	// 投票相关指标
	votesGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "votes_count",
			Help: "Number of votes",
		},
		[]string{"period"},
	)

	// 错误率指标
	errorRateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "error_rate",
			Help: "Error rate by type",
		},
		[]string{"error_type", "severity"},
	)

	// 事件队列长度
	eventQueueLengthGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "event_queue_length",
			Help: "Current length of event queue",
		},
	)
)



// MetricsCollector 指标收集器
type MetricsCollector struct {
	redisClient *redis.Client
	logger      *logrus.Logger
	metrics     map[string]*EventMetrics
	mutex       sync.RWMutex
}

// EventMetrics 事件指标
type EventMetrics struct {
	Count             int64         `json:"count"`
	LastOccurred      time.Time     `json:"last_occurred"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	ErrorCount        int64         `json:"error_count"`
	ErrorRate         float64       `json:"error_rate"`
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(redisClient *redis.Client, logger *logrus.Logger) *MetricsCollector {
	collector := &MetricsCollector{
		redisClient: redisClient,
		logger:      logger,
		metrics:     make(map[string]*EventMetrics),
	}

	// 启动定期指标更新
	go collector.startPeriodicMetricsUpdate()

	return collector
}

// Handle 处理事件指标收集
func (c *MetricsCollector) Handle(event shared.Event) error {
	start := time.Now()

	// 更新事件计数器
	eventCounter.WithLabelValues(event.GetType(), "processed").Inc()

	// 记录处理时间
	processingTime := time.Since(start)
	eventProcessingDuration.WithLabelValues(event.GetType(), "metrics_collector").Observe(processingTime.Seconds())

	// 更新内存中的指标
	c.updateEventMetrics(event.GetType(), processingTime)

	// 根据事件类型更新特定指标
	switch event.GetType() {
	case "user.registered":
		c.handleUserRegisteredMetrics(event)
	case "user.logged_in":
		c.handleUserLoginMetrics(event)
	case "prediction.created":
		c.handlePredictionCreatedMetrics(event)
	case "vote.cast":
		c.handleVoteCastMetrics(event)
	case "error.encountered":
		c.handleErrorMetrics(event)
	}

	return nil
}

// updateEventMetrics 更新事件指标
func (c *MetricsCollector) updateEventMetrics(eventType string, processingTime time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics[eventType] == nil {
		c.metrics[eventType] = &EventMetrics{}
	}

	metrics := c.metrics[eventType]
	metrics.Count++
	metrics.LastOccurred = time.Now()

	// 计算平均处理时间
	if metrics.AvgProcessingTime == 0 {
		metrics.AvgProcessingTime = processingTime
	} else {
		metrics.AvgProcessingTime = (metrics.AvgProcessingTime + processingTime) / 2
	}
}

// handleUserRegisteredMetrics 处理用户注册指标
func (c *MetricsCollector) handleUserRegisteredMetrics(event shared.Event) {
	ctx := context.Background()

	// 更新每日注册数
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("metrics:registrations:daily:%s", today)
	count, _ := c.redisClient.Incr(ctx, dailyKey)
	c.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	// 更新 Prometheus 指标
	activeUsersGauge.WithLabelValues("registrations_today").Set(float64(count))

	c.logger.WithFields(logrus.Fields{
		"event_type":          event.GetType(),
		"daily_registrations": count,
	}).Debug("User registration metrics updated")
}

// handleUserLoginMetrics 处理用户登录指标
func (c *MetricsCollector) handleUserLoginMetrics(event shared.Event) {
	payload, ok := event.GetPayload().(*UserLoggedInPayload)
	if !ok {
		return
	}

	ctx := context.Background()

	// 更新每日活跃用户
	today := time.Now().Format("2006-01-02")
	dauKey := fmt.Sprintf("metrics:dau:%s", today)
	c.redisClient.SAdd(ctx, dauKey, payload.UserID)
	c.redisClient.Expire(ctx, dauKey, 7*24*time.Hour)

	// 获取 DAU 数量
	dauCount, _ := c.redisClient.SCard(ctx, dauKey)
	activeUsersGauge.WithLabelValues("daily").Set(float64(dauCount))

	// 更新每周活跃用户
	week := time.Now().Format("2006-W02")
	wauKey := fmt.Sprintf("metrics:wau:%s", week)
	c.redisClient.SAdd(ctx, wauKey, payload.UserID)
	c.redisClient.Expire(ctx, wauKey, 4*7*24*time.Hour)

	wauCount, _ := c.redisClient.SCard(ctx, wauKey)
	activeUsersGauge.WithLabelValues("weekly").Set(float64(wauCount))

	// 更新每月活跃用户
	month := time.Now().Format("2006-01")
	mauKey := fmt.Sprintf("metrics:mau:%s", month)
	c.redisClient.SAdd(ctx, mauKey, payload.UserID)
	c.redisClient.Expire(ctx, mauKey, 12*30*24*time.Hour)

	mauCount, _ := c.redisClient.SCard(ctx, mauKey)
	activeUsersGauge.WithLabelValues("monthly").Set(float64(mauCount))

	c.logger.WithFields(logrus.Fields{
		"event_type": event.GetType(),
		"user_id":    payload.UserID,
		"dau":        dauCount,
		"wau":        wauCount,
		"mau":        mauCount,
	}).Debug("User login metrics updated")
}

// handlePredictionCreatedMetrics 处理预测创建指标
func (c *MetricsCollector) handlePredictionCreatedMetrics(event shared.Event) {
	payload, ok := event.GetPayload().(*PredictionCreatedPayload)
	if !ok {
		return
	}

	ctx := context.Background()

	// 更新锦标赛预测数
	tournamentKey := fmt.Sprintf("metrics:predictions:tournament:%s", payload.Tournament)
	count, _ := c.redisClient.Incr(ctx, tournamentKey)

	predictionsGauge.WithLabelValues(payload.Tournament, "active").Set(float64(count))

	// 更新每日预测数
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("metrics:predictions:daily:%s", today)
	dailyCount, _ := c.redisClient.Incr(ctx, dailyKey)
	c.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	predictionsGauge.WithLabelValues("all", "daily").Set(float64(dailyCount))

	c.logger.WithFields(logrus.Fields{
		"event_type":             event.GetType(),
		"prediction_id":          payload.PredictionID,
		"tournament":             payload.Tournament,
		"tournament_predictions": count,
		"daily_predictions":      dailyCount,
	}).Debug("Prediction creation metrics updated")
}

// handleVoteCastMetrics 处理投票指标
func (c *MetricsCollector) handleVoteCastMetrics(event shared.Event) {
	payload, ok := event.GetPayload().(*VoteCastPayload)
	if !ok {
		return
	}

	ctx := context.Background()

	// 更新每日投票数
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("metrics:votes:daily:%s", today)
	count, _ := c.redisClient.Incr(ctx, dailyKey)
	c.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	votesGauge.WithLabelValues("daily").Set(float64(count))

	// 更新每小时投票数
	hour := time.Now().Format("2006-01-02-15")
	hourlyKey := fmt.Sprintf("metrics:votes:hourly:%s", hour)
	hourlyCount, _ := c.redisClient.Incr(ctx, hourlyKey)
	c.redisClient.Expire(ctx, hourlyKey, 24*time.Hour)

	votesGauge.WithLabelValues("hourly").Set(float64(hourlyCount))

	c.logger.WithFields(logrus.Fields{
		"event_type":    event.GetType(),
		"vote_id":       payload.VoteID,
		"prediction_id": payload.PredictionID,
		"daily_votes":   count,
		"hourly_votes":  hourlyCount,
	}).Debug("Vote cast metrics updated")
}

// handleErrorMetrics 处理错误指标
func (c *MetricsCollector) handleErrorMetrics(event shared.Event) {
	payload, ok := event.GetPayload().(*ErrorEncounteredPayload)
	if !ok {
		return
	}

	ctx := context.Background()

	// 更新错误计数
	errorKey := fmt.Sprintf("metrics:errors:%s:%s", payload.ErrorType, payload.Severity)
	count, _ := c.redisClient.Incr(ctx, errorKey)

	errorRateGauge.WithLabelValues(payload.ErrorType, payload.Severity).Set(float64(count))

	// 更新每日错误数
	today := time.Now().Format("2006-01-02")
	dailyErrorKey := fmt.Sprintf("metrics:errors:daily:%s", today)
	dailyCount, _ := c.redisClient.Incr(ctx, dailyErrorKey)
	c.redisClient.Expire(ctx, dailyErrorKey, 7*24*time.Hour)

	errorRateGauge.WithLabelValues("all", "daily").Set(float64(dailyCount))

	// 更新事件指标中的错误计数
	c.mutex.Lock()
	if c.metrics[event.GetType()] == nil {
		c.metrics[event.GetType()] = &EventMetrics{}
	}
	c.metrics[event.GetType()].ErrorCount++
	c.mutex.Unlock()

	c.logger.WithFields(logrus.Fields{
		"event_type":   event.GetType(),
		"error_type":   payload.ErrorType,
		"error_code":   payload.ErrorCode,
		"severity":     payload.Severity,
		"error_count":  count,
		"daily_errors": dailyCount,
	}).Warn("Error metrics updated")
}

// startPeriodicMetricsUpdate 启动定期指标更新
func (c *MetricsCollector) startPeriodicMetricsUpdate() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.updatePeriodicMetrics()
	}
}

// updatePeriodicMetrics 更新定期指标
func (c *MetricsCollector) updatePeriodicMetrics() {
	ctx := context.Background()

	// 更新事件队列长度（模拟）
	queueLength, _ := c.redisClient.LLen(ctx, "events:queue")
	eventQueueLengthGauge.Set(float64(queueLength))

	// 计算错误率
	c.calculateErrorRates()

	c.logger.Debug("Periodic metrics updated")
}

// calculateErrorRates 计算错误率
func (c *MetricsCollector) calculateErrorRates() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for eventType, metrics := range c.metrics {
		if metrics.Count > 0 {
			errorRate := float64(metrics.ErrorCount) / float64(metrics.Count) * 100
			metrics.ErrorRate = errorRate

			// 更新 Prometheus 指标
			errorRateGauge.WithLabelValues(eventType, "rate").Set(errorRate)
		}
	}
}

// GetMetrics 获取指标数据
func (c *MetricsCollector) GetMetrics() map[string]*EventMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 创建副本以避免并发访问问题
	result := make(map[string]*EventMetrics)
	for k, v := range c.metrics {
		result[k] = &EventMetrics{
			Count:             v.Count,
			LastOccurred:      v.LastOccurred,
			AvgProcessingTime: v.AvgProcessingTime,
			ErrorCount:        v.ErrorCount,
			ErrorRate:         v.ErrorRate,
		}
	}

	return result
}

// GetMetricsByType 获取指定类型的指标
func (c *MetricsCollector) GetMetricsByType(eventType string) *EventMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if metrics, exists := c.metrics[eventType]; exists {
		return &EventMetrics{
			Count:             metrics.Count,
			LastOccurred:      metrics.LastOccurred,
			AvgProcessingTime: metrics.AvgProcessingTime,
			ErrorCount:        metrics.ErrorCount,
			ErrorRate:         metrics.ErrorRate,
		}
	}

	return nil
}

// ResetMetrics 重置指标
func (c *MetricsCollector) ResetMetrics() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.metrics = make(map[string]*EventMetrics)
	c.logger.Info("Metrics reset")
}

// GetSystemMetrics 获取系统级指标
func (c *MetricsCollector) GetSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 获取今日活跃用户数
	today := time.Now().Format("2006-01-02")
	dauKey := fmt.Sprintf("metrics:dau:%s", today)
	dau, _ := c.redisClient.SCard(ctx, dauKey)
	metrics["dau"] = dau

	// 获取今日注册数
	regKey := fmt.Sprintf("metrics:registrations:daily:%s", today)
	registrations, _ := c.redisClient.Get(ctx, regKey)
	metrics["daily_registrations"] = registrations

	// 获取今日预测数
	predKey := fmt.Sprintf("metrics:predictions:daily:%s", today)
	predictions, _ := c.redisClient.Get(ctx, predKey)
	metrics["daily_predictions"] = predictions

	// 获取今日投票数
	voteKey := fmt.Sprintf("metrics:votes:daily:%s", today)
	votes, _ := c.redisClient.Get(ctx, voteKey)
	metrics["daily_votes"] = votes

	// 获取今日错误数
	errorKey := fmt.Sprintf("metrics:errors:daily:%s", today)
	errors, _ := c.redisClient.Get(ctx, errorKey)
	metrics["daily_errors"] = errors

	// 获取事件队列长度
	queueLength, _ := c.redisClient.LLen(ctx, "events:queue")
	metrics["event_queue_length"] = queueLength

	return metrics, nil
}
