package routes

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/adapters/http/handlers"
	httpmw "backend-go/internal/adapters/http/middleware"
)

// RegisterCacheRoutes 注册缓存管理路由
func RegisterCacheRoutes(r *gin.RouterGroup, cacheHandler *handlers.CacheHandler, auth *httpmw.AuthMiddleware) {
	// 缓存管理路由组（需要管理员权限）
	cache := r.Group("/cache")
	cache.Use(auth.RequireAuth())  // 需要认证
	cache.Use(auth.RequireAdmin()) // 需要管理员权限
	{
		// 缓存统计和监控
		cache.GET("/stats", cacheHandler.GetCacheStats)
		cache.GET("/metrics", cacheHandler.GetCacheMetrics)
		cache.GET("/report", cacheHandler.GetCacheReport)
		cache.GET("/hit-rate", cacheHandler.CheckHitRate)

		// 排行榜缓存管理
		leaderboard := cache.Group("/leaderboard")
		{
			leaderboard.POST("/invalidate", cacheHandler.InvalidateLeaderboard)
			leaderboard.POST("/refresh", cacheHandler.RefreshLeaderboard)
		}

		// 批量操作
		cache.POST("/prewarm", cacheHandler.PrewarmCache)
		cache.POST("/batch-invalidate", cacheHandler.BatchInvalidate)

		// 事件触发的缓存失效
		cache.POST("/invalidate-points", cacheHandler.InvalidateOnPointsUpdate)
	}
}

// RegisterPublicCacheRoutes 注册公共缓存路由（用于健康检查等）
func RegisterPublicCacheRoutes(r *gin.RouterGroup, cacheHandler *handlers.CacheHandler) {
	// 公共缓存信息（不需要认证，但信息有限）
	cache := r.Group("/cache")
	{
		// 基础健康检查
		cache.GET("/health", func(c *gin.Context) {
			isHealthy := cacheHandler.CheckHitRate
			c.JSON(200, gin.H{
				"status":        "ok",
				"cache_healthy": isHealthy,
			})
		})
	}
}
