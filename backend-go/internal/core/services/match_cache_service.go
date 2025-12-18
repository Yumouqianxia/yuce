package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend-go/internal/core/domain/match"
	"backend-go/pkg/cache"
	"github.com/sirupsen/logrus"
)

// MatchCacheService 比赛缓存服务
type MatchCacheService struct {
	cache     cache.LayeredCacheService
	matchRepo match.Repository
	logger    *logrus.Logger
}

// NewMatchCacheService 创建比赛缓存服务实例
func NewMatchCacheService(
	cache cache.LayeredCacheService,
	matchRepo match.Repository,
	logger *logrus.Logger,
) *MatchCacheService {
	if logger == nil {
		logger = logrus.New()
	}

	return &MatchCacheService{
		cache:     cache,
		matchRepo: matchRepo,
		logger:    logger,
	}
}

// 缓存键常量
const (
	// 比赛详情缓存键前缀
	MatchDetailKeyPrefix = "match:detail:"

	// 比赛列表缓存键前缀
	MatchListKeyPrefix = "match:list:"

	// 热点比赛缓存键前缀
	HotMatchKeyPrefix = "match:hot:"

	// 即将开始的比赛缓存键
	UpcomingMatchesKey = "match:upcoming"

	// 正在进行的比赛缓存键
	LiveMatchesKey = "match:live"

	// 已结束的比赛缓存键前缀
	FinishedMatchesKeyPrefix = "match:finished:"
)

// 缓存TTL常量
const (
	// 比赛详情缓存时间 (5分钟)
	MatchDetailTTL = 5 * time.Minute

	// 热点比赛缓存时间 (1分钟)
	HotMatchTTL = 1 * time.Minute

	// 比赛列表缓存时间 (2分钟)
	MatchListTTL = 2 * time.Minute

	// 正在进行比赛缓存时间 (30秒)
	LiveMatchTTL = 30 * time.Second

	// 已结束比赛缓存时间 (10分钟)
	FinishedMatchTTL = 10 * time.Minute
)

// GetMatch 获取比赛详情 (带缓存)
func (mcs *MatchCacheService) GetMatch(ctx context.Context, id uint) (*match.Match, error) {
	key := fmt.Sprintf("%s%d", MatchDetailKeyPrefix, id)

	// 尝试从缓存获取
	data, err := mcs.cache.Get(ctx, key)
	if err == nil {
		var m match.Match
		if err := json.Unmarshal(data, &m); err == nil {
			mcs.logger.WithField("match_id", id).Debug("Match found in cache")
			return &m, nil
		}
		mcs.logger.WithError(err).Warn("Failed to unmarshal cached match")
	}

	// 从数据库获取
	m, err := mcs.matchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if data, err := json.Marshal(m); err == nil {
		ttl := MatchDetailTTL
		// 热点比赛使用更短的缓存时间
		if mcs.isHotMatch(m) {
			ttl = HotMatchTTL
		}

		if err := mcs.cache.Set(ctx, key, data, ttl); err != nil {
			mcs.logger.WithError(err).Warn("Failed to cache match")
		}
	}

	mcs.logger.WithField("match_id", id).Debug("Match loaded from database")
	return m, nil
}

// ListMatches 获取比赛列表 (带缓存)
func (mcs *MatchCacheService) ListMatches(ctx context.Context, filter match.ListFilter) ([]match.Match, error) {
	key := mcs.buildListCacheKey(filter)

	// 尝试从缓存获取
	data, err := mcs.cache.Get(ctx, key)
	if err == nil {
		var matches []match.Match
		if err := json.Unmarshal(data, &matches); err == nil {
			mcs.logger.WithField("filter", filter).Debug("Match list found in cache")
			return matches, nil
		}
		mcs.logger.WithError(err).Warn("Failed to unmarshal cached match list")
	}

	// 从数据库获取
	matches, err := mcs.matchRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if data, err := json.Marshal(matches); err == nil {
		if err := mcs.cache.Set(ctx, key, data, MatchListTTL); err != nil {
			mcs.logger.WithError(err).Warn("Failed to cache match list")
		}
	}

	mcs.logger.WithField("filter", filter).Debug("Match list loaded from database")
	return matches, nil
}

