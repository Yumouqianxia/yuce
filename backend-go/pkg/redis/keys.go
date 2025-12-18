package redis

import (
	"fmt"
	"strings"
	"time"
)

// CacheKeyManager 缓存键管理器
type CacheKeyManager struct {
	prefix    string
	separator string
}

// NewCacheKeyManager 创建缓存键管理器
func NewCacheKeyManager(prefix string) *CacheKeyManager {
	return &CacheKeyManager{
		prefix:    prefix,
		separator: ":",
	}
}

// 预定义的缓存键前缀
const (
	// 用户相关
	KeyPrefixUser        = "user"
	KeyPrefixUserSession = "user:session"
	KeyPrefixUserProfile = "user:profile"
	KeyPrefixUserStats   = "user:stats"

	// 比赛相关
	KeyPrefixMatch      = "match"
	KeyPrefixMatchList  = "match:list"
	KeyPrefixMatchStats = "match:stats"

	// 预测相关
	KeyPrefixPrediction     = "prediction"
	KeyPrefixPredictionList = "prediction:list"
	KeyPrefixPredictionVote = "prediction:vote"

	// 排行榜相关
	KeyPrefixLeaderboard    = "leaderboard"
	KeyPrefixLeaderboardTop = "leaderboard:top"

	// 投票相关
	KeyPrefixVote      = "vote"
	KeyPrefixVoteCount = "vote:count"

	// 统计相关
	KeyPrefixStats       = "stats"
	KeyPrefixStatsDaily  = "stats:daily"
	KeyPrefixStatsWeekly = "stats:weekly"

	// 缓存相关
	KeyPrefixCache     = "cache"
	KeyPrefixCacheTemp = "cache:temp"

	// 锁相关
	KeyPrefixLock            = "lock"
	KeyPrefixLockDistributed = "lock:distributed"

	// 会话相关
	KeyPrefixSession     = "session"
	KeyPrefixSessionAuth = "session:auth"

	// 通知相关
	KeyPrefixNotification      = "notification"
	KeyPrefixNotificationQueue = "notification:queue"
)

// 预定义的过期时间
const (
	// 短期缓存 (5分钟)
	ExpirationShort = 5 * time.Minute

	// 中期缓存 (30分钟)
	ExpirationMedium = 30 * time.Minute

	// 长期缓存 (2小时)
	ExpirationLong = 2 * time.Hour

	// 日缓存 (24小时)
	ExpirationDaily = 24 * time.Hour

	// 周缓存 (7天)
	ExpirationWeekly = 7 * 24 * time.Hour

	// 会话缓存 (30天)
	ExpirationSession = 30 * 24 * time.Hour

	// 临时缓存 (1分钟)
	ExpirationTemp = 1 * time.Minute

	// 排行榜缓存 (5分钟，根据需求文档)
	ExpirationLeaderboard = 5 * time.Minute

	// 比赛数据缓存 (1分钟，根据需求文档)
	ExpirationMatchData = 1 * time.Minute
)

// 用户相关键生成

// UserKey 生成用户键
func (km *CacheKeyManager) UserKey(userID uint) string {
	return km.buildKey(KeyPrefixUser, fmt.Sprintf("%d", userID))
}

// UserSessionKey 生成用户会话键
func (km *CacheKeyManager) UserSessionKey(sessionID string) string {
	return km.buildKey(KeyPrefixUserSession, sessionID)
}

// UserProfileKey 生成用户资料键
func (km *CacheKeyManager) UserProfileKey(userID uint) string {
	return km.buildKey(KeyPrefixUserProfile, fmt.Sprintf("%d", userID))
}

// UserStatsKey 生成用户统计键
func (km *CacheKeyManager) UserStatsKey(userID uint, tournament string) string {
	return km.buildKey(KeyPrefixUserStats, fmt.Sprintf("%d", userID), tournament)
}

// 比赛相关键生成

// MatchKey 生成比赛键
func (km *CacheKeyManager) MatchKey(matchID uint) string {
	return km.buildKey(KeyPrefixMatch, fmt.Sprintf("%d", matchID))
}

