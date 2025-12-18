package handlers

import (
	"context"
	"fmt"
	"time"

	"backend-go/internal/core/domain/shared"
	"backend-go/pkg/redis"
	"github.com/sirupsen/logrus"
)

// Event payload types for statistics handler
type UserRegisteredPayload struct {
	UserID             uint   `json:"user_id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	RegistrationSource string `json:"registration_source"`
}

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

type MatchViewedPayload struct {
	MatchID      uint          `json:"match_id"`
	UserID       uint          `json:"user_id"`
	Tournament   string        `json:"tournament"`
	ViewDuration time.Duration `json:"view_duration"`
}

type LeaderboardViewedPayload struct {
	UserID     uint   `json:"user_id"`
	Tournament string `json:"tournament"`
	UserRank   int    `json:"user_rank"`
	UserPoints int    `json:"user_points"`
}

type PageViewedPayload struct {
	UserID   uint   `json:"user_id"`
	PagePath string `json:"page_path"`
	Referrer string `json:"referrer"`
}

type FeatureUsedPayload struct {
	UserID      uint          `json:"user_id"`
	FeatureName string        `json:"feature_name"`
	Action      string        `json:"action"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
}

type ErrorEncounteredPayload struct {
	UserID       uint   `json:"user_id"`
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Severity     string `json:"severity"`
}



// StatisticsHandler 统计事件处理器
type StatisticsHandler struct {
	redisClient *redis.Client
	logger      *logrus.Logger
}

// NewStatisticsHandler 创建统计事件处理器
func NewStatisticsHandler(redisClient *redis.Client, logger *logrus.Logger) *StatisticsHandler {
	return &StatisticsHandler{
		redisClient: redisClient,
		logger:      logger,
	}
}

// Handle 处理事件
func (h *StatisticsHandler) Handle(event shared.Event) error {
	ctx := context.Background()

	switch event.GetType() {
	case "user.registered":
		return h.handleUserRegistered(ctx, event)
	case "user.logged_in":
		return h.handleUserLoggedIn(ctx, event)
	case "prediction.created":
		return h.handlePredictionCreated(ctx, event)
	case "vote.cast":
		return h.handleVoteCast(ctx, event)
	case "match.viewed":
		return h.handleMatchViewed(ctx, event)
	case "leaderboard.viewed":
		return h.handleLeaderboardViewed(ctx, event)
	case "page.viewed":
		return h.handlePageViewed(ctx, event)
	case "feature.used":
		return h.handleFeatureUsed(ctx, event)
	case "error.encountered":
		return h.handleErrorEncountered(ctx, event)
	default:
		h.logger.WithField("event_type", event.GetType()).Debug("Unhandled event type in statistics handler")
		return nil
	}
}

// handleUserRegistered 处理用户注册统计
func (h *StatisticsHandler) handleUserRegistered(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*UserRegisteredPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for user registered event")
	}

	// 更新每日注册统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:registrations:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily registration count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour) // 保留7�?
	// 更新每月注册统计
	month := time.Now().Format("2006-01")
	monthlyKey := fmt.Sprintf("stats:registrations:monthly:%s", month)
	if _, err := h.redisClient.Incr(ctx, monthlyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment monthly registration count")
	}
	h.redisClient.Expire(ctx, monthlyKey, 365*24*time.Hour) // 保留1�?
	// 按注册来源统计
	sourceKey := fmt.Sprintf("stats:registrations:source:%s", payload.RegistrationSource)
	if _, err := h.redisClient.Incr(ctx, sourceKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment registration source count")
	}

	// 更新总注册数
	totalKey := "stats:registrations:total"
	if _, err := h.redisClient.Incr(ctx, totalKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment total registration count")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": payload.UserID,
		"source":  payload.RegistrationSource,
	}).Info("User registration statistics updated")

	return nil
}