// GetUpcomingMatches 获取即将开始的比赛 (带缓存)
func (mcs *MatchCacheService) GetUpcomingMatches(ctx context.Context, limit int) ([]match.Match, error) {
	key := fmt.Sprintf("%s:%d", UpcomingMatchesKey, limit)

	// 尝试从缓存获取
	data, err := mcs.cache.Get(ctx, key)
	if err == nil {
		var matches []match.Match
		if err := json.Unmarshal(data, &matches); err == nil {
			mcs.logger.Debug("Upcoming matches found in cache")
			return matches, nil
		}
	}

	// 从数据库获取
	matches, err := mcs.matchRepo.GetUpcoming(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if data, err := json.Marshal(matches); err == nil {
		if err := mcs.cache.Set(ctx, key, data, MatchListTTL); err != nil {
			mcs.logger.WithError(err).Warn("Failed to cache upcoming matches")
		}
	}

	return matches, nil
}

// GetLiveMatches 获取正在进行的比赛 (带缓存)
func (mcs *MatchCacheService) GetLiveMatches(ctx context.Context) ([]match.Match, error) {
	key := LiveMatchesKey

	// 尝试从缓存获取
	data, err := mcs.cache.Get(ctx, key)
	if err == nil {
		var matches []match.Match
		if err := json.Unmarshal(data, &matches); err == nil {
			mcs.logger.Debug("Live matches found in cache")
			return matches, nil
		}
	}

	// 从数据库获取
	matches, err := mcs.matchRepo.GetLive(ctx)
	if err != nil {
		return nil, err
	}

	// 缓存结果 (较短TTL，因为比赛状态变化频繁)
	if data, err := json.Marshal(matches); err == nil {
		if err := mcs.cache.Set(ctx, key, data, LiveMatchTTL); err != nil {
			mcs.logger.WithError(err).Warn("Failed to cache live matches")
		}
	}

	return matches, nil
}

// GetFinishedMatches 获取已结束的比赛 (带缓存)
func (mcs *MatchCacheService) GetFinishedMatches(ctx context.Context, limit int) ([]match.Match, error) {
	key := fmt.Sprintf("%s%d", FinishedMatchesKeyPrefix, limit)

	// 尝试从缓存获取
	data, err := mcs.cache.Get(ctx, key)
	if err == nil {
		var matches []match.Match
		if err := json.Unmarshal(data, &matches); err == nil {
			mcs.logger.Debug("Finished matches found in cache")
			return matches, nil
		}
	}

	// 从数据库获取
	matches, err := mcs.matchRepo.GetFinished(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果 (较长TTL，因为已结束的比赛不会变化)
	if data, err := json.Marshal(matches); err == nil {
		if err := mcs.cache.Set(ctx, key, data, FinishedMatchTTL); err != nil {
			mcs.logger.WithError(err).Warn("Failed to cache finished matches")
		}
	}

	return matches, nil
}

// InvalidateMatch 使比赛相关缓存失效
func (mcs *MatchCacheService) InvalidateMatch(ctx context.Context, id uint) error {
	// 删除比赛详情缓存
	detailKey := fmt.Sprintf("%s%d", MatchDetailKeyPrefix, id)
	if err := mcs.cache.Delete(ctx, detailKey); err != nil {
		mcs.logger.WithError(err).Warn("Failed to invalidate match detail cache")
	}

	// 删除热点比赛缓存
	hotKey := fmt.Sprintf("%s%d", HotMatchKeyPrefix, id)
	if err := mcs.cache.Delete(ctx, hotKey); err != nil {
		mcs.logger.WithError(err).Warn("Failed to invalidate hot match cache")
	}

	return nil
}

// InvalidateMatchLists 使比赛列表缓存失效
func (mcs *MatchCacheService) InvalidateMatchLists(ctx context.Context) error {
	patterns := []string{
		MatchListKeyPrefix + "*",
		UpcomingMatchesKey + "*",
		LiveMatchesKey,
		FinishedMatchesKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		if err := mcs.cache.DeletePattern(ctx, pattern); err != nil {
			mcs.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate match list cache")
		}
	}

	return nil
}

// InvalidateAllMatchCaches 使所有比赛缓存失效
func (mcs *MatchCacheService) InvalidateAllMatchCaches(ctx context.Context) error {
	patterns := []string{
		MatchDetailKeyPrefix + "*",
		MatchListKeyPrefix + "*",
		HotMatchKeyPrefix + "*",
		UpcomingMatchesKey + "*",
		LiveMatchesKey,
		FinishedMatchesKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		if err := mcs.cache.DeletePattern(ctx, pattern); err != nil {
			mcs.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate match cache")
		}
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (mcs *MatchCacheService) GetCacheStats() cache.CacheStats {
	return mcs.cache.GetStats()
}

// buildListCacheKey 构建列表缓存键
func (mcs *MatchCacheService) buildListCacheKey(filter match.ListFilter) string {
	key := MatchListKeyPrefix

	if filter.Tournament != "" {
		key += fmt.Sprintf("tournament:%s:", filter.Tournament)
	}

	if filter.Status != "" {
		key += fmt.Sprintf("status:%s:", filter.Status)
	}

	if filter.StartDate != nil {
		key += fmt.Sprintf("start:%d:", filter.StartDate.Unix())
	}

	if filter.EndDate != nil {
		key += fmt.Sprintf("end:%d:", filter.EndDate.Unix())
	}

	key += fmt.Sprintf("limit:%d:offset:%d", filter.Limit, filter.Offset)

	return key
}

// isHotMatch 判断是否为热点比赛
func (mcs *MatchCacheService) isHotMatch(m *match.Match) bool {
	// 正在进行的比赛被认为是热点比赛
	if m.Status == match.MatchStatusLive {
		return true
	}

	// 即将开始的比赛 (1小时内) 被认为是热点比赛
	if m.Status == match.MatchStatusUpcoming {
		return time.Until(m.StartTime) <= time.Hour
	}

	return false
}