// MatchListKey 生成比赛列表键
func (km *CacheKeyManager) MatchListKey(tournament string, status string) string {
	return km.buildKey(KeyPrefixMatchList, tournament, status)
}

// MatchStatsKey 生成比赛统计键
func (km *CacheKeyManager) MatchStatsKey(matchID uint) string {
	return km.buildKey(KeyPrefixMatchStats, fmt.Sprintf("%d", matchID))
}

// 预测相关键生成

// PredictionKey 生成预测键
func (km *CacheKeyManager) PredictionKey(predictionID uint) string {
	return km.buildKey(KeyPrefixPrediction, fmt.Sprintf("%d", predictionID))
}

// PredictionListKey 生成预测列表键
func (km *CacheKeyManager) PredictionListKey(matchID uint, sortBy string) string {
	return km.buildKey(KeyPrefixPredictionList, fmt.Sprintf("%d", matchID), sortBy)
}

// PredictionVoteKey 生成预测投票键
func (km *CacheKeyManager) PredictionVoteKey(predictionID uint) string {
	return km.buildKey(KeyPrefixPredictionVote, fmt.Sprintf("%d", predictionID))
}

// UserPredictionKey 生成用户预测键
func (km *CacheKeyManager) UserPredictionKey(userID uint, matchID uint) string {
	return km.buildKey(KeyPrefixPrediction, "user", fmt.Sprintf("%d", userID), fmt.Sprintf("%d", matchID))
}

// 排行榜相关键生成

// LeaderboardKey 生成排行榜键
func (km *CacheKeyManager) LeaderboardKey(tournament string) string {
	return km.buildKey(KeyPrefixLeaderboard, tournament)
}

// LeaderboardTopKey 生成排行榜前N名键
func (km *CacheKeyManager) LeaderboardTopKey(tournament string, limit int) string {
	return km.buildKey(KeyPrefixLeaderboardTop, tournament, fmt.Sprintf("%d", limit))
}

// 投票相关键生成

// VoteKey 生成投票键
func (km *CacheKeyManager) VoteKey(userID uint, predictionID uint) string {
	return km.buildKey(KeyPrefixVote, fmt.Sprintf("%d", userID), fmt.Sprintf("%d", predictionID))
}

// VoteCountKey 生成投票计数键
func (km *CacheKeyManager) VoteCountKey(predictionID uint) string {
	return km.buildKey(KeyPrefixVoteCount, fmt.Sprintf("%d", predictionID))
}

// 统计相关键生成

// StatsKey 生成统计键
func (km *CacheKeyManager) StatsKey(category string, date string) string {
	return km.buildKey(KeyPrefixStats, category, date)
}

// DailyStatsKey 生成日统计键
func (km *CacheKeyManager) DailyStatsKey(date string) string {
	return km.buildKey(KeyPrefixStatsDaily, date)
}

// WeeklyStatsKey 生成周统计键
func (km *CacheKeyManager) WeeklyStatsKey(week string) string {
	return km.buildKey(KeyPrefixStatsWeekly, week)
}

// 缓存相关键生成

// CacheKey 生成通用缓存键
func (km *CacheKeyManager) CacheKey(category string, identifier string) string {
	return km.buildKey(KeyPrefixCache, category, identifier)
}

// TempCacheKey 生成临时缓存键
func (km *CacheKeyManager) TempCacheKey(identifier string) string {
	return km.buildKey(KeyPrefixCacheTemp, identifier)
}

// 锁相关键生成

// LockKey 生成锁键
func (km *CacheKeyManager) LockKey(resource string) string {
	return km.buildKey(KeyPrefixLock, resource)
}

// DistributedLockKey 生成分布式锁键
func (km *CacheKeyManager) DistributedLockKey(resource string) string {
	return km.buildKey(KeyPrefixLockDistributed, resource)
}

// 会话相关键生成

// SessionKey 生成会话键
func (km *CacheKeyManager) SessionKey(sessionID string) string {
	return km.buildKey(KeyPrefixSession, sessionID)
}

// AuthSessionKey 生成认证会话键
func (km *CacheKeyManager) AuthSessionKey(userID uint) string {
	return km.buildKey(KeyPrefixSessionAuth, fmt.Sprintf("%d", userID))
}

