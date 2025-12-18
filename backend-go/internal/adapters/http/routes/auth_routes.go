package routes

import (
	"backend-go/internal/adapters/http/handlers"
	"backend-go/internal/adapters/http/middleware"
	"backend-go/internal/core/domain/user"
	"github.com/gin-gonic/gin"
)

// AuthRoutes 认证路由配置
type AuthRoutes struct {
	authHandler    *handlers.AuthHandler
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthRoutes 创建认证路由
func NewAuthRoutes(userService user.Service) *AuthRoutes {
	return &AuthRoutes{
		authHandler:    handlers.NewAuthHandler(userService),
		authMiddleware: middleware.NewAuthMiddleware(userService),
	}
}

// RegisterRoutes 注册认证相关路由
func (r *AuthRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	// 认证路由组
	auth := rg.Group("/auth")
	{
		// 公开路由（不需要认证）
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)

		// 需要认证的路由
		authenticated := auth.Group("")
		authenticated.Use(r.authMiddleware.RequireAuth())
		{
			authenticated.GET("/profile", r.authHandler.GetProfile)
			authenticated.PATCH("/profile", r.authHandler.UpdateProfile)
			authenticated.POST("/change-password", r.authHandler.ChangePassword)
			authenticated.POST("/logout", r.authHandler.Logout)
		}
	}
}

// GetAuthMiddleware 获取认证中间件（供其他路由使用）
func (r *AuthRoutes) GetAuthMiddleware() *middleware.AuthMiddleware {
	return r.authMiddleware
}
