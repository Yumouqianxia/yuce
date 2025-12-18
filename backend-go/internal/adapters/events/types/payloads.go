package types

import "time"

// Event payload types shared across different handlers
// These types are defined here to avoid duplication and import cycles

// UserRegisteredPayload represents the payload for user registration events
type UserRegisteredPayload struct {
	UserID             uint   `json:"user_id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	RegistrationSource string `json:"registration_source"`
}

// UserLoggedInPayload represents the payload for user login events
type UserLoggedInPayload struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	LoginMethod string `json:"login_method"`
	LoginSource string `json:"login_source"`
	LoginCount  int    `json:"login_count"`
}

// VoteCastPayload represents the payload for vote cast events
type VoteCastPayload struct {
	VoteID       uint `json:"vote_id"`
	UserID       uint `json:"user_id"`
	PredictionID uint `json:"prediction_id"`
	VoterID      uint `json:"voter_id"`
	NewVoteCount int  `json:"new_vote_count"`
}

// RankingChangedPayload represents the payload for ranking change events
type RankingChangedPayload struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Tournament string `json:"tournament"`
	OldRank    int    `json:"old_rank"`
	NewRank    int    `json:"new_rank"`
	RankChange int    `json:"rank_change"`
}

// MatchViewedPayload represents the payload for match viewed events
type MatchViewedPayload struct {
	MatchID      uint          `json:"match_id"`
	UserID       uint          `json:"user_id"`
	Tournament   string        `json:"tournament"`
	ViewDuration time.Duration `json:"view_duration"`
}

// PredictionCreatedPayload represents the payload for prediction created events
type PredictionCreatedPayload struct {
	PredictionID     uint          `json:"prediction_id"`
	UserID           uint          `json:"user_id"`
	MatchID          uint          `json:"match_id"`
	Tournament       string        `json:"tournament"`
	TimeToMatchStart time.Duration `json:"time_to_match_start"`
}

// LeaderboardViewedPayload represents the payload for leaderboard viewed events
type LeaderboardViewedPayload struct {
	UserID     uint   `json:"user_id"`
	Tournament string `json:"tournament"`
	UserRank   int    `json:"user_rank"`
	UserPoints int    `json:"user_points"`
}

// PageViewedPayload represents the payload for page viewed events
type PageViewedPayload struct {
	UserID   uint   `json:"user_id"`
	PagePath string `json:"page_path"`
	Referrer string `json:"referrer"`
}

// FeatureUsedPayload represents the payload for feature used events
type FeatureUsedPayload struct {
	UserID      uint          `json:"user_id"`
	FeatureName string        `json:"feature_name"`
	Action      string        `json:"action"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
}

// ErrorEncounteredPayload represents the payload for error encountered events
type ErrorEncounteredPayload struct {
	UserID       uint   `json:"user_id"`
	ErrorType    string `json:"error_type"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Severity     string `json:"severity"`
}