// 通知相关键生成

// NotificationKey 生成通知键
func (km *CacheKeyManager) NotificationKey(userID uint) string {
	return km.buildKey(KeyPrefixNotification, fmt.Sprintf("%d", userID))
}

// NotificationQueueKey 生成通知队列键
func (km *CacheKeyManager) NotificationQueueKey(queueType string) string {
	return km.buildKey(KeyPrefixNotificationQueue, queueType)
}

// 辅助方法

// buildKey 构建缓存键
func (km *CacheKeyManager) buildKey(parts ...string) string {
	allParts := []string{}

	if km.prefix != "" {
		allParts = append(allParts, km.prefix)
	}

	for _, part := range parts {
		if part != "" {
			allParts = append(allParts, part)
		}
	}

	return strings.Join(allParts, km.separator)
}

// ParseKey 解析缓存键
func (km *CacheKeyManager) ParseKey(key string) []string {
	if km.prefix != "" && strings.HasPrefix(key, km.prefix+km.separator) {
		key = strings.TrimPrefix(key, km.prefix+km.separator)
	}

	return strings.Split(key, km.separator)
}

// IsKeyType 检查键是否为指定类型
func (km *CacheKeyManager) IsKeyType(key string, keyType string) bool {
	parts := km.ParseKey(key)
	return len(parts) > 0 && parts[0] == keyType
}

// GetKeyPattern 获取键模式
func (km *CacheKeyManager) GetKeyPattern(keyType string) string {
	return km.buildKey(keyType, "*")
}

// 全局键管理器实例
var defaultKeyManager = NewCacheKeyManager("prediction_system")

// 全局键生成函数

// UserKey 生成用户键
func UserKey(userID uint) string {
	return defaultKeyManager.UserKey(userID)
}

// UserSessionKey 生成用户会话键
func UserSessionKey(sessionID string) string {
	return defaultKeyManager.UserSessionKey(sessionID)
}

// UserProfileKey 生成用户资料键
func UserProfileKey(userID uint) string {
	return defaultKeyManager.UserProfileKey(userID)
}

// UserStatsKey 生成用户统计键
func UserStatsKey(userID uint, tournament string) string {
	return defaultKeyManager.UserStatsKey(userID, tournament)
}

// MatchKey 生成比赛键
func MatchKey(matchID uint) string {
	return defaultKeyManager.MatchKey(matchID)
}

// MatchListKey 生成比赛列表键
func MatchListKey(tournament string, status string) string {
	return defaultKeyManager.MatchListKey(tournament, status)
}

// MatchStatsKey 生成比赛统计键
func MatchStatsKey(matchID uint) string {
	return defaultKeyManager.MatchStatsKey(matchID)
}

// PredictionKey 生成预测键
func PredictionKey(predictionID uint) string {
	return defaultKeyManager.PredictionKey(predictionID)
}

// PredictionListKey 生成预测列表键
func PredictionListKey(matchID uint, sortBy string) string {
	return defaultKeyManager.PredictionListKey(matchID, sortBy)
}

// PredictionVoteKey 生成预测投票键
func PredictionVoteKey(predictionID uint) string {
	return defaultKeyManager.PredictionVoteKey(predictionID)
}

// UserPredictionKey 生成用户预测键
func UserPredictionKey(userID uint, matchID uint) string {
	return defaultKeyManager.UserPredictionKey(userID, matchID)
}

// LeaderboardKey 生成排行榜键
func LeaderboardKey(tournament string) string {
	return defaultKeyManager.LeaderboardKey(tournament)
}

// LeaderboardTopKey 生成排行榜前N名键
func LeaderboardTopKey(tournament string, limit int) string {
	return defaultKeyManager.LeaderboardTopKey(tournament, limit)
}

// VoteKey 生成投票键
func VoteKey(userID uint, predictionID uint) string {
	return defaultKeyManager.VoteKey(userID, predictionID)
}

// VoteCountKey 生成投票计数键
func VoteCountKey(predictionID uint) string {
	return defaultKeyManager.VoteCountKey(predictionID)
}