// handleUserLoggedIn 处理用户登录统计
func (h *StatisticsHandler) handleUserLoggedIn(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*UserLoggedInPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for user logged in event")
	}

	// 更新每日活跃用户
	today := time.Now().Format("2006-01-02")
	dauKey := fmt.Sprintf("stats:dau:%s", today)
	if err := h.redisClient.SAdd(ctx, dauKey, payload.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to add user to DAU set")
	}
	h.redisClient.Expire(ctx, dauKey, 7*24*time.Hour)

	// 更新每月活跃用户
	month := time.Now().Format("2006-01")
	mauKey := fmt.Sprintf("stats:mau:%s", month)
	if err := h.redisClient.SAdd(ctx, mauKey, payload.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to add user to MAU set")
	}
	h.redisClient.Expire(ctx, mauKey, 365*24*time.Hour)

	// 按登录方式统计
	methodKey := fmt.Sprintf("stats:logins:method:%s", payload.LoginMethod)
	if _, err := h.redisClient.Incr(ctx, methodKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment login method count")
	}

	// 按登录来源统计
	sourceKey := fmt.Sprintf("stats:logins:source:%s", payload.LoginSource)
	if _, err := h.redisClient.Incr(ctx, sourceKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment login source count")
	}

	// 更新用户登录次数
	userLoginKey := fmt.Sprintf("stats:user:logins:%d", payload.UserID)
	if err := h.redisClient.Set(ctx, userLoginKey, payload.LoginCount, 0); err != nil {
		h.logger.WithError(err).Error("Failed to update user login count")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      payload.UserID,
		"login_method": payload.LoginMethod,
		"login_source": payload.LoginSource,
		"login_count":  payload.LoginCount,
	}).Info("User login statistics updated")

	return nil
}

// handlePredictionCreated 处理预测创建统计
func (h *StatisticsHandler) handlePredictionCreated(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*PredictionCreatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for prediction created event")
	}

	// 更新每日预测统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:predictions:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily prediction count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	// 按锦标赛统计
	tournamentKey := fmt.Sprintf("stats:predictions:tournament:%s", payload.Tournament)
	if _, err := h.redisClient.Incr(ctx, tournamentKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment tournament prediction count")
	}

	// 更新用户预测统计
	userPredKey := fmt.Sprintf("stats:user:predictions:%d", payload.UserID)
	if _, err := h.redisClient.Incr(ctx, userPredKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment user prediction count")
	}

	// 按预测时间距离比赛开始时间统计
	timeCategory := h.categorizeTimeToMatch(payload.TimeToMatchStart)
	timingKey := fmt.Sprintf("stats:predictions:timing:%s", timeCategory)
	if _, err := h.redisClient.Incr(ctx, timingKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment prediction timing count")
	}

	h.logger.WithFields(logrus.Fields{
		"prediction_id":   payload.PredictionID,
		"user_id":         payload.UserID,
		"match_id":        payload.MatchID,
		"tournament":      payload.Tournament,
		"time_to_match":   payload.TimeToMatchStart,
		"timing_category": timeCategory,
	}).Info("Prediction creation statistics updated")

	return nil
}

// handleVoteCast 处理投票统计
func (h *StatisticsHandler) handleVoteCast(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*VoteCastPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for vote cast event")
	}

	// 更新每日投票统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:votes:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily vote count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	// 更新用户投票统计
	userVoteKey := fmt.Sprintf("stats:user:votes:%d", payload.VoterID)
	if _, err := h.redisClient.Incr(ctx, userVoteKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment user vote count")
	}

	// 更新预测获得投票统计
	predVoteKey := fmt.Sprintf("stats:prediction:votes:%d", payload.PredictionID)
	if err := h.redisClient.Set(ctx, predVoteKey, payload.NewVoteCount, 0); err != nil {
		h.logger.WithError(err).Error("Failed to update prediction vote count")
	}

	h.logger.WithFields(logrus.Fields{
		"vote_id":        payload.VoteID,
		"voter_id":       payload.VoterID,
		"prediction_id":  payload.PredictionID,
		"new_vote_count": payload.NewVoteCount,
	}).Info("Vote cast statistics updated")

	return nil
}

