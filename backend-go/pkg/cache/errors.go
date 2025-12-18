package cache

import "errors"

var (
	// ErrCacheNotFound 缓存未找到错误
	ErrCacheNotFound = errors.New("cache not found")

	// ErrCacheExpired 缓存已过期错误
	ErrCacheExpired = errors.New("cache expired")

	// ErrInvalidKey 无效缓存键错误
	ErrInvalidKey = errors.New("invalid cache key")

	// ErrInvalidValue 无效缓存值错误
	ErrInvalidValue = errors.New("invalid cache value")

	// ErrCacheUnavailable 缓存服务不可用错误
	ErrCacheUnavailable = errors.New("cache service unavailable")
)
