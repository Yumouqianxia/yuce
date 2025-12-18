// Package cache provides a comprehensive caching system with multi-layer support.
//
// This package implements a flexible caching architecture that supports both
// in-memory and Redis-based caching strategies. It provides interfaces and
// implementations for various caching strategies.
package cache

import (
	"context"
	"time"
)

// CacheService 缓存服务接口
type CacheService interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存值
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// DeletePattern 批量删除匹配模式的缓存键
	DeletePattern(ctx context.Context, pattern string) error

	// Exists 检查缓存是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// SetTTL 设置缓存过期时间
	SetTTL(ctx context.Context, key string, ttl time.Duration) error

	// GetTTL 获取缓存剩余过期时间
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

// LayeredCacheService 分层缓存服务接口
type LayeredCacheService interface {
	CacheService

	// GetFromMemory 从内存缓存获取
	GetFromMemory(ctx context.Context, key string) ([]byte, error)

	// GetFromRedis 从Redis缓存获取
	GetFromRedis(ctx context.Context, key string) ([]byte, error)

	// SetToMemory 设置到内存缓存
	SetToMemory(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// SetToRedis 设置到Redis缓存
	SetToRedis(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// InvalidateMemory 清除内存缓存
	InvalidateMemory(ctx context.Context, pattern string) error

	// GetStats 获取缓存统计信息
	GetStats() CacheStats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	MemoryHits   int64   `json:"memory_hits"`
	MemoryMisses int64   `json:"memory_misses"`
	RedisHits    int64   `json:"redis_hits"`
	RedisMisses  int64   `json:"redis_misses"`
	TotalHits    int64   `json:"total_hits"`
	TotalMisses  int64   `json:"total_misses"`
	HitRate      float64 `json:"hit_rate"`
}
