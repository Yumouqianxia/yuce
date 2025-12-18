// Package domain contains the core business entities and domain logic.
//
// This package defines the fundamental business concepts and rules of the
// prediction system. It includes entities like Match, Prediction, User, and Vote,
// along with their associated business logic and validation rules.
//
// The domain layer is independent of external concerns like databases, web
// frameworks, or external services. It focuses purely on business logic and
// maintains the integrity of business rules.
package domain

import (
	"time"
)

// MatchStatus represents the current state of a match in the system.
//
// This enumeration defines all possible states a match can be in throughout
// its lifecycle, from initial creation to completion or cancellation.
type MatchStatus string

// Match status constants define the possible states of a match.
const (
	MatchStatusUpcoming  MatchStatus = "UPCOMING"  // Match is scheduled but not started
	MatchStatusLive      MatchStatus = "LIVE"      // Match is currently in progress
	MatchStatusFinished  MatchStatus = "FINISHED"  // Match has completed with results
	MatchStatusCancelled MatchStatus = "CANCELLED" // Match was cancelled before completion
)

// Tournament represents the type of tournament or competition.
//
// This enumeration categorizes matches into different tournament types,
// which may have different rules, scoring systems, or importance levels.
type Tournament string

// Tournament constants define the available tournament types.
const (
	TournamentSpring Tournament = "SPRING" // Spring season tournament
	TournamentSummer Tournament = "SUMMER" // Summer season tournament
	TournamentWorlds Tournament = "WORLDS" // World championship tournament
)

// Winner represents the possible outcomes of a match.
//
// This enumeration defines who won the match, supporting both team victories
// and draw scenarios where applicable.
type Winner string

// Winner constants define the possible match outcomes.
const (
	WinnerA    Winner = "A"    // Team A won the match
	WinnerB    Winner = "B"    // Team B won the match
	WinnerDraw Winner = "DRAW" // Match ended in a draw
)

// Match represents a sports match or game between two teams.
//
// This is the core entity for the prediction system, representing a competitive
// match between two teams. Users can make predictions about the outcome of
// matches, and the system tracks the actual results for scoring purposes.
//
// The Match entity includes all necessary information for match management:
// team information, scheduling, status tracking, and result recording.
//
// Business Rules:
//   - Predictions can only be made for upcoming matches
//   - Match results can only be set when the match is finished
//   - Team names must be non-empty and different from each other
//   - Start time must be in the future for new matches
//   - Scores must be non-negative integers
//
// State Transitions:
//
//	UPCOMING -> LIVE -> FINISHED
//	UPCOMING -> CANCELLED
//	LIVE -> CANCELLED (in exceptional cases)
type Match struct {
	ID          uint        `gorm:"primaryKey" json:"id"`                              // Unique identifier
	TeamA       string      `gorm:"column:team_a;size:100;not null" json:"optionA"`    // First team name (前端兼容字段名)
	TeamB       string      `gorm:"column:team_b;size:100;not null" json:"optionB"`    // Second team name (前端兼容字段名)
	Tournament  Tournament  `gorm:"column:tournament;size:50;default:SPRING" json:"-"` // Tournament type (internal only)
	SportTypeID *uint       `gorm:"column:sport_type_id;index" json:"sportTypeId"`     // 运动类型ID
	Status      MatchStatus `gorm:"column:status;default:UPCOMING" json:"-"`           // Current match status (internal only)
	StartTime   time.Time   `gorm:"column:start_time;not null" json:"matchTime"`       // Scheduled start time (前端兼容字段名)
	Winner      string      `gorm:"column:winner;size:10" json:"winner"`               // Winner identifier ('A', 'B', or empty)
	ScoreA      int         `gorm:"column:score_a;default:0" json:"scoreA"`            // Team A final score (前端兼容字段名)
	ScoreB      int         `gorm:"column:score_b;default:0" json:"scoreB"`            // Team B final score (前端兼容字段名)
	CreatedAt   time.Time   `gorm:"column:created_at" json:"createdAt"`                // Record creation timestamp (前端兼容字段名)
	UpdatedAt   time.Time   `gorm:"column:updated_at" json:"updatedAt"`                // Record last update timestamp (前端兼容字段名)

	// 添加前端需要的字段
	Title              string `gorm:"-" json:"title"`          // 比赛标题 (计算字段)
	Description        string `gorm:"-" json:"description"`    // 比赛描述 (计算字段)
	IsActive           bool   `gorm:"-" json:"isActive"`       // 是否活跃 (计算字段)
	MatchType          string `gorm:"-" json:"matchType"`      // 比赛类型 (计算字段)
	Series             string `gorm:"-" json:"series"`         // 比赛系列 (计算字段)
	Year               int    `gorm:"-" json:"year"`           // 年份 (计算字段)
	FrontendTournament string `gorm:"-" json:"tournamentType"` // 前端兼容的赛事类型 (计算字段)
	FrontendStatus     string `gorm:"-" json:"status"`         // 前端兼容的状态 (计算字段)

	// Associations
	Predictions []Prediction `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"predictions,omitempty"` // Associated predictions
}

