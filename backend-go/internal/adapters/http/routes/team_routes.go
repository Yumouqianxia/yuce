package routes

import (
	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/ports"
	"github.com/gin-gonic/gin"
)

// TeamRoutes 战队路由
type TeamRoutes struct {
	handler        *handlers.TeamHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewTeamRoutes 创建路由
func NewTeamRoutes(service ports.TeamService, authMiddleware *middleware.AuthMiddleware) *TeamRoutes {
	return &TeamRoutes{
		handler:        handlers.NewTeamHandler(service),
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册
func (r *TeamRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	teams := rg.Group("/teams")

	// 公开获取列表
	teams.GET("", r.handler.ListTeams)
	teams.GET("/:id", r.handler.GetTeam)

	// 管理员操作
	admin := teams.Group("")
	admin.Use(r.authMiddleware.RequireAuth(), r.authMiddleware.RequireAdmin())
	{
		admin.POST("", r.handler.CreateTeam)
		admin.PUT("/:id", r.handler.UpdateTeam)
		admin.DELETE("/:id", r.handler.DeleteTeam)
	}
}