// handleMatchViewed 处理比赛查看统计
func (h *StatisticsHandler) handleMatchViewed(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*MatchViewedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for match viewed event")
	}

	// 更新比赛查看次数
	matchViewKey := fmt.Sprintf("stats:match:views:%d", payload.MatchID)
	if _, err := h.redisClient.Incr(ctx, matchViewKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment match view count")
	}

	// 更新每日比赛查看统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:match_views:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily match view count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	// 按锦标赛统计
	tournamentKey := fmt.Sprintf("stats:match_views:tournament:%s", payload.Tournament)
	if _, err := h.redisClient.Incr(ctx, tournamentKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment tournament match view count")
	}

	h.logger.WithFields(logrus.Fields{
		"match_id":      payload.MatchID,
		"user_id":       payload.UserID,
		"tournament":    payload.Tournament,
		"view_duration": payload.ViewDuration,
	}).Info("Match view statistics updated")

	return nil
}

// handleLeaderboardViewed 处理排行榜查看统计
func (h *StatisticsHandler) handleLeaderboardViewed(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*LeaderboardViewedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for leaderboard viewed event")
	}

	// 更新每日排行榜查看统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:leaderboard_views:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily leaderboard view count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	// 按锦标赛统计
	tournamentKey := fmt.Sprintf("stats:leaderboard_views:tournament:%s", payload.Tournament)
	if _, err := h.redisClient.Incr(ctx, tournamentKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment tournament leaderboard view count")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     payload.UserID,
		"tournament":  payload.Tournament,
		"user_rank":   payload.UserRank,
		"user_points": payload.UserPoints,
	}).Info("Leaderboard view statistics updated")

	return nil
}

// handlePageViewed 处理页面访问统计
func (h *StatisticsHandler) handlePageViewed(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*PageViewedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for page viewed event")
	}

	// 更新页面访问统计
	pageKey := fmt.Sprintf("stats:page_views:%s", payload.PagePath)
	if _, err := h.redisClient.Incr(ctx, pageKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment page view count")
	}

	// 更新每日页面访问统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:page_views:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily page view count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	h.logger.WithFields(logrus.Fields{
		"user_id":   payload.UserID,
		"page_path": payload.PagePath,
		"referrer":  payload.Referrer,
	}).Debug("Page view statistics updated")

	return nil
}

// handleFeatureUsed 处理功能使用统计
func (h *StatisticsHandler) handleFeatureUsed(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*FeatureUsedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for feature used event")
	}

	// 更新功能使用统计
	featureKey := fmt.Sprintf("stats:feature_usage:%s:%s", payload.FeatureName, payload.Action)
	if _, err := h.redisClient.Incr(ctx, featureKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment feature usage count")
	}

	// 按成�?失败统计
	statusKey := fmt.Sprintf("stats:feature_usage:%s:%s:%t", payload.FeatureName, payload.Action, payload.Success)
	if _, err := h.redisClient.Incr(ctx, statusKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment feature usage status count")
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":      payload.UserID,
		"feature_name": payload.FeatureName,
		"action":       payload.Action,
		"success":      payload.Success,
		"duration":     payload.Duration,
	}).Info("Feature usage statistics updated")

	return nil
}

// handleErrorEncountered 处理错误统计
func (h *StatisticsHandler) handleErrorEncountered(ctx context.Context, event shared.Event) error {
	payload, ok := event.GetPayload().(*ErrorEncounteredPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for error encountered event")
	}

	// 更新错误统计
	errorKey := fmt.Sprintf("stats:errors:%s:%s", payload.ErrorType, payload.ErrorCode)
	if _, err := h.redisClient.Incr(ctx, errorKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment error count")
	}

	// 按严重程度统计
	severityKey := fmt.Sprintf("stats:errors:severity:%s", payload.Severity)
	if _, err := h.redisClient.Incr(ctx, severityKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment error severity count")
	}

	// 更新每日错误统计
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("stats:errors:daily:%s", today)
	if _, err := h.redisClient.Incr(ctx, dailyKey); err != nil {
		h.logger.WithError(err).Error("Failed to increment daily error count")
	}
	h.redisClient.Expire(ctx, dailyKey, 7*24*time.Hour)

	h.logger.WithFields(logrus.Fields{
		"user_id":       payload.UserID,
		"error_type":    payload.ErrorType,
		"error_code":    payload.ErrorCode,
		"error_message": payload.ErrorMessage,
		"severity":      payload.Severity,
	}).Warn("Error encountered statistics updated")

	return nil
}