// TableName returns the database table name for the Match entity.
//
// This method implements the GORM Tabler interface to specify a custom
// table name for the Match entity in the database.
func (Match) TableName() string {
	return "matches"
}

// IsUpcoming returns true if the match is in upcoming status.
//
// This method checks if the match is scheduled but has not yet started.
// Upcoming matches are eligible for predictions and can be modified.
func (m *Match) IsUpcoming() bool {
	return m.Status == MatchStatusUpcoming
}

// IsLive returns true if the match is currently in progress.
//
// Live matches are actively being played and cannot accept new predictions
// or modifications to existing predictions.
func (m *Match) IsLive() bool {
	return m.Status == MatchStatusLive
}

// IsFinished returns true if the match has completed.
//
// Finished matches have final results and are used for calculating
// prediction accuracy and user scores.
func (m *Match) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

// CanAcceptPredictions returns true if the match can accept new predictions.
//
// This method implements the business rule that predictions can only be made
// for upcoming matches that haven't started yet. It checks both the match
// status and the current time against the scheduled start time.
//
// Returns:
//   - true if the match is upcoming and the start time is in the future
//   - false if the match has started, finished, or been cancelled
func (m *Match) CanAcceptPredictions() bool {
	return m.Status == MatchStatusUpcoming && time.Now().Before(m.StartTime)
}

// CanPredict 检查是否可以预测（别名方法）
func (m *Match) CanPredict() bool {
	return m.CanAcceptPredictions()
}

// SetResult 设置比赛结果
func (m *Match) SetResult(scoreA, scoreB int, winner string) error {
	if winner != "" && winner != "A" && winner != "B" {
		return ErrInvalidWinner
	}

	m.ScoreA = scoreA
	m.ScoreB = scoreB
	m.Winner = winner
	m.Status = MatchStatusFinished

	return nil
}

// GetWinnerTeam 获取获胜队伍名称
func (m *Match) GetWinnerTeam() string {
	if m.Winner == "A" {
		return m.TeamA
	} else if m.Winner == "B" {
		return m.TeamB
	}
	return ""
}

// IsValidStatus 检查状态是否有效
func IsValidMatchStatus(status string) bool {
	validStatuses := []MatchStatus{
		MatchStatusUpcoming,
		MatchStatusLive,
		MatchStatusFinished,
		MatchStatusCancelled,
	}

	for _, s := range validStatuses {
		if string(s) == status {
			return true
		}
	}
	return false
}

// IsValidTournament 检查赛事是否有效
func IsValidTournament(tournament string) bool {
	validTournaments := []Tournament{
		TournamentSpring,
		TournamentSummer,
		TournamentWorlds,
	}

	for _, t := range validTournaments {
		if string(t) == tournament {
			return true
		}
	}
	return false
}

