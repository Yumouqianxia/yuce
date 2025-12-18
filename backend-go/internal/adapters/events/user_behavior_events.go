package events

import (
	"time"

	"backend-go/internal/core/domain/shared"
)

// UserBehaviorEvent 用户行为事件基础结构
type UserBehaviorEvent struct {
	*shared.BaseEvent
	UserID    uint                   `json:"user_id"`
	SessionID string                 `json:"session_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserBehaviorEvent 创建用户行为事件
func NewUserBehaviorEvent(eventType string, userID uint, payload interface{}) *UserBehaviorEvent {
	return &UserBehaviorEvent{
		BaseEvent: &shared.BaseEvent{
			Type:      eventType,
			Payload:   payload,
			Timestamp: time.Now(),
		},
		UserID:   userID,
		Metadata: make(map[string]interface{}),
	}
}

// WithSession 添加会话信息
func (e *UserBehaviorEvent) WithSession(sessionID, ipAddress, userAgent string) *UserBehaviorEvent {
	e.SessionID = sessionID
	e.IPAddress = ipAddress
	e.UserAgent = userAgent
	return e
}

// WithMetadata 添加元数据
func (e *UserBehaviorEvent) WithMetadata(key string, value interface{}) *UserBehaviorEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// 用户行为事件类型常量
const (
	// 用户生命周期事件
	EventUserRegistered      = "user.registered"
	EventUserLoggedIn        = "user.logged_in"
	EventUserLoggedOut       = "user.logged_out"
	EventUserProfileUpdated  = "user.profile_updated"
	EventUserPasswordChanged = "user.password_changed"
	EventUserDeactivated     = "user.deactivated"

	// 预测行为事件
	EventPredictionCreated = "prediction.created"
	EventPredictionUpdated = "prediction.updated"
	EventPredictionDeleted = "prediction.deleted"
	EventPredictionViewed  = "prediction.viewed"

	// 投票行为事件
	EventVoteCast    = "vote.cast"
	EventVoteRemoved = "vote.removed"
	EventVoteChanged = "vote.changed"

	// 比赛互动事件
	EventMatchViewed       = "match.viewed"
	EventMatchSubscribed   = "match.subscribed"
	EventMatchUnsubscribed = "match.unsubscribed"

	// 排行榜事件
	EventLeaderboardViewed = "leaderboard.viewed"
	EventRankingChanged    = "ranking.changed"

	// 系统交互事件
	EventPageViewed       = "page.viewed"
	EventFeatureUsed      = "feature.used"
	EventErrorEncountered = "error.encountered"
	EventSearchPerformed  = "search.performed"
)

// 事件载荷结构

// UserLoggedOutPayload 用户登出事件载荷
type UserLoggedOutPayload struct {
	UserID          uint          `json:"user_id"`
	Username        string        `json:"username"`
	SessionDuration time.Duration `json:"session_duration"`
	LoggedOutAt     time.Time     `json:"logged_out_at"`
}

// UserProfileUpdatedPayload 用户资料更新事件载荷
type UserProfileUpdatedPayload struct {
	UserID        uint                   `json:"user_id"`
	Username      string                 `json:"username"`
	UpdatedFields []string               `json:"updated_fields"`
	OldValues     map[string]interface{} `json:"old_values"`
	NewValues     map[string]interface{} `json:"new_values"`
	UpdatedAt     time.Time              `json:"updated_at"`
}



// PredictionUpdatedPayload 预测更新事件载荷
type PredictionUpdatedPayload struct {
	PredictionID       uint          `json:"prediction_id"`
	UserID             uint          `json:"user_id"`
	MatchID            uint          `json:"match_id"`
	OldPredictedWinner string        `json:"old_predicted_winner"`
	NewPredictedWinner string        `json:"new_predicted_winner"`
	OldPredictedScoreA int           `json:"old_predicted_score_a"`
	NewPredictedScoreA int           `json:"new_predicted_score_a"`
	OldPredictedScoreB int           `json:"old_predicted_score_b"`
	NewPredictedScoreB int           `json:"new_predicted_score_b"`
	ModificationCount  int           `json:"modification_count"`
	TimeToMatchStart   time.Duration `json:"time_to_match_start"`
	UpdatedAt          time.Time     `json:"updated_at"`
}



// VoteRemovedPayload 取消投票事件载荷
type VoteRemovedPayload struct {
	UserID       uint      `json:"user_id"`
	PredictionID uint      `json:"prediction_id"`
	MatchID      uint      `json:"match_id"`
	VoterID      uint      `json:"voter_id"`
	NewVoteCount int       `json:"new_vote_count"`
	RemovedAt    time.Time `json:"removed_at"`
}













// SearchPerformedPayload 搜索执行事件载荷
type SearchPerformedPayload struct {
	UserID         uint          `json:"user_id"`
	SearchQuery    string        `json:"search_query"`
	SearchType     string        `json:"search_type"` // matches, users, predictions
	ResultCount    int           `json:"result_count"`
	SearchDuration time.Duration `json:"search_duration"`
	PerformedAt    time.Time     `json:"performed_at"`
}
