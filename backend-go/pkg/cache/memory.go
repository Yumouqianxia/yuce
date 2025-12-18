package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data   sync.Map
	expiry sync.Map
	mu     sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Value     []byte
	ExpiresAt time.Time
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{}

	// 启动清理过期数据的goroutine
	go mc.cleanupExpired()

	return mc
}

// Get 获取缓存值
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, ok := mc.data.Load(key)
	if !ok {
		return nil, ErrCacheNotFound
	}

	item := value.(*CacheItem)

	// 检查是否过期
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		mc.data.Delete(key)
		mc.expiry.Delete(key)
		return nil, ErrCacheNotFound
	}

	return item.Value, nil
}

// Set 设置缓存值
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
		mc.expiry.Store(key, expiresAt)
	}

	item := &CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	mc.data.Store(key, item)
	return nil
}

// Delete 删除缓存
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.data.Delete(key)
	mc.expiry.Delete(key)
	return nil
}

// DeletePattern 批量删除匹配模式的缓存键
func (mc *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// 简单的通配符匹配实现
	mc.data.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if matchPattern(keyStr, pattern) {
			mc.data.Delete(key)
			mc.expiry.Delete(key)
		}
		return true
	})
	return nil
}

// Exists 检查缓存是否存在
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, err := mc.Get(ctx, key)
	return err == nil, nil
}

// SetTTL 设置缓存过期时间
func (mc *MemoryCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	value, ok := mc.data.Load(key)
	if !ok {
		return ErrCacheNotFound
	}

	item := value.(*CacheItem)
	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl)
		mc.expiry.Store(key, item.ExpiresAt)
	} else {
		item.ExpiresAt = time.Time{}
		mc.expiry.Delete(key)
	}

	mc.data.Store(key, item)
	return nil
}

// GetTTL 获取缓存剩余过期时间
func (mc *MemoryCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	value, ok := mc.data.Load(key)
	if !ok {
		return 0, ErrCacheNotFound
	}

	item := value.(*CacheItem)
	if item.ExpiresAt.IsZero() {
		return -1, nil // 永不过期
	}

	remaining := time.Until(item.ExpiresAt)
	if remaining <= 0 {
		mc.data.Delete(key)
		mc.expiry.Delete(key)
		return 0, ErrCacheNotFound
	}

	return remaining, nil
}

// cleanupExpired 清理过期的缓存项
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		mc.expiry.Range(func(key, value interface{}) bool {
			expiresAt := value.(time.Time)
			if now.After(expiresAt) {
				mc.data.Delete(key)
				mc.expiry.Delete(key)
			}
			return true
		})
	}
}

// Clear 清空所有缓存
func (mc *MemoryCache) Clear() {
	mc.data = sync.Map{}
	mc.expiry = sync.Map{}
}

// Size 获取缓存项数量
func (mc *MemoryCache) Size() int {
	count := 0
	mc.data.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// matchPattern 简单的通配符匹配
func matchPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// 简单实现：支持前缀匹配
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(str) >= len(prefix) && str[:len(prefix)] == prefix
	}

	return str == pattern
}