// FillComputedFields 填充计算字段以兼容前端
func (m *Match) FillComputedFields() {
	// 生成标题
	m.Title = m.TeamA + " vs " + m.TeamB

	// 生成描述
	m.Description = m.TeamA + " vs " + m.TeamB

	// 设置是否活跃 (未完成的比赛视为活跃)
	m.IsActive = m.Status != MatchStatusFinished && m.Status != MatchStatusCancelled

	// 设置比赛类型 (根据状态推断)
	if m.Status == MatchStatusFinished {
		m.MatchType = "regular"
	} else {
		m.MatchType = "regular"
	}

	// 设置系列
	m.Series = "BO3"

	// 设置年份
	m.Year = m.StartTime.Year()

	// 设置前端兼容的赛事类型
	switch m.Tournament {
	case TournamentSpring:
		m.FrontendTournament = "spring"
	case TournamentSummer:
		m.FrontendTournament = "summer"
	case TournamentWorlds:
		m.FrontendTournament = "annual"
	default:
		m.FrontendTournament = "summer" // 默认值
	}

	// 设置前端兼容的状态
	switch m.Status {
	case MatchStatusUpcoming:
		m.FrontendStatus = "not_started"
	case MatchStatusLive:
		m.FrontendStatus = "in_progress"
	case MatchStatusFinished:
		m.FrontendStatus = "completed"
	case MatchStatusCancelled:
		m.FrontendStatus = "cancelled"
	default:
		m.FrontendStatus = "not_started" // 默认值
	}
}

// MatchResponse 用于API响应的Match结构体，包含前端兼容的字段
type MatchResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	OptionA         string    `json:"optionA"`
	OptionB         string    `json:"optionB"`
	MatchTime       time.Time `json:"matchTime"`
	Status          string    `json:"status"` // 转换为前端格式
	MatchType       string    `json:"matchType"`
	Series          string    `json:"series"`
	Winner          string    `json:"winner"`
	ScoreA          int       `json:"scoreA"`
	ScoreB          int       `json:"scoreB"`
	IsActive        bool      `json:"isActive"`
	TournamentType  string    `json:"tournamentType"` // 转换为前端格式
	TournamentStage string    `json:"tournamentStage"`
	Year            int       `json:"year"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToResponse 将Match转换为前端兼容的响应格式
func (m *Match) ToResponse() MatchResponse {
	// 转换状态为前端格式
	var frontendStatus string
	switch m.Status {
	case MatchStatusUpcoming:
		frontendStatus = "not_started"
	case MatchStatusLive:
		frontendStatus = "in_progress"
	case MatchStatusFinished:
		frontendStatus = "completed"
	case MatchStatusCancelled:
		frontendStatus = "cancelled"
	default:
		frontendStatus = "not_started"
	}

	// 转换赛事类型为前端格式
	var frontendTournament string
	switch m.Tournament {
	case TournamentSpring:
		frontendTournament = "spring"
	case TournamentSummer:
		frontendTournament = "summer"
	case TournamentWorlds:
		frontendTournament = "annual"
	default:
		frontendTournament = "spring"
	}

	return MatchResponse{
		ID:              m.ID,
		Title:           m.TeamA + " vs " + m.TeamB,
		Description:     m.TeamA + " vs " + m.TeamB,
		OptionA:         m.TeamA,
		OptionB:         m.TeamB,
		MatchTime:       m.StartTime,
		Status:          frontendStatus,
		MatchType:       "regular",
		Series:          "BO3",
		Winner:          m.Winner,
		ScoreA:          m.ScoreA,
		ScoreB:          m.ScoreB,
		IsActive:        m.Status != MatchStatusFinished && m.Status != MatchStatusCancelled,
		TournamentType:  frontendTournament,
		TournamentStage: "regular",
		Year:            m.StartTime.Year(),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
