package shared

import (
	"time"
)

// Event 事件接口
type Event interface {
	GetType() string
	GetPayload() interface{}
	GetTimestamp() time.Time
}

// BaseEvent 基础事件
type BaseEvent struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// GetType 获取事件类型
func (e *BaseEvent) GetType() string {
	return e.Type
}

// GetPayload 获取事件载荷
func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

// GetTimestamp 获取事件时间戳
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// NewEvent 创建新事件
func NewEvent(eventType string, payload interface{}) Event {
	return &BaseEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// 事件类型常量
const (
	// 用户事件
	EventUserRegistered = "user.registered"
	EventUserLoggedIn   = "user.logged_in"
	EventUserUpdated    = "user.updated"

	// 比赛事件
	EventMatchCreated       = "match.created"
	EventMatchStarted       = "match.started"
	EventMatchFinished      = "match.finished"
	EventMatchUpdated       = "match.updated"
	EventMatchStatusChanged = "match.status_changed"
	EventMatchCancelled     = "match.cancelled"
	EventMatchScoreUpdated  = "match.score_updated"

	// 预测事件
	EventPredictionCreated = "prediction.created"
	EventPredictionUpdated = "prediction.updated"
	EventPredictionVoted   = "prediction.voted"
	EventPredictionUnvoted = "prediction.unvoted"

	// 积分事件
	EventPointsCalculated   = "points.calculated"
	EventLeaderboardUpdated = "leaderboard.updated"
)

// 事件载荷结构

// UserRegisteredPayload 用户注册事件载荷
type UserRegisteredPayload struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserLoggedInPayload 用户登录事件载荷
type UserLoggedInPayload struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// MatchCreatedPayload 比赛创建事件载荷
type MatchCreatedPayload struct {
	MatchID    uint      `json:"match_id"`
	TeamA      string    `json:"team_a"`
	TeamB      string    `json:"team_b"`
	Tournament string    `json:"tournament"`
	StartTime  time.Time `json:"start_time"`
}

// MatchStartedPayload 比赛开始事件载荷
type MatchStartedPayload struct {
	MatchID   uint      `json:"match_id"`
	StartTime time.Time `json:"start_time"`
}

// MatchFinishedPayload 比赛结束事件载荷
type MatchFinishedPayload struct {
	MatchID uint   `json:"match_id"`
	Winner  string `json:"winner"`
	ScoreA  int    `json:"score_a"`
	ScoreB  int    `json:"score_b"`
}

// MatchStatusChangedPayload 比赛状态变更事件载荷
type MatchStatusChangedPayload struct {
	MatchID   uint   `json:"match_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

// MatchCancelledPayload 比赛取消事件载荷
type MatchCancelledPayload struct {
	MatchID uint   `json:"match_id"`
	Reason  string `json:"reason"`
}

// MatchScoreUpdatedPayload 比赛比分更新事件载荷
type MatchScoreUpdatedPayload struct {
	MatchID   uint `json:"match_id"`
	OldScoreA int  `json:"old_score_a"`
	OldScoreB int  `json:"old_score_b"`
	NewScoreA int  `json:"new_score_a"`
	NewScoreB int  `json:"new_score_b"`
}

// PredictionCreatedPayload 预测创建事件载荷
type PredictionCreatedPayload struct {
	PredictionID    uint   `json:"prediction_id"`
	UserID          uint   `json:"user_id"`
	MatchID         uint   `json:"match_id"`
	PredictedWinner string `json:"predicted_winner"`
}

// PredictionVotedPayload 预测投票事件载荷
type PredictionVotedPayload struct {
	PredictionID uint `json:"prediction_id"`
	UserID       uint `json:"user_id"`
	VoteCount    int  `json:"vote_count"`
}

// PointsCalculatedPayload 积分计算事件载荷
type PointsCalculatedPayload struct {
	MatchID     uint                   `json:"match_id"`
	Predictions []PredictionPointsInfo `json:"predictions"`
}

// PredictionPointsInfo 预测积分信息
type PredictionPointsInfo struct {
	PredictionID uint `json:"prediction_id"`
	UserID       uint `json:"user_id"`
	Points       int  `json:"points"`
	IsCorrect    bool `json:"is_correct"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(event Event) error
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(event Event) error
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
}
