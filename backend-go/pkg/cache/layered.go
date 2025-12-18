package cache

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"backend-go/pkg/redis"
	"github.com/sirupsen/logrus"
)

// LayeredCache 分层缓存实现 (内存 + Redis)
type LayeredCache struct {
	memory      *MemoryCache
	redisClient *redis.Client
	logger      *logrus.Logger

	// 统计信息
	memoryHits   int64
	memoryMisses int64
	redisHits    int64
	redisMisses  int64
}

// NewLayeredCache 创建分层缓存实例
func NewLayeredCache(redisClient *redis.Client, logger *logrus.Logger) LayeredCacheService {
	if logger == nil {
		logger = logrus.New()
	}

	return &LayeredCache{
		memory:      NewMemoryCache(),
		redisClient: redisClient,
		logger:      logger,
	}
}

// Get 获取缓存值 (先查内存，再查Redis)
func (lc *LayeredCache) Get(ctx context.Context, key string) ([]byte, error) {
	// 1. 先从内存缓存获取
	value, err := lc.GetFromMemory(ctx, key)
	if err == nil {
		atomic.AddInt64(&lc.memoryHits, 1)
		return value, nil
	}

	atomic.AddInt64(&lc.memoryMisses, 1)

	// 2. 从Redis获取
	value, err = lc.GetFromRedis(ctx, key)
	if err == nil {
		atomic.AddInt64(&lc.redisHits, 1)

		// 将数据回写到内存缓存 (较短的TTL)
		memoryTTL := 30 * time.Second
		if setErr := lc.SetToMemory(ctx, key, value, memoryTTL); setErr != nil {
			lc.logger.WithError(setErr).Warn("Failed to set memory cache")
		}

		return value, nil
	}

	atomic.AddInt64(&lc.redisMisses, 1)
	return nil, ErrCacheNotFound
}

// Set 设置缓存值 (同时设置内存和Redis)
func (lc *LayeredCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// 设置到内存缓存 (较短的TTL)
	memoryTTL := ttl
	if ttl > time.Minute {
		memoryTTL = time.Minute // 内存缓存最多1分钟
	}

	if err := lc.SetToMemory(ctx, key, value, memoryTTL); err != nil {
		lc.logger.WithError(err).Warn("Failed to set memory cache")
	}

	// 设置到Redis缓存
	return lc.SetToRedis(ctx, key, value, ttl)
}

// Delete 删除缓存
func (lc *LayeredCache) Delete(ctx context.Context, key string) error {
	// 删除内存缓存
	if err := lc.memory.Delete(ctx, key); err != nil {
		lc.logger.WithError(err).Warn("Failed to delete memory cache")
	}

	// 删除Redis缓存
	rdb := lc.redisClient.GetRedisClient()
	return rdb.Del(ctx, key).Err()
}

// DeletePattern 批量删除匹配模式的缓存键
func (lc *LayeredCache) DeletePattern(ctx context.Context, pattern string) error {
	// 删除内存缓存
	if err := lc.memory.DeletePattern(ctx, pattern); err != nil {
		lc.logger.WithError(err).Warn("Failed to delete memory cache pattern")
	}

	// 删除Redis缓存
	rdb := lc.redisClient.GetRedisClient()

	// 获取匹配的键
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return rdb.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists 检查缓存是否存在
func (lc *LayeredCache) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查内存缓存
	exists, err := lc.memory.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// 检查Redis缓存
	rdb := lc.redisClient.GetRedisClient()
	count, err := rdb.Exists(ctx, key).Result()
	return count > 0, err
}

// SetTTL 设置缓存过期时间
func (lc *LayeredCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	// 设置内存缓存TTL
	memoryTTL := ttl
	if ttl > time.Minute {
		memoryTTL = time.Minute
	}

	if err := lc.memory.SetTTL(ctx, key, memoryTTL); err != nil {
		lc.logger.WithError(err).Warn("Failed to set memory cache TTL")
	}

	// 设置Redis缓存TTL
	rdb := lc.redisClient.GetRedisClient()
	return rdb.Expire(ctx, key, ttl).Err()
}

// GetTTL 获取缓存剩余过期时间
func (lc *LayeredCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	// 优先返回Redis的TTL
	rdb := lc.redisClient.GetRedisClient()
	return rdb.TTL(ctx, key).Result()
}

// GetFromMemory 从内存缓存获取
func (lc *LayeredCache) GetFromMemory(ctx context.Context, key string) ([]byte, error) {
	return lc.memory.Get(ctx, key)
}

// GetFromRedis 从Redis缓存获取
func (lc *LayeredCache) GetFromRedis(ctx context.Context, key string) ([]byte, error) {
	rdb := lc.redisClient.GetRedisClient()
	result, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

// SetToMemory 设置到内存缓存
func (lc *LayeredCache) SetToMemory(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return lc.memory.Set(ctx, key, value, ttl)
}

// SetToRedis 设置到Redis缓存
func (lc *LayeredCache) SetToRedis(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	rdb := lc.redisClient.GetRedisClient()
	return rdb.Set(ctx, key, value, ttl).Err()
}

// InvalidateMemory 清除内存缓存
func (lc *LayeredCache) InvalidateMemory(ctx context.Context, pattern string) error {
	return lc.memory.DeletePattern(ctx, pattern)
}

// GetStats 获取缓存统计信息
func (lc *LayeredCache) GetStats() CacheStats {
	memoryHits := atomic.LoadInt64(&lc.memoryHits)
	memoryMisses := atomic.LoadInt64(&lc.memoryMisses)
	redisHits := atomic.LoadInt64(&lc.redisHits)
	redisMisses := atomic.LoadInt64(&lc.redisMisses)

	totalHits := memoryHits + redisHits
	totalMisses := memoryMisses + redisMisses
	total := totalHits + totalMisses

	var hitRate float64
	if total > 0 {
		hitRate = float64(totalHits) / float64(total)
	}

	return CacheStats{
		MemoryHits:   memoryHits,
		MemoryMisses: memoryMisses,
		RedisHits:    redisHits,
		RedisMisses:  redisMisses,
		TotalHits:    totalHits,
		TotalMisses:  totalMisses,
		HitRate:      hitRate,
	}
}

// ResetStats 重置统计信息
func (lc *LayeredCache) ResetStats() {
	atomic.StoreInt64(&lc.memoryHits, 0)
	atomic.StoreInt64(&lc.memoryMisses, 0)
	atomic.StoreInt64(&lc.redisHits, 0)
	atomic.StoreInt64(&lc.redisMisses, 0)
}
