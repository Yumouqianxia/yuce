package routes

import (
	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/match"
	"github.com/gin-gonic/gin"
)

// MatchRoutes 比赛路由
type MatchRoutes struct {
	matchHandler   *handlers.MatchHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewMatchRoutes 创建比赛路由
func NewMatchRoutes(matchService match.Service, authMiddleware *middleware.AuthMiddleware) *MatchRoutes {
	return &MatchRoutes{
		matchHandler:   handlers.NewMatchHandler(matchService),
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册比赛路由
func (r *MatchRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	matches := rg.Group("/matches")

	// 公开路由 - 不需要认证
	matches.GET("", r.matchHandler.ListMatches)                 // 获取比赛列表
	matches.GET("/:id", r.matchHandler.GetMatch)                // 获取比赛详情
	matches.GET("/upcoming", r.matchHandler.GetUpcomingMatches) // 获取即将开始的比赛
	matches.GET("/live", r.matchHandler.GetLiveMatches)         // 获取正在进行的比赛
	matches.GET("/finished", r.matchHandler.GetFinishedMatches) // 获取已结束的比赛

	// 写操作仅管理员
	adminOnly := matches.Group("")
	adminOnly.Use(r.authMiddleware.RequireAuth())
	adminOnly.Use(r.authMiddleware.RequireAdmin())
	{
		adminOnly.POST("", r.matchHandler.CreateMatch)            // 创建比赛
		adminOnly.PUT("/:id", r.matchHandler.UpdateMatch)         // 更新比赛
		adminOnly.POST("/:id/start", r.matchHandler.StartMatch)   // 开始比赛
		adminOnly.POST("/:id/result", r.matchHandler.SetResult)   // 设置比赛结果
		adminOnly.POST("/:id/cancel", r.matchHandler.CancelMatch) // 取消比赛
	}
}
