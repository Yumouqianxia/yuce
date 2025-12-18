package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService Redis 缓存服务接口
type CacheService interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, pairs ...interface{}) error
	MDelete(ctx context.Context, keys ...string) error

	// JSON 操作
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// 哈希操作
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDelete(ctx context.Context, key string, fields ...string) error

	// 列表操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)

	// 集合操作
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SRem(ctx context.Context, key string, members ...interface{}) error

	// 有序集合操作
	ZAdd(ctx context.Context, key string, members ...redis.Z) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRem(ctx context.Context, key string, members ...interface{}) error
	ZScore(ctx context.Context, key string, member string) (float64, error)

	// 过期时间操作
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 高级操作
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	DecrementBy(ctx context.Context, key string, value int64) (int64, error)

	// 分布式锁
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error

	// 缓存模式
	GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error)
	Remember(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error)

	// 缓存失效
	InvalidatePattern(ctx context.Context, pattern string) error
	FlushDB(ctx context.Context) error
}

// cacheService 缓存服务实现
type cacheService struct {
	client *Client
}

// NewCacheService 创建缓存服务
func NewCacheService(client *Client) CacheService {
	return &cacheService{
		client: client,
	}
}

// GetCacheService 获取全局缓存服务实例
func GetCacheService() CacheService {
	return NewCacheService(GetClient())
}

// 基础操作实现

func (s *cacheService) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("get", time.Since(start), nil)
	}()

	result := s.client.rdb.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			s.client.metrics.RecordCacheMiss(key)
			return "", ErrKeyNotFound
		}
		s.client.metrics.RecordOperation("get", time.Since(start), err)
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	s.client.metrics.RecordCacheHit(key)
	return result.Val(), nil
}

func (s *cacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("set", time.Since(start), nil)
	}()

	result := s.client.rdb.Set(ctx, key, value, expiration)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("set", time.Since(start), err)
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("delete", time.Since(start), nil)
	}()

	result := s.client.rdb.Del(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("delete", time.Since(start), err)
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("exists", time.Since(start), nil)
	}()

	result := s.client.rdb.Exists(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("exists", time.Since(start), err)
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	return result.Val() > 0, nil
}

// 批量操作实现

func (s *cacheService) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("mget", time.Since(start), nil)
	}()

	result := s.client.rdb.MGet(ctx, keys...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("mget", time.Since(start), err)
		return nil, fmt.Errorf("failed to mget keys: %w", err)
	}

	return result.Val(), nil
}

func (s *cacheService) MSet(ctx context.Context, pairs ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("mset", time.Since(start), nil)
	}()

	result := s.client.rdb.MSet(ctx, pairs...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("mset", time.Since(start), err)
		return fmt.Errorf("failed to mset: %w", err)
	}

	return nil
}

func (s *cacheService) MDelete(ctx context.Context, keys ...string) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("mdel", time.Since(start), nil)
	}()

	result := s.client.rdb.Del(ctx, keys...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("mdel", time.Since(start), err)
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// JSON 操作实现

func (s *cacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	return s.Set(ctx, key, data, expiration)
}

// 哈希操作实现

func (s *cacheService) HGet(ctx context.Context, key, field string) (string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("hget", time.Since(start), nil)
	}()

	result := s.client.rdb.HGet(ctx, key, field)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		s.client.metrics.RecordOperation("hget", time.Since(start), err)
		return "", fmt.Errorf("failed to hget %s.%s: %w", key, field, err)
	}

	return result.Val(), nil
}