// StatsKey 生成统计键
func StatsKey(category string, date string) string {
	return defaultKeyManager.StatsKey(category, date)
}

// DailyStatsKey 生成日统计键
func DailyStatsKey(date string) string {
	return defaultKeyManager.DailyStatsKey(date)
}

// WeeklyStatsKey 生成周统计键
func WeeklyStatsKey(week string) string {
	return defaultKeyManager.WeeklyStatsKey(week)
}

// CacheKey 生成通用缓存键
func CacheKey(category string, identifier string) string {
	return defaultKeyManager.CacheKey(category, identifier)
}

// TempCacheKey 生成临时缓存键
func TempCacheKey(identifier string) string {
	return defaultKeyManager.TempCacheKey(identifier)
}

// LockKey 生成锁键
func LockKey(resource string) string {
	return defaultKeyManager.LockKey(resource)
}

// DistributedLockKey 生成分布式锁键
func DistributedLockKey(resource string) string {
	return defaultKeyManager.DistributedLockKey(resource)
}

// SessionKey 生成会话键
func SessionKey(sessionID string) string {
	return defaultKeyManager.SessionKey(sessionID)
}

// AuthSessionKey 生成认证会话键
func AuthSessionKey(userID uint) string {
	return defaultKeyManager.AuthSessionKey(userID)
}

// NotificationKey 生成通知键
func NotificationKey(userID uint) string {
	return defaultKeyManager.NotificationKey(userID)
}

// NotificationQueueKey 生成通知队列键
func NotificationQueueKey(queueType string) string {
	return defaultKeyManager.NotificationQueueKey(queueType)
}

// GetKeyPattern 获取键模式
func GetKeyPattern(keyType string) string {
	return defaultKeyManager.GetKeyPattern(keyType)
}

// IsKeyType 检查键是否为指定类型
func IsKeyType(key string, keyType string) bool {
	return defaultKeyManager.IsKeyType(key, keyType)
}

// ParseKey 解析缓存键
func ParseKey(key string) []string {
	return defaultKeyManager.ParseKey(key)
}

// 缓存策略配置
type CacheStrategy struct {
	Key        string
	Expiration time.Duration
	Tags       []string
}

// 预定义缓存策略
var (
	// 排行榜缓存策略 (5分钟过期)
	LeaderboardCacheStrategy = CacheStrategy{
		Expiration: ExpirationLeaderboard,
		Tags:       []string{"leaderboard", "ranking"},
	}

	// 比赛数据缓存策略 (1分钟过期)
	MatchDataCacheStrategy = CacheStrategy{
		Expiration: ExpirationMatchData,
		Tags:       []string{"match", "data"},
	}

	// 用户资料缓存策略 (30分钟过期)
	UserProfileCacheStrategy = CacheStrategy{
		Expiration: ExpirationMedium,
		Tags:       []string{"user", "profile"},
	}

	// 预测列表缓存策略 (5分钟过期)
	PredictionListCacheStrategy = CacheStrategy{
		Expiration: ExpirationShort,
		Tags:       []string{"prediction", "list"},
	}

	// 统计数据缓存策略 (2小时过期)
	StatsCacheStrategy = CacheStrategy{
		Expiration: ExpirationLong,
		Tags:       []string{"stats", "analytics"},
	}
)

// GetCacheStrategy 根据键类型获取缓存策略
func GetCacheStrategy(keyType string) CacheStrategy {
	switch keyType {
	case KeyPrefixLeaderboard, KeyPrefixLeaderboardTop:
		return LeaderboardCacheStrategy
	case KeyPrefixMatch, KeyPrefixMatchList, KeyPrefixMatchStats:
		return MatchDataCacheStrategy
	case KeyPrefixUserProfile:
		return UserProfileCacheStrategy
	case KeyPrefixPredictionList:
		return PredictionListCacheStrategy
	case KeyPrefixStats, KeyPrefixStatsDaily, KeyPrefixStatsWeekly:
		return StatsCacheStrategy
	default:
		return CacheStrategy{
			Expiration: ExpirationMedium,
			Tags:       []string{"default"},
		}
	}
}
