package routes

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
)

// RegisterLeaderboardRoutes 注册排行榜路由
func RegisterLeaderboardRoutes(r *gin.RouterGroup, handler *handlers.LeaderboardHandler, authMiddleware *middleware.AuthMiddleware) {
	leaderboard := r.Group("/leaderboard")
	{
		// 公开路由
		leaderboard.GET("", handler.GetLeaderboard)                                 // 获取排行榜
		leaderboard.GET("/stats", handler.GetLeaderboardStats)                      // 获取排行榜统计
		leaderboard.GET("/users/:user_id/rank", handler.GetUserRank)                // 获取用户排名
		leaderboard.GET("/ranks/:rank/around", handler.GetUsersAroundRank)          // 获取排名周围的用户
		leaderboard.GET("/users/:user_id/points-history", handler.GetPointsHistory) // 获取用户积分历史

		// 需要认证的路由
		authenticated := leaderboard.Group("")
		authenticated.Use(authMiddleware.RequireAuth())
		{
			authenticated.POST("/refresh", handler.RefreshLeaderboard)                              // 刷新排行榜缓存
			authenticated.POST("/matches/:match_id/calculate-points", handler.CalculateMatchPoints) // 计算比赛积分
		}
	}
}