func (s *cacheService) HSet(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("hset", time.Since(start), nil)
	}()

	result := s.client.rdb.HSet(ctx, key, values...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("hset", time.Since(start), err)
		return fmt.Errorf("failed to hset %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("hgetall", time.Since(start), nil)
	}()

	result := s.client.rdb.HGetAll(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("hgetall", time.Since(start), err)
		return nil, fmt.Errorf("failed to hgetall %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) HDelete(ctx context.Context, key string, fields ...string) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("hdel", time.Since(start), nil)
	}()

	result := s.client.rdb.HDel(ctx, key, fields...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("hdel", time.Since(start), err)
		return fmt.Errorf("failed to hdel %s: %w", key, err)
	}

	return nil
}

// 列表操作实现

func (s *cacheService) LPush(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("lpush", time.Since(start), nil)
	}()

	result := s.client.rdb.LPush(ctx, key, values...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("lpush", time.Since(start), err)
		return fmt.Errorf("failed to lpush %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) RPush(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("rpush", time.Since(start), nil)
	}()

	result := s.client.rdb.RPush(ctx, key, values...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("rpush", time.Since(start), err)
		return fmt.Errorf("failed to rpush %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) LPop(ctx context.Context, key string) (string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("lpop", time.Since(start), nil)
	}()

	result := s.client.rdb.LPop(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		s.client.metrics.RecordOperation("lpop", time.Since(start), err)
		return "", fmt.Errorf("failed to lpop %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) RPop(ctx context.Context, key string) (string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("rpop", time.Since(start), nil)
	}()

	result := s.client.rdb.RPop(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		s.client.metrics.RecordOperation("rpop", time.Since(start), err)
		return "", fmt.Errorf("failed to rpop %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("lrange", time.Since(startTime), nil)
	}()

	result := s.client.rdb.LRange(ctx, key, start, stop)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("lrange", time.Since(startTime), err)
		return nil, fmt.Errorf("failed to lrange %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) LLen(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("llen", time.Since(start), nil)
	}()

	result := s.client.rdb.LLen(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("llen", time.Since(start), err)
		return 0, fmt.Errorf("failed to llen %s: %w", key, err)
	}

	return result.Val(), nil
}

// 集合操作实现

func (s *cacheService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("sadd", time.Since(start), nil)
	}()

	result := s.client.rdb.SAdd(ctx, key, members...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("sadd", time.Since(start), err)
		return fmt.Errorf("failed to sadd %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("smembers", time.Since(start), nil)
	}()

	result := s.client.rdb.SMembers(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("smembers", time.Since(start), err)
		return nil, fmt.Errorf("failed to smembers %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("sismember", time.Since(start), nil)
	}()

	result := s.client.rdb.SIsMember(ctx, key, member)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("sismember", time.Since(start), err)
		return false, fmt.Errorf("failed to sismember %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) SRem(ctx context.Context, key string, members ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("srem", time.Since(start), nil)
	}()

	result := s.client.rdb.SRem(ctx, key, members...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("srem", time.Since(start), err)
		return fmt.Errorf("failed to srem %s: %w", key, err)
	}

	return nil
}

// 有序集合操作实现

func (s *cacheService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("zadd", time.Since(start), nil)
	}()

	result := s.client.rdb.ZAdd(ctx, key, members...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("zadd", time.Since(start), err)
		return fmt.Errorf("failed to zadd %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("zrange", time.Since(startTime), nil)
	}()

	result := s.client.rdb.ZRange(ctx, key, start, stop)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("zrange", time.Since(startTime), err)
		return nil, fmt.Errorf("failed to zrange %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	startTime := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("zrange_with_scores", time.Since(startTime), nil)
	}()

	result := s.client.rdb.ZRangeWithScores(ctx, key, start, stop)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("zrange_with_scores", time.Since(startTime), err)
		return nil, fmt.Errorf("failed to zrange with scores %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) ZRem(ctx context.Context, key string, members ...interface{}) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("zrem", time.Since(start), nil)
	}()

	result := s.client.rdb.ZRem(ctx, key, members...)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("zrem", time.Since(start), err)
		return fmt.Errorf("failed to zrem %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) ZScore(ctx context.Context, key string, member string) (float64, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("zscore", time.Since(start), nil)
	}()

	result := s.client.rdb.ZScore(ctx, key, member)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return 0, ErrKeyNotFound
		}
		s.client.metrics.RecordOperation("zscore", time.Since(start), err)
		return 0, fmt.Errorf("failed to zscore %s: %w", key, err)
	}

	return result.Val(), nil
}

// 过期时间操作实现

func (s *cacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("expire", time.Since(start), nil)
	}()

	result := s.client.rdb.Expire(ctx, key, expiration)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("expire", time.Since(start), err)
		return fmt.Errorf("failed to expire %s: %w", key, err)
	}

	return nil
}

func (s *cacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("ttl", time.Since(start), nil)
	}()

	result := s.client.rdb.TTL(ctx, key)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("ttl", time.Since(start), err)
		return 0, fmt.Errorf("failed to ttl %s: %w", key, err)
	}

	return result.Val(), nil
}

// 高级操作实现

func (s *cacheService) Increment(ctx context.Context, key string) (int64, error) {
	return s.IncrementBy(ctx, key, 1)
}

func (s *cacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("incrby", time.Since(start), nil)
	}()

	result := s.client.rdb.IncrBy(ctx, key, value)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("incrby", time.Since(start), err)
		return 0, fmt.Errorf("failed to incrby %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) Decrement(ctx context.Context, key string) (int64, error) {
	return s.DecrementBy(ctx, key, 1)
}

func (s *cacheService) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("decrby", time.Since(start), nil)
	}()

	result := s.client.rdb.DecrBy(ctx, key, value)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("decrby", time.Since(start), err)
		return 0, fmt.Errorf("failed to decrby %s: %w", key, err)
	}

	return result.Val(), nil
}

// 分布式锁实现

func (s *cacheService) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("lock", time.Since(start), nil)
	}()

	lockKey := fmt.Sprintf("lock:%s", key)
	result := s.client.rdb.SetNX(ctx, lockKey, "locked", expiration)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("lock", time.Since(start), err)
		return false, fmt.Errorf("failed to acquire lock %s: %w", key, err)
	}

	return result.Val(), nil
}

func (s *cacheService) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	return s.Delete(ctx, lockKey)
}

// 缓存模式实现

func (s *cacheService) GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// 尝试从缓存获取
	value, err := s.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	if err != ErrKeyNotFound {
		return nil, err
	}

	// 缓存未命中，执行函数获取值
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	if err := s.Set(ctx, key, result, expiration); err != nil {
		s.client.logger.WithError(err).Warnf("Failed to set cache for key: %s", key)
	}

	return result, nil
}

func (s *cacheService) Remember(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	return s.GetOrSet(ctx, key, expiration, fn)
}

// 缓存失效实现

func (s *cacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("invalidate_pattern", time.Since(start), nil)
	}()

	// 使用 SCAN 命令查找匹配的键
	var cursor uint64
	var keys []string

	for {
		result := s.client.rdb.Scan(ctx, cursor, pattern, 100)
		if err := result.Err(); err != nil {
			s.client.metrics.RecordOperation("invalidate_pattern", time.Since(start), err)
			return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
		}

		scanKeys, newCursor := result.Val()
		keys = append(keys, scanKeys...)
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	// 删除找到的键
	if len(keys) > 0 {
		return s.MDelete(ctx, keys...)
	}

	return nil
}

func (s *cacheService) FlushDB(ctx context.Context) error {
	start := time.Now()
	defer func() {
		s.client.metrics.RecordOperation("flushdb", time.Since(start), nil)
	}()

	result := s.client.rdb.FlushDB(ctx)
	if err := result.Err(); err != nil {
		s.client.metrics.RecordOperation("flushdb", time.Since(start), err)
		return fmt.Errorf("failed to flush database: %w", err)
	}

	return nil
}
