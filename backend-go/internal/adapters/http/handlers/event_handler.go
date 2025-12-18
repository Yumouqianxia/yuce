package handlers

import (
	"net/http"
	"strconv"
	"time"

	"backend-go/internal/adapters/events"
	"backend-go/internal/adapters/events/replay"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// EventHandler 事件处理器
type EventHandler struct {
	eventManager *events.EventManager
	logger       *logrus.Logger
}

// NewEventHandler 创建事件处理器
func NewEventHandler(eventManager *events.EventManager, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		eventManager: eventManager,
		logger:       logger,
	}
}

// GetStatistics 获取统计数据
// @Summary 获取统计数据
// @Description 获取指定类型和时间段的统计数据
// @Tags events
// @Accept json
// @Produce json
// @Param type query string true "统计类型" Enums(registrations,logins,predictions,votes,errors)
// @Param period query string true "时间段" Enums(daily,weekly,monthly)
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/statistics [get]
func (h *EventHandler) GetStatistics(c *gin.Context) {
	statType := c.Query("type")
	period := c.Query("period")

	if statType == "" {
		response.Error(c, http.StatusBadRequest, "统计类型不能为空", "MISSING_STAT_TYPE")
		return
	}

	if period == "" {
		response.Error(c, http.StatusBadRequest, "时间段不能为空", "MISSING_PERIOD")
		return
	}

	stats, err := h.eventManager.GetStatistics(c.Request.Context(), statType, period)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get statistics")
		response.Error(c, http.StatusInternalServerError, "获取统计数据失败", "GET_STATISTICS_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "Statistics retrieved successfully", stats)
}

// GetMetrics 获取事件指标
// @Summary 获取事件指标
// @Description 获取事件处理指标数据
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/events/metrics [get]
func (h *EventHandler) GetMetrics(c *gin.Context) {
	metrics := h.eventManager.GetMetrics()
	response.Success(c, http.StatusOK, "Metrics retrieved successfully", metrics)
}

// GetSystemMetrics 获取系统指标
// @Summary 获取系统指标
// @Description 获取系统级别的指标数据
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/events/system-metrics [get]
func (h *EventHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := h.eventManager.GetSystemMetrics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get system metrics")
		response.Error(c, http.StatusInternalServerError, "获取系统指标失败", "GET_SYSTEM_METRICS_FAILED")
		return
	}

	response.Success(c, http.StatusOK, "System metrics retrieved successfully", metrics)
}

// ReplayEventsRequest 重放事件请求
type ReplayEventsRequest struct {
	EventType    string `json:"event_type,omitempty" example:"user.registered"`
	StartTime    string `json:"start_time" binding:"required" example:"2024-01-01T00:00:00Z"`
	EndTime      string `json:"end_time" binding:"required" example:"2024-01-02T00:00:00Z"`
	UserID       uint   `json:"user_id,omitempty" example:"1"`
	BatchSize    int    `json:"batch_size,omitempty" example:"100"`
	DelayBetween int    `json:"delay_between,omitempty" example:"1000"` // 毫秒
	DryRun       bool   `json:"dry_run,omitempty" example:"true"`
}

// ReplayEvents 重放事件
// @Summary 重放事件
// @Description 重放指定条件的历史事件
// @Tags events
// @Accept json
// @Produce json
// @Param request body ReplayEventsRequest true "重放请求"
// @Success 200 {object} response.Response{data=replay.ReplayResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/replay [post]
func (h *EventHandler) ReplayEvents(c *gin.Context) {
	var req ReplayEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数无效", err.Error())
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "开始时间格式无效", err.Error())
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "结束时间格式无效", err.Error())
		return
	}

	// 构建重放选项
	options := replay.ReplayOptions{
		EventType:    req.EventType,
		StartTime:    startTime,
		EndTime:      endTime,
		UserID:       req.UserID,
		BatchSize:    req.BatchSize,
		DelayBetween: time.Duration(req.DelayBetween) * time.Millisecond,
		DryRun:       req.DryRun,
	}

	// 执行重放
	result, err := h.eventManager.ReplayEvents(c.Request.Context(), options)
	if err != nil {
		h.logger.WithError(err).Error("Failed to replay events")
		response.Error(c, http.StatusInternalServerError, "重放事件失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Events replayed successfully", result)
}

// ReplayFailedEvents 重放失败的事件
// @Summary 重放失败的事件
// @Description 重放处理失败的事件
// @Tags events
// @Accept json
// @Produce json
// @Param limit query int false "限制数量" default(100)
// @Success 200 {object} response.Response{data=replay.ReplayResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/replay-failed [post]
func (h *EventHandler) ReplayFailedEvents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "限制数量无效", err.Error())
		return
	}

	if limit <= 0 || limit > 1000 {
		response.Error(c, http.StatusBadRequest, "限制数量必须在1-1000之间", "INVALID_LIMIT_RANGE")
		return
	}

	result, err := h.eventManager.ReplayFailedEvents(c.Request.Context(), limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to replay failed events")
		response.Error(c, http.StatusInternalServerError, "重放失败事件失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Failed events replayed successfully", result)
}

