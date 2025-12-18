package match

import (
	"backend-go/internal/core/domain"
	"time"
)

// Match 比赛实体 - 使用主域的 Match 类型
type Match = domain.Match

// MatchStatus 比赛状态枚举 - 使用主域的枚举
type MatchStatus = domain.MatchStatus

// Tournament 赛事枚举 - 使用主域的枚举
type Tournament = domain.Tournament

// Winner 获胜者枚举 - 使用主域的枚举
type Winner = domain.Winner

// 重新导出常量
const (
	MatchStatusUpcoming  = domain.MatchStatusUpcoming
	MatchStatusLive      = domain.MatchStatusLive
	MatchStatusFinished  = domain.MatchStatusFinished
	MatchStatusCancelled = domain.MatchStatusCancelled
)

const (
	TournamentSpring = domain.TournamentSpring
	TournamentSummer = domain.TournamentSummer
	TournamentWorlds = domain.TournamentWorlds
)

const (
	WinnerA    = domain.WinnerA
	WinnerB    = domain.WinnerB
	WinnerDraw = domain.WinnerDraw
)

// CreateMatchRequest 创建比赛请求
type CreateMatchRequest struct {
	TeamA      string     `json:"team_a" validate:"required,max=100"`
	TeamB      string     `json:"team_b" validate:"required,max=100"`
	Tournament Tournament `json:"tournament" validate:"required"`
	StartTime  time.Time  `json:"start_time" validate:"required"`
}

// UpdateMatchRequest 更新比赛请求
type UpdateMatchRequest struct {
	TeamA      string     `json:"team_a" validate:"max=100"`
	TeamB      string     `json:"team_b" validate:"max=100"`
	Tournament Tournament `json:"tournament"`
	StartTime  *time.Time `json:"start_time"`
}

// SetResultRequest 设置比赛结果请求
type SetResultRequest struct {
	ScoreA int    `json:"score_a" validate:"min=0"`
	ScoreB int    `json:"score_b" validate:"min=0"`
	Winner string `json:"winner" validate:"oneof=A B ''"`
}