// categorizeTimeToMatch 将预测时间距离比赛开始时间分类
func (h *StatisticsHandler) categorizeTimeToMatch(duration time.Duration) string {
	hours := duration.Hours()

	if hours < 1 {
		return "last_hour"
	} else if hours < 24 {
		return "same_day"
	} else if hours < 72 {
		return "within_3_days"
	} else if hours < 168 {
		return "within_week"
	} else {
		return "more_than_week"
	}
}

// GetStatistics 获取统计数据
func (h *StatisticsHandler) GetStatistics(ctx context.Context, statType string, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	switch statType {
	case "registrations":
		return h.getRegistrationStats(ctx, period)
	case "logins":
		return h.getLoginStats(ctx, period)
	case "predictions":
		return h.getPredictionStats(ctx, period)
	case "votes":
		return h.getVoteStats(ctx, period)
	case "errors":
		return h.getErrorStats(ctx, period)
	default:
		return stats, fmt.Errorf("unsupported statistics type: %s", statType)
	}
}

// getRegistrationStats 获取注册统计
func (h *StatisticsHandler) getRegistrationStats(ctx context.Context, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取总注册数
	total, err := h.redisClient.Get(ctx, "stats:registrations:total")
	if err == nil {
		stats["total"] = total
	}

	// 获取每日/每月统计
	if period == "daily" {
		today := time.Now().Format("2006-01-02")
		dailyKey := fmt.Sprintf("stats:registrations:daily:%s", today)
		daily, err := h.redisClient.Get(ctx, dailyKey)
		if err == nil {
			stats["today"] = daily
		}
	} else if period == "monthly" {
		month := time.Now().Format("2006-01")
		monthlyKey := fmt.Sprintf("stats:registrations:monthly:%s", month)
		monthly, err := h.redisClient.Get(ctx, monthlyKey)
		if err == nil {
			stats["this_month"] = monthly
		}
	}

	return stats, nil
}

// getLoginStats 获取登录统计
func (h *StatisticsHandler) getLoginStats(ctx context.Context, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if period == "daily" {
		today := time.Now().Format("2006-01-02")
		dauKey := fmt.Sprintf("stats:dau:%s", today)
		dau, err := h.redisClient.SCard(ctx, dauKey)
		if err == nil {
			stats["dau"] = dau
		}
	} else if period == "monthly" {
		month := time.Now().Format("2006-01")
		mauKey := fmt.Sprintf("stats:mau:%s", month)
		mau, err := h.redisClient.SCard(ctx, mauKey)
		if err == nil {
			stats["mau"] = mau
		}
	}

	return stats, nil
}

// getPredictionStats 获取预测统计
func (h *StatisticsHandler) getPredictionStats(ctx context.Context, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if period == "daily" {
		today := time.Now().Format("2006-01-02")
		dailyKey := fmt.Sprintf("stats:predictions:daily:%s", today)
		daily, err := h.redisClient.Get(ctx, dailyKey)
		if err == nil {
			stats["today"] = daily
		}
	}

	return stats, nil
}

// getVoteStats 获取投票统计
func (h *StatisticsHandler) getVoteStats(ctx context.Context, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if period == "daily" {
		today := time.Now().Format("2006-01-02")
		dailyKey := fmt.Sprintf("stats:votes:daily:%s", today)
		daily, err := h.redisClient.Get(ctx, dailyKey)
		if err == nil {
			stats["today"] = daily
		}
	}

	return stats, nil
}

// getErrorStats 获取错误统计
func (h *StatisticsHandler) getErrorStats(ctx context.Context, period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if period == "daily" {
		today := time.Now().Format("2006-01-02")
		dailyKey := fmt.Sprintf("stats:errors:daily:%s", today)
		daily, err := h.redisClient.Get(ctx, dailyKey)
		if err == nil {
			stats["today"] = daily
		}
	}

	return stats, nil
}
