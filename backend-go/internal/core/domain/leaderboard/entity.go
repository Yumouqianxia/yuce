package leaderboard

import (
	"time"

	"backend-go/internal/core/domain/user"
)

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Points     int       `json:"points"`
	Rank       int       `json:"rank"`
	Tournament string    `json:"tournament"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LeaderboardStats 排行榜统计信息
type LeaderboardStats struct {
	TotalUsers   int       `json:"total_users"`
	TopScore     int       `json:"top_score"`
	AverageScore float64   `json:"average_score"`
	LastUpdated  time.Time `json:"last_updated"`
	Tournament   string    `json:"tournament"`
}

// UserRankInfo 用户排名信息
type UserRankInfo struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Points       int    `json:"points"`
	Rank         int    `json:"rank"`
	Tournament   string `json:"tournament"`
	RankChange   int    `json:"rank_change"`   // 排名变化（正数表示上升，负数表示下降）
	PointsChange int    `json:"points_change"` // 积分变化
}

// Tournament 锦标赛类型
type Tournament string

const (
	TournamentSpring Tournament = "SPRING"
	TournamentSummer Tournament = "SUMMER"
	TournamentGlobal Tournament = "GLOBAL"
)

// IsValidTournament 检查锦标赛类型是否有效
func IsValidTournament(tournament string) bool {
	switch Tournament(tournament) {
	case TournamentSpring, TournamentSummer, TournamentGlobal:
		return true
	default:
		return false
	}
}

// GetDisplayName 获取显示名称
func (e *LeaderboardEntry) GetDisplayName() string {
	if e.Nickname != "" {
		return e.Nickname
	}
	return e.Username
}

// ToUserLeaderboardEntry 转换为用户排行榜条目
func (e *LeaderboardEntry) ToUserLeaderboardEntry() *user.LeaderboardEntry {
	return &user.LeaderboardEntry{
		UserID:     e.UserID,
		Username:   e.Username,
		Nickname:   e.Nickname,
		Avatar:     e.Avatar,
		Points:     e.Points,
		Rank:       e.Rank,
		Tournament: e.Tournament,
	}
}

// FromUserLeaderboardEntry 从用户排行榜条目创建
func FromUserLeaderboardEntry(entry *user.LeaderboardEntry) *LeaderboardEntry {
	return &LeaderboardEntry{
		UserID:     entry.UserID,
		Username:   entry.Username,
		Nickname:   entry.Nickname,
		Avatar:     entry.Avatar,
		Points:     entry.Points,
		Rank:       entry.Rank,
		Tournament: entry.Tournament,
		UpdatedAt:  time.Now(),
	}
}