// GetReplayStatus 获取重放状态
// @Summary 获取重放状态
// @Description 获取事件重放系统的状态信息
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/v1/events/replay-status [get]
func (h *EventHandler) GetReplayStatus(c *gin.Context) {
	status, err := h.eventManager.GetReplayStatus(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get replay status")
		response.Error(c, http.StatusInternalServerError, "获取重放状态失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Replay status retrieved successfully", status)
}

// PublishTestEventRequest 发布测试事件请求
type PublishTestEventRequest struct {
	EventType string                 `json:"event_type" binding:"required" example:"user.registered"`
	UserID    uint                   `json:"user_id" binding:"required" example:"1"`
	Payload   map[string]interface{} `json:"payload" binding:"required"`
}

// PublishTestEvent 发布测试事件
// @Summary 发布测试事件
// @Description 发布一个测试事件用于调试和测试
// @Tags events
// @Accept json
// @Produce json
// @Param request body PublishTestEventRequest true "测试事件请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/events/test [post]
func (h *EventHandler) PublishTestEvent(c *gin.Context) {
	var req PublishTestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数无效", err.Error())
		return
	}

	// 根据事件类型发布相应的测试事件
	var err error
	switch req.EventType {
	case events.EventUserRegistered:
		err = h.publishTestUserRegistered(req.UserID, req.Payload)
	case events.EventUserLoggedIn:
		err = h.publishTestUserLoggedIn(req.UserID, req.Payload)
	case events.EventPredictionCreated:
		err = h.publishTestPredictionCreated(req.UserID, req.Payload)
	case events.EventVoteCast:
		err = h.publishTestVoteCast(req.UserID, req.Payload)
	case events.EventErrorEncountered:
		err = h.publishTestErrorEncountered(req.UserID, req.Payload)
	default:
		response.Error(c, http.StatusBadRequest, "不支持的事件类型", "UNSUPPORTED_EVENT_TYPE")
		return
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to publish test event")
		response.Error(c, http.StatusInternalServerError, "发布测试事件失败", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Test event published successfully", gin.H{"message": "测试事件发布成功"})
}

// publishTestUserRegistered 发布测试用户注册事件
func (h *EventHandler) publishTestUserRegistered(userID uint, payload map[string]interface{}) error {
	username := getStringFromPayload(payload, "username", "test_user")
	email := getStringFromPayload(payload, "email", "test@example.com")
	nickname := getStringFromPayload(payload, "nickname", "Test User")
	source := getStringFromPayload(payload, "source", "test")

	return h.eventManager.PublishUserRegistered(userID, username, email, nickname, source)
}

// publishTestUserLoggedIn 发布测试用户登录事件
func (h *EventHandler) publishTestUserLoggedIn(userID uint, payload map[string]interface{}) error {
	username := getStringFromPayload(payload, "username", "test_user")
	loginMethod := getStringFromPayload(payload, "login_method", "password")
	loginSource := getStringFromPayload(payload, "login_source", "test")
	loginCount := getIntFromPayload(payload, "login_count", 1)

	return h.eventManager.PublishUserLoggedIn(userID, username, loginMethod, loginSource, loginCount)
}

// publishTestPredictionCreated 发布测试预测创建事件
func (h *EventHandler) publishTestPredictionCreated(userID uint, payload map[string]interface{}) error {
	predictionID := uint(getIntFromPayload(payload, "prediction_id", 1))
	matchID := uint(getIntFromPayload(payload, "match_id", 1))
	predictedWinner := getStringFromPayload(payload, "predicted_winner", "A")
	scoreA := getIntFromPayload(payload, "predicted_score_a", 2)
	scoreB := getIntFromPayload(payload, "predicted_score_b", 1)
	tournament := getStringFromPayload(payload, "tournament", "SPRING")
	timeToStart := time.Duration(getIntFromPayload(payload, "time_to_start_hours", 2)) * time.Hour

	return h.eventManager.PublishPredictionCreated(predictionID, userID, matchID, predictedWinner, scoreA, scoreB, tournament, timeToStart)
}

// publishTestVoteCast 发布测试投票事件
func (h *EventHandler) publishTestVoteCast(userID uint, payload map[string]interface{}) error {
	voteID := uint(getIntFromPayload(payload, "vote_id", 1))
	predictionID := uint(getIntFromPayload(payload, "prediction_id", 1))
	matchID := uint(getIntFromPayload(payload, "match_id", 1))
	voterID := uint(getIntFromPayload(payload, "voter_id", 2))
	newVoteCount := getIntFromPayload(payload, "new_vote_count", 5)

	return h.eventManager.PublishVoteCast(voteID, userID, predictionID, matchID, voterID, newVoteCount)
}

// publishTestErrorEncountered 发布测试错误事件
func (h *EventHandler) publishTestErrorEncountered(userID uint, payload map[string]interface{}) error {
	errorType := getStringFromPayload(payload, "error_type", "test_error")
	errorCode := getStringFromPayload(payload, "error_code", "TEST_ERROR")
	errorMessage := getStringFromPayload(payload, "error_message", "This is a test error")
	severity := getStringFromPayload(payload, "severity", "low")

	context := make(map[string]interface{})
	if ctx, ok := payload["context"].(map[string]interface{}); ok {
		context = ctx
	}

	return h.eventManager.PublishErrorEncountered(userID, errorType, errorCode, errorMessage, severity, context)
}

// 辅助函数
func getStringFromPayload(payload map[string]interface{}, key, defaultValue string) string {
	if value, ok := payload[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntFromPayload(payload map[string]interface{}, key string, defaultValue int) int {
	if value, ok := payload[key].(float64); ok {
		return int(value)
	}
	if value, ok := payload[key].(int); ok {
		return value
	}
	return defaultValue
}
