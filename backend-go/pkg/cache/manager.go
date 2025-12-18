package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CacheManager 缓存管理器
type CacheManager struct {
	cache  LayeredCacheService
	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewCacheManager 创建缓存管理器实例
func NewCacheManager(cache LayeredCacheService, logger *logrus.Logger) *CacheManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &CacheManager{
		cache:  cache,
		logger: logger,
	}
}

// BatchInvalidate 批量使缓存失效
func (cm *CacheManager) BatchInvalidate(ctx context.Context, patterns []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var errors []error

	for _, pattern := range patterns {
		if err := cm.cache.DeletePattern(ctx, pattern); err != nil {
			cm.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate cache pattern")
			errors = append(errors, fmt.Errorf("pattern %s: %w", pattern, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch invalidation failed: %v", errors)
	}

	return nil
}

// WarmupCache 预热缓存
func (cm *CacheManager) WarmupCache(ctx context.Context, warmupFunc func(ctx context.Context) error) error {
	cm.logger.Info("Starting cache warmup")
	start := time.Now()

	err := warmupFunc(ctx)

	duration := time.Since(start)
	if err != nil {
		cm.logger.WithError(err).WithField("duration", duration).Error("Cache warmup failed")
		return err
	}

	cm.logger.WithField("duration", duration).Info("Cache warmup completed")
	return nil
}

// ScheduledInvalidation 定时缓存失效
func (cm *CacheManager) ScheduledInvalidation(ctx context.Context, interval time.Duration, patterns []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	cm.logger.WithFields(logrus.Fields{
		"interval": interval,
		"patterns": patterns,
	}).Info("Starting scheduled cache invalidation")

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Stopping scheduled cache invalidation")
			return
		case <-ticker.C:
			if err := cm.BatchInvalidate(ctx, patterns); err != nil {
				cm.logger.WithError(err).Error("Scheduled cache invalidation failed")
			} else {
				cm.logger.Debug("Scheduled cache invalidation completed")
			}
		}
	}
}

// GetCacheHealth 获取缓存健康状态
func (cm *CacheManager) GetCacheHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"timestamp": time.Now(),
		"stats":     cm.cache.GetStats(),
	}

	// 测试缓存读写
	testKey := "health_check_" + fmt.Sprintf("%d", time.Now().UnixNano())
	testValue := []byte("test")

	// 写入测试
	writeStart := time.Now()
	writeErr := cm.cache.Set(ctx, testKey, testValue, time.Minute)
	writeDuration := time.Since(writeStart)

	health["write_test"] = map[string]interface{}{
		"success":  writeErr == nil,
		"duration": writeDuration,
		"error":    writeErr,
	}

	// 读取测试
	if writeErr == nil {
		readStart := time.Now()
		_, readErr := cm.cache.Get(ctx, testKey)
		readDuration := time.Since(readStart)

		health["read_test"] = map[string]interface{}{
			"success":  readErr == nil,
			"duration": readDuration,
			"error":    readErr,
		}

		// 清理测试数据
		cm.cache.Delete(ctx, testKey)
	}

	return health
}

// ClearExpiredKeys 清理过期键 (仅对内存缓存有效)
func (cm *CacheManager) ClearExpiredKeys(ctx context.Context) error {
	cm.logger.Info("Clearing expired cache keys")

	// 对于分层缓存，主要清理内存缓存
	// Redis会自动清理过期键
	if err := cm.cache.InvalidateMemory(ctx, "*"); err != nil {
		cm.logger.WithError(err).Error("Failed to clear expired keys")
		return err
	}

	cm.logger.Info("Expired cache keys cleared")
	return nil
}

// GetCacheSize 获取缓存大小信息
func (cm *CacheManager) GetCacheSize(ctx context.Context) map[string]interface{} {
	stats := cm.cache.GetStats()

	return map[string]interface{}{
		"memory_items": "N/A", // 内存缓存项数量需要额外实现
		"redis_items":  "N/A", // Redis项数量需要额外实现
		"hit_rate":     stats.HitRate,
		"total_hits":   stats.TotalHits,
		"total_misses": stats.TotalMisses,
	}
}